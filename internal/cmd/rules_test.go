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
