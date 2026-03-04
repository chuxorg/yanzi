// Package store provides a SQLite persistence layer.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/core/model"
	_ "modernc.org/sqlite"
)

const schemaMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TEXT NOT NULL
);
`

// Store provides CRUD and migration operations for intent persistence.
type Store struct {
	db *sql.DB
}

// Open creates a SQLite-backed Store with required runtime pragmas enabled.
func Open(path string) (*Store, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("sqlite path is required")
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA busy_timeout=5000;`); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Migrate(ctx context.Context) error {
	if s.db == nil {
		return errors.New("store not initialized")
	}
	if _, err := s.db.ExecContext(ctx, schemaMigrationsTable); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	paths, err := listMigrationFiles()
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return errors.New("no migration files found")
	}

	sort.Strings(paths)
	for _, path := range paths {
		version := filepath.Base(path)
		applied, err := s.isMigrationApplied(ctx, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}

		tx, err := s.db.BeginTx(ctx, nil)
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

// listMigrationFiles collects migration SQL files from the migrations directory.
func listMigrationFiles() ([]string, error) {
	entries, err := os.ReadDir("migrations")
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

func (s *Store) isMigrationApplied(ctx context.Context, version string) (bool, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM schema_migrations WHERE version = ?`, version).Scan(&count); err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}
	return count > 0, nil
}

func (s *Store) CreateIntent(ctx context.Context, record model.IntentRecord) error {
	var title any
	if record.Title != "" {
		title = record.Title
	}
	var meta any
	if len(record.Meta) > 0 {
		meta = string(record.Meta)
	}
	var prevHash any
	if record.PrevHash != "" {
		prevHash = record.PrevHash
	}

	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO intents (id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ID,
		record.CreatedAt,
		record.Author,
		record.SourceType,
		title,
		record.Prompt,
		record.Response,
		meta,
		prevHash,
		record.Hash,
	)
	return err
}

func (s *Store) GetIntent(ctx context.Context, id string) (model.IntentRecord, error) {
	var record model.IntentRecord
	var title sql.NullString
	var meta sql.NullString
	var prevHash sql.NullString
	row := s.db.QueryRowContext(ctx, `SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash FROM intents WHERE id = ?`, id)
	if err := row.Scan(
		&record.ID,
		&record.CreatedAt,
		&record.Author,
		&record.SourceType,
		&title,
		&record.Prompt,
		&record.Response,
		&meta,
		&prevHash,
		&record.Hash,
	); err != nil {
		return record, err
	}

	if title.Valid {
		record.Title = title.String
	}
	if meta.Valid && meta.String != "" {
		record.Meta = []byte(meta.String)
	}
	if prevHash.Valid {
		record.PrevHash = prevHash.String
	}
	return record, nil
}

// GetIntentByHash loads an intent by its hash for chain traversal.
func (s *Store) GetIntentByHash(ctx context.Context, hash string) (model.IntentRecord, error) {
	var record model.IntentRecord
	var title sql.NullString
	var meta sql.NullString
	var prevHash sql.NullString
	row := s.db.QueryRowContext(ctx, `SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash FROM intents WHERE hash = ?`, hash)
	if err := row.Scan(
		&record.ID,
		&record.CreatedAt,
		&record.Author,
		&record.SourceType,
		&title,
		&record.Prompt,
		&record.Response,
		&meta,
		&prevHash,
		&record.Hash,
	); err != nil {
		return record, err
	}

	if title.Valid {
		record.Title = title.String
	}
	if meta.Valid && meta.String != "" {
		record.Meta = []byte(meta.String)
	}
	if prevHash.Valid {
		record.PrevHash = prevHash.String
	}
	return record, nil
}

func (s *Store) ListIntents(ctx context.Context, limit int) ([]model.IntentRecord, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.QueryContext(ctx, `SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash FROM intents ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var intents []model.IntentRecord
	for rows.Next() {
		var record model.IntentRecord
		var title sql.NullString
		var meta sql.NullString
		var prevHash sql.NullString
		if err := rows.Scan(
			&record.ID,
			&record.CreatedAt,
			&record.Author,
			&record.SourceType,
			&title,
			&record.Prompt,
			&record.Response,
			&meta,
			&prevHash,
			&record.Hash,
		); err != nil {
			return nil, err
		}
		if title.Valid {
			record.Title = title.String
		}
		if meta.Valid && meta.String != "" {
			record.Meta = []byte(meta.String)
		}
		if prevHash.Valid {
			record.PrevHash = prevHash.String
		}
		intents = append(intents, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return intents, nil
}
