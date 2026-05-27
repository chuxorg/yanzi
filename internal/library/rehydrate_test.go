package yanzilibrary

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/chuxorg/yanzi/internal/core/hash"
	"github.com/chuxorg/yanzi/internal/core/model"
)

func TestRehydrateProjectReturnsCheckpointBoundaryAndSubsequentIntents(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envDBPath, "")

	if _, err := Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	db, err := InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	if _, err := CreateProject("alpha", ""); err != nil {
		t.Fatalf("CreateProject alpha: %v", err)
	}
	if _, err := CreateProject("beta", ""); err != nil {
		t.Fatalf("CreateProject beta: %v", err)
	}

	seedRehydrateCheckpoint(t, db, "alpha", "2025-01-01T00:00:10Z", "checkpoint summary")
	seedRehydrateIntent(t, db, rehydrateTestIntent{
		ID:        "before-checkpoint",
		CreatedAt: "2025-01-01T00:00:09Z",
		Project:   "alpha",
		Author:    "Ada",
		Prompt:    "ignore me",
		Response:  "ignore me",
	})
	seedRehydrateIntent(t, db, rehydrateTestIntent{
		ID:        "after-1",
		CreatedAt: "2025-01-01T00:00:11Z",
		Project:   "alpha",
		Author:    "Ada",
		Title:     "JWT validation edge cases",
		Prompt:    "Need to determine why refresh tokens fail after reconnect.",
		Response:  "Issue appears related to clock skew and token leeway handling.",
		Meta: map[string]string{
			"project": "alpha",
			"area":    "auth",
		},
	})
	seedRehydrateIntent(t, db, rehydrateTestIntent{
		ID:        "after-2",
		CreatedAt: "2025-01-01T00:00:12Z",
		Project:   "alpha",
		Author:    "Ada",
		Prompt:    "Second prompt",
		Response:  "Second response",
	})
	seedRehydrateIntent(t, db, rehydrateTestIntent{
		ID:        "other-project",
		CreatedAt: "2025-01-01T00:00:13Z",
		Project:   "beta",
		Author:    "Byron",
		Prompt:    "should not appear",
		Response:  "should not appear",
	})

	payload, err := RehydrateProject("alpha")
	if err != nil {
		t.Fatalf("RehydrateProject: %v", err)
	}

	if payload.Project != "alpha" {
		t.Fatalf("expected project alpha, got %q", payload.Project)
	}
	if payload.Fallback {
		t.Fatalf("did not expect fallback payload: %+v", payload)
	}
	if payload.LatestCheckpoint == nil {
		t.Fatal("expected checkpoint payload")
	}
	if payload.LatestCheckpoint.Summary != "checkpoint summary" {
		t.Fatalf("unexpected checkpoint summary: %+v", payload.LatestCheckpoint)
	}
	if len(payload.Intents) != 2 {
		t.Fatalf("expected 2 post-checkpoint intents, got %d", len(payload.Intents))
	}
	if payload.Intents[0].ID != "after-1" || payload.Intents[1].ID != "after-2" {
		t.Fatalf("unexpected intent ordering: %+v", payload.Intents)
	}
	if payload.Intents[0].Title != "JWT validation edge cases" {
		t.Fatalf("expected title to round-trip, got %+v", payload.Intents[0])
	}

	meta := decodeLibraryMeta(t, payload.Intents[0].Meta)
	if meta["project"] != "alpha" || meta["area"] != "auth" {
		t.Fatalf("unexpected merged meta: %+v", meta)
	}
}

func TestRehydrateProjectFallsBackToRecentProjectIntents(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envDBPath, "")

	if _, err := Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	db, err := InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	if _, err := CreateProject("alpha", ""); err != nil {
		t.Fatalf("CreateProject alpha: %v", err)
	}
	if _, err := CreateProject("beta", ""); err != nil {
		t.Fatalf("CreateProject beta: %v", err)
	}

	for i := 1; i <= 12; i++ {
		seedRehydrateIntent(t, db, rehydrateTestIntent{
			ID:        "cap-" + strconv.Itoa(i),
			CreatedAt: rehydrateTestTimestamp(i),
			Project:   "alpha",
			Author:    "Ada",
			Prompt:    "prompt " + strconv.Itoa(i),
			Response:  "response " + strconv.Itoa(i),
		})
	}
	seedRehydrateIntent(t, db, rehydrateTestIntent{
		ID:        "beta-cap",
		CreatedAt: "2025-01-01T00:00:13Z",
		Project:   "beta",
		Author:    "Byron",
		Prompt:    "ignore",
		Response:  "ignore",
	})

	payload, err := RehydrateProject("alpha")
	if err != nil {
		t.Fatalf("RehydrateProject fallback: %v", err)
	}

	if !payload.Fallback {
		t.Fatalf("expected fallback payload: %+v", payload)
	}
	if payload.LatestCheckpoint != nil {
		t.Fatalf("did not expect checkpoint: %+v", payload.LatestCheckpoint)
	}
	if payload.FallbackReason != ErrCheckpointNotFound.Error() {
		t.Fatalf("unexpected fallback reason: %q", payload.FallbackReason)
	}
	if payload.FallbackLimit != DefaultRehydrateFallbackLimit {
		t.Fatalf("unexpected fallback limit: %d", payload.FallbackLimit)
	}
	if len(payload.Intents) != DefaultRehydrateFallbackLimit {
		t.Fatalf("expected %d intents, got %d", DefaultRehydrateFallbackLimit, len(payload.Intents))
	}
	if payload.Intents[0].ID != "cap-3" || payload.Intents[len(payload.Intents)-1].ID != "cap-12" {
		t.Fatalf("unexpected fallback window/order: %+v", payload.Intents)
	}
}

func TestRehydrateCompatibilityAfterLegacyDatabaseUpgrade(t *testing.T) {
	home := t.TempDir()
	dbPath := filepath.Join(home, "state", "legacy.db")
	t.Setenv("HOME", home)
	t.Setenv(envDBPath, dbPath)

	createLegacyRehydrateDatabase(t, dbPath)

	payload, err := RehydrateProject("alpha")
	if err != nil {
		t.Fatalf("RehydrateProject upgraded legacy db: %v", err)
	}
	if payload.Fallback {
		t.Fatalf("did not expect fallback after legacy upgrade: %+v", payload)
	}
	if payload.LatestCheckpoint == nil || payload.LatestCheckpoint.Summary != "legacy checkpoint" {
		t.Fatalf("unexpected checkpoint after legacy upgrade: %+v", payload.LatestCheckpoint)
	}
	if len(payload.Intents) != 1 || payload.Intents[0].ID != "legacy-after" {
		t.Fatalf("unexpected post-checkpoint intents after legacy upgrade: %+v", payload.Intents)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open upgraded db: %v", err)
	}
	defer db.Close()
	for _, column := range []string{"class", "type", "content", "metadata"} {
		var count int
		if err := db.QueryRow(`SELECT COUNT(1) FROM pragma_table_info('intents') WHERE name = ?`, column).Scan(&count); err != nil {
			t.Fatalf("check upgraded column %s: %v", column, err)
		}
		if count != 1 {
			t.Fatalf("expected upgraded intents column %s", column)
		}
	}
}

func TestRehydrateCompatibilityRepeatedCheckpointComposition(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envDBPath, "")

	if _, err := Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	db, err := InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	if _, err := CreateProject("alpha", ""); err != nil {
		t.Fatalf("CreateProject alpha: %v", err)
	}
	seedRehydrateCheckpoint(t, db, "alpha", "2025-01-01T00:00:10Z", "stable checkpoint")
	seedRehydrateIntent(t, db, rehydrateTestIntent{
		ID:        "before",
		CreatedAt: "2025-01-01T00:00:09Z",
		Project:   "alpha",
		Author:    "Ada",
		Prompt:    "before prompt",
		Response:  "before response",
	})
	seedRehydrateIntent(t, db, rehydrateTestIntent{
		ID:        "after-a",
		CreatedAt: "2025-01-01T00:00:11Z",
		Project:   "alpha",
		Author:    "Ada",
		Prompt:    "after a",
		Response:  "response a",
	})
	seedRehydrateIntent(t, db, rehydrateTestIntent{
		ID:        "after-b",
		CreatedAt: "2025-01-01T00:00:12Z",
		Project:   "alpha",
		Author:    "Ada",
		Prompt:    "after b",
		Response:  "response b",
	})

	first, err := RehydrateProject("alpha")
	if err != nil {
		t.Fatalf("first RehydrateProject: %v", err)
	}
	second, err := RehydrateProject("alpha")
	if err != nil {
		t.Fatalf("second RehydrateProject: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated rehydration differed: first=%+v second=%+v", first, second)
	}
	if len(first.Intents) != 2 || first.Intents[0].ID != "after-a" || first.Intents[1].ID != "after-b" {
		t.Fatalf("unexpected repeated rehydration composition: %+v", first.Intents)
	}
}

type rehydrateTestIntent struct {
	ID        string
	CreatedAt string
	Project   string
	Author    string
	Title     string
	Prompt    string
	Response  string
	Meta      map[string]string
}

func seedRehydrateCheckpoint(t *testing.T, db *sql.DB, project, createdAt, summary string) {
	t.Helper()

	checkpoint := Checkpoint{
		Project:     project,
		Summary:     summary,
		CreatedAt:   createdAt,
		ArtifactIDs: []string{},
	}
	hashValue, err := HashCheckpoint(checkpoint)
	if err != nil {
		t.Fatalf("HashCheckpoint: %v", err)
	}

	artifactIDsJSON, err := json.Marshal([]string{})
	if err != nil {
		t.Fatalf("marshal artifact ids: %v", err)
	}

	if _, err := db.ExecContext(
		context.Background(),
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

func seedRehydrateIntent(t *testing.T, db *sql.DB, input rehydrateTestIntent) {
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

	if _, err := db.ExecContext(
		context.Background(),
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

func decodeLibraryMeta(t *testing.T, raw json.RawMessage) map[string]string {
	t.Helper()

	meta := map[string]string{}
	if err := json.Unmarshal(raw, &meta); err != nil {
		t.Fatalf("decode meta: %v", err)
	}
	return meta
}

func nullIfEmptyString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func rehydrateTestTimestamp(second int) string {
	return "2025-01-01T00:00:" + strconv.FormatInt(int64(second/10), 10) + strconv.Itoa(second%10) + "Z"
}

func createLegacyRehydrateDatabase(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create legacy db dir: %v", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	defer db.Close()

	statements := []string{
		`CREATE TABLE schema_version (version INTEGER NOT NULL, applied_at TIMESTAMP NOT NULL);`,
		`INSERT INTO schema_version (version, applied_at) VALUES (1, '2026-01-01T00:00:00Z');`,
		`CREATE TABLE schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT NOT NULL);`,
		`INSERT INTO schema_migrations (version, applied_at) VALUES ('0001_create_intent_table.sql', '2026-01-01T00:00:00Z');`,
		`INSERT INTO schema_migrations (version, applied_at) VALUES ('0002_create_projects_table.sql', '2026-01-01T00:00:00Z');`,
		`INSERT INTO schema_migrations (version, applied_at) VALUES ('0003_create_checkpoints_table.sql', '2026-01-01T00:00:00Z');`,
		`CREATE TABLE intents (id TEXT PRIMARY KEY, created_at TEXT NOT NULL, author TEXT NOT NULL, source_type TEXT NOT NULL, title TEXT, prompt TEXT NOT NULL, response TEXT NOT NULL, meta TEXT, prev_hash TEXT, hash TEXT NOT NULL);`,
		`CREATE TABLE projects (name TEXT PRIMARY KEY, description TEXT, created_at TEXT NOT NULL, prev_hash TEXT, hash TEXT NOT NULL);`,
		`CREATE TABLE checkpoints (hash TEXT PRIMARY KEY, project TEXT NOT NULL, summary TEXT NOT NULL, created_at TEXT NOT NULL, artifact_ids TEXT NOT NULL, previous_checkpoint_id TEXT);`,
		`INSERT INTO projects (name, description, created_at, prev_hash, hash) VALUES ('alpha', '', '2025-01-01T00:00:00Z', NULL, 'project-hash');`,
		`INSERT INTO checkpoints (hash, project, summary, created_at, artifact_ids, previous_checkpoint_id) VALUES ('checkpoint-hash', 'alpha', 'legacy checkpoint', '2025-01-01T00:00:10Z', '[]', NULL);`,
		`INSERT INTO intents (id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash) VALUES ('legacy-before', '2025-01-01T00:00:09Z', 'Ada', 'cli', 'Before', 'before prompt', 'before response', '{"project":"alpha"}', NULL, 'legacy-before-hash');`,
		`INSERT INTO intents (id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash) VALUES ('legacy-after', '2025-01-01T00:00:11Z', 'Ada', 'cli', 'After', 'after prompt', 'after response', '{"project":"alpha"}', NULL, 'legacy-after-hash');`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("exec legacy statement %q: %v", statement, err)
		}
	}
}
