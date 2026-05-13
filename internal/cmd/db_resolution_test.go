package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

func TestOpenLocalDBPrefersEnvOverride(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configDBPath := filepath.Join(home, "config", "yanzi.db")
	writeTestConfigWithDBPath(t, home, configDBPath)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	envDBPath := filepath.Join(home, "env", "yanzi.db")
	t.Setenv(config.LocalDBPathEnvVar, envDBPath)

	db, err := openLocalDB(cfg)
	if err != nil {
		t.Fatalf("openLocalDB: %v", err)
	}
	defer db.Close()

	if got := yanzilibrary.ResolvedDBPath(); got != envDBPath {
		t.Fatalf("expected env db path %q, got %q", envDBPath, got)
	}
	if _, err := os.Stat(envDBPath); err != nil {
		t.Fatalf("stat env db path: %v", err)
	}
	if _, err := os.Stat(configDBPath); !os.IsNotExist(err) {
		t.Fatalf("expected config db path to remain unused, stat err=%v", err)
	}
}

func TestOpenLocalDBAndLibraryShareResolvedPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	sharedDBPath := filepath.Join(home, "shared", "yanzi.db")
	writeTestConfigWithDBPath(t, home, sharedDBPath)
	t.Setenv(config.LocalDBPathEnvVar, "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	db, err := openLocalDB(cfg)
	if err != nil {
		t.Fatalf("openLocalDB: %v", err)
	}
	_ = db.Close()

	if _, err := yanzilibrary.CreateProject("alpha", ""); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if got := yanzilibrary.ResolvedDBPath(); got != sharedDBPath {
		t.Fatalf("expected shared db path %q, got %q", sharedDBPath, got)
	}
	if _, err := os.Stat(sharedDBPath); err != nil {
		t.Fatalf("stat shared db path: %v", err)
	}
}
