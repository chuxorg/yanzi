package registry

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/storage"
	"github.com/chuxorg/yanzi/internal/storage/postgres"
	"github.com/chuxorg/yanzi/internal/storage/sqlite"
)

// Options contains provider construction inputs.
type Options struct{}

// Open returns the configured storage provider based on the effective provider
// name resolved from config and environment variables.
func Open(ctx context.Context, cfg config.Config, opts Options) (storage.Provider, error) {
	provider := config.EffectiveStorageProvider(cfg)
	switch provider {
	case string(storage.ProviderPostgres):
		p, err := postgres.NewProvider(cfg.Storage.Postgres)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize storage provider: %w", err)
		}
		return p, nil
	default:
		path, err := config.EffectiveLocalDBPath(cfg)
		if err != nil {
			return nil, err
		}
		p, _, err := sqlite.Open(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize storage provider: %w", err)
		}
		return p, nil
	}
}

// OpenAtPath returns the SQLite provider at a specific path.
func OpenAtPath(ctx context.Context, path string, opts Options) (storage.Provider, bool, error) {
	return sqlite.Open(ctx, path)
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

// ValidateProviderName reports whether the provider name is supported.
func ValidateProviderName(name string) error {
	switch strings.TrimSpace(name) {
	case "", string(storage.ProviderSQLite), string(storage.ProviderPostgres):
		return nil
	default:
		return storage.ErrUnsupportedProvider
	}
}
