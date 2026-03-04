package cmd

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/chuxorg/chux-yanzi-cli/internal/config"
	"github.com/chuxorg/chux-yanzi-cli/internal/core/hash"
	"github.com/chuxorg/chux-yanzi-cli/internal/core/model"
	yanzilibrary "github.com/chuxorg/chux-yanzi-cli/internal/library"
	_ "modernc.org/sqlite"
)

func TestExportMarkdownNoActiveProject(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)

	err := RunExport([]string{"--format", "markdown"}, "v1.2.3")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no active project") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExportMarkdownChronological(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	db := openConfiguredDBForExportTest(t)
	defer db.Close()
	seedProject(t, db, "alpha")

	seedIntentWithSource(t, db, "cap-1", "2025-01-01T00:00:01Z", "alpha", "engineer", "cli", "prompt 1", "response 1")
	seedCheckpointForExport(t, db, "alpha", "2025-01-01T00:00:02Z", "checkpoint 1")
	seedIntentWithSource(t, db, "evt-1", "2025-01-01T00:00:03Z", "alpha", "engineer", "meta-command", "@yanzi pause", "true")
	seedIntentWithSource(t, db, "cap-2", "2025-01-01T00:00:04Z", "alpha", "reviewer", "cli", "prompt 2", "response 2")

	if err := RunExport([]string{"--format", "markdown"}, "v9.9.9"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	path := filepath.Join(workdir, "YANZI_LOG.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	output := string(data)

	if !strings.Contains(output, "# Yanzi Agent Log") {
		t.Fatalf("missing header: %q", output)
	}
	if !strings.Contains(output, "Project: alpha") {
		t.Fatalf("missing project: %q", output)
	}
	if !strings.Contains(output, "Version: v9.9.9") {
		t.Fatalf("missing version: %q", output)
	}

	idxCap1 := strings.Index(output, "### Capture: cap-1")
	idxCheckpoint := strings.Index(output, "## Checkpoint:")
	idxEvent := strings.Index(output, "### Event: @yanzi pause")
	idxCap2 := strings.Index(output, "### Capture: cap-2")
	if idxCap1 == -1 || idxCheckpoint == -1 || idxEvent == -1 || idxCap2 == -1 {
		t.Fatalf("missing expected sections: %q", output)
	}
	if !(idxCap1 < idxCheckpoint && idxCheckpoint < idxEvent && idxEvent < idxCap2) {
		t.Fatalf("unexpected order: %q", output)
	}

	if !strings.Contains(output, "**Prompt**\n```text\nprompt 1\n```") {
		t.Fatalf("missing prompt block: %q", output)
	}
	if !strings.Contains(output, "**Response**\n```text\nresponse 1\n```") {
		t.Fatalf("missing response block: %q", output)
	}
}

func TestExportMarkdownRendersSortedMetadata(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	db := openConfiguredDBForExportTest(t)
	defer db.Close()
	seedProject(t, db, "alpha")
	seedIntentWithMeta(t, db, "cap-meta", "2025-01-01T00:00:01Z", "engineer", "cli", "prompt", "response", map[string]string{
		"project":       "alpha",
		"decision_type": "refactor",
		"area":          "auth",
		"tags":          "migration,security",
	})

	if err := RunExport([]string{"--format", "markdown"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.md"))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	output := string(data)
	metaBlock := "Metadata:\n  area: auth\n  decision_type: refactor\n  tags: migration,security\n"
	if !strings.Contains(output, metaBlock) {
		t.Fatalf("expected sorted metadata block, got: %q", output)
	}
}

func TestExportJSONNoActiveProject(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)

	err := RunExport([]string{"--format", "json"}, "v1.2.3")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no active project") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExportJSONCanonicalShapeAndChronology(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	db := openConfiguredDBForExportTest(t)
	defer db.Close()
	seedProject(t, db, "alpha")
	seedIntentWithMeta(t, db, "cap-1", "2025-01-01T00:00:01Z", "engineer", "cli", "prompt 1", "response 1", map[string]string{
		"project":       "alpha",
		"decision_type": "refactor",
		"area":          "auth",
	})
	seedCheckpointForExport(t, db, "alpha", "2025-01-01T00:00:02Z", "checkpoint 1")
	seedIntentWithSource(t, db, "evt-1", "2025-01-01T00:00:03Z", "alpha", "engineer", "meta-command", "@yanzi pause", "true")

	if err := RunExport([]string{"--format", "json"}, "v9.9.9"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	path := filepath.Join(workdir, "YANZI_LOG.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal export json: %v", err)
	}
	if got := payload["schema_version"]; got != float64(1) {
		t.Fatalf("expected schema_version=1, got %v", got)
	}
	if got := payload["project"]; got != "alpha" {
		t.Fatalf("expected project alpha, got %v", got)
	}
	if got := payload["version"]; got != "v9.9.9" {
		t.Fatalf("expected version v9.9.9, got %v", got)
	}
	exportedAt, ok := payload["exported_at"].(string)
	if !ok || exportedAt == "" {
		t.Fatalf("expected exported_at string, got %T %v", payload["exported_at"], payload["exported_at"])
	}
	if _, err := time.Parse(time.RFC3339, exportedAt); err != nil {
		t.Fatalf("expected RFC3339 exported_at, got %q (%v)", exportedAt, err)
	}

	events, ok := payload["events"].([]any)
	if !ok {
		t.Fatalf("expected events array, got %T", payload["events"])
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	first := events[0].(map[string]any)
	if first["type"] != "capture" {
		t.Fatalf("expected first event capture, got %v", first["type"])
	}
	if _, ok := first["metadata"]; !ok {
		t.Fatalf("expected capture metadata field")
	}
	if _, exists := first["summary"]; exists {
		t.Fatalf("did not expect checkpoint fields on capture")
	}

	second := events[1].(map[string]any)
	if second["type"] != "checkpoint" {
		t.Fatalf("expected second event checkpoint, got %v", second["type"])
	}

	third := events[2].(map[string]any)
	if third["type"] != "meta" {
		t.Fatalf("expected third event meta, got %v", third["type"])
	}
	if third["command"] != "@yanzi pause" {
		t.Fatalf("unexpected meta command: %v", third["command"])
	}
	if third["value"] != "true" {
		t.Fatalf("unexpected meta value: %v", third["value"])
	}

	// encoding/json emits map keys in sorted order; ensure deterministic metadata key ordering.
	jsonText := string(data)
	areaIdx := strings.Index(jsonText, "\"area\": \"auth\"")
	decisionIdx := strings.Index(jsonText, "\"decision_type\": \"refactor\"")
	if areaIdx == -1 || decisionIdx == -1 || areaIdx > decisionIdx {
		t.Fatalf("expected sorted metadata keys in output json, got: %s", jsonText)
	}
}

func TestExportJSONOmitMetadataWhenOnlyProjectAndAllowNullMetaValue(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	db := openConfiguredDBForExportTest(t)
	defer db.Close()
	seedProject(t, db, "alpha")
	seedIntentWithMeta(t, db, "cap-1", "2025-01-01T00:00:01Z", "engineer", "cli", "prompt 1", "response 1", map[string]string{
		"project": "alpha",
	})
	seedIntentWithSource(t, db, "evt-1", "2025-01-01T00:00:02Z", "alpha", "engineer", "meta-command", "@yanzi role Architect", "   ")

	if err := RunExport([]string{"--format", "json"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.json"))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal export json: %v", err)
	}
	events := payload["events"].([]any)
	capture := events[0].(map[string]any)
	if _, ok := capture["metadata"]; ok {
		t.Fatalf("did not expect metadata field when only system metadata exists")
	}
	metaEvent := events[1].(map[string]any)
	if _, ok := metaEvent["value"]; !ok {
		t.Fatalf("expected value field on meta event")
	}
	if metaEvent["value"] != nil {
		t.Fatalf("expected null meta value, got %v", metaEvent["value"])
	}
}

func TestExportJSONNoEventsProducesEmptyArray(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	db := openConfiguredDBForExportTest(t)
	defer db.Close()
	seedProject(t, db, "alpha")

	if err := RunExport([]string{"--format", "json"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.json"))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal export json: %v", err)
	}
	events, ok := payload["events"].([]any)
	if !ok {
		t.Fatalf("expected events array, got %T", payload["events"])
	}
	if len(events) != 0 {
		t.Fatalf("expected empty events array, got %d", len(events))
	}
}

func TestExportHTMLNoActiveProject(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)

	err := RunExport([]string{"--format", "html"}, "v1.2.3")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no active project") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExportHTMLCanonicalRenderAndCounts(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	db := openConfiguredDBForExportTest(t)
	defer db.Close()
	seedProject(t, db, "alpha")
	seedIntentWithMeta(t, db, "cap-1", "2025-01-01T00:00:01Z", "engineer", "cli", "line1\nline2", "result\nok", map[string]string{
		"project":       "alpha",
		"decision_type": "refactor",
		"area":          "auth",
	})
	seedCheckpointForExport(t, db, "alpha", "2025-01-01T00:00:02Z", "checkpoint 1")
	seedIntentWithSource(t, db, "evt-1", "2025-01-01T00:00:03Z", "alpha", "engineer", "meta-command", "@yanzi pause", "true")

	if err := RunExport([]string{"--format", "html"}, "v9.9.9"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	path := filepath.Join(workdir, "YANZI_LOG.html")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	output := string(data)

	if !strings.Contains(output, "<!doctype html>") {
		t.Fatalf("missing html doctype: %q", output)
	}
	if !strings.Contains(output, "Project:</span> alpha") {
		t.Fatalf("missing project header: %q", output)
	}
	if !strings.Contains(output, "Version:</span> v9.9.9") {
		t.Fatalf("missing version header: %q", output)
	}
	if !strings.Contains(output, "Total events: 3") || !strings.Contains(output, "Total captures: 1") || !strings.Contains(output, "Total checkpoints: 1") {
		t.Fatalf("missing counts: %q", output)
	}

	idxCapture := strings.Index(output, "Capture: cap-1")
	idxCheckpoint := strings.Index(output, "Checkpoint: ")
	idxMeta := strings.Index(output, "Event:</span> @yanzi pause")
	if idxCapture == -1 || idxCheckpoint == -1 || idxMeta == -1 {
		t.Fatalf("missing expected timeline sections: %q", output)
	}
	if !(idxCapture < idxCheckpoint && idxCheckpoint < idxMeta) {
		t.Fatalf("unexpected timeline order: %q", output)
	}

	if !strings.Contains(output, "<th>Metadata Key</th><th>Value</th>") {
		t.Fatalf("missing metadata table header: %q", output)
	}
	areaIdx := strings.Index(output, "<td>area</td><td>auth</td>")
	decisionIdx := strings.Index(output, "<td>decision_type</td><td>refactor</td>")
	if areaIdx == -1 || decisionIdx == -1 || areaIdx > decisionIdx {
		t.Fatalf("metadata order not deterministic: %q", output)
	}

	if !strings.Contains(output, "<pre>line1\nline2</pre>") {
		t.Fatalf("prompt pre block did not preserve whitespace: %q", output)
	}
	if !strings.Contains(output, "<pre>result\nok</pre>") {
		t.Fatalf("response pre block did not preserve whitespace: %q", output)
	}
}

func TestExportHTMLOmitsMetadataWhenOnlyProject(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	db := openConfiguredDBForExportTest(t)
	defer db.Close()
	seedProject(t, db, "alpha")
	seedIntentWithMeta(t, db, "cap-1", "2025-01-01T00:00:01Z", "engineer", "cli", "prompt", "response", map[string]string{
		"project": "alpha",
	})

	if err := RunExport([]string{"--format", "html"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.html"))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	output := string(data)
	if strings.Contains(output, "<th>Metadata Key</th><th>Value</th>") {
		t.Fatalf("did not expect metadata table for project-only metadata: %q", output)
	}
}

func TestExportMarkdownNoCapturesRecorded(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	db := openConfiguredDBForExportTest(t)
	defer db.Close()
	seedProject(t, db, "alpha")

	if err := RunExport([]string{"--format", "markdown"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.md"))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	output := string(data)
	if !strings.Contains(output, "No captures recorded.") {
		t.Fatalf("expected no captures message: %q", output)
	}
}

func TestExportMarkdownOnlyCheckpoints(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	db := openConfiguredDBForExportTest(t)
	defer db.Close()
	seedProject(t, db, "alpha")
	seedCheckpointForExport(t, db, "alpha", "2025-01-01T00:00:02Z", "checkpoint only")

	if err := RunExport([]string{"--format", "markdown"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.md"))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	output := string(data)
	if !strings.Contains(output, "## Checkpoint:") {
		t.Fatalf("expected checkpoint output: %q", output)
	}
	if strings.Contains(output, "No captures recorded.") {
		t.Fatalf("did not expect no captures message: %q", output)
	}
}

func openConfiguredDBForExportTest(t *testing.T) *sql.DB {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	db, err := sql.Open("sqlite", cfg.DBPath)
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

func seedIntentWithSource(t *testing.T, db *sql.DB, id, createdAt, project, author, sourceType, prompt, response string) {
	t.Helper()
	seedIntentWithMeta(t, db, id, createdAt, author, sourceType, prompt, response, map[string]string{"project": project})
}

func seedIntentWithMeta(t *testing.T, db *sql.DB, id, createdAt, author, sourceType, prompt, response string, metaPayload map[string]string) {
	t.Helper()
	keys := make([]string, 0, len(metaPayload))
	for key := range metaPayload {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	ordered := make(map[string]string, len(metaPayload))
	for _, key := range keys {
		ordered[key] = metaPayload[key]
	}
	meta, err := json.Marshal(ordered)
	if err != nil {
		t.Fatalf("encode meta: %v", err)
	}
	record := model.IntentRecord{
		ID:         id,
		CreatedAt:  createdAt,
		Author:     author,
		SourceType: sourceType,
		Prompt:     prompt,
		Response:   response,
		Meta:       meta,
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

func seedCheckpointForExport(t *testing.T, db *sql.DB, project, createdAt, summary string) {
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
