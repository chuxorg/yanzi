package registry

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/storage"
	"github.com/chuxorg/yanzi/internal/storage/sqlite"
)

// Options contains provider construction inputs that are not user-facing config.
type Options struct {
	Migrations fs.FS
}

// Open returns the configured storage provider.
//
// CAP-001 Phase 1 supports SQLite only and preserves existing local db_path
// resolution. No provider config key is active yet.
func Open(ctx context.Context, cfg config.Config, opts Options) (storage.Provider, error) {
	path, err := config.EffectiveLocalDBPath(cfg)
	if err != nil {
		return nil, err
	}
	provider, _, err := sqlite.Open(ctx, path, opts.Migrations)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// OpenAtPath returns the SQLite provider at a specific path.
func OpenAtPath(ctx context.Context, path string, opts Options) (storage.Provider, bool, error) {
	return sqlite.Open(ctx, path, opts.Migrations)
}

// EnsureLocalStateDir preserves existing local SQLite directory creation.
func EnsureLocalStateDir() error {
	dir, err := config.StateDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("ensure db dir: %w", err)
	}
	return nil
}

// ValidateProviderName rejects future provider names until implementations exist.
func ValidateProviderName(name string) error {
	switch strings.TrimSpace(name) {
	case "", string(storage.ProviderSQLite):
		return nil
	default:
		return storage.ErrUnsupportedProvider
	}
}
