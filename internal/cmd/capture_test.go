package cmd

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/config"
	_ "modernc.org/sqlite"
)

func TestKVPairsToJSONLastValueWins(t *testing.T) {
	pairs := kvPairs{"area=auth", "area=billing", "tags=migration,security"}

	raw, err := pairs.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}

	var meta map[string]string
	if err := json.Unmarshal(raw, &meta); err != nil {
		t.Fatalf("unmarshal meta: %v", err)
	}
	if meta["area"] != "billing" {
		t.Fatalf("expected last value to win for duplicate key, got %q", meta["area"])
	}
	if meta["tags"] != "migration,security" {
		t.Fatalf("unexpected tags value: %q", meta["tags"])
	}
}

func TestKVPairsToJSONMalformedArgument(t *testing.T) {
	pairs := kvPairs{"missing-separator"}

	_, err := pairs.ToJSON()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "expected key=value") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCaptureStoresProfileMetadata(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	if err := RunCapture([]string{
		"--author", "Ada",
		"--prompt", "Hello",
		"--response", "World",
		"--profile", "engineer",
		"--meta", "area=auth",
	}); err != nil {
		t.Fatalf("RunCapture: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var metaText string
	if err := db.QueryRow(`SELECT meta FROM intents ORDER BY rowid DESC LIMIT 1`).Scan(&metaText); err != nil {
		t.Fatalf("query meta: %v", err)
	}

	var meta map[string]string
	if err := json.Unmarshal([]byte(metaText), &meta); err != nil {
		t.Fatalf("decode meta: %v", err)
	}
	if meta["profile"] != "engineer" {
		t.Fatalf("expected profile metadata, got %#v", meta)
	}
	if meta["area"] != "auth" {
		t.Fatalf("expected existing metadata to remain, got %#v", meta)
	}
	if meta["type"] == "context" {
		t.Fatalf("did not expect generic capture to force context metadata: %#v", meta)
	}
}
