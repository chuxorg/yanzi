package sqliteruntime

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestOpenConfiguresWALAndBusyTimeout(t *testing.T) {
	path := filepath.Join(t.TempDir(), "yanzi.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	var journalMode string
	if err := db.QueryRow(`PRAGMA journal_mode;`).Scan(&journalMode); err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Fatalf("expected WAL mode, got %q", journalMode)
	}

	var timeout int
	if err := db.QueryRow(`PRAGMA busy_timeout;`).Scan(&timeout); err != nil {
		t.Fatalf("query busy_timeout: %v", err)
	}
	if timeout != int(busyTimeout/time.Millisecond) {
		t.Fatalf("expected busy_timeout %d, got %d", int(busyTimeout/time.Millisecond), timeout)
	}
}

func TestExecContextRetriesTemporaryWriteLock(t *testing.T) {
	withShortSQLiteRetryConfig(t, 400*time.Millisecond, 20*time.Millisecond, 8, 5)

	path := filepath.Join(t.TempDir(), "yanzi.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open writer db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE intents (id TEXT PRIMARY KEY, value TEXT NOT NULL);`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	blocker, err := Open(path)
	if err != nil {
		t.Fatalf("Open blocker db: %v", err)
	}
	defer blocker.Close()

	tx, err := blocker.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("BeginTx blocker: %v", err)
	}
	if _, err := tx.ExecContext(context.Background(), `INSERT INTO intents (id, value) VALUES (?, ?)`, "held", "lock"); err != nil {
		t.Fatalf("seed blocker row: %v", err)
	}

	go func() {
		time.Sleep(120 * time.Millisecond)
		_ = tx.Commit()
	}()

	if _, err := ExecContext(context.Background(), db, path, "insert intent", `INSERT INTO intents (id, value) VALUES (?, ?)`, "released", "ok"); err != nil {
		t.Fatalf("ExecContext after transient lock: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(1) FROM intents`).Scan(&count); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 rows after recovery, got %d", count)
	}
}

func TestExecContextReturnsActionableLockErrorAndDatabaseRecovers(t *testing.T) {
	withShortSQLiteRetryConfig(t, 40*time.Millisecond, 10*time.Millisecond, 2, 2)

	path := filepath.Join(t.TempDir(), "yanzi.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open writer db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE intents (id TEXT PRIMARY KEY, value TEXT NOT NULL);`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	blocker, err := Open(path)
	if err != nil {
		t.Fatalf("Open blocker db: %v", err)
	}
	defer blocker.Close()

	tx, err := blocker.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("BeginTx blocker: %v", err)
	}
	if _, err := tx.ExecContext(context.Background(), `INSERT INTO intents (id, value) VALUES (?, ?)`, "held", "lock"); err != nil {
		t.Fatalf("seed blocker row: %v", err)
	}

	err = func() error {
		_, execErr := ExecContext(context.Background(), db, path, "insert intent", `INSERT INTO intents (id, value) VALUES (?, ?)`, "blocked", "nope")
		return execErr
	}()
	if err == nil {
		t.Fatal("expected lock error")
	}
	if !strings.Contains(err.Error(), "sqlite database is locked by another writer; retry shortly") {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit blocker: %v", err)
	}

	if _, err := ExecContext(context.Background(), db, path, "insert intent", `INSERT INTO intents (id, value) VALUES (?, ?)`, "after", "ok"); err != nil {
		t.Fatalf("ExecContext after lock release: %v", err)
	}
}

func TestOpenReportsActionablePathError(t *testing.T) {
	tempDir := t.TempDir()
	parentFile := filepath.Join(tempDir, "not-a-dir")
	if err := os.WriteFile(parentFile, []byte("x"), 0o600); err != nil {
		t.Fatalf("write parent file: %v", err)
	}

	_, err := Open(filepath.Join(parentFile, "yanzi.db"))
	if err == nil {
		t.Fatal("expected path error")
	}
	if !strings.Contains(err.Error(), "unable to open sqlite database") && !strings.Contains(err.Error(), "prepare sqlite directory") {
		t.Fatalf("unexpected path error: %v", err)
	}
}

func withShortSQLiteRetryConfig(t *testing.T, timeout, delay time.Duration, openRetries, writeRetries int) {
	t.Helper()

	originalTimeout := busyTimeout
	originalDelay := retryDelay
	originalOpenRetries := openRetryCount
	originalWriteRetries := writeRetryCount

	busyTimeout = timeout
	retryDelay = delay
	openRetryCount = openRetries
	writeRetryCount = writeRetries

	t.Cleanup(func() {
		busyTimeout = originalTimeout
		retryDelay = originalDelay
		openRetryCount = originalOpenRetries
		writeRetryCount = originalWriteRetries
	})
}
