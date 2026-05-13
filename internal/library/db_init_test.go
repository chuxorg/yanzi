package yanzilibrary

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/chuxorg/yanzi/internal/config"
	_ "modernc.org/sqlite"
)

func TestInitializeCreatesRuntimeState(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envDBPath, "")

	initialized, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if !initialized {
		t.Fatalf("expected first initialization")
	}

	dir := filepath.Join(home, defaultDBDirName)
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("stat runtime dir: %v", err)
	}

	dbPath := filepath.Join(dir, defaultDBFile)
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("stat db file: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	assertTableExists(t, db, "schema_version")
	assertTableExists(t, db, "schema_migrations")
	assertTableExists(t, db, "intents")
	assertTableExists(t, db, "projects")
	assertTableExists(t, db, "checkpoints")

	var version int
	if err := db.QueryRow(`SELECT version FROM schema_version LIMIT 1`).Scan(&version); err != nil {
		t.Fatalf("query schema_version: %v", err)
	}
	if version != 1 {
		t.Fatalf("expected schema version 1, got %d", version)
	}
}

func TestInitializeIsIdempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envDBPath, "")

	first, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize first: %v", err)
	}
	if !first {
		t.Fatalf("expected first initialization")
	}

	second, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize second: %v", err)
	}
	if second {
		t.Fatalf("expected subsequent initialization to be silent")
	}

	dbPath := filepath.Join(home, defaultDBDirName, defaultDBFile)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow(`SELECT COUNT(1) FROM schema_version`).Scan(&count); err != nil {
		t.Fatalf("count schema_version rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected single schema_version row, got %d", count)
	}
}

func TestInitializeRecreatesDatabaseWhenDeleted(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envDBPath, "")

	if _, err := Initialize(); err != nil {
		t.Fatalf("Initialize first: %v", err)
	}

	dbPath := filepath.Join(home, defaultDBDirName, defaultDBFile)
	if err := os.Remove(dbPath); err != nil {
		t.Fatalf("remove db: %v", err)
	}

	initialized, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize after delete: %v", err)
	}
	if !initialized {
		t.Fatalf("expected initialization after db deletion")
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("stat recreated db: %v", err)
	}
}

func TestInitDBUsesConfigDBPathWhenEnvMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(config.LocalDBPathEnvVar, "")

	configPath := filepath.Join(home, ".yanzi", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	configDBPath := filepath.Join(home, "data", "config.db")
	if err := os.WriteFile(configPath, []byte("mode: local\ndb_path: "+configDBPath+"\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	db, err := InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	if got := ResolvedDBPath(); got != configDBPath {
		t.Fatalf("expected resolved config path %q, got %q", configDBPath, got)
	}
	if _, err := os.Stat(configDBPath); err != nil {
		t.Fatalf("stat config db: %v", err)
	}
}

func TestInitDBPrefersEnvOverrideOverConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".yanzi", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	configDBPath := filepath.Join(home, "data", "config.db")
	if err := os.WriteFile(configPath, []byte("mode: local\ndb_path: "+configDBPath+"\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	envDBPath := filepath.Join(home, "override", "env.db")
	t.Setenv(config.LocalDBPathEnvVar, envDBPath)

	db, err := InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	if got := ResolvedDBPath(); got != envDBPath {
		t.Fatalf("expected env db path %q, got %q", envDBPath, got)
	}
	if _, err := os.Stat(envDBPath); err != nil {
		t.Fatalf("stat env db: %v", err)
	}
	if _, err := os.Stat(configDBPath); !os.IsNotExist(err) {
		t.Fatalf("expected config db to remain unused, stat err=%v", err)
	}
}

func assertTableExists(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	var count int
	if err := db.QueryRow(`SELECT COUNT(1) FROM sqlite_master WHERE type = 'table' AND name = ?`, name).Scan(&count); err != nil {
		t.Fatalf("query sqlite_master for %s: %v", name, err)
	}
	if count != 1 {
		t.Fatalf("expected table %s to exist", name)
	}
}
