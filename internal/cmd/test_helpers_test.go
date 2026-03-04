package cmd

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

func writeTestConfig(t *testing.T, home string) {
	t.Helper()

	stateDir := filepath.Join(home, ".yanzi")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatalf("create state dir: %v", err)
	}
	dbPath := filepath.Join(stateDir, "yanzi.db")
	configPath := filepath.Join(stateDir, "config.yaml")
	content := []byte("mode: local\ndb_path: " + dbPath + "\n")
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func writeStateFile(t *testing.T, dir, project string) {
	t.Helper()

	stateDir := filepath.Join(dir, ".yanzi")
	path := filepath.Join(stateDir, "state.json")
	payload := map[string]string{"active_project": project}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("encode state: %v", err)
	}
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatalf("create state dir: %v", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		t.Fatalf("write state file: %v", err)
	}
}

func createTestProject(t *testing.T, name string) {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if err := os.Setenv("YANZI_DB_PATH", cfg.DBPath); err != nil {
		t.Fatalf("set YANZI_DB_PATH: %v", err)
	}

	if _, err := yanzilibrary.CreateProject(name, ""); err != nil {
		t.Fatalf("create project: %v", err)
	}
}

func seedProject(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO projects (name, description, created_at, prev_hash, hash) VALUES (?, ?, ?, ?, ?)`,
		name,
		nil,
		"2025-01-01T00:00:00Z",
		nil,
		"seed-hash",
	)
	if err != nil {
		t.Fatalf("seed project: %v", err)
	}
}

func captureStdout(fn func() error) (string, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return "", err
	}

	stdout := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = stdout
	}()

	runErr := fn()
	_ = writer.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, reader)
	_ = reader.Close()

	return buf.String(), runErr
}
