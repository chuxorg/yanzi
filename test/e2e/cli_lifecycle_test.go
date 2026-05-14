package e2e_test

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestYanziProjectCaptureCheckpointExportAndRehydrate(t *testing.T) {
	h := newCLIHarness(t)

	projectName := "qa-e2e-project"

	res := h.run("project", "create", projectName)
	h.requireExitCode(res, 0)
	h.requireContains(res.Stdout, "Project created: "+projectName)

	res = h.run("project", "list")
	h.requireExitCode(res, 0)
	h.requireContains(res.Stdout, "Name\tCreatedAt\tDescription")
	h.requireContains(res.Stdout, projectName)

	res = h.run("project", "use", projectName)
	h.requireExitCode(res, 0)
	h.requireContains(res.Stdout, "Active project set to "+projectName+".")

	res = h.run("project", "current")
	h.requireExitCode(res, 0)
	h.requireContains(res.Stdout, "Active project: "+projectName)

	promptFile := filepath.Join(repoRoot(t), "test", "fixtures", "export", "prompt.txt")
	responseFile := filepath.Join(repoRoot(t), "test", "fixtures", "export", "response.txt")
	res = h.run(
		"capture",
		"--author", "qa-agent",
		"--prompt-file", promptFile,
		"--response-file", responseFile,
		"--title", "QA capture",
		"--meta", "phase=qa",
	)
	h.requireExitCode(res, 0)
	h.requireContains(res.Stdout, "id:")
	h.requireContains(res.Stdout, "hash:")

	res = h.run("list", "--author", "qa-agent")
	h.requireExitCode(res, 0)
	h.requireContains(res.Stdout, "qa-agent")
	h.requireContains(res.Stdout, "QA capture")

	res = h.run("checkpoint", "create", "--summary", "QA baseline")
	h.requireExitCode(res, 0)
	h.requireContains(res.Stdout, "id:")
	h.requireContains(res.Stdout, "summary: QA baseline")

	res = h.run("checkpoint", "list")
	h.requireExitCode(res, 0)
	h.requireContains(res.Stdout, "Project: "+projectName)
	h.requireContains(res.Stdout, "CreatedAt\tSummary")
	h.requireContains(res.Stdout, "QA baseline")

	res = h.run("export", "--format", "markdown")
	h.requireExitCode(res, 0)
	h.requireContains(res.Stdout, "Exported YANZI_LOG.md")
	md := h.requireFile("YANZI_LOG.md")
	compareOrUpdateSnapshot(t, "test/snapshots/export_markdown.snap", normalizeExportSnapshot(md))

	res = h.run("export", "--format", "json")
	h.requireExitCode(res, 0)
	h.requireContains(res.Stdout, "Exported YANZI_LOG.json")
	jsonContent := h.requireFile("YANZI_LOG.json")
	var payload map[string]any
	if err := json.Unmarshal([]byte(jsonContent), &payload); err != nil {
		t.Fatalf("parse json export: %v", err)
	}
	if payload["project"] == "" {
		t.Fatalf("json export missing project")
	}

	res = h.run("export", "--format", "html")
	h.requireExitCode(res, 0)
	h.requireContains(res.Stdout, "Exported YANZI_LOG.html")
	html := h.requireFile("YANZI_LOG.html")
	compareOrUpdateSnapshot(t, "test/snapshots/export_html.snap", normalizeExportSnapshot(html))

	res = h.run("rehydrate")
	h.requireExitCode(res, 0)
	h.requireContains(res.Stdout, "Project: "+projectName)
	h.requireContains(res.Stdout, "Post-Checkpoint Continuity")

	res = h.run("rehydrate", "--format", "json")
	h.requireExitCode(res, 0)
	var rehydrate map[string]any
	if err := json.Unmarshal([]byte(res.Stdout), &rehydrate); err != nil {
		t.Fatalf("parse rehydrate json: %v\nstdout:\n%s", err, res.Stdout)
	}
	if rehydrate["project"] != projectName {
		t.Fatalf("rehydrate project mismatch: got %v want %s", rehydrate["project"], projectName)
	}
}
