package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

func TestRunIntentAddAndList(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	addOutput, err := captureStdout(func() error {
		return RunIntent([]string{"add", "--type", "decision", "--title", "Export direction", "--content", "Keep export deterministic."})
	})
	if err != nil {
		t.Fatalf("RunIntent add: %v", err)
	}
	if !strings.Contains(addOutput, "ID\tTYPE\tTITLE\tCREATED") {
		t.Fatalf("unexpected add output: %q", addOutput)
	}
	if !strings.Contains(addOutput, "\tdecision\tExport direction\t") {
		t.Fatalf("unexpected add row: %q", addOutput)
	}

	listOutput, err := captureStdout(func() error {
		return RunIntent([]string{"list", "--type", "decision"})
	})
	if err != nil {
		t.Fatalf("RunIntent list: %v", err)
	}
	if !strings.Contains(listOutput, "ID\tTYPE\tTITLE\tCREATED") {
		t.Fatalf("unexpected list output: %q", listOutput)
	}
	if !strings.Contains(listOutput, "Export direction") {
		t.Fatalf("expected listed artifact: %q", listOutput)
	}
}

func TestRunContextAddGlobal(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)

	output, err := captureStdout(func() error {
		return RunContext([]string{"add", "--type", "note", "--title", "Test Note", "--content", "This is a test.", "--scope", "global"})
	})
	if err != nil {
		t.Fatalf("RunContext add: %v", err)
	}
	if !strings.Contains(output, "ID\tTYPE\tSCOPE\tPROJECT\tTITLE\tCREATED") {
		t.Fatalf("unexpected output: %q", output)
	}
	if !strings.Contains(output, "\tnote\tglobal\t-\tTest Note\t") {
		t.Fatalf("unexpected row: %q", output)
	}
}

func TestRunContextAddProjectFromFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	contentPath := filepath.Join(home, "go_errors.md")
	if err := os.WriteFile(contentPath, []byte("Return wrapped errors."), 0o644); err != nil {
		t.Fatalf("write content file: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunContext([]string{"add", "--type", "coding_standard", "--title", "Go Error Handling", "--file", contentPath, "--scope", "project"})
	})
	if err != nil {
		t.Fatalf("RunContext add: %v", err)
	}
	if !strings.Contains(output, "\tcoding_standard\tproject\talpha\tGo Error Handling\t") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestRunContextRejectsInvalidType(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	err := RunContext([]string{"add", "--type", "architecture", "--title", "Bad type", "--content", "content"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "invalid context type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunContextRejectsInvalidScope(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	err := RunContext([]string{"add", "--type", "note", "--title", "Bad scope", "--content", "content", "--scope", "team"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "invalid context scope") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunContextRejectsProjectScopeWithoutCurrentProject(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)

	err := RunContext([]string{"add", "--type", "note", "--title", "Needs project", "--content", "content", "--scope", "project"})
	if err == nil {
		t.Fatal("expected project scope error")
	}
	if !strings.Contains(err.Error(), "project-scoped context requires an active project") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunContextListShowsVisibleArtifacts(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	if _, err := yanzilibrary.CreateContextArtifact("", "note", yanzilibrary.ContextScopeGlobal, "Global note", "Global content", ""); err != nil {
		t.Fatalf("CreateContextArtifact global: %v", err)
	}
	if _, err := yanzilibrary.CreateContextArtifact("alpha", "requirement", yanzilibrary.ContextScopeProject, "Project requirement", "Project content", ""); err != nil {
		t.Fatalf("CreateContextArtifact project: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunContext([]string{"list"})
	})
	if err != nil {
		t.Fatalf("RunContext list: %v", err)
	}
	if !strings.Contains(output, "ID\tTYPE\tSCOPE\tPROJECT\tTITLE\tCREATED") {
		t.Fatalf("unexpected output: %q", output)
	}
	if !strings.Contains(output, "Global note") || !strings.Contains(output, "Project requirement") {
		t.Fatalf("expected visible artifacts in list: %q", output)
	}
}

func TestRunContextShow(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	artifact, err := yanzilibrary.CreateContextArtifact("alpha", "reference", yanzilibrary.ContextScopeProject, "API Link", "https://example.test", `{"owner":"docs"}`)
	if err != nil {
		t.Fatalf("CreateContextArtifact: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunContext([]string{"show", shortArtifactID(artifact.ID)})
	})
	if err != nil {
		t.Fatalf("RunContext show: %v", err)
	}
	if !strings.Contains(output, "ID: "+artifact.ID) {
		t.Fatalf("expected full id in output: %q", output)
	}
	if !strings.Contains(output, "Type: reference") || !strings.Contains(output, "Project: alpha") {
		t.Fatalf("unexpected show output: %q", output)
	}
	if !strings.Contains(output, "https://example.test") {
		t.Fatalf("expected full content in show output: %q", output)
	}
}
