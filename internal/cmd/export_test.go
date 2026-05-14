package cmd

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/core/hash"
	"github.com/chuxorg/yanzi/internal/core/model"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
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
	if !strings.Contains(output, "## Continuity Summary") || !strings.Contains(output, "- Continuity mode: checkpoint") || !strings.Contains(output, "- Protocol annotations: 1") {
		t.Fatalf("missing continuity summary: %q", output)
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

	intentDirEntries, err := os.ReadDir(filepath.Join(workdir, "Intent"))
	if err != nil {
		t.Fatalf("read intent export dir: %v", err)
	}
	if len(intentDirEntries) != 0 {
		t.Fatalf("expected empty intent export dir, got %d files", len(intentDirEntries))
	}

	contextDirEntries, err := os.ReadDir(filepath.Join(workdir, "Context"))
	if err != nil {
		t.Fatalf("read context export dir: %v", err)
	}
	if len(contextDirEntries) != 0 {
		t.Fatalf("expected empty context export dir, got %d files", len(contextDirEntries))
	}
}

func TestExportMarkdownWritesArtifactDirectories(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	createTestProject(t, "alpha")
	if _, err := yanzilibrary.CreateArtifact("alpha", yanzilibrary.ArtifactClassIntent, "decision", "Export scope", "Export intent artifacts.", ""); err != nil {
		t.Fatalf("CreateArtifact intent: %v", err)
	}
	if _, err := yanzilibrary.CreateContextArtifact("alpha", "process_rule", yanzilibrary.ContextScopeProject, "Release policy", "Never rewrite history.", ""); err != nil {
		t.Fatalf("CreateArtifact context: %v", err)
	}

	if err := RunExport([]string{"--format", "markdown"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	intentEntries, err := os.ReadDir(filepath.Join(workdir, "Intent"))
	if err != nil {
		t.Fatalf("read intent dir: %v", err)
	}
	if len(intentEntries) != 1 {
		t.Fatalf("expected 1 intent artifact file, got %d", len(intentEntries))
	}
	if !strings.HasSuffix(intentEntries[0].Name(), "-export-scope.md") {
		t.Fatalf("unexpected intent filename: %s", intentEntries[0].Name())
	}

	contextEntries, err := os.ReadDir(filepath.Join(workdir, "Context"))
	if err != nil {
		t.Fatalf("read context dir: %v", err)
	}
	if len(contextEntries) != 1 {
		t.Fatalf("expected 1 context artifact file, got %d", len(contextEntries))
	}
	if !strings.HasSuffix(contextEntries[0].Name(), "-release-policy.md") {
		t.Fatalf("unexpected context filename: %s", contextEntries[0].Name())
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

func TestExportMarkdownIncludesProfileMetadata(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	db := openConfiguredDBForExportTest(t)
	defer db.Close()
	seedProject(t, db, "alpha")
	seedIntentWithMeta(t, db, "cap-profile", "2025-01-01T00:00:01Z", "engineer", "cli", "prompt", "response", map[string]string{
		"project": "alpha",
		"profile": "engineer",
	})

	if err := RunExport([]string{"--format", "markdown", "--profile", "engineer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.md"))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	if !strings.Contains(string(data), "Metadata:\n  profile: engineer\n") {
		t.Fatalf("expected profile metadata block, got: %q", string(data))
	}
}

func TestExportMarkdownMetaFiltersRuleArtifacts(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")
	createTestProject(t, "alpha")

	captureRuleArtifactsForExportTest(t, workdir)

	if err := RunExport([]string{"--meta", "type=context", "--meta", "subtype=rules", "--format", "markdown"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.md"))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	output := string(data)
	if !strings.Contains(output, "Canonical system rules") || !strings.Contains(output, "# System Rules") {
		t.Fatalf("expected rule artifact content in markdown export: %q", output)
	}
	if strings.Contains(output, "General context note") || strings.Contains(output, "# Project Notes") {
		t.Fatalf("did not expect non-rule artifact in markdown export: %q", output)
	}
	if strings.Contains(output, "## Checkpoint:") || strings.Contains(output, "### Event:") {
		t.Fatalf("did not expect checkpoints or meta events in filtered export: %q", output)
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
	if got := payload["kind"]; got != jsonKindHistoryExport {
		t.Fatalf("expected history export kind, got %v", got)
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
	if _, err := time.Parse(time.RFC3339Nano, exportedAt); err != nil {
		t.Fatalf("expected RFC3339 exported_at, got %q (%v)", exportedAt, err)
	}
	summary, ok := payload["summary"].(map[string]any)
	if !ok {
		t.Fatalf("expected summary object, got %T", payload["summary"])
	}
	if summary["continuity_mode"] != "checkpoint" || summary["total_protocol_annotations"] != float64(1) {
		t.Fatalf("unexpected summary payload: %#v", summary)
	}
	latestCheckpoint, ok := summary["latest_checkpoint"].(map[string]any)
	if !ok || latestCheckpoint["project"] != "alpha" {
		t.Fatalf("unexpected latest checkpoint summary payload: %#v", summary["latest_checkpoint"])
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
	if strings.Index(jsonText, "\"schema_version\"") > strings.Index(jsonText, "\"kind\"") || strings.Index(jsonText, "\"kind\"") > strings.Index(jsonText, "\"project\"") {
		t.Fatalf("unexpected top-level field ordering: %s", jsonText)
	}
	areaIdx := strings.Index(jsonText, "\"area\": \"auth\"")
	decisionIdx := strings.Index(jsonText, "\"decision_type\": \"refactor\"")
	if areaIdx == -1 || decisionIdx == -1 || areaIdx > decisionIdx {
		t.Fatalf("expected sorted metadata keys in output json, got: %s", jsonText)
	}
}

func TestExportJSONIncludesProfileMetadata(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	db := openConfiguredDBForExportTest(t)
	defer db.Close()
	seedProject(t, db, "alpha")
	seedIntentWithMeta(t, db, "cap-profile", "2025-01-01T00:00:01Z", "engineer", "cli", "prompt", "response", map[string]string{
		"project": "alpha",
		"profile": "engineer",
	})

	if err := RunExport([]string{"--format", "json", "--profile", "engineer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.json"))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	if !strings.Contains(string(data), `"profile": "engineer"`) {
		t.Fatalf("expected profile metadata in json export: %q", string(data))
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
	if payload["kind"] != jsonKindHistoryExport {
		t.Fatalf("unexpected kind: %#v", payload["kind"])
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

func TestExportJSONMetaFiltersRuleArtifacts(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")
	createTestProject(t, "alpha")

	captureRuleArtifactsForExportTest(t, workdir)

	if err := RunExport([]string{"--meta", "type=context", "--meta", "subtype=rules", "--format", "json"}, "v1.0.0"); err != nil {
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
	if len(events) != 1 {
		t.Fatalf("expected 1 filtered event, got %d", len(events))
	}
	event := events[0].(map[string]any)
	if event["type"] != "capture" {
		t.Fatalf("expected filtered event to be capture, got %v", event["type"])
	}
	if event["response"] != "Canonical system rules" {
		t.Fatalf("unexpected filtered capture response: %v", event["response"])
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

func TestExportHTMLOpenInvokesBrowser(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	db := openConfiguredDBForExportTest(t)
	defer db.Close()
	seedProject(t, db, "alpha")

	var openedPath string
	original := openExportInBrowser
	openExportInBrowser = func(path string) error {
		openedPath = path
		return nil
	}
	defer func() {
		openExportInBrowser = original
	}()

	if err := RunExport([]string{"--format", "html", "--open"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}
	if openedPath != "YANZI_LOG.html" {
		t.Fatalf("expected browser open for YANZI_LOG.html, got %q", openedPath)
	}
	if _, err := os.Stat(filepath.Join(workdir, "YANZI_LOG.html")); err != nil {
		t.Fatalf("expected html export file to exist: %v", err)
	}
}

func TestExportHTMLOpenReturnsUsefulErrorAfterWritingFile(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	db := openConfiguredDBForExportTest(t)
	defer db.Close()
	seedProject(t, db, "alpha")

	original := openExportInBrowser
	openExportInBrowser = func(path string) error {
		return errors.New("launcher unavailable")
	}
	defer func() {
		openExportInBrowser = original
	}()

	err := RunExport([]string{"--format", "html", "--open"}, "v1.0.0")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "open export in browser: launcher unavailable") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(workdir, "YANZI_LOG.html")); statErr != nil {
		t.Fatalf("expected html export file to still exist: %v", statErr)
	}
}

func TestExportMarkdownOpenRejected(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	err := RunExport([]string{"--format", "markdown", "--open"}, "v1.0.0")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--open is only supported with --format html") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExportJSONOpenRejected(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")

	err := RunExport([]string{"--format", "json", "--open"}, "v1.0.0")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--open is only supported with --format html") {
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
	if !strings.Contains(output, "Total artifacts: 1") || !strings.Contains(output, "Total events: 3") || !strings.Contains(output, "Checkpoints: 1") {
		t.Fatalf("missing counts: %q", output)
	}
	if !strings.Contains(output, "Continuity Mode") || !strings.Contains(output, "checkpoint") || !strings.Contains(output, "Open Work") {
		t.Fatalf("missing continuity summary cards: %q", output)
	}
	if !strings.Contains(output, "Checkpoint:</span> <span class=\"mono-inline\">") || !strings.Contains(output, "checkpoint 1") {
		t.Fatalf("missing sticky header checkpoint summary: %q", output)
	}
	if !strings.Contains(output, "position:sticky") {
		t.Fatalf("expected sticky header styling: %q", output)
	}
	if !strings.Contains(output, "id=\"event-search\"") || !strings.Contains(output, "Showing 3 of 3 events") {
		t.Fatalf("missing search UI: %q", output)
	}
	if !strings.Contains(output, ".timeline::before") || !strings.Contains(output, "class=\"timeline-marker\"") {
		t.Fatalf("missing timeline rail and markers: %q", output)
	}
	if !strings.Contains(output, ".timeline-marker{position:absolute;left:-97px;top:20px;width:22px;height:22px") {
		t.Fatalf("timeline markers were not reduced in size: %q", output)
	}
	if strings.Contains(output, "class=\"timeline-stamp\"") || strings.Contains(output, ".timeline-stamp{") {
		t.Fatalf("timeline timestamp labels should not be rendered: %q", output)
	}

	idxCapture := strings.Index(output, "Artifact: <span class=\"mono-inline\">cap-1</span>")
	idxCheckpoint := strings.Index(output, "Checkpoint: <span class=\"mono-inline\">")
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

	if !strings.Contains(output, "Show details") || !strings.Contains(output, "artifact-toggle") {
		t.Fatalf("missing artifact collapse controls: %q", output)
	}
	if !strings.Contains(output, "Copy prompt") || !strings.Contains(output, "Copy response") || !strings.Contains(output, "Copy capture ID") || !strings.Contains(output, "Copy checkpoint ID") || !strings.Contains(output, "Copy hash") {
		t.Fatalf("missing copy controls: %q", output)
	}
	if !strings.Contains(output, "<p class=\"artifact-preview\">line1 line2</p>") {
		t.Fatalf("missing prompt preview snippet: %q", output)
	}
	if !strings.Contains(output, "id=\"artifact-body-0\" hidden") {
		t.Fatalf("artifact body should be collapsed by default: %q", output)
	}
	if !strings.Contains(output, "class=\"timeline-entry timeline-entry-checkpoint event-card\"") || !strings.Contains(output, "CHECKPOINT") {
		t.Fatalf("checkpoint styling was not rendered: %q", output)
	}
	if !strings.Contains(output, "timeline-entry-checkpoint .timeline-marker") || !strings.Contains(output, "class=\"timeline-divider\"") || !strings.Contains(output, "class=\"checkpoint-divider\">Checkpoint boundary: checkpoint 1</div>") {
		t.Fatalf("checkpoint boundary marker was not rendered: %q", output)
	}
	if !strings.Contains(output, "class=\"timeline-entry event-card\"") || !strings.Contains(output, "class=\"capture timeline-card\"") {
		t.Fatalf("capture timeline layout was not rendered: %q", output)
	}
	if !strings.Contains(output, "class=\"timeline-entry timeline-entry-meta event-card\"") {
		t.Fatalf("timeline stamps or meta entry layout missing: %q", output)
	}
	if !strings.Contains(output, "<span class=\"badge badge-muted\">Capture</span>") ||
		!strings.Contains(output, "<span class=\"badge badge-muted\">Prompt</span>") ||
		!strings.Contains(output, "<span class=\"badge badge-muted\">Response</span>") ||
		!strings.Contains(output, "<span class=\"badge badge-muted\">Hash</span>") {
		t.Fatalf("missing capture semantic badges: %q", output)
	}
	if !strings.Contains(output, "<span class=\"badge badge-accent\">Role: engineer</span>") ||
		!strings.Contains(output, "<span class=\"badge badge-accent\">Source: cli</span>") ||
		!strings.Contains(output, "<span class=\"badge badge-muted\">Metadata</span>") {
		t.Fatalf("missing role/source/metadata badges: %q", output)
	}
	if !strings.Contains(output, "<span class=\"badge badge-strong\">Checkpoint</span>") ||
		!strings.Contains(output, "<span class=\"badge badge-strong\">Boundary</span>") ||
		!strings.Contains(output, "<span class=\"badge badge-strong\">Rehydration Anchor</span>") {
		t.Fatalf("missing checkpoint semantic badges: %q", output)
	}
	if !strings.Contains(output, "data-search=\"capture 2025-01-01T00:00:01Z cap-1 engineer") {
		t.Fatalf("missing capture search corpus: %q", output)
	}
	if !strings.Contains(output, "Role: engineer Source: cli Metadata") ||
		!strings.Contains(output, "Checkpoint Boundary Rehydration Anchor Hash") {
		t.Fatalf("badge text should be included in the search corpus: %q", output)
	}
	if !strings.Contains(output, "id=\"prompt-0\" class=\"content-block\"") || !strings.Contains(output, "id=\"response-0\" class=\"content-block\"") {
		t.Fatalf("prompt and response blocks should be rendered inside the artifact body: %q", output)
	}
	if !strings.Contains(output, "<pre>line1\nline2</pre>") {
		t.Fatalf("prompt pre block did not preserve whitespace: %q", output)
	}
	if !strings.Contains(output, "<pre>result\nok</pre>") {
		t.Fatalf("response pre block did not preserve whitespace: %q", output)
	}
	if !strings.Contains(output, "class=\"js-timestamp timeline-time\" data-timestamp=\"2025-01-01T00:00:01Z\" title=\"2025-01-01T00:00:01Z\">2025-01-01T00:00:01Z</span>") {
		t.Fatalf("missing raw timestamp tooltip hook: %q", output)
	}
	if !strings.Contains(output, "Intl.DateTimeFormat") || !strings.Contains(output, "formatTimestamps()") {
		t.Fatalf("expected client-side timestamp formatting: %q", output)
	}
	if !strings.Contains(output, "min-width:110px;height:34px") {
		t.Fatalf("expected consistent button styling: %q", output)
	}
	if !strings.Contains(output, ".timeline-time{font-weight:700;white-space:nowrap}") {
		t.Fatalf("expected non-wrapping bold timestamp styling: %q", output)
	}
	if !strings.Contains(output, "navigator.clipboard") || !strings.Contains(output, "document.execCommand('copy')") {
		t.Fatalf("expected clipboard copy with fallback: %q", output)
	}
	if strings.Contains(output, "<script src=") || strings.Contains(output, "<link rel=\"stylesheet\"") {
		t.Fatalf("html export must remain standalone: %q", output)
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
	if strings.Contains(output, "Metadata</span>") {
		t.Fatalf("did not expect metadata badge for project-only metadata: %q", output)
	}
}

func TestExportHTMLMetaFiltersRuleArtifacts(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")
	createTestProject(t, "alpha")

	captureRuleArtifactsForExportTest(t, workdir)

	if err := RunExport([]string{"--meta", "type=context", "--meta", "subtype=rules", "--format", "html"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.html"))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	output := string(data)
	if !strings.Contains(output, "Canonical system rules") || !strings.Contains(output, "Showing 1 of 1 events") {
		t.Fatalf("expected filtered html export output: %q", output)
	}
	if !strings.Contains(output, "<span class=\"badge badge-accent\">SYSTEM RULE</span>") {
		t.Fatalf("expected system rule label in html export: %q", output)
	}
	if strings.Contains(output, "<th>Metadata Key</th><th>Value</th>") {
		t.Fatalf("did not expect full metadata table in rule html export: %q", output)
	}
	if strings.Contains(output, "General context note") {
		t.Fatalf("did not expect non-rule artifact in html export: %q", output)
	}
}

func TestExportClaudeContext(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")
	createTestProject(t, "alpha")

	if _, err := yanzilibrary.CreateContextArtifact("", "process_rule", yanzilibrary.ContextScopeGlobal, "System Rules", "Never rewrite history.", `{"owner":"ops"}`); err != nil {
		t.Fatalf("CreateContextArtifact global: %v", err)
	}
	if _, err := yanzilibrary.CreateContextArtifact("alpha", "reference", yanzilibrary.ContextScopeProject, "API Reference", "https://example.test", ""); err != nil {
		t.Fatalf("CreateContextArtifact project: %v", err)
	}

	if err := RunExport([]string{"--format", "claude-context"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, "CLAUDE_CONTEXT.md"))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	output := string(data)
	if !strings.Contains(output, "# Claude Context") || !strings.Contains(output, "## process_rule") || !strings.Contains(output, "## reference") {
		t.Fatalf("expected grouped claude context export: %q", output)
	}
	if !strings.Contains(output, "- Scope: global") || !strings.Contains(output, "Never rewrite history.") {
		t.Fatalf("expected minimal context metadata and content: %q", output)
	}
}

func TestExportRequiresFormat(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	createTestProject(t, "alpha")
	writeStateFile(t, workdir, "alpha")

	err := RunExport([]string{}, "v1.0.0")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "--format is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExportClaudeContextRespectsMetaFilters(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	createTestProject(t, "alpha")
	if err := os.MkdirAll(filepath.Join(workdir, ".yanzi"), 0o700); err != nil {
		t.Fatalf("mkdir .yanzi: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workdir, ".yanzi", "project"), []byte("alpha\n"), 0o644); err != nil {
		t.Fatalf("write project binding: %v", err)
	}

	packPath := filepath.Join(workdir, "vibe-coder.yaml")
	contentPath := filepath.Join(workdir, "go.md")
	if err := os.WriteFile(contentPath, []byte("Return wrapped errors."), 0o644); err != nil {
		t.Fatalf("write content: %v", err)
	}
	packYAML := "name: vibe-coder\nversion: 1.0\ncontext:\n  - type: coding_standard\n    title: Go Standards\n    file: go.md\n"
	if err := os.WriteFile(packPath, []byte(packYAML), 0o644); err != nil {
		t.Fatalf("write pack: %v", err)
	}
	if err := RunPack([]string{"apply", packPath}); err != nil {
		t.Fatalf("RunPack apply: %v", err)
	}
	if _, err := yanzilibrary.CreateContextArtifact("alpha", "reference", yanzilibrary.ContextScopeProject, "API Reference", "https://example.test", ""); err != nil {
		t.Fatalf("CreateContextArtifact: %v", err)
	}

	if err := RunExport([]string{"--format", "claude-context", "--meta", "pack=vibe-coder"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, "CLAUDE_CONTEXT.md"))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	output := string(data)
	if !strings.Contains(output, "Go Standards") {
		t.Fatalf("expected pack artifact in export: %q", output)
	}
	if strings.Contains(output, "API Reference") {
		t.Fatalf("did not expect non-pack artifact in export: %q", output)
	}
}

func TestExportRejectsRemovedContextRetrievalFlags(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")
	createTestProject(t, "alpha")

	for _, flag := range []string{"--type", "--fields", "--order", "--limit"} {
		err := RunExport([]string{"--format", "claude-context", flag, "value"}, "v1.0.0")
		if err == nil {
			t.Fatalf("expected error for %s", flag)
		}
		if !strings.Contains(err.Error(), "flag provided but not defined") {
			t.Fatalf("expected unknown flag error for %s, got: %v", flag, err)
		}
	}
}

func TestExportHTMLShowsProfileRuleLabel(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")
	createTestProject(t, "alpha")

	rulesPath := filepath.Join(workdir, "ENGINEER_RULES.md")
	_ = os.WriteFile(rulesPath, []byte("Engineer rule body"), 0o644)
	if err := RunRules([]string{"add", rulesPath, "--scope", "project", "--profile", "engineer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add: %v", err)
	}

	if err := RunExport([]string{"--meta", "type=context", "--meta", "subtype=rules", "--format", "html", "--profile", "engineer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.html"))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	output := string(data)
	if !strings.Contains(output, "<span class=\"badge badge-accent\">PROFILE: engineer</span>") {
		t.Fatalf("expected profile rule label in html export: %q", output)
	}
	if strings.Contains(output, "<th>Metadata Key</th><th>Value</th>") {
		t.Fatalf("did not expect full metadata table in profile rule html export: %q", output)
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
	t.Setenv("YANZI_DB_PATH", cfg.DBPath)
	db, err := yanzilibrary.InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
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

func captureRuleArtifactsForExportTest(t *testing.T, workdir string) {
	t.Helper()

	systemRulesPath := filepath.Join(workdir, "SYSTEM_RULES.md")
	if err := os.WriteFile(systemRulesPath, []byte("# System Rules\nAlways verify changes.\n"), 0o644); err != nil {
		t.Fatalf("write SYSTEM_RULES.md: %v", err)
	}
	notesPath := filepath.Join(workdir, "project-notes.md")
	if err := os.WriteFile(notesPath, []byte("# Project Notes\nGeneral context.\n"), 0o644); err != nil {
		t.Fatalf("write project-notes.md: %v", err)
	}

	if err := RunCapture([]string{
		"--author", "human",
		"--title", "System Rules",
		"--prompt-file", systemRulesPath,
		"--response", "Canonical system rules",
		"--meta", "type=context",
		"--meta", "subtype=rules",
		"--meta", "scope=global",
		"--meta", "priority=critical",
	}); err != nil {
		t.Fatalf("RunCapture rules: %v", err)
	}

	if err := RunCapture([]string{
		"--author", "human",
		"--title", "Project Note",
		"--prompt-file", notesPath,
		"--response", "General context note",
		"--meta", "type=context",
		"--meta", "subtype=note",
		"--meta", "scope=project",
	}); err != nil {
		t.Fatalf("RunCapture note: %v", err)
	}
}
