package yanzilibrary

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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

func TestRehydrateProjectRejectsMissingProject(t *testing.T) {
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

	if _, err := RehydrateProject("missing"); err == nil {
		t.Fatal("expected error")
	} else {
		var notFound ProjectNotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestRehydrateProjectOrdersEqualTimestampsByID(t *testing.T) {
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

	seedRehydrateCheckpoint(t, db, "alpha", "2025-01-01T00:00:10Z", "checkpoint summary")
	seedRehydrateIntent(t, db, rehydrateTestIntent{
		ID:        "intent-b",
		CreatedAt: "2025-01-01T00:00:11Z",
		Project:   "alpha",
		Author:    "Ada",
		Prompt:    "prompt b",
		Response:  "response b",
	})
	seedRehydrateIntent(t, db, rehydrateTestIntent{
		ID:        "intent-a",
		CreatedAt: "2025-01-01T00:00:11Z",
		Project:   "alpha",
		Author:    "Ada",
		Prompt:    "prompt a",
		Response:  "response a",
	})

	payload, err := RehydrateProject("alpha")
	if err != nil {
		t.Fatalf("RehydrateProject: %v", err)
	}
	if len(payload.Intents) != 2 {
		t.Fatalf("expected 2 intents, got %d", len(payload.Intents))
	}
	if payload.Intents[0].ID != "intent-a" || payload.Intents[1].ID != "intent-b" {
		t.Fatalf("unexpected intent order: %+v", payload.Intents)
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
