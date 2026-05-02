package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/client"
	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

func TestRunDeleteHidesFromListAndIncludeDeletedShowsIt(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	record := createTestIntentRecord(t, createIntentInput{
		Author:     "alice",
		SourceType: "cli",
		Title:      "Delete me",
		Prompt:     "prompt",
		Response:   "response",
	})

	if _, err := captureStdout(func() error {
		return RunDelete([]string{record.ID})
	}); err != nil {
		t.Fatalf("RunDelete: %v", err)
	}

	listOutput, err := captureStdout(func() error {
		return RunList([]string{})
	})
	if err != nil {
		t.Fatalf("RunList: %v", err)
	}
	if strings.Contains(listOutput, "Delete me") {
		t.Fatalf("did not expect tombstoned record in default list: %q", listOutput)
	}

	listDeletedOutput, err := captureStdout(func() error {
		return RunList([]string{"--include-deleted"})
	})
	if err != nil {
		t.Fatalf("RunList include deleted: %v", err)
	}
	if !strings.Contains(listDeletedOutput, "Delete me") {
		t.Fatalf("expected tombstoned record with --include-deleted: %q", listDeletedOutput)
	}
}

func TestRunRestoreMakesDeletedRecordVisibleAgain(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	record := createTestIntentRecord(t, createIntentInput{
		Author:     "alice",
		SourceType: "cli",
		Title:      "Restore me",
		Prompt:     "prompt",
		Response:   "response",
	})

	if err := RunDelete([]string{record.ID}); err != nil {
		t.Fatalf("RunDelete: %v", err)
	}
	if err := RunRestore([]string{record.ID}); err != nil {
		t.Fatalf("RunRestore: %v", err)
	}

	listOutput, err := captureStdout(func() error {
		return RunList([]string{})
	})
	if err != nil {
		t.Fatalf("RunList: %v", err)
	}
	if !strings.Contains(listOutput, "Restore me") {
		t.Fatalf("expected restored record in list: %q", listOutput)
	}
}

func TestRunDeleteCascadeOnlyWhenRequested(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	root := createTestIntentRecord(t, createIntentInput{
		Author:     "alice",
		SourceType: "cli",
		Title:      "root",
		Prompt:     "p1",
		Response:   "r1",
	})
	child := createTestIntentRecord(t, createIntentInput{
		Author:     "alice",
		SourceType: "cli",
		Title:      "child",
		Prompt:     "p2",
		Response:   "r2",
		PrevHash:   root.Hash,
	})
	grandchild := createTestIntentRecord(t, createIntentInput{
		Author:     "alice",
		SourceType: "cli",
		Title:      "grandchild",
		Prompt:     "p3",
		Response:   "r3",
		PrevHash:   child.Hash,
	})
	if grandchild.ID == "" {
		t.Fatal("expected grandchild record id")
	}

	if err := RunDelete([]string{root.ID}); err != nil {
		t.Fatalf("RunDelete: %v", err)
	}

	listOutput, err := captureStdout(func() error {
		return RunList([]string{})
	})
	if err != nil {
		t.Fatalf("RunList: %v", err)
	}
	if strings.Contains(listOutput, "root") {
		t.Fatalf("did not expect root after delete: %q", listOutput)
	}
	if !strings.Contains(listOutput, "child") || !strings.Contains(listOutput, "grandchild") {
		t.Fatalf("expected descendants to remain without cascade: %q", listOutput)
	}

	if err := RunDelete([]string{child.ID, "--cascade"}); err != nil {
		t.Fatalf("RunDelete cascade: %v", err)
	}
	listAfterCascade, err := captureStdout(func() error {
		return RunList([]string{})
	})
	if err != nil {
		t.Fatalf("RunList after cascade: %v", err)
	}
	if strings.Contains(listAfterCascade, "child") || strings.Contains(listAfterCascade, "grandchild") {
		t.Fatalf("did not expect descendants after cascade delete: %q", listAfterCascade)
	}
}

func TestRunDeleteCheckpointProtectionRequiresForce(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	artifact, err := yanzilibrary.CreateArtifact("alpha", yanzilibrary.ArtifactClassIntent, "decision", "Protected", "checkpoint-bound", "")
	if err != nil {
		t.Fatalf("CreateArtifact: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	db, err := openLocalDB(cfg)
	if err != nil {
		t.Fatalf("openLocalDB: %v", err)
	}
	defer db.Close()
	if _, err := yanzilibrary.CreateCheckpoint(context.Background(), db, "alpha", "snapshot", []string{artifact.ID}); err != nil {
		t.Fatalf("CreateCheckpoint: %v", err)
	}

	err = RunDelete([]string{artifact.ID})
	if err == nil {
		t.Fatal("expected checkpoint protection error")
	}
	if !strings.Contains(err.Error(), "use --force") {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := RunDelete([]string{artifact.ID, "--force"}); err != nil {
		t.Fatalf("RunDelete force: %v", err)
	}
}

func TestRunExportExcludesDeletedByDefaultAndIncludesWithFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withCwd(t, home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	record := createTestIntentRecord(t, createIntentInput{
		Author:     "alice",
		SourceType: "cli",
		Title:      "Export tombstone",
		Prompt:     "prompt",
		Response:   "response",
	})
	if err := RunDelete([]string{record.ID}); err != nil {
		t.Fatalf("RunDelete: %v", err)
	}

	if err := RunExport([]string{"--format", "markdown"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport: %v", err)
	}
	defaultOutput := readExportFile(t, home, "YANZI_LOG.md")
	if strings.Contains(defaultOutput, "Export tombstone") {
		t.Fatalf("did not expect deleted record in default export: %q", defaultOutput)
	}

	if err := RunExport([]string{"--format", "markdown", "--include-deleted"}, "v1.0.0"); err != nil {
		t.Fatalf("RunExport include-deleted: %v", err)
	}
	includedOutput := readExportFile(t, home, "YANZI_LOG.md")
	if !strings.Contains(includedOutput, record.ID) || !strings.Contains(includedOutput, "deleted: true") || !strings.Contains(includedOutput, "response") {
		t.Fatalf("expected deleted record in include-deleted export: %q", includedOutput)
	}
}

func createTestIntentRecord(t *testing.T, input createIntentInput) client.IntentRecord {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	db, err := openLocalDB(cfg)
	if err != nil {
		t.Fatalf("openLocalDB: %v", err)
	}
	defer db.Close()

	input.Meta, err = attachProjectMeta(input.Meta, "alpha")
	if err != nil {
		t.Fatalf("attachProjectMeta: %v", err)
	}
	record, err := buildLocalIntent(input)
	if err != nil {
		t.Fatalf("buildLocalIntent: %v", err)
	}
	if err := createLocalIntent(context.Background(), db, record); err != nil {
		t.Fatalf("createLocalIntent: %v", err)
	}
	return record
}

func readExportFile(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("read export file: %v", err)
	}
	return string(data)
}
