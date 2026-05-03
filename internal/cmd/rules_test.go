package cmd

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/config"
	_ "modernc.org/sqlite"
)

func TestRunRulesAddCapturesRulesMetadata(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)

	rulesPath := filepath.Join(home, "SYSTEM_RULES.md")
	if err := os.WriteFile(rulesPath, []byte("# Rules\nAlways verify.\n"), 0o644); err != nil {
		t.Fatalf("write rules file: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunRules([]string{"add", rulesPath, "--priority", "critical"}, "v1.0.0")
	})
	if err != nil {
		t.Fatalf("RunRules add: %v", err)
	}
	if !strings.Contains(output, "id: ") || !strings.Contains(output, "hash: ") {
		t.Fatalf("unexpected add output: %q", output)
	}

	db := openRulesTestDB(t, home)
	defer db.Close()

	var (
		title    string
		prompt   string
		response string
		metaText string
	)
	if err := db.QueryRow(`SELECT title, prompt, response, meta FROM intents ORDER BY rowid DESC LIMIT 1`).Scan(&title, &prompt, &response, &metaText); err != nil {
		t.Fatalf("query rules capture: %v", err)
	}
	if title != "SYSTEM_RULES.md" {
		t.Fatalf("expected title from file name, got %q", title)
	}
	if prompt != "# Rules\nAlways verify.\n" {
		t.Fatalf("unexpected prompt content: %q", prompt)
	}
	if response != rulesResponse {
		t.Fatalf("expected wrapper response %q, got %q", rulesResponse, response)
	}

	var meta map[string]string
	if err := json.Unmarshal([]byte(metaText), &meta); err != nil {
		t.Fatalf("decode meta: %v", err)
	}
	if meta["type"] != "context" || meta["subtype"] != "rules" {
		t.Fatalf("unexpected rule metadata: %#v", meta)
	}
	if meta["scope"] != "global" {
		t.Fatalf("expected default global scope, got %#v", meta)
	}
	if meta["priority"] != "critical" {
		t.Fatalf("expected priority metadata, got %#v", meta)
	}
}

func TestRunRulesListFiltersRules(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	createTestProject(t, "alpha")
	writeStateFile(t, workdir, "alpha")

	globalRulesPath := filepath.Join(workdir, "SYSTEM_RULES.md")
	if err := os.WriteFile(globalRulesPath, []byte("# Global Rules\n"), 0o644); err != nil {
		t.Fatalf("write global rules: %v", err)
	}
	projectRulesPath := filepath.Join(workdir, "ALPHA_RULES.md")
	if err := os.WriteFile(projectRulesPath, []byte("# Project Rules\n"), 0o644); err != nil {
		t.Fatalf("write project rules: %v", err)
	}

	if err := RunRules([]string{"add", globalRulesPath}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add global: %v", err)
	}
	if err := RunRules([]string{"add", projectRulesPath, "--scope", "project", "--profile", "alpha"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add project: %v", err)
	}
	if err := RunCapture([]string{
		"--author", "human",
		"--title", "Project Note",
		"--prompt", "note",
		"--response", "General context note",
		"--meta", "type=context",
		"--meta", "subtype=note",
		"--meta", "scope=project",
	}); err != nil {
		t.Fatalf("RunCapture note: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunRules([]string{"list", "--scope", "project", "--profile", "alpha"}, "v1.0.0")
	})
	if err != nil {
		t.Fatalf("RunRules list: %v", err)
	}
	if !strings.Contains(output, "ALPHA_RULES.md") {
		t.Fatalf("expected project rules in list output: %q", output)
	}
	if strings.Contains(output, "SYSTEM_RULES.md") || strings.Contains(output, "Project Note") {
		t.Fatalf("did not expect non-matching records in list output: %q", output)
	}
}

func TestRunRulesExportFiltersRules(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	createTestProject(t, "alpha")
	writeStateFile(t, workdir, "alpha")

	rulesPath := filepath.Join(workdir, "SYSTEM_RULES.md")
	if err := os.WriteFile(rulesPath, []byte("# System Rules\nAlways verify changes.\n"), 0o644); err != nil {
		t.Fatalf("write rules file: %v", err)
	}
	if err := RunRules([]string{"add", rulesPath, "--profile", "default"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add: %v", err)
	}
	if err := RunCapture([]string{
		"--author", "human",
		"--title", "Project Note",
		"--prompt", "# Project Notes\nGeneral context.\n",
		"--response", "General context note",
		"--meta", "type=context",
		"--meta", "subtype=note",
		"--meta", "scope=global",
	}); err != nil {
		t.Fatalf("RunCapture note: %v", err)
	}

	if err := RunRules([]string{"export", "--format", "markdown", "--profile", "default"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules export: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.md"))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	output := string(data)
	if !strings.Contains(output, "# System Rules") || !strings.Contains(output, rulesResponse) {
		t.Fatalf("expected rules artifact in export output: %q", output)
	}
	if strings.Contains(output, "Project Note") || strings.Contains(output, "General context note") {
		t.Fatalf("did not expect non-rule artifact in export output: %q", output)
	}
}

func openRulesTestDB(t *testing.T, home string) *sql.DB {
	t.Helper()
	if err := os.Setenv("HOME", home); err != nil {
		t.Fatalf("set HOME: %v", err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db
}
