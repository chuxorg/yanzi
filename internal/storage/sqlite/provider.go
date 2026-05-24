package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/storage"
	_ "modernc.org/sqlite"
)

const schemaVersionTable = `
CREATE TABLE IF NOT EXISTS schema_version (
	version INTEGER NOT NULL,
	applied_at TIMESTAMP NOT NULL
);
`

const schemaMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TEXT NOT NULL
);
`

// Provider is the embedded SQLite storage provider.
type Provider struct {
	path string
	db   *sql.DB
}

// Open initializes a SQLite provider at path using the provided migration files.
func Open(ctx context.Context, path string, migrations fs.FS) (*Provider, bool, error) {
	if strings.TrimSpace(path) == "" {
		return nil, false, errors.New("sqlite path is required")
	}
	if migrations == nil {
		return nil, false, errors.New("sqlite migrations fs is required")
	}
	if err := ensureDBFile(path); err != nil {
		return nil, false, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, false, fmt.Errorf("open db: %w", err)
	}
	provider := &Provider{path: path, db: db}

	if err := provider.configure(); err != nil {
		_ = provider.Close()
		return nil, false, err
	}

	initialized, err := provider.ensureSchemaVersion(ctx)
	if err != nil {
		_ = provider.Close()
		return nil, false, err
	}
	if err := provider.migrate(ctx, migrations); err != nil {
		_ = provider.Close()
		return nil, false, err
	}

	return provider, initialized, nil
}

// Name returns the provider identifier.
func (p *Provider) Name() storage.ProviderName {
	return storage.ProviderSQLite
}

// SQLDB exposes the current SQLite handle for existing local call sites.
func (p *Provider) SQLDB() *sql.DB {
	if p == nil {
		return nil
	}
	return p.db
}

// Close closes the provider handle.
func (p *Provider) Close() error {
	if p == nil || p.db == nil {
		return nil
	}
	return p.db.Close()
}

// Health reports internal readiness for the provider.
func (p *Provider) Health(ctx context.Context) storage.Health {
	health := storage.Health{Provider: storage.ProviderSQLite, Path: p.path, Status: storage.HealthReady}
	if p == nil || p.db == nil {
		health.Status = storage.HealthUnavailable
		health.Error = storage.ErrProviderUnavailable.Error()
		return health
	}
	if err := p.db.PingContext(ctx); err != nil {
		health.Status = storage.HealthUnavailable
		health.Error = err.Error()
	}
	return health
}

func (p *Provider) Artifacts() bool    { return true }
func (p *Provider) Projects() bool     { return true }
func (p *Provider) Checkpoints() bool  { return true }
func (p *Provider) Verification() bool { return true }
func (p *Provider) ImportExport() bool { return true }

func (p *Provider) configure() error {
	if err := p.db.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	if _, err := p.db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		return err
	}
	if _, err := p.db.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		return err
	}
	if _, err := p.db.Exec(`PRAGMA busy_timeout=5000;`); err != nil {
		return err
	}
	return nil
}

func ensureDBFile(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create db directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("create db file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close db file: %w", err)
	}
	return nil
}

func (p *Provider) ensureSchemaVersion(ctx context.Context) (bool, error) {
	if _, err := p.db.ExecContext(ctx, schemaVersionTable); err != nil {
		return false, fmt.Errorf("create schema_version: %w", err)
	}

	var count int
	if err := p.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM schema_version`).Scan(&count); err != nil {
		return false, fmt.Errorf("read schema_version: %w", err)
	}
	if count > 0 {
		return false, nil
	}

	if _, err := p.db.ExecContext(ctx, `INSERT INTO schema_version (version, applied_at) VALUES (?, ?)`, 1, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return false, fmt.Errorf("write schema_version: %w", err)
	}
	return true, nil
}

func (p *Provider) migrate(ctx context.Context, migrations fs.FS) error {
	if p == nil || p.db == nil {
		return errors.New("database is nil")
	}
	if _, err := p.db.ExecContext(ctx, schemaMigrationsTable); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	paths, err := listMigrationFiles(migrations)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return errors.New("no migration files found")
	}

	sort.Strings(paths)
	for _, path := range paths {
		version := filepath.Base(path)
		applied, err := p.isMigrationApplied(ctx, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		contents, err := fs.ReadFile(migrations, path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}

		tx, err := p.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`, version, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", version, err)
		}
	}

	return nil
}

func listMigrationFiles(fsys fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(fsys, "migrations")
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		paths = append(paths, filepath.Join("migrations", name))
	}
	return paths, nil
}

func (p *Provider) isMigrationApplied(ctx context.Context, version string) (bool, error) {
	var count int
	if err := p.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM schema_migrations WHERE version = ?`, version).Scan(&count); err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}
	return count > 0, nil
}

// FromDB wraps an existing SQLite handle with provider operations.
func FromDB(db *sql.DB) *Provider {
	return &Provider{db: db}
}
