package cmd

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/core/hash"
	"github.com/chuxorg/yanzi/internal/core/model"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
	_ "modernc.org/sqlite"
)

const (
	intentTableSQL = `CREATE TABLE IF NOT EXISTS intents (
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
);`
	projectTableSQL = `CREATE TABLE IF NOT EXISTS projects (
	name TEXT PRIMARY KEY,
	description TEXT,
	created_at TEXT NOT NULL,
	prev_hash TEXT,
	hash TEXT NOT NULL
);`
	checkpointTableSQL = `CREATE TABLE IF NOT EXISTS checkpoints (
	hash TEXT PRIMARY KEY,
	project TEXT NOT NULL,
	summary TEXT NOT NULL,
	created_at TEXT NOT NULL,
	artifact_ids TEXT NOT NULL,
	previous_checkpoint_id TEXT
);`
)

func TestRehydrateWithArtifacts(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeStateFile(t, workdir, "alpha")
	db := openTestDB(t, workdir)
	defer db.Close()

	seedProject(t, db, "alpha")
	seedCheckpoint(t, db, "alpha", "2025-01-01T00:00:00Z", "first checkpoint")
	seedIntent(t, db, "01", "2025-01-01T00:00:01Z", "alpha")
	seedIntent(t, db, "02", "2025-01-01T00:00:02Z", "alpha")

	output, err := captureStdout(func() error {
		return RunRehydrate([]string{})
	})
	if err != nil {
		t.Fatalf("RunRehydrate: %v", err)
	}

	if !strings.Contains(output, "Project: alpha") {
		t.Fatalf("missing project: %q", output)
	}
	if !strings.Contains(output, "Latest Checkpoint:") {
		t.Fatalf("missing checkpoint header: %q", output)
	}
	if !strings.Contains(output, "* CreatedAt: 2025-01-01T00:00:00Z") {
		t.Fatalf("missing created_at: %q", output)
	}
	if !strings.Contains(output, "* Summary: first checkpoint") {
		t.Fatalf("missing summary: %q", output)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var artifactLines []string
	for _, line := range lines {
		if strings.HasPrefix(line, "1. ") || strings.HasPrefix(line, "2. ") {
			artifactLines = append(artifactLines, line)
		}
	}
	if len(artifactLines) != 2 {
		t.Fatalf("expected 2 artifacts, got %v", artifactLines)
	}
	if !strings.Contains(artifactLines[0], "01 2025-01-01T00:00:01Z intent") {
		t.Fatalf("unexpected first artifact: %q", artifactLines[0])
	}
	if !strings.Contains(artifactLines[1], "02 2025-01-01T00:00:02Z intent") {
		t.Fatalf("unexpected second artifact: %q", artifactLines[1])
	}
}

func TestRehydrateNoArtifacts(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeStateFile(t, workdir, "alpha")
	db := openTestDB(t, workdir)
	defer db.Close()

	seedProject(t, db, "alpha")
	seedCheckpoint(t, db, "alpha", "2025-01-01T00:00:00Z", "first checkpoint")

	output, err := captureStdout(func() error {
		return RunRehydrate([]string{})
	})
	if err != nil {
		t.Fatalf("RunRehydrate: %v", err)
	}
	if !strings.Contains(output, "(none)") {
		t.Fatalf("expected none output, got %q", output)
	}
}

func TestRehydrateNoActiveProject(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)

	err := RunRehydrate([]string{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no active project") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRehydrateNoCheckpoint(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeStateFile(t, workdir, "alpha")
	db := openTestDB(t, workdir)
	defer db.Close()

	seedProject(t, db, "alpha")

	err := RunRehydrate([]string{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no checkpoint") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func openTestDB(t *testing.T, dir string) *sql.DB {
	t.Helper()

	path := filepath.Join(dir, "yanzi-library.db")
	t.Setenv("YANZI_DB_PATH", path)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if _, err := db.Exec(intentTableSQL); err != nil {
		t.Fatalf("create intents: %v", err)
	}
	if _, err := db.Exec(projectTableSQL); err != nil {
		t.Fatalf("create projects: %v", err)
	}
	if _, err := db.Exec(checkpointTableSQL); err != nil {
		t.Fatalf("create checkpoints: %v", err)
	}
	return db
}

func seedCheckpoint(t *testing.T, db *sql.DB, project, createdAt, summary string) {
	t.Helper()
	checkpoint := yanzilibrary.Checkpoint{
		Project:     project,
		Summary:     summary,
		CreatedAt:   createdAt,
		ArtifactIDs: []string{},
	}
	hashValue, err := yanzilibrary.HashCheckpoint(checkpoint)
	if err != nil {
		t.Fatalf("hash checkpoint: %v", err)
	}
	artifactIDs, err := json.Marshal([]string{})
	if err != nil {
		t.Fatalf("encode artifacts: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO checkpoints (hash, project, summary, created_at, artifact_ids, previous_checkpoint_id)
		VALUES (?, ?, ?, ?, ?, ?)`,
		hashValue,
		project,
		summary,
		createdAt,
		string(artifactIDs),
		nil,
	)
	if err != nil {
		t.Fatalf("seed checkpoint: %v", err)
	}
}

func seedIntent(t *testing.T, db *sql.DB, id, createdAt, project string) {
	t.Helper()
	meta, err := json.Marshal(map[string]string{"project": project})
	if err != nil {
		t.Fatalf("encode meta: %v", err)
	}
	record := model.IntentRecord{
		ID:         id,
		CreatedAt:  createdAt,
		Author:     "tester",
		SourceType: "cli",
		Title:      "",
		Prompt:     "prompt",
		Response:   "response",
		Meta:       meta,
		PrevHash:   "",
	}
	sum, err := hash.HashIntent(record)
	if err != nil {
		t.Fatalf("hash intent: %v", err)
	}
	record.Hash = sum
	_, err = db.Exec(
		`INSERT INTO intents (id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ID,
		record.CreatedAt,
		record.Author,
		record.SourceType,
		nil,
		record.Prompt,
		record.Response,
		string(record.Meta),
		nil,
		record.Hash,
	)
	if err != nil {
		t.Fatalf("seed intent: %v", err)
	}
}

func withCwd(t *testing.T, dir string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil && !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("restore wd: %v", err)
		}
	})
}
