package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBootstrapDryRun(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withCwd(t, home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	if err := os.MkdirAll(filepath.Join(home, ".yanzi"), 0o700); err != nil {
		t.Fatalf("mkdir bootstrap dir: %v", err)
	}
	rulesPath := filepath.Join(home, "system-rules.md")
	if err := os.WriteFile(rulesPath, []byte("Never rewrite history."), 0o644); err != nil {
		t.Fatalf("write rules file: %v", err)
	}
	bootstrap := "documents:\n  - type: governance\n    title: System Rules\n    path: system-rules.md\n    scope: global\n"
	if err := os.WriteFile(filepath.Join(home, ".yanzi", "bootstrap.yaml"), []byte(bootstrap), 0o644); err != nil {
		t.Fatalf("write bootstrap config: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunBootstrap([]string{"--dry-run"})
	})
	if err != nil {
		t.Fatalf("RunBootstrap dry-run: %v", err)
	}
	if !strings.Contains(output, "Validated documents: 1") || !strings.Contains(output, "- process_rule: 1") {
		t.Fatalf("unexpected bootstrap dry-run output: %q", output)
	}
}

func TestRunBootstrapLoadsContextArtifacts(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withCwd(t, home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	if err := os.MkdirAll(filepath.Join(home, ".yanzi"), 0o700); err != nil {
		t.Fatalf("mkdir bootstrap dir: %v", err)
	}
	rulesPath := filepath.Join(home, "system-rules.md")
	if err := os.WriteFile(rulesPath, []byte("Never rewrite history."), 0o644); err != nil {
		t.Fatalf("write rules file: %v", err)
	}
	bootstrap := "documents:\n  - type: governance\n    title: System Rules\n    path: system-rules.md\n    scope: project\n"
	if err := os.WriteFile(filepath.Join(home, ".yanzi", "bootstrap.yaml"), []byte(bootstrap), 0o644); err != nil {
		t.Fatalf("write bootstrap config: %v", err)
	}

	if _, err := captureStdout(func() error {
		return RunBootstrap([]string{})
	}); err != nil {
		t.Fatalf("RunBootstrap: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunContext([]string{"list", "--type", "process_rule"})
	})
	if err != nil {
		t.Fatalf("RunContext list: %v", err)
	}
	if !strings.Contains(output, "System Rules") {
		t.Fatalf("expected bootstrapped context artifact: %q", output)
	}
}
