package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

func TestCheckpointCreateSuccess(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	output, err := captureStdout(func() error {
		return RunCheckpoint([]string{"create", "--summary", "first checkpoint"})
	})
	if err != nil {
		t.Fatalf("RunCheckpoint create: %v", err)
	}
	if !strings.Contains(output, "id: ") {
		t.Fatalf("expected id output, got %q", output)
	}
	if !strings.Contains(output, "summary: first checkpoint") {
		t.Fatalf("expected summary output, got %q", output)
	}
}

func TestCheckpointCreateNoActiveProject(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)

	err := RunCheckpoint([]string{"create", "--summary", "first checkpoint"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no active project") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckpointListEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	output, err := captureStdout(func() error {
		return RunCheckpoint([]string{"list"})
	})
	if err != nil {
		t.Fatalf("RunCheckpoint list: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected header only, got %q", output)
	}
	if lines[0] != "Index\tCreatedAt\tSummary" {
		t.Fatalf("unexpected header: %q", lines[0])
	}
}

func TestCheckpointListPopulated(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")
	createTestCheckpoint(t, "alpha", "first")
	createTestCheckpoint(t, "alpha", "second")

	output, err := captureStdout(func() error {
		return RunCheckpoint([]string{"list"})
	})
	if err != nil {
		t.Fatalf("RunCheckpoint list: %v", err)
	}
	if !strings.Contains(output, "first") || !strings.Contains(output, "second") {
		t.Fatalf("expected summaries in output, got %q", output)
	}
}

func createTestCheckpoint(t *testing.T, project, summary string) {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	ctx := context.Background()
	db, err := openLocalDB(cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := yanzilibrary.CreateCheckpoint(ctx, db, project, summary, []string{}); err != nil {
		t.Fatalf("create checkpoint: %v", err)
	}
}
