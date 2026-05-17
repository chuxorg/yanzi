package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
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

func TestRunCaptureHandlesLightConcurrentWriters(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	db, err := openLocalDB(cfg)
	if err != nil {
		t.Fatalf("openLocalDB: %v", err)
	}
	defer db.Close()

	locker, err := yanzilibrary.InitDBAtPath(cfg.DBPath)
	if err != nil {
		t.Fatalf("InitDBAtPath: %v", err)
	}
	defer locker.Close()

	tx, err := locker.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	if _, err := tx.ExecContext(context.Background(), `INSERT INTO projects (name, description, created_at, prev_hash, hash) VALUES (?, ?, ?, ?, ?)`,
		"lock-holder", "", time.Now().UTC().Format(time.RFC3339Nano), nil, "lock-holder-hash"); err != nil {
		t.Fatalf("seed lock holder: %v", err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 2)
	runCapture := func(prompt string) {
		defer wg.Done()
		errCh <- RunCapture([]string{
			"--author", "Ada",
			"--prompt", prompt,
			"--response", "ok",
		})
	}

	wg.Add(2)
	go runCapture("prompt one")
	go runCapture("prompt two")
	time.Sleep(120 * time.Millisecond)
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit lock holder: %v", err)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("RunCapture concurrent writer: %v", err)
		}
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(1) FROM intents WHERE source_type = 'cli'`).Scan(&count); err != nil {
		t.Fatalf("count intents: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 captured intents, got %d", count)
	}
}
