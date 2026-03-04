package store

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chuxorg/yanzi/internal/core/model"
)

func TestStoreMigrateAndCRUD(t *testing.T) {
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatalf("mkdir migrations: %v", err)
	}

	migration := `
CREATE TABLE IF NOT EXISTS intents (
	id TEXT PRIMARY KEY,
	created_at TEXT NOT NULL,
	author TEXT NOT NULL,
	source_type TEXT NOT NULL,
	title TEXT,
	prompt TEXT NOT NULL,
	response TEXT NOT NULL,
	meta TEXT,
	prev_hash TEXT,
	hash TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS intents_hash_idx ON intents(hash);
`
	migrationPath := filepath.Join(migrationsDir, "001_init.sql")
	if err := os.WriteFile(migrationPath, []byte(migration), 0o644); err != nil {
		t.Fatalf("write migration: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	dbPath := filepath.Join(tempDir, "test.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	ctx := context.Background()
	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	intent := model.IntentRecord{
		ID:         "01HZYFQ7T9ZV54X2G4A8M4J2C1",
		CreatedAt:  time.Now().UTC().Format(time.RFC3339Nano),
		Author:     "alice",
		SourceType: "cli",
		Title:      "hello",
		Prompt:     "prompt",
		Response:   "response",
		Meta:       json.RawMessage(`{"env":"prod"}`),
		PrevHash:   "prevhash",
		Hash:       "hashvalue",
	}

	if err := store.CreateIntent(ctx, intent); err != nil {
		t.Fatalf("create intent: %v", err)
	}

	loaded, err := store.GetIntent(ctx, intent.ID)
	if err != nil {
		t.Fatalf("get intent: %v", err)
	}
	if loaded.ID != intent.ID || loaded.Hash != intent.Hash || loaded.Title != intent.Title {
		t.Fatalf("unexpected loaded intent: %+v", loaded)
	}
	if string(loaded.Meta) != string(intent.Meta) {
		t.Fatalf("expected meta %s, got %s", intent.Meta, loaded.Meta)
	}

	byHash, err := store.GetIntentByHash(ctx, intent.Hash)
	if err != nil {
		t.Fatalf("get intent by hash: %v", err)
	}
	if byHash.ID != intent.ID {
		t.Fatalf("expected id %s, got %s", intent.ID, byHash.ID)
	}

	list, err := store.ListIntents(ctx, 10)
	if err != nil {
		t.Fatalf("list intents: %v", err)
	}
	if len(list) != 1 || list[0].ID != intent.ID {
		t.Fatalf("unexpected list result: %+v", list)
	}
}

func TestOpenEmptyPath(t *testing.T) {
	if _, err := Open(" "); err == nil {
		t.Fatalf("expected error for empty sqlite path")
	}
}
