package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunProjectCreate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)

	output, err := captureStdout(func() error {
		return RunProject([]string{"create", "alpha"})
	})
	if err != nil {
		t.Fatalf("RunProject create: %v", err)
	}

	if !strings.Contains(output, "Project created: alpha") {
		t.Fatalf("expected confirmation output, got %q", output)
	}
}

func TestRunProjectList(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)

	if err := RunProject([]string{"create", "alpha"}); err != nil {
		t.Fatalf("RunProject create: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunProject([]string{"list"})
	})
	if err != nil {
		t.Fatalf("RunProject list: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected header + rows, got %q", output)
	}
	if lines[0] != "Name\tCreatedAt\tDescription" {
		t.Fatalf("unexpected header: %q", lines[0])
	}
	if !strings.Contains(output, "alpha") {
		t.Fatalf("expected project name in output, got %q", output)
	}
}

func TestRunProjectCreateDuplicate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)

	if err := RunProject([]string{"create", "alpha"}); err != nil {
		t.Fatalf("initial create failed: %v", err)
	}

	err := RunProject([]string{"create", "alpha"})
	if err == nil {
		t.Fatal("expected duplicate project error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "already exists") {
		t.Fatalf("expected clear duplicate error, got %v", err)
	}
}

func TestRunInitCreatesProjectAndBinding(t *testing.T) {
	home := t.TempDir()
	repo := filepath.Join(home, "alpha")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	t.Setenv("HOME", home)
	withCwd(t, repo)
	writeTestConfig(t, home)

	output, err := captureStdout(func() error {
		return RunInit(nil)
	})
	if err != nil {
		t.Fatalf("RunInit: %v", err)
	}
	if !strings.Contains(output, "Project created: alpha") {
		t.Fatalf("unexpected init output: %q", output)
	}

	binding, err := os.ReadFile(filepath.Join(repo, ".yanzi", "project"))
	if err != nil {
		t.Fatalf("read binding: %v", err)
	}
	if strings.TrimSpace(string(binding)) != "alpha" {
		t.Fatalf("unexpected binding content: %q", string(binding))
	}
}

func TestLoadActiveProjectPrefersDirectoryBinding(t *testing.T) {
	home := t.TempDir()
	repo := filepath.Join(home, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".yanzi"), 0o755); err != nil {
		t.Fatalf("mkdir repo binding dir: %v", err)
	}
	t.Setenv("HOME", home)
	withCwd(t, repo)
	writeTestConfig(t, home)
	writeStateFile(t, home, "global-project")
	if err := os.WriteFile(filepath.Join(repo, ".yanzi", "project"), []byte("bound-project\n"), 0o644); err != nil {
		t.Fatalf("write binding: %v", err)
	}

	project, err := loadActiveProject()
	if err != nil {
		t.Fatalf("loadActiveProject: %v", err)
	}
	if project != "bound-project" {
		t.Fatalf("expected bound project, got %q", project)
	}
}
