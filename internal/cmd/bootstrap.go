package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type bootstrapConfig struct {
	Documents []bootstrapDocument `yaml:"documents"`
}

type bootstrapDocument struct {
	Type     string            `yaml:"type"`
	Title    string            `yaml:"title"`
	Path     string            `yaml:"path"`
	File     string            `yaml:"file"`
	Scope    string            `yaml:"scope"`
	Metadata map[string]string `yaml:"metadata"`
}

func RunBootstrap(args []string) error {
	fs := flag.NewFlagSet("bootstrap", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dryRun := fs.Bool("dry-run", false, "validate bootstrap documents without loading them")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("usage: yanzi bootstrap [--dry-run]")
	}

	configPath := filepath.Join(".yanzi", "bootstrap.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read bootstrap config: %w", err)
	}

	var cfg bootstrapConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse bootstrap config: %w", err)
	}
	if len(cfg.Documents) == 0 {
		return errors.New("bootstrap config contains no documents")
	}

	typeCounts := map[string]int{}
	failures := make([]string, 0)

	for i, doc := range cfg.Documents {
		if err := validateBootstrapDocument(configPath, doc); err != nil {
			failures = append(failures, fmt.Sprintf("document %d: %v", i+1, err))
			continue
		}
		typeCounts[normalizeContextType(doc.Type)]++
	}
	if len(failures) > 0 {
		printBootstrapSummary("Validated", 0, typeCounts, failures)
		return errors.New("bootstrap validation failed")
	}
	if *dryRun {
		printBootstrapSummary("Validated", len(cfg.Documents), typeCounts, nil)
		return nil
	}

	loaded := 0
	failures = failures[:0]
	for i, doc := range cfg.Documents {
		args, err := bootstrapContextArgs(configPath, doc)
		if err != nil {
			failures = append(failures, fmt.Sprintf("document %d: %v", i+1, err))
			continue
		}
		if err := runSilenced(func() error { return runContextAdd(args) }); err != nil {
			failures = append(failures, fmt.Sprintf("document %d: %v", i+1, err))
			continue
		}
		loaded++
	}

	printBootstrapSummary("Loaded", loaded, typeCounts, failures)
	if len(failures) > 0 {
		return errors.New("bootstrap completed with failures")
	}
	return nil
}

func runSilenced(fn func() error) error {
	reader, writer, err := os.Pipe()
	if err != nil {
		return err
	}
	stdout := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = stdout
	}()

	runErr := fn()
	_ = writer.Close()
	_, _ = io.Copy(io.Discard, reader)
	_ = reader.Close()
	return runErr
}

func validateBootstrapDocument(configPath string, doc bootstrapDocument) error {
	artifactType := normalizeContextType(doc.Type)
	if artifactType == "" {
		return errors.New("type is required")
	}
	title := strings.TrimSpace(doc.Title)
	if title == "" {
		return errors.New("title is required")
	}
	resolvedPath, err := resolveBootstrapPath(configPath, doc)
	if err != nil {
		return err
	}
	info, err := os.Stat(resolvedPath)
	if err != nil {
		return fmt.Errorf("stat %s: %w", resolvedPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", resolvedPath)
	}
	return nil
}

func bootstrapContextArgs(configPath string, doc bootstrapDocument) ([]string, error) {
	resolvedPath, err := resolveBootstrapPath(configPath, doc)
	if err != nil {
		return nil, err
	}
	args := []string{
		"--type", normalizeContextType(doc.Type),
		"--title", strings.TrimSpace(doc.Title),
		"--file", resolvedPath,
	}
	if trimmed := strings.TrimSpace(doc.Scope); trimmed != "" {
		args = append(args, "--scope", trimmed)
	}
	if len(doc.Metadata) > 0 {
		payload, err := json.Marshal(doc.Metadata)
		if err != nil {
			return nil, fmt.Errorf("encode metadata: %w", err)
		}
		args = append(args, "--metadata", string(payload))
	}
	return args, nil
}

func resolveBootstrapPath(configPath string, doc bootstrapDocument) (string, error) {
	pathValue := strings.TrimSpace(doc.Path)
	if pathValue == "" {
		pathValue = strings.TrimSpace(doc.File)
	}
	if pathValue == "" {
		return "", errors.New("path is required")
	}
	if filepath.IsAbs(pathValue) {
		return pathValue, nil
	}
	bootstrapRelative := filepath.Join(filepath.Dir(configPath), pathValue)
	if _, err := os.Stat(bootstrapRelative); err == nil {
		return bootstrapRelative, nil
	}
	return pathValue, nil
}

func printBootstrapSummary(action string, loaded int, typeCounts map[string]int, failures []string) {
	fmt.Printf("%s documents: %d\n", action, loaded)
	if len(typeCounts) == 0 {
		fmt.Println("Types loaded: none")
	} else {
		keys := make([]string, 0, len(typeCounts))
		for key := range typeCounts {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		fmt.Println("Types loaded:")
		for _, key := range keys {
			fmt.Printf("- %s: %d\n", key, typeCounts[key])
		}
	}
	if len(failures) == 0 {
		fmt.Println("Failures: none")
		return
	}
	fmt.Println("Failures:")
	for _, failure := range failures {
		fmt.Printf("- %s\n", failure)
	}
}
