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

func TestRunRulesAddStoresProfileMetadata(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	rulesPath := filepath.Join(home, "SYSTEM_RULES.md")
	if err := os.WriteFile(rulesPath, []byte("# Rules\nAlways verify.\n"), 0o644); err != nil {
		t.Fatalf("write rules file: %v", err)
	}

	if err := RunRules([]string{"add", rulesPath, "--profile", "engineer", "--priority", "critical"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var metaText string
	if err := db.QueryRow(`SELECT meta FROM intents ORDER BY rowid DESC LIMIT 1`).Scan(&metaText); err != nil {
		t.Fatalf("query meta: %v", err)
	}
	var meta map[string]string
	if err := json.Unmarshal([]byte(metaText), &meta); err != nil {
		t.Fatalf("decode meta: %v", err)
	}
	if meta["subtype"] != "rules" || meta["profile"] != "engineer" || meta["priority"] != "critical" {
		t.Fatalf("unexpected rule metadata: %#v", meta)
	}
}

func TestRunRulesListFiltersByProfile(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	createTestProject(t, "alpha")
	writeStateFile(t, workdir, "alpha")

	rulesA := filepath.Join(workdir, "ENGINEER_RULES.md")
	rulesB := filepath.Join(workdir, "REVIEWER_RULES.md")
	_ = os.WriteFile(rulesA, []byte("# Engineer\n"), 0o644)
	_ = os.WriteFile(rulesB, []byte("# Reviewer\n"), 0o644)

	if err := RunRules([]string{"add", rulesA, "--profile", "engineer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add engineer: %v", err)
	}
	if err := RunRules([]string{"add", rulesB, "--profile", "reviewer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add reviewer: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunRules([]string{"list", "--profile", "engineer"}, "v1.0.0")
	})
	if err != nil {
		t.Fatalf("RunRules list: %v", err)
	}
	if !strings.Contains(output, "ENGINEER_RULES.md") || !strings.Contains(output, "profile=engineer") {
		t.Fatalf("expected engineer rule in output: %q", output)
	}
	if strings.Contains(output, "REVIEWER_RULES.md") || strings.Contains(output, "profile=reviewer") {
		t.Fatalf("did not expect reviewer rule in output: %q", output)
	}
}

func TestRunRulesListFiltersByScope(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	createTestProject(t, "alpha")
	writeStateFile(t, workdir, "alpha")

	globalRules := filepath.Join(workdir, "SYSTEM_RULES.md")
	projectRules := filepath.Join(workdir, "PROJECT_RULES.md")
	_ = os.WriteFile(globalRules, []byte("# Global\n"), 0o644)
	_ = os.WriteFile(projectRules, []byte("# Project\n"), 0o644)

	if err := RunRules([]string{"add", globalRules, "--scope", "global"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add global: %v", err)
	}
	if err := RunRules([]string{"add", projectRules, "--scope", "project"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add project: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunRules([]string{"list", "--scope", "global"}, "v1.0.0")
	})
	if err != nil {
		t.Fatalf("RunRules list scope: %v", err)
	}
	if !strings.Contains(output, "SYSTEM_RULES.md") || !strings.Contains(output, "scope=global") {
		t.Fatalf("expected global rule in output: %q", output)
	}
	if strings.Contains(output, "PROJECT_RULES.md") || strings.Contains(output, "scope=project") {
		t.Fatalf("did not expect project rule in global scope output: %q", output)
	}
}

func TestRunRulesExportFiltersByProfile(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	createTestProject(t, "alpha")
	writeStateFile(t, workdir, "alpha")

	rulesA := filepath.Join(workdir, "ENGINEER_RULES.md")
	rulesB := filepath.Join(workdir, "REVIEWER_RULES.md")
	_ = os.WriteFile(rulesA, []byte("# Engineer Rules\n"), 0o644)
	_ = os.WriteFile(rulesB, []byte("# Reviewer Rules\n"), 0o644)

	if err := RunRules([]string{"add", rulesA, "--profile", "engineer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add engineer: %v", err)
	}
	if err := RunRules([]string{"add", rulesB, "--profile", "reviewer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules add reviewer: %v", err)
	}

	if err := RunRules([]string{"export", "--format", "markdown", "--profile", "engineer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules export markdown: %v", err)
	}
	md, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.md"))
	if err != nil {
		t.Fatalf("read markdown export: %v", err)
	}
	if !strings.Contains(string(md), "profile: engineer") || strings.Contains(string(md), "profile: reviewer") {
		t.Fatalf("unexpected markdown export: %q", string(md))
	}
	if err := RunRules([]string{"export", "--format", "json", "--profile", "engineer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules export json: %v", err)
	}
	js, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.json"))
	if err != nil {
		t.Fatalf("read json export: %v", err)
	}
	if !strings.Contains(string(js), `"profile": "engineer"`) || strings.Contains(string(js), `"profile": "reviewer"`) {
		t.Fatalf("unexpected json export: %q", string(js))
	}
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

func TestRunRulesExportComposeIncludesHTMLSections(t *testing.T) {
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

	if err := RunRules([]string{"export", "--format", "html", "--compose", "--profile", "engineer"}, "v1.0.0"); err != nil {
		t.Fatalf("RunRules export compose html: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(workdir, "YANZI_LOG.html"))
	if err != nil {
		t.Fatalf("read html export: %v", err)
	}
	output := string(data)
	systemIdx := strings.Index(output, "SYSTEM RULES")
	profileIdx := strings.Index(output, "PROFILE: engineer")
	globalIdx := strings.Index(output, "# Global Rules")
	engineerIdx := strings.Index(output, "# Engineer Rules")
	if systemIdx == -1 || profileIdx == -1 || globalIdx == -1 || engineerIdx == -1 {
		t.Fatalf("missing composed html sections: %q", output)
	}
	if !(systemIdx < profileIdx && globalIdx < engineerIdx) {
		t.Fatalf("expected system section before profile section: %q", output)
	}
}
