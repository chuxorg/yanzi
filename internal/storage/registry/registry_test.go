package registry_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/storage"
	"github.com/chuxorg/yanzi/internal/storage/registry"
)

func TestOpenReturnsSQLiteProviderFromCurrentConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "configured.db")
	provider, err := registry.Open(context.Background(), config.Config{Mode: config.ModeLocal, DBPath: path}, registry.Options{})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer provider.Close()

	if provider.Name() != storage.ProviderSQLite {
		t.Fatalf("expected sqlite provider, got %q", provider.Name())
	}
	if provider.SQLDB() == nil {
		t.Fatalf("expected provider SQLDB handle")
	}
}

func TestValidateProviderNameRejectsFutureProviders(t *testing.T) {
	if err := registry.ValidateProviderName(""); err != nil {
		t.Fatalf("empty provider should select default sqlite: %v", err)
	}
	if err := registry.ValidateProviderName("sqlite"); err != nil {
		t.Fatalf("sqlite provider should be valid: %v", err)
	}
	if err := registry.ValidateProviderName("postgres"); err != nil {
		t.Fatalf("postgres should be a valid provider name: %v", err)
	}
	if err := registry.ValidateProviderName("mysql"); err == nil {
		t.Fatalf("expected unsupported provider error for unknown provider")
	}
}
