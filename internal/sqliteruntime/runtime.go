package sqliteruntime

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	busyTimeout     = 5 * time.Second
	retryDelay      = 75 * time.Millisecond
	openRetryCount  = 6
	writeRetryCount = 4
)

// PreparePath ensures the parent directory and database file exist.
func PreparePath(path string) error {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return errors.New("sqlite path is required")
	}

	dir := filepath.Dir(trimmed)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return normalizeError("prepare sqlite directory", trimmed, err)
	}

	file, err := os.OpenFile(trimmed, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return normalizeError("prepare sqlite file", trimmed, err)
	}
	if err := file.Close(); err != nil {
		return normalizeError("close sqlite file", trimmed, err)
	}
	return nil
}

// Open returns a configured SQLite handle with deterministic runtime pragmas.
func Open(path string) (*sql.DB, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return nil, errors.New("sqlite path is required")
	}
	if err := PreparePath(trimmed); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", trimmed)
	if err != nil {
		return nil, normalizeError("open sqlite database", trimmed, err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxIdleTime(0)
	db.SetConnMaxLifetime(0)

	if err := retry(context.Background(), openRetryCount, trimmed, "connect sqlite database", func() error {
		return db.Ping()
	}); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := configurePragmas(db, trimmed); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// ExecContext runs a write operation with bounded busy retries.
func ExecContext(ctx context.Context, db *sql.DB, path, operation, query string, args ...any) (sql.Result, error) {
	var result sql.Result
	err := retry(ctx, writeRetryCount, path, operation, func() error {
		var execErr error
		result, execErr = db.ExecContext(ctx, query, args...)
		return execErr
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ExecTxContext runs a transactional write operation with bounded busy retries.
func ExecTxContext(ctx context.Context, tx *sql.Tx, path, operation, query string, args ...any) (sql.Result, error) {
	var result sql.Result
	err := retry(ctx, writeRetryCount, path, operation, func() error {
		var execErr error
		result, execErr = tx.ExecContext(ctx, query, args...)
		return execErr
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// RunTx retries the whole transaction when SQLite reports bounded contention.
func RunTx(ctx context.Context, db *sql.DB, path, operation string, fn func(*sql.Tx) error) error {
	return retry(ctx, writeRetryCount, path, operation, func() error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		committed := false
		defer func() {
			if !committed {
				_ = tx.Rollback()
			}
		}()

		if err := fn(tx); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		committed = true
		return nil
	})
}

// NormalizeError rewrites low-signal SQLite errors into actionable operator diagnostics.
func NormalizeError(operation, path string, err error) error {
	return normalizeError(operation, path, err)
}

func configurePragmas(db *sql.DB, path string) error {
	pragmas := []struct {
		operation string
		query     string
	}{
		{operation: "enable sqlite WAL mode", query: `PRAGMA journal_mode=WAL;`},
		{operation: "enable sqlite foreign keys", query: `PRAGMA foreign_keys=ON;`},
		{operation: "set sqlite busy timeout", query: fmt.Sprintf(`PRAGMA busy_timeout=%d;`, busyTimeout/time.Millisecond)},
	}

	for _, pragma := range pragmas {
		if err := retry(context.Background(), openRetryCount, path, pragma.operation, func() error {
			_, execErr := db.Exec(pragma.query)
			return execErr
		}); err != nil {
			return err
		}
	}
	return nil
}

func retry(ctx context.Context, attempts int, path, operation string, fn func() error) error {
	if attempts <= 0 {
		attempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if ctx != nil && ctx.Err() != nil {
			return ctx.Err()
		}

		err := fn()
		if err == nil {
			return nil
		}
		if !isBusyError(err) {
			return normalizeError(operation, path, err)
		}
		lastErr = err
		if attempt == attempts {
			break
		}
		if sleepErr := sleepWithContext(ctx, retryDelay); sleepErr != nil {
			return sleepErr
		}
	}

	return normalizeError(operation, path, lastErr)
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()

	if ctx == nil {
		<-timer.C
		return nil
	}

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func normalizeError(operation, path string, err error) error {
	if err == nil {
		return nil
	}
	trimmedPath := strings.TrimSpace(path)
	lower := strings.ToLower(err.Error())

	switch {
	case isBusyError(err):
		if trimmedPath == "" {
			return fmt.Errorf("%s: sqlite database is locked by another writer; retry shortly: %w", operation, err)
		}
		return fmt.Errorf("%s at %s: sqlite database is locked by another writer; retry shortly: %w", operation, trimmedPath, err)
	case strings.Contains(lower, "permission denied"), strings.Contains(lower, "readonly"), strings.Contains(lower, "read-only"):
		if trimmedPath == "" {
			return fmt.Errorf("%s: cannot access sqlite database; check filesystem permissions: %w", operation, err)
		}
		return fmt.Errorf("%s at %s: cannot access sqlite database; check filesystem permissions: %w", operation, trimmedPath, err)
	case strings.Contains(lower, "unable to open database file"), strings.Contains(lower, "not a directory"):
		if trimmedPath == "" {
			return fmt.Errorf("%s: unable to open sqlite database; check that the path exists and is writable: %w", operation, err)
		}
		return fmt.Errorf("%s at %s: unable to open sqlite database; check that the path exists and is writable: %w", operation, trimmedPath, err)
	case strings.Contains(lower, "file is not a database"), strings.Contains(lower, "malformed"), strings.Contains(lower, "corrupt"):
		if trimmedPath == "" {
			return fmt.Errorf("%s: sqlite database appears unreadable or corrupted: %w", operation, err)
		}
		return fmt.Errorf("%s at %s: sqlite database appears unreadable or corrupted: %w", operation, trimmedPath, err)
	default:
		if trimmedPath == "" {
			return fmt.Errorf("%s: %w", operation, err)
		}
		return fmt.Errorf("%s at %s: %w", operation, trimmedPath, err)
	}
}

func isBusyError(err error) bool {
	if err == nil {
		return false
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "database is locked") ||
		strings.Contains(lower, "database table is locked") ||
		strings.Contains(lower, "sqlite_busy") ||
		strings.Contains(lower, "sqlite_locked") ||
		strings.Contains(lower, "busy")
}
