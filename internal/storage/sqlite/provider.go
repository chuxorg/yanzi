package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
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

// Provider is the embedded SQLite storage provider.
type Provider struct {
	path string
	db   *sql.DB
}

// Open initializes a SQLite provider at path, running embedded migrations.
func Open(ctx context.Context, path string) (*Provider, bool, error) {
	if path == "" {
		return nil, false, fmt.Errorf("sqlite path is required")
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
	if err := RunMigrations(db); err != nil {
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

// FromDB wraps an existing SQLite handle with provider operations.
func FromDB(db *sql.DB) *Provider {
	return &Provider{db: db}
}
