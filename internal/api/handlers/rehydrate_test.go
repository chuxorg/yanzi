package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/core/hash"
	"github.com/chuxorg/yanzi/internal/core/model"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
	"github.com/chuxorg/yanzi/internal/storage"
	_ "modernc.org/sqlite"
)

func TestRehydrateHandlerReturnsDeterministicPayload(t *testing.T) {
	workdir := t.TempDir()
	withHandlerWorkdir(t, workdir)
	t.Setenv("HOME", workdir)
	writeHandlerProjectBinding(t, workdir, "alpha")
	db := openHandlerRehydrateDB(t, workdir)
	defer db.Close()

	seedHandlerProject(t, db, "alpha")
	seedHandlerCheckpoint(t, db, "alpha", "2026-01-01T00:00:10Z", "checkpoint summary")
	seedHandlerIntent(t, db, handlerRehydrateSeedIntent{
		ID:        "intent-b",
		CreatedAt: "2026-01-01T00:00:11Z",
		Project:   "alpha",
		Author:    "Ada",
		Title:     "Follow up",
		Prompt:    "prompt b",
		Response:  "response b",
		Meta: map[string]string{
			"project": "alpha",
			"area":    "auth",
		},
	})
	seedHandlerIntent(t, db, handlerRehydrateSeedIntent{
		ID:        "intent-a",
		CreatedAt: "2026-01-01T00:00:11Z",
		Project:   "alpha",
		Author:    "Ada",
		Title:     "Primary",
		Prompt:    strings.Repeat("Prompt a ", 30),
		Response:  strings.Repeat("Response a ", 30),
		Meta: map[string]string{
			"project": "alpha",
			"area":    "auth",
		},
	})

	provider := &stubProvider{
		health: storageHealthReady(),
		db:     db,
	}
	handler := NewRehydrateHandler(Dependencies{
		LoadConfig: func() (config.Config, error) {
			return config.Config{Mode: config.ModeLocal}, nil
		},
		OpenProvider: func(context.Context, config.Config) (storage.Provider, error) {
			return provider, nil
		},
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v0/rehydrate", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", rec.Code, rec.Body.String())
	}

	var resp models.RehydrateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Project != "alpha" || !resp.HasCheckpoint || resp.Fallback {
		t.Fatalf("unexpected top-level payload: %+v", resp)
	}
	if resp.Checkpoint == nil || resp.Checkpoint.Summary != "checkpoint summary" {
		t.Fatalf("unexpected checkpoint payload: %+v", resp.Checkpoint)
	}
	if len(resp.Intents) != 2 {
		t.Fatalf("expected 2 intents, got %d", len(resp.Intents))
	}
	if resp.Intents[0].ID != "intent-a" || resp.Intents[1].ID != "intent-b" {
		t.Fatalf("unexpected ordering: %+v", resp.Intents)
	}
	if resp.Intents[0].PromptSnippet == "" || !strings.HasSuffix(resp.Intents[0].PromptSnippet, "...") {
		t.Fatalf("expected truncated prompt snippet: %+v", resp.Intents[0])
	}
	if resp.Intents[0].Metadata["area"] != "auth" || resp.Intents[0].Metadata["project"] != "alpha" {
		t.Fatalf("unexpected metadata: %+v", resp.Intents[0].Metadata)
	}
	if !provider.closed {
		t.Fatal("expected provider to be closed")
	}
}

func TestRehydrateHandlerFallsBackWithoutCheckpoint(t *testing.T) {
	workdir := t.TempDir()
	withHandlerWorkdir(t, workdir)
	t.Setenv("HOME", workdir)
	writeHandlerProjectBinding(t, workdir, "alpha")
	db := openHandlerRehydrateDB(t, workdir)
	defer db.Close()

	seedHandlerProject(t, db, "alpha")
	seedHandlerIntent(t, db, handlerRehydrateSeedIntent{
		ID:        "fallback-1",
		CreatedAt: "2026-01-01T00:00:01Z",
		Project:   "alpha",
		Author:    "Ada",
		Prompt:    "prompt",
		Response:  "response",
		Meta:      map[string]string{"project": "alpha"},
	})

	provider := &stubProvider{
		health: storageHealthReady(),
		db:     db,
	}
	handler := NewRehydrateHandler(Dependencies{
		LoadConfig: func() (config.Config, error) {
			return config.Config{Mode: config.ModeLocal}, nil
		},
		OpenProvider: func(context.Context, config.Config) (storage.Provider, error) {
			return provider, nil
		},
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v0/rehydrate", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", rec.Code, rec.Body.String())
	}
	var resp models.RehydrateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Fallback || resp.FallbackReason != yanzilibrary.ErrCheckpointNotFound.Error() {
		t.Fatalf("unexpected fallback payload: %+v", resp)
	}
	if resp.Checkpoint != nil {
		t.Fatalf("did not expect checkpoint payload: %+v", resp.Checkpoint)
	}
}

func TestRehydrateHandlerReturnsMissingProject(t *testing.T) {
	workdir := t.TempDir()
	withHandlerWorkdir(t, workdir)
	t.Setenv("HOME", workdir)
	writeHandlerProjectBinding(t, workdir, "missing")
	db := openHandlerRehydrateDB(t, workdir)
	defer db.Close()

	handler := NewRehydrateHandler(Dependencies{
		LoadConfig: func() (config.Config, error) {
			return config.Config{Mode: config.ModeLocal}, nil
		},
		OpenProvider: func(context.Context, config.Config) (storage.Provider, error) {
			return &stubProvider{health: storageHealthReady(), db: db}, nil
		},
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v0/rehydrate", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%q", rec.Code, rec.Body.String())
	}
	var resp responses.ErrorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error.Code != "project_not_found" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}

func TestRehydrateHandlerRequiresActiveProject(t *testing.T) {
	workdir := t.TempDir()
	withHandlerWorkdir(t, workdir)
	t.Setenv("HOME", workdir)
	db := openHandlerRehydrateDB(t, workdir)
	defer db.Close()

	handler := NewRehydrateHandler(Dependencies{
		LoadConfig: func() (config.Config, error) {
			return config.Config{Mode: config.ModeLocal}, nil
		},
		OpenProvider: func(context.Context, config.Config) (storage.Provider, error) {
			return &stubProvider{health: storageHealthReady(), db: db}, nil
		},
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v0/rehydrate", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%q", rec.Code, rec.Body.String())
	}
	var resp responses.ErrorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error.Code != "active_project_not_set" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}

func openHandlerRehydrateDB(t *testing.T, dir string) *sql.DB {
	t.Helper()
	path := filepath.Join(dir, "yanzi.db")
	t.Setenv("YANZI_DB_PATH", path)
	db, err := yanzilibrary.InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	return db
}

func withHandlerWorkdir(t *testing.T, dir string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})
}

func writeHandlerProjectBinding(t *testing.T, dir, project string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".yanzi"), 0o700); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".yanzi", "project"), []byte(project+"\n"), 0o644); err != nil {
		t.Fatalf("write project binding: %v", err)
	}
}

func seedHandlerProject(t *testing.T, _ *sql.DB, name string) {
	t.Helper()
	if _, err := yanzilibrary.CreateProject(name, ""); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
}

func seedHandlerCheckpoint(t *testing.T, db *sql.DB, project, createdAt, summary string) {
	t.Helper()
	checkpoint := yanzilibrary.Checkpoint{
		Project:     project,
		Summary:     summary,
		CreatedAt:   createdAt,
		ArtifactIDs: []string{},
	}
	hashValue, err := yanzilibrary.HashCheckpoint(checkpoint)
	if err != nil {
		t.Fatalf("HashCheckpoint: %v", err)
	}
	artifactIDsJSON, err := json.Marshal([]string{})
	if err != nil {
		t.Fatalf("marshal artifact ids: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO checkpoints (hash, project, summary, created_at, artifact_ids, previous_checkpoint_id)
		VALUES (?, ?, ?, ?, ?, ?)`,
		hashValue,
		project,
		summary,
		createdAt,
		string(artifactIDsJSON),
		nil,
	); err != nil {
		t.Fatalf("seed checkpoint: %v", err)
	}
}

type handlerRehydrateSeedIntent struct {
	ID        string
	CreatedAt string
	Project   string
	Author    string
	Title     string
	Prompt    string
	Response  string
	Meta      map[string]string
}

func seedHandlerIntent(t *testing.T, db *sql.DB, input handlerRehydrateSeedIntent) {
	t.Helper()
	meta := input.Meta
	if meta == nil {
		meta = map[string]string{"project": input.Project}
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal meta: %v", err)
	}
	record := model.IntentRecord{
		ID:         input.ID,
		CreatedAt:  input.CreatedAt,
		Author:     input.Author,
		SourceType: "cli",
		Title:      input.Title,
		Prompt:     input.Prompt,
		Response:   input.Response,
		Meta:       metaJSON,
	}
	record.Hash, err = hash.HashIntent(record)
	if err != nil {
		t.Fatalf("HashIntent: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO intents (id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ID,
		record.CreatedAt,
		record.Author,
		record.SourceType,
		nullIfEmptyString(record.Title),
		record.Prompt,
		record.Response,
		string(record.Meta),
		nil,
		record.Hash,
	); err != nil {
		t.Fatalf("seed intent: %v", err)
	}
}

func storageHealthReady() storage.Health {
	return storage.Health{Provider: storage.ProviderSQLite, Status: storage.HealthReady}
}

func nullIfEmptyString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
