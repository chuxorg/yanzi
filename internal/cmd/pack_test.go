package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
	"gopkg.in/yaml.v3"
)

func TestRunPackApplyIsIdempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withCwd(t, home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	if err := os.MkdirAll(filepath.Join(home, ".yanzi"), 0o700); err != nil {
		t.Fatalf("mkdir .yanzi: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".yanzi", "project"), []byte("alpha\n"), 0o644); err != nil {
		t.Fatalf("write project binding: %v", err)
	}

	contentPath := filepath.Join(home, "go.md")
	if err := os.WriteFile(contentPath, []byte("Return wrapped errors."), 0o644); err != nil {
		t.Fatalf("write content: %v", err)
	}
	packPath := filepath.Join(home, "vibe-coder.yaml")
	packYAML := "name: vibe-coder\nversion: 1.0\ncontext:\n  - type: coding_standard\n    title: Go Standards\n    file: go.md\n"
	if err := os.WriteFile(packPath, []byte(packYAML), 0o644); err != nil {
		t.Fatalf("write pack: %v", err)
	}

	firstOutput, err := captureStdout(func() error {
		return RunPack([]string{"apply", packPath})
	})
	if err != nil {
		t.Fatalf("first apply: %v", err)
	}
	if !strings.Contains(firstOutput, "Go Standards (coding_standard): applied") {
		t.Fatalf("unexpected first apply output: %q", firstOutput)
	}

	secondOutput, err := captureStdout(func() error {
		return RunPack([]string{"apply", packPath})
	})
	if err != nil {
		t.Fatalf("second apply: %v", err)
	}
	if !strings.Contains(secondOutput, "Go Standards (coding_standard): already exists") {
		t.Fatalf("unexpected second apply output: %q", secondOutput)
	}

	artifacts, err := yanzilibrary.ListVisibleContextArtifacts("alpha", "coding_standard", "", "", false)
	if err != nil {
		t.Fatalf("list visible artifacts: %v", err)
	}
	if len(artifacts) != 1 {
		t.Fatalf("expected 1 artifact after repeated apply, got %d", len(artifacts))
	}
}

func TestRunPackExportWritesYamlAndSidecars(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withCwd(t, home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	if err := os.MkdirAll(filepath.Join(home, ".yanzi"), 0o700); err != nil {
		t.Fatalf("mkdir .yanzi: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, ".yanzi", "project"), []byte("alpha\n"), 0o644); err != nil {
		t.Fatalf("write project binding: %v", err)
	}

	if _, err := yanzilibrary.CreateContextArtifact("", "process_rule", yanzilibrary.ContextScopeGlobal, "System Rules", "Never rewrite history.", ""); err != nil {
		t.Fatalf("create global context: %v", err)
	}
	if _, err := yanzilibrary.CreateContextArtifact("alpha", "coding_standard", yanzilibrary.ContextScopeProject, "Go Standards", "Return wrapped errors.", ""); err != nil {
		t.Fatalf("create project context: %v", err)
	}

	outputPath := filepath.Join(home, "packs", "alpha.yaml")
	output, err := captureStdout(func() error {
		return RunPack([]string{"export", "--output", outputPath})
	})
	if err != nil {
		t.Fatalf("RunPack export: %v", err)
	}
	if !strings.Contains(output, "Exported "+outputPath) {
		t.Fatalf("unexpected export output: %q", output)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read export yaml: %v", err)
	}
	var pack packDefinition
	if err := yaml.Unmarshal(data, &pack); err != nil {
		t.Fatalf("unmarshal pack yaml: %v", err)
	}
	if pack.Name != "alpha" || pack.Version != "1.0" {
		t.Fatalf("unexpected pack header: %+v", pack)
	}
	if len(pack.Context) != 2 {
		t.Fatalf("expected 2 pack entries, got %d", len(pack.Context))
	}
	for _, entry := range pack.Context {
		if strings.TrimSpace(entry.File) == "" {
			t.Fatalf("expected sidecar file in entry: %+v", entry)
		}
		sidecarPath := filepath.Join(filepath.Dir(outputPath), entry.File)
		if _, err := os.Stat(sidecarPath); err != nil {
			t.Fatalf("missing sidecar file %s: %v", sidecarPath, err)
		}
	}
}
