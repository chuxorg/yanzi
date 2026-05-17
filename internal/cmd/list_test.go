package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunListMetaFiltersRuleArtifacts(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	createTestProject(t, "alpha")
	writeStateFile(t, workdir, "alpha")

	systemRulesPath := filepath.Join(workdir, "system-rules.md")
	if err := os.WriteFile(systemRulesPath, []byte("# System Rules\nAlways verify changes.\n"), 0o644); err != nil {
		t.Fatalf("write system-rules.md: %v", err)
	}
	notesPath := filepath.Join(workdir, "notes.md")
	if err := os.WriteFile(notesPath, []byte("This is a project note.\n"), 0o644); err != nil {
		t.Fatalf("write notes.md: %v", err)
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

	output, err := captureStdout(func() error {
		return RunList([]string{"--meta", "type=context", "--meta", "subtype=rules"})
	})
	if err != nil {
		t.Fatalf("RunList: %v", err)
	}

	if !strings.Contains(output, "System Rules") {
		t.Fatalf("expected rules artifact in list output: %q", output)
	}
	if strings.Contains(output, "Project Note") {
		t.Fatalf("did not expect non-rule artifact in list output: %q", output)
	}
}

func TestRunListProfileFilterAndMetadataVisibility(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	createTestProject(t, "alpha")
	writeStateFile(t, workdir, "alpha")

	if err := RunCapture([]string{
		"--author", "human",
		"--title", "Engineer Note",
		"--prompt", "prompt",
		"--response", "response",
		"--profile", "engineer",
		"--meta", "area=auth",
	}); err != nil {
		t.Fatalf("RunCapture engineer: %v", err)
	}
	if err := RunCapture([]string{
		"--author", "human",
		"--title", "Reviewer Note",
		"--prompt", "prompt",
		"--response", "response",
		"--profile", "reviewer",
	}); err != nil {
		t.Fatalf("RunCapture reviewer: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunList([]string{"--profile", "engineer"})
	})
	if err != nil {
		t.Fatalf("RunList: %v", err)
	}
	if !strings.Contains(output, "ID\tCreated_At\tAuthor\tSource\tTitle\tMetadata") {
		t.Fatalf("expected metadata column in list output: %q", output)
	}
	if !strings.Contains(output, "Engineer Note") || !strings.Contains(output, "profile=engineer") {
		t.Fatalf("expected engineer profile metadata in list output: %q", output)
	}
	if strings.Contains(output, "Reviewer Note") || strings.Contains(output, "profile=reviewer") {
		t.Fatalf("did not expect reviewer record in filtered list output: %q", output)
	}
}

func TestRunListIsScopedToActiveProject(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	createTestProject(t, "alpha")
	createTestProject(t, "beta")
	writeStateFile(t, workdir, "alpha")

	if err := RunCapture([]string{
		"--author", "human",
		"--title", "Alpha Note",
		"--prompt", "alpha",
		"--response", "alpha",
	}); err != nil {
		t.Fatalf("RunCapture alpha: %v", err)
	}

	writeStateFile(t, workdir, "beta")
	if err := RunCapture([]string{
		"--author", "human",
		"--title", "Beta Note",
		"--prompt", "beta",
		"--response", "beta",
	}); err != nil {
		t.Fatalf("RunCapture beta: %v", err)
	}

	writeStateFile(t, workdir, "alpha")
	output, err := captureStdout(func() error {
		return RunList([]string{})
	})
	if err != nil {
		t.Fatalf("RunList: %v", err)
	}
	if !strings.Contains(output, "Alpha Note") {
		t.Fatalf("expected alpha record in list output: %q", output)
	}
	if strings.Contains(output, "Beta Note") {
		t.Fatalf("did not expect beta record in alpha-scoped list output: %q", output)
	}
	if !strings.Contains(output, "Project: alpha") {
		t.Fatalf("expected project header in scoped list output: %q", output)
	}
}

func TestRunListAllProjectsIncludesProjectColumn(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	createTestProject(t, "alpha")
	createTestProject(t, "beta")
	writeStateFile(t, workdir, "alpha")

	if err := RunCapture([]string{
		"--author", "human",
		"--title", "Alpha Note",
		"--prompt", "alpha",
		"--response", "alpha",
	}); err != nil {
		t.Fatalf("RunCapture alpha: %v", err)
	}

	writeStateFile(t, workdir, "beta")
	if err := RunCapture([]string{
		"--author", "human",
		"--title", "Beta Note",
		"--prompt", "beta",
		"--response", "beta",
	}); err != nil {
		t.Fatalf("RunCapture beta: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunList([]string{"--all-projects"})
	})
	if err != nil {
		t.Fatalf("RunList all projects: %v", err)
	}
	if !strings.Contains(output, "Project: All projects") {
		t.Fatalf("expected all-projects header: %q", output)
	}
	if !strings.Contains(output, "ID\tCreated_At\tProject\tAuthor\tSource\tTitle\tMetadata") {
		t.Fatalf("expected project column in all-projects output: %q", output)
	}
	if !strings.Contains(output, "Alpha Note") || !strings.Contains(output, "Beta Note") {
		t.Fatalf("expected both projects in output: %q", output)
	}
	if !strings.Contains(output, "\talpha\t") || !strings.Contains(output, "\tbeta\t") {
		t.Fatalf("expected explicit project values in output: %q", output)
	}
}

func TestRunListAllProjectsUsesDeterministicOrdering(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	createTestProject(t, "alpha")
	createTestProject(t, "beta")
	writeStateFile(t, workdir, "alpha")

	if err := RunCapture([]string{
		"--author", "human",
		"--title", "Older Note",
		"--prompt", "older",
		"--response", "older",
	}); err != nil {
		t.Fatalf("RunCapture older: %v", err)
	}

	writeStateFile(t, workdir, "beta")
	if err := RunCapture([]string{
		"--author", "human",
		"--title", "Newer Note",
		"--prompt", "newer",
		"--response", "newer",
	}); err != nil {
		t.Fatalf("RunCapture newer: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunList([]string{"--all-projects"})
	})
	if err != nil {
		t.Fatalf("RunList all projects: %v", err)
	}
	if strings.Index(output, "Newer Note") > strings.Index(output, "Older Note") {
		t.Fatalf("expected newest-first ordering in output: %q", output)
	}
}
