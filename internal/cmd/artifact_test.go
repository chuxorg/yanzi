package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestRunContextAddFromFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	contentPath := filepath.Join(home, "policy.md")
	if err := os.WriteFile(contentPath, []byte("Document the release policy."), 0o644); err != nil {
		t.Fatalf("write content file: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunContext([]string{"add", "--type", "policy", "--title", "Release policy", "--file", contentPath})
	})
	if err != nil {
		t.Fatalf("RunContext add: %v", err)
	}
	if !strings.Contains(output, "\tpolicy\tRelease policy\t") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestRunContextRejectsInvalidType(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	err := RunContext([]string{"add", "--type", "note", "--title", "Bad type", "--content", "content"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "invalid context type") {
		t.Fatalf("unexpected error: %v", err)
	}
}
