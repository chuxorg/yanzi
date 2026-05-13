package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
	"gopkg.in/yaml.v3"
)

type packDefinition struct {
	Name    string      `yaml:"name"`
	Seed    string      `yaml:"seed"`
	Version string      `yaml:"version"`
	Context []packEntry `yaml:"context"`
}

type packEntry struct {
	Type  string `yaml:"type"`
	Title string `yaml:"title"`
	File  string `yaml:"file"`
	Scope string `yaml:"scope,omitempty"`
}

// RunPack applies or exports portable context packs.
//
// Problem:
// Reusing the same stored context across projects is tedious when each item
// must be added manually.
//
// Solution:
// RunPack exposes `apply` for idempotent pack loading and `export` for
// producing a pack definition plus sidecar files from visible context.
//
// Arguments:
//
//	args starts with `apply` or `export` followed by the corresponding flags.
func RunPack(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: yanzi pack <apply|export> [args]")
	}

	switch args[0] {
	case "apply":
		return runPackApply(args[1:])
	case "export":
		return runPackExport(args[1:])
	default:
		return errors.New("usage: yanzi pack <apply|export> [args]")
	}
}

func runPackApply(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: yanzi pack apply <file>")
	}

	packPath, err := resolvePackPath(args[0])
	if err != nil {
		return err
	}
	definition, err := loadPackDefinition(packPath)
	if err != nil {
		return err
	}
	if len(definition.Context) == 0 {
		return errors.New("pack contains no context entries")
	}

	activeProject, err := loadActiveProject()
	if err != nil {
		return err
	}
	existing, err := yanzilibrary.ListVisibleContextArtifacts(activeProject, "", "", "", false)
	if err != nil {
		return err
	}

	for _, entry := range definition.Context {
		if err := validatePackEntry(entry); err != nil {
			return err
		}
		entry.Type = normalizeContextType(entry.Type)
		if packEntryExists(existing, entry) {
			fmt.Printf("%s (%s): already exists\n", strings.TrimSpace(entry.Title), entry.Type)
			continue
		}

		contentPath := resolvePackEntryPath(packPath, entry.File)
		contextArgs := []string{
			"--type", entry.Type,
			"--title", strings.TrimSpace(entry.Title),
			"--file", contentPath,
		}
		if scope := strings.TrimSpace(entry.Scope); scope != "" {
			contextArgs = append(contextArgs, "--scope", scope)
		}
		if metadata, err := packMetadataJSON(definition); err != nil {
			return err
		} else if metadata != "" {
			contextArgs = append(contextArgs, "--metadata", metadata)
		}
		if err := runSilenced(func() error { return runContextAdd(contextArgs) }); err != nil {
			return fmt.Errorf("apply %s: %w", entry.Title, err)
		}
		fmt.Printf("%s (%s): applied\n", strings.TrimSpace(entry.Title), entry.Type)

		content, err := os.ReadFile(contentPath)
		if err != nil {
			return fmt.Errorf("read pack file: %w", err)
		}
		existing = append(existing, yanzilibrary.Artifact{
			Type:    entry.Type,
			Title:   strings.TrimSpace(entry.Title),
			Content: string(content),
		})
	}

	return nil
}

func runPackExport(args []string) error {
	fs := flag.NewFlagSet("pack export", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	output := fs.String("output", "", "pack output file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 || strings.TrimSpace(*output) == "" {
		return errors.New("usage: yanzi pack export --output <file>")
	}

	activeProject, err := loadActiveProject()
	if err != nil {
		return err
	}
	if activeProject == "" {
		return errors.New("no active project set")
	}

	artifacts, err := yanzilibrary.ListVisibleContextArtifacts(activeProject, "", "", "", false)
	if err != nil {
		return err
	}

	outputPath := strings.TrimSpace(*output)
	outputDir := filepath.Dir(outputPath)
	if outputDir == "" {
		outputDir = "."
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	pack := packDefinition{
		Name:    activeProject,
		Version: "1.0",
		Context: make([]packEntry, 0, len(artifacts)),
	}
	for i, artifact := range artifacts {
		fileName := packExportFilename(i, artifact)
		if err := os.WriteFile(filepath.Join(outputDir, fileName), []byte(artifact.Content), 0o644); err != nil {
			return fmt.Errorf("write pack content file: %w", err)
		}
		pack.Context = append(pack.Context, packEntry{
			Type:  artifact.Type,
			Title: artifact.Title,
			File:  fileName,
			Scope: artifact.Scope,
		})
	}

	data, err := yaml.Marshal(pack)
	if err != nil {
		return fmt.Errorf("encode pack: %w", err)
	}
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("write pack file: %w", err)
	}

	fmt.Printf("Exported %s\n", outputPath)
	return nil
}

func resolvePackPath(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errors.New("pack file is required")
	}
	if trimmed == "default" {
		return "", errors.New("default pack is not configured")
	}
	return trimmed, nil
}

func loadPackDefinition(path string) (packDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return packDefinition{}, fmt.Errorf("read pack file: %w", err)
	}
	var definition packDefinition
	if err := yaml.Unmarshal(data, &definition); err != nil {
		return packDefinition{}, fmt.Errorf("parse pack file: %w", err)
	}
	return definition, nil
}

func validatePackEntry(entry packEntry) error {
	if strings.TrimSpace(entry.Type) == "" {
		return errors.New("pack context type is required")
	}
	if strings.TrimSpace(entry.Title) == "" {
		return errors.New("pack context title is required")
	}
	if strings.TrimSpace(entry.File) == "" {
		return errors.New("pack context file is required")
	}
	return nil
}

func resolvePackEntryPath(packPath, fileValue string) string {
	if filepath.IsAbs(fileValue) {
		return fileValue
	}
	return filepath.Join(filepath.Dir(packPath), fileValue)
}

func packEntryExists(existing []yanzilibrary.Artifact, entry packEntry) bool {
	artifactType := normalizeContextType(entry.Type)
	title := strings.TrimSpace(entry.Title)
	for _, artifact := range existing {
		if artifact.Type == artifactType && strings.TrimSpace(artifact.Title) == title {
			return true
		}
	}
	return false
}

func packExportFilename(index int, artifact yanzilibrary.Artifact) string {
	return fmt.Sprintf("%02d-%s-%s.md", index+1, artifact.Type, slugify(artifact.Title))
}

func packMetadataJSON(definition packDefinition) (string, error) {
	payload := map[string]string{}
	if name := strings.TrimSpace(definition.Name); name != "" {
		payload["pack"] = name
	}
	if seed := strings.TrimSpace(definition.Seed); seed != "" {
		payload["seed"] = seed
	}
	if len(payload) == 0 {
		return "", nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode pack metadata: %w", err)
	}
	return string(data), nil
}
