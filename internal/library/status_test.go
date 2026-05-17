package yanzilibrary

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/chuxorg/yanzi/internal/sqliteruntime"
)

func TestLoadProjectStatusSummarizesContinuity(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	db := openTestLibraryDB(t, workdir)
	defer db.Close()

	if _, err := CreateProject("alpha", ""); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if _, err := CreateArtifact("alpha", ArtifactClassIntent, "task", "Open task", "Continue implementation.", `{"status":"in_progress"}`); err != nil {
		t.Fatalf("CreateArtifact task: %v", err)
	}
	if _, err := CreateArtifact("alpha", ArtifactClassIntent, "change_request", "Closed change", "Already done.", `{"status":"resolved"}`); err != nil {
		t.Fatalf("CreateArtifact change_request: %v", err)
	}
	if _, err := CreateContextArtifact("", "process_rule", ContextScopeGlobal, "Rule", "Never rewrite history.", ""); err != nil {
		t.Fatalf("CreateContextArtifact global: %v", err)
	}
	if _, err := CreateContextArtifact("alpha", "reference", ContextScopeProject, "API", "https://example.test", ""); err != nil {
		t.Fatalf("CreateContextArtifact project: %v", err)
	}

	seedStatusIntent(t, db, statusSeedIntent{ID: "cap-1", CreatedAt: "2025-01-01T00:00:01Z", Project: "alpha", Prompt: "Investigate auth drift", Response: "Started review", Author: "Ada"})
	seedStatusIntent(t, db, statusSeedIntent{ID: "evt-1", CreatedAt: "2025-01-01T00:00:02Z", Project: "alpha", Prompt: "@yanzi pause", Response: "true", SourceType: "meta-command", Author: "Ada"})
	seedStatusCheckpoint(t, db, "alpha", "2025-01-01T00:00:03Z", "checkpoint 1")
	seedStatusIntent(t, db, statusSeedIntent{ID: "cap-2", CreatedAt: "2025-01-01T00:00:04Z", Project: "alpha", Prompt: "Resume auth work", Response: "Narrowed to retry path", Title: "Auth retry", Author: "Ada"})

	status, err := LoadProjectStatus("alpha", 4)
	if err != nil {
		t.Fatalf("LoadProjectStatus: %v", err)
	}

	if status.Project != "alpha" {
		t.Fatalf("unexpected project: %#v", status.Project)
	}
	if status.ContinuityMode != "checkpoint" {
		t.Fatalf("unexpected continuity mode: %#v", status.ContinuityMode)
	}
	if status.ContinuityDepth != 1 {
		t.Fatalf("unexpected continuity depth: %d", status.ContinuityDepth)
	}
	if status.TotalCaptures != 2 || status.TotalProtocolAnnotations != 1 || status.TotalCheckpoints != 1 {
		t.Fatalf("unexpected counts: %#v", status)
	}
	if status.TotalIntentArtifacts != 2 || status.VisibleContextArtifacts != 2 {
		t.Fatalf("unexpected artifact visibility counts: %#v", status)
	}
	if status.LastActivityAt != "2025-01-01T00:00:04Z" || status.LastCaptureAt != "2025-01-01T00:00:04Z" {
		t.Fatalf("unexpected last activity timestamps: %#v", status)
	}
	if status.LatestCheckpoint == nil || status.LatestCheckpoint.Summary != "checkpoint 1" {
		t.Fatalf("unexpected latest checkpoint: %#v", status.LatestCheckpoint)
	}
	if len(status.UnresolvedWork) != 1 || status.UnresolvedWork[0].Title != "Open task" {
		t.Fatalf("unexpected unresolved work: %#v", status.UnresolvedWork)
	}
	if len(status.RecentActivity) != 4 {
		t.Fatalf("unexpected recent activity count: %#v", status.RecentActivity)
	}
	if status.RecentActivity[0].Kind != "capture" || status.RecentActivity[1].Kind != "checkpoint" || status.RecentActivity[2].Kind != "protocol_annotation" || status.RecentActivity[3].Kind != "capture" {
		t.Fatalf("unexpected recent activity ordering: %#v", status.RecentActivity)
	}
}

func TestLoadProjectStatusFallsBackWithoutCheckpoint(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	db := openTestLibraryDB(t, workdir)
	defer db.Close()

	if _, err := CreateProject("alpha", ""); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	seedStatusIntent(t, db, statusSeedIntent{ID: "cap-1", CreatedAt: "2025-01-01T00:00:01Z", Project: "alpha", Prompt: "Need continuity", Response: "Use fallback", Author: "Ada"})

	status, err := LoadProjectStatus("alpha", 3)
	if err != nil {
		t.Fatalf("LoadProjectStatus: %v", err)
	}

	if status.ContinuityMode != "fallback" {
		t.Fatalf("unexpected continuity mode: %#v", status.ContinuityMode)
	}
	if status.ContinuityDepth != 1 {
		t.Fatalf("unexpected fallback depth: %d", status.ContinuityDepth)
	}
	if status.LatestCheckpoint != nil {
		t.Fatalf("did not expect checkpoint: %#v", status.LatestCheckpoint)
	}
}

func openTestLibraryDB(t *testing.T, dir string) *sql.DB {
	t.Helper()

	path := dir + "/yanzi-library.db"
	t.Setenv("YANZI_DB_PATH", path)
	db, err := InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	return db
}

type statusSeedIntent struct {
	ID         string
	CreatedAt  string
	Project    string
	Author     string
	Title      string
	Prompt     string
	Response   string
	SourceType string
	Metadata   map[string]string
}

func seedStatusIntent(t *testing.T, db *sql.DB, input statusSeedIntent) {
	t.Helper()

	systemMeta := map[string]string{"project": input.Project}
	systemMetaJSON, err := json.Marshal(systemMeta)
	if err != nil {
		t.Fatalf("encode system meta: %v", err)
	}
	var metadataText any
	if len(input.Metadata) > 0 {
		metadataJSON, err := json.Marshal(input.Metadata)
		if err != nil {
			t.Fatalf("encode metadata: %v", err)
		}
		metadataText = string(metadataJSON)
	}
	sourceType := input.SourceType
	if sourceType == "" {
		sourceType = "cli"
	}
	if _, err := sqliteruntime.ExecContext(
		context.Background(),
		db,
		ResolvedDBPath(),
		"seed status intent",
		`INSERT INTO intents (id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.ID,
		input.CreatedAt,
		input.Author,
		sourceType,
		nullIfBlank(input.Title),
		input.Prompt,
		input.Response,
		string(systemMetaJSON),
		nil,
		input.ID+"-hash",
		metadataText,
	); err != nil {
		t.Fatalf("seed status intent: %v", err)
	}
}

func seedStatusCheckpoint(t *testing.T, db *sql.DB, project, createdAt, summary string) {
	t.Helper()
	if _, err := sqliteruntime.ExecContext(
		context.Background(),
		db,
		ResolvedDBPath(),
		"seed checkpoint",
		`INSERT INTO checkpoints (hash, project, summary, created_at, artifact_ids, previous_checkpoint_id)
		VALUES (?, ?, ?, ?, ?, ?)`,
		project+"-"+summary,
		project,
		summary,
		createdAt,
		"[]",
		nil,
	); err != nil {
		t.Fatalf("seed checkpoint: %v", err)
	}
}

func nullIfBlank(value string) any {
	if value == "" {
		return nil
	}
	return value
}
