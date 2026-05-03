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
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	rulesPath := filepath.Join(home, "SYSTEM_RULES.md")
	if err := os.WriteFile(rulesPath, []byte("# Rules\nAlways verify.\n"), 0o644); err != nil {
		t.Fatalf("write rules file: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunRules([]string{"add", rulesPath, "--priority", "critical", "--profile", "engineer"}, "v1.0.0")
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
	if meta["profile"] != "engineer" {
		t.Fatalf("expected profile metadata, got %#v", meta)
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

func TestRunRulesExportComposeSystemOnly(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	createTestProject(t, "alpha")
	writeStateFile(t, workdir, "alpha")

	systemRules := filepath.Join(workdir, "SYSTEM_RULES.md")
	engineerRules := filepath.Join(workdir, "ENGINEER_RULES.md")
	_ = os.WriteFile(systemRules, []byte("# Global Rules\nAlways verify changes.\n"), 0o644)
	_ = os.WriteFile(engineerRules, []byte("# Engineer Rules\nPrefer narrow diffs.\n"), 0o644)

	if err := RunRules([]string{"add", systemRules}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add system: %v", err)
	}
	if err := RunRules([]string{"add", engineerRules, "--scope", "project", "--profile", "engineer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add engineer: %v", err)
	}

	if err := RunRules([]string{"export", "--format", "markdown", "--compose"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules export compose: %v", err)
	}
	md, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.md"))
	if err != nil {
		t.Fatalf("read markdown export: %v", err)
	}
	output := string(md)
	if !strings.Contains(output, "# SYSTEM RULES") || !strings.Contains(output, "# Global Rules") {
		t.Fatalf("expected system rules section, got: %q", output)
	}
	if strings.Contains(output, "# PROFILE: engineer") || strings.Contains(output, "# Engineer Rules") {
		t.Fatalf("did not expect profile rules without profile filter: %q", output)
	}
	if strings.Contains(output, "Metadata:") {
		t.Fatalf("did not expect metadata noise in composed export: %q", output)
	}
}

func TestRunRulesExportComposeIncludesProfileSection(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	createTestProject(t, "alpha")
	writeStateFile(t, workdir, "alpha")

	systemRules := filepath.Join(workdir, "SYSTEM_RULES.md")
	engineerRules := filepath.Join(workdir, "ENGINEER_RULES.md")
	reviewerRules := filepath.Join(workdir, "REVIEWER_RULES.md")
	_ = os.WriteFile(systemRules, []byte("# Global Rules\nAlways verify changes.\n"), 0o644)
	_ = os.WriteFile(engineerRules, []byte("# Engineer Rules\nPrefer narrow diffs.\n"), 0o644)
	_ = os.WriteFile(reviewerRules, []byte("# Reviewer Rules\nAsk for explicit sign-off.\n"), 0o644)

	if err := RunRules([]string{"add", systemRules}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add system: %v", err)
	}
	if err := RunRules([]string{"add", engineerRules, "--scope", "project", "--profile", "engineer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add engineer: %v", err)
	}
	if err := RunRules([]string{"add", reviewerRules, "--scope", "project", "--profile", "reviewer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add reviewer: %v", err)
	}

	if err := RunRules([]string{"export", "--format", "markdown", "--compose", "--profile", "engineer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules export compose profile: %v", err)
	}
	md, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.md"))
	if err != nil {
		t.Fatalf("read markdown export: %v", err)
	}
	output := string(md)
	systemIdx := strings.Index(output, "# SYSTEM RULES")
	profileIdx := strings.Index(output, "# PROFILE: engineer")
	globalIdx := strings.Index(output, "# Global Rules")
	engineerIdx := strings.Index(output, "# Engineer Rules")
	if systemIdx == -1 || profileIdx == -1 || globalIdx == -1 || engineerIdx == -1 {
		t.Fatalf("missing composed sections: %q", output)
	}
	if !(systemIdx < profileIdx && globalIdx < engineerIdx) {
		t.Fatalf("expected system rules before profile rules: %q", output)
	}
	if strings.Contains(output, "# Reviewer Rules") {
		t.Fatalf("did not expect other profile rules in composed export: %q", output)
	}
	if strings.Contains(output, "Metadata:") {
		t.Fatalf("did not expect metadata noise in composed export: %q", output)
	}
}

func TestRunRulesExportComposeIgnoredForJSON(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	createTestProject(t, "alpha")
	writeStateFile(t, workdir, "alpha")

	engineerRules := filepath.Join(workdir, "ENGINEER_RULES.md")
	_ = os.WriteFile(engineerRules, []byte("# Engineer Rules\nPrefer narrow diffs.\n"), 0o644)
	if err := RunRules([]string{"add", engineerRules, "--scope", "project", "--profile", "engineer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add engineer: %v", err)
	}

	if err := RunRules([]string{"export", "--format", "json", "--compose", "--profile", "engineer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules export json compose: %v", err)
	}
	js, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.json"))
	if err != nil {
		t.Fatalf("read json export: %v", err)
	}
	output := string(js)
	if !strings.Contains(output, `"schema_version": 1`) || !strings.Contains(output, `"profile": "engineer"`) {
		t.Fatalf("expected standard json export shape, got: %q", output)
	}
	if strings.Contains(output, "# SYSTEM RULES") || strings.Contains(output, "# PROFILE: engineer") {
		t.Fatalf("did not expect composed markdown structure in json export: %q", output)
	}
}
