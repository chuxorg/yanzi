package yanzilibrary

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/storage/registry"
)

const (
	envDBPath        = config.LocalDBPathEnvVar
	defaultDBDirName = ".yanzi"
	defaultDBFile    = "yanzi.db"
)

var (
	resolvedDBPath string
	resolvedMu     sync.RWMutex
)

// Initialize ensures the default yanzi runtime directory, database, and schema exist.
// It returns true when this run performed first-time schema initialization.
func Initialize() (bool, error) {
	path, err := resolveDBPath()
	if err != nil {
		return false, err
	}

	db, initialized, err := openInitializedDB(path)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = db.Close()
	}()

	setResolvedDBPath(path)
	return initialized, nil
}

// InitDB resolves the database path, ensures migrations, and returns a SQLite handle.
func InitDB() (*sql.DB, error) {
	path, err := resolveDBPath()
	if err != nil {
		return nil, err
	}

	return InitDBAtPath(path)
}

// InitDBAtPath ensures migrations and returns a SQLite handle for the provided path.
func InitDBAtPath(path string) (*sql.DB, error) {
	db, _, err := openInitializedDB(path)
	if err != nil {
		return nil, err
	}

	setResolvedDBPath(path)
	return db, nil
}

// ResolvedDBPath returns the most recently resolved database path.
func ResolvedDBPath() string {
	resolvedMu.RLock()
	defer resolvedMu.RUnlock()
	return resolvedDBPath
}

// setResolvedDBPath updates the in-memory record of the database path used by InitDB.
func setResolvedDBPath(path string) {
	resolvedMu.Lock()
	resolvedDBPath = path
	resolvedMu.Unlock()
}

// resolveDBPath determines the SQLite path using the shared CLI/library precedence rules.
func resolveDBPath() (string, error) {
	if override := strings.TrimSpace(os.Getenv(config.LocalDBPathEnvVar)); override != "" {
		return override, nil
	}

	cfg, err := config.Load()
	if err != nil {
		return "", err
	}

	path, err := config.EffectiveLocalDBPath(cfg)
	if err != nil {
		return "", err
	}

	dir, err := config.StateDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("ensure db dir: %w", err)
	}
	return path, nil
}

func openInitializedDB(path string) (*sql.DB, bool, error) {
	provider, initialized, err := registry.OpenAtPath(context.Background(), path, registry.Options{Migrations: MigrationsFS()})
	if err != nil {
		return nil, false, err
	}
	db := provider.SQLDB()
	if db == nil {
		_ = provider.Close()
		return nil, false, fmt.Errorf("storage provider returned nil database")
	}
	return db, initialized, nil
}
