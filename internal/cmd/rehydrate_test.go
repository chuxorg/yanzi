package cmd

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	hash TEXT NOT NULL,
	class TEXT,
	type TEXT,
	content TEXT,
	metadata TEXT
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

func TestRehydrateWithRichArtifacts(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeStateFile(t, workdir, "alpha")
	db := openTestDB(t, workdir)
	defer db.Close()

	seedProject(t, db, "alpha")
	seedCheckpoint(t, db, "alpha", "2025-01-01T00:00:00Z", "first checkpoint")
	seedIntentRecord(t, db, rehydrateSeedIntent{
		ID:        "01",
		CreatedAt: "2025-01-01T00:00:01Z",
		Project:   "alpha",
		Author:    "Ada",
		Title:     "JWT validation edge cases",
		Prompt:    strings.Repeat("Need to determine why refresh tokens fail after reconnect. ", 5),
		Response:  strings.Repeat("Issue appears related to clock skew and token leeway handling. ", 5),
		Meta: map[string]string{
			"project": "alpha",
			"area":    "auth",
			"phase":   "validation",
		},
	})
	seedIntentRecord(t, db, rehydrateSeedIntent{
		ID:        "other-project",
		CreatedAt: "2025-01-01T00:00:02Z",
		Project:   "beta",
		Author:    "Byron",
		Prompt:    "should not appear",
		Response:  "should not appear",
		Meta: map[string]string{
			"project": "beta",
		},
	})

	output, err := captureStdout(func() error {
		return RunRehydrate([]string{})
	})
	if err != nil {
		t.Fatalf("RunRehydrate: %v", err)
	}

	if !strings.Contains(output, "Project: alpha") {
		t.Fatalf("missing project: %q", output)
	}
	if !strings.Contains(output, "Continuity Summary") || !strings.Contains(output, "Mode: checkpoint") || !strings.Contains(output, "Open work: 0") {
		t.Fatalf("missing continuity summary: %q", output)
	}
	if !strings.Contains(output, "Checkpoint") || !strings.Contains(output, "Summary: first checkpoint") {
		t.Fatalf("missing checkpoint block: %q", output)
	}
	if !strings.Contains(output, "Post-Checkpoint Continuity") {
		t.Fatalf("missing continuity header: %q", output)
	}
	if !strings.Contains(output, "[1] 2025-01-01T00:00:01Z") {
		t.Fatalf("missing artifact index/timestamp: %q", output)
	}
	if !strings.Contains(output, "Author: Ada") {
		t.Fatalf("missing author: %q", output)
	}
	if !strings.Contains(output, "Title: JWT validation edge cases") {
		t.Fatalf("missing title: %q", output)
	}
	if !strings.Contains(output, "Prompt:\nNeed to determine why refresh tokens fail") {
		t.Fatalf("missing prompt snippet: %q", output)
	}
	if !strings.Contains(output, "...") {
		t.Fatalf("expected truncation ellipsis: %q", output)
	}
	if !strings.Contains(output, "Response:\nIssue appears related to clock skew") {
		t.Fatalf("missing response snippet: %q", output)
	}
	if !strings.Contains(output, "Meta:\narea=auth phase=validation project=alpha") {
		t.Fatalf("missing metadata summary: %q", output)
	}
	if strings.Contains(output, "other-project") || strings.Contains(output, "should not appear") {
		t.Fatalf("unexpected cross-project continuity in output: %q", output)
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

func TestRehydrateDryRun(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeStateFile(t, workdir, "alpha")
	db := openTestDB(t, workdir)
	defer db.Close()

	seedProject(t, db, "alpha")
	seedCheckpointWithArtifacts(t, db, "alpha", "2025-01-01T00:00:00Z", "first checkpoint", []string{"ctx-1", "ctx-2"})
	seedIntentRecord(t, db, rehydrateSeedIntent{
		ID:        "dry-run-1",
		CreatedAt: "2025-01-01T00:00:01Z",
		Project:   "alpha",
		Prompt:    "prompt",
		Response:  "response",
		Meta:      map[string]string{"project": "alpha"},
	})

	output, err := captureStdout(func() error {
		return RunRehydrate([]string{"--dry-run"})
	})
	if err != nil {
		t.Fatalf("RunRehydrate dry-run: %v", err)
	}
	if !strings.Contains(output, "Checkpoints to load: 1") ||
		!strings.Contains(output, "Continuity mode: checkpoint") ||
		!strings.Contains(output, "Continuity depth: 1") ||
		!strings.Contains(output, "Context count: 2") ||
		!strings.Contains(output, "Last checkpoint summary: first checkpoint") ||
		!strings.Contains(output, "Intents to load: 1") {
		t.Fatalf("unexpected dry-run output: %q", output)
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

func TestRehydrateNoCheckpointFallsBack(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeStateFile(t, workdir, "alpha")
	db := openTestDB(t, workdir)
	defer db.Close()

	seedProject(t, db, "alpha")
	for i := 1; i <= 12; i++ {
		seedIntentRecord(t, db, rehydrateSeedIntent{
			ID:        fmtID("cap", i),
			CreatedAt: fmtTimestamp(i),
			Project:   "alpha",
			Author:    "Ada",
			Prompt:    "prompt " + fmtID("", i),
			Response:  "response " + fmtID("", i),
			Meta:      map[string]string{"project": "alpha"},
		})
	}

	output, err := captureStdout(func() error {
		return RunRehydrate([]string{})
	})
	if err != nil {
		t.Fatalf("RunRehydrate fallback: %v", err)
	}
	if !strings.Contains(output, "Warning: No checkpoint found for active project.") {
		t.Fatalf("missing fallback warning: %q", output)
	}
	if !strings.Contains(output, "Showing last 10 captures instead.") {
		t.Fatalf("missing fallback limit text: %q", output)
	}
	if strings.Contains(output, "[11]") || strings.Contains(output, "2025-01-01T00:00:01Z") || strings.Contains(output, "2025-01-01T00:00:02Z") {
		t.Fatalf("expected fallback window to cap output: %q", output)
	}
	if !strings.Contains(output, "[1] 2025-01-01T00:00:03Z") || !strings.Contains(output, "[10] 2025-01-01T00:00:12Z") {
		t.Fatalf("expected chronological fallback window: %q", output)
	}
}

func TestRehydrateJSONOutput(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeStateFile(t, workdir, "alpha")
	db := openTestDB(t, workdir)
	defer db.Close()

	seedProject(t, db, "alpha")
	seedCheckpointWithArtifacts(t, db, "alpha", "2025-01-01T00:00:00Z", "first checkpoint", []string{"ctx-1"})
	seedIntentRecord(t, db, rehydrateSeedIntent{
		ID:        "json-1",
		CreatedAt: "2025-01-01T00:00:01Z",
		Project:   "alpha",
		Author:    "Ada",
		Title:     "JSON example",
		Prompt:    "Prompt body",
		Response:  "Response body",
		Meta: map[string]string{
			"project": "alpha",
			"area":    "auth",
		},
	})

	output, err := captureStdout(func() error {
		return RunRehydrate([]string{"--format", "json"})
	})
	if err != nil {
		t.Fatalf("RunRehydrate json: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("decode json output: %v\noutput=%s", err, output)
	}

	if payload["project"] != "alpha" {
		t.Fatalf("unexpected project: %#v", payload["project"])
	}
	if payload["schema_version"] != float64(machineContractSchemaVersion) || payload["kind"] != jsonKindRehydrate {
		t.Fatalf("unexpected contract identity: %#v", payload)
	}
	if strings.Index(output, "\"schema_version\"") > strings.Index(output, "\"kind\"") || strings.Index(output, "\"kind\"") > strings.Index(output, "\"project\"") {
		t.Fatalf("unexpected field ordering: %s", output)
	}
	if payload["has_checkpoint"] != true || payload["fallback"] != false {
		t.Fatalf("unexpected checkpoint/fallback flags: %#v", payload)
	}
	checkpoint, ok := payload["checkpoint"].(map[string]any)
	if !ok || checkpoint["summary"] != "first checkpoint" {
		t.Fatalf("unexpected checkpoint payload: %#v", payload["checkpoint"])
	}
	if checkpoint["project"] != "alpha" {
		t.Fatalf("unexpected checkpoint project: %#v", checkpoint)
	}
	intents, ok := payload["intents"].([]any)
	if !ok || len(intents) != 1 {
		t.Fatalf("unexpected intents payload: %#v", payload["intents"])
	}
	intent, ok := intents[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected intent entry: %#v", intents[0])
	}
	if intent["id"] != "json-1" || intent["title"] != "JSON example" {
		t.Fatalf("unexpected intent identity: %#v", intent)
	}
	if intent["prompt_snippet"] != "Prompt body" || intent["response_snippet"] != "Response body" {
		t.Fatalf("unexpected snippets: %#v", intent)
	}
	metadata, ok := intent["metadata"].(map[string]any)
	if !ok || metadata["area"] != "auth" || metadata["project"] != "alpha" {
		t.Fatalf("unexpected metadata: %#v", intent["metadata"])
	}
}

func TestRehydrateJSONFallbackOutput(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeStateFile(t, workdir, "alpha")
	db := openTestDB(t, workdir)
	defer db.Close()

	seedProject(t, db, "alpha")
	seedIntentRecord(t, db, rehydrateSeedIntent{
		ID:        "fallback-json-1",
		CreatedAt: "2025-01-01T00:00:01Z",
		Project:   "alpha",
		Prompt:    "Prompt body",
		Response:  "Response body",
		Meta:      map[string]string{"project": "alpha"},
	})

	output, err := captureStdout(func() error {
		return RunRehydrate([]string{"--format", "json"})
	})
	if err != nil {
		t.Fatalf("RunRehydrate json fallback: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("decode json output: %v\noutput=%s", err, output)
	}
	if payload["has_checkpoint"] != false || payload["fallback"] != true {
		t.Fatalf("unexpected fallback flags: %#v", payload)
	}
	if payload["schema_version"] != float64(machineContractSchemaVersion) || payload["kind"] != jsonKindRehydrate {
		t.Fatalf("unexpected contract identity: %#v", payload)
	}
	if strings.Index(output, "\"schema_version\"") > strings.Index(output, "\"kind\"") || strings.Index(output, "\"kind\"") > strings.Index(output, "\"project\"") {
		t.Fatalf("unexpected field ordering: %s", output)
	}
	if payload["fallback_reason"] != yanzilibrary.ErrCheckpointNotFound.Error() {
		t.Fatalf("unexpected fallback reason: %#v", payload["fallback_reason"])
	}
	if payload["checkpoint"] != nil {
		t.Fatalf("did not expect checkpoint object: %#v", payload["checkpoint"])
	}
}

func openTestDB(t *testing.T, dir string) *sql.DB {
	t.Helper()

	path := filepath.Join(dir, "yanzi-library.db")
	t.Setenv("YANZI_DB_PATH", path)
	db, err := yanzilibrary.InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	return db
}

func seedCheckpoint(t *testing.T, db *sql.DB, project, createdAt, summary string) {
	t.Helper()
	seedCheckpointWithArtifacts(t, db, project, createdAt, summary, []string{})
}

func seedCheckpointWithArtifacts(t *testing.T, db *sql.DB, project, createdAt, summary string, artifactIDs []string) {
	t.Helper()
	checkpoint := yanzilibrary.Checkpoint{
		Project:     project,
		Summary:     summary,
		CreatedAt:   createdAt,
		ArtifactIDs: artifactIDs,
	}
	hashValue, err := yanzilibrary.HashCheckpoint(checkpoint)
	if err != nil {
		t.Fatalf("hash checkpoint: %v", err)
	}
	artifactIDsJSON, err := json.Marshal(artifactIDs)
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
		string(artifactIDsJSON),
		nil,
	)
	if err != nil {
		t.Fatalf("seed checkpoint: %v", err)
	}
}

type rehydrateSeedIntent struct {
	ID        string
	CreatedAt string
	Project   string
	Author    string
	Title     string
	Prompt    string
	Response  string
	Meta      map[string]string
}

func seedIntentRecord(t *testing.T, db *sql.DB, input rehydrateSeedIntent) {
	t.Helper()

	meta := input.Meta
	if meta == nil {
		meta = map[string]string{"project": input.Project}
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("encode meta: %v", err)
	}
	author := input.Author
	if author == "" {
		author = "tester"
	}
	record := model.IntentRecord{
		ID:         input.ID,
		CreatedAt:  input.CreatedAt,
		Author:     author,
		SourceType: "cli",
		Title:      input.Title,
		Prompt:     input.Prompt,
		Response:   input.Response,
		Meta:       metaJSON,
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
		nullIfEmpty(record.Title),
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

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func fmtID(prefix string, n int) string {
	if prefix == "" {
		return strconv.Itoa(n)
	}
	return prefix + "-" + strconv.Itoa(n)
}

func fmtTimestamp(second int) string {
	return "2025-01-01T00:00:" + fmt.Sprintf("%02d", second) + "Z"
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
