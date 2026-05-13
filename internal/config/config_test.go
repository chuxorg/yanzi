package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDefaultsWhenMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Mode != ModeLocal {
		t.Fatalf("expected mode %q, got %q", ModeLocal, cfg.Mode)
	}
	wantDB := filepath.Join(home, ".yanzi", "yanzi.db")
	if cfg.DBPath != wantDB {
		t.Fatalf("expected db_path %q, got %q", wantDB, cfg.DBPath)
	}
	if cfg.BaseURL != "" {
		t.Fatalf("expected empty base_url, got %q", cfg.BaseURL)
	}
}

func TestLoadInvalidMode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".yanzi", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("mode: nope\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
	if !strings.Contains(err.Error(), "invalid mode") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadHTTPRequiresBaseURL(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".yanzi", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("mode: http\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing base_url")
	}
	if !strings.Contains(err.Error(), "base_url is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadTrimsValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".yanzi", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	data := "mode: http\nbase_url: ' https://example.com/ '\n"
	if err := os.WriteFile(configPath, []byte(data), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.BaseURL != "https://example.com/" {
		t.Fatalf("expected trimmed base_url, got %q", cfg.BaseURL)
	}
}

func TestLoadRejectsMultipleDocuments(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".yanzi", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	data := "mode: local\n---\nmode: local\n"
	if err := os.WriteFile(configPath, []byte(data), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for multiple YAML documents")
	}
	if !strings.Contains(err.Error(), "multiple YAML documents") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEffectiveLocalDBPathPrefersEnvOverride(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(LocalDBPathEnvVar, "/tmp/yanzi-env.db")

	path, err := EffectiveLocalDBPath(Config{DBPath: "/tmp/yanzi-config.db"})
	if err != nil {
		t.Fatalf("EffectiveLocalDBPath: %v", err)
	}
	if path != "/tmp/yanzi-env.db" {
		t.Fatalf("expected env path, got %q", path)
	}
}

func TestEffectiveLocalDBPathFallsBackToConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(LocalDBPathEnvVar, "")

	path, err := EffectiveLocalDBPath(Config{DBPath: "/tmp/yanzi-config.db"})
	if err != nil {
		t.Fatalf("EffectiveLocalDBPath: %v", err)
	}
	if path != "/tmp/yanzi-config.db" {
		t.Fatalf("expected config path, got %q", path)
	}
}

func TestEffectiveLocalDBPathFallsBackToDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(LocalDBPathEnvVar, "")

	path, err := EffectiveLocalDBPath(Config{})
	if err != nil {
		t.Fatalf("EffectiveLocalDBPath: %v", err)
	}
	want := filepath.Join(home, ".yanzi", "yanzi.db")
	if path != want {
		t.Fatalf("expected default path %q, got %q", want, path)
	}
}
