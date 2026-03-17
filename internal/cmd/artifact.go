package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

func RunIntent(args []string) error {
	return runArtifactCommand(yanzilibrary.ArtifactClassIntent, "note", args)
}

func RunContext(args []string) error {
	return runArtifactCommand(yanzilibrary.ArtifactClassContext, "requirement", args)
}

func runArtifactCommand(class, defaultType string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: yanzi %s <add|list> [args]", class)
	}

	switch args[0] {
	case "add":
		return runArtifactAdd(class, defaultType, args[1:])
	case "list":
		return runArtifactList(class, args[1:])
	default:
		return fmt.Errorf("usage: yanzi %s <add|list> [args]", class)
	}
}

func runArtifactAdd(class, defaultType string, args []string) error {
	fs := flag.NewFlagSet(class+" add", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	artifactType := fs.String("type", defaultType, "artifact type")
	title := fs.String("title", "", "artifact title")
	filePath := fs.String("file", "", "file path for content")
	contentFlag := fs.String("content", "", "artifact content")
	metadata := fs.String("metadata", "", "optional artifact metadata")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return fmt.Errorf("usage: yanzi %s add --title <title> [--type <type>] [--content <text> | --file <path>]", class)
	}
	if strings.TrimSpace(*title) == "" {
		return errors.New("--title is required")
	}

	content, err := resolveArtifactContent(*contentFlag, *filePath)
	if err != nil {
		return err
	}

	project, err := loadActiveProject()
	if err != nil {
		return err
	}
	if project == "" {
		return errors.New("no active project set")
	}

	artifact, err := yanzilibrary.CreateArtifact(project, class, *artifactType, *title, content, *metadata)
	if err != nil {
		return err
	}

	fmt.Printf("ID\tTYPE\tTITLE\tCREATED\n")
	fmt.Printf("%s\t%s\t%s\t%s\n", artifact.ID, artifact.Type, artifact.Title, artifact.CreatedAt)
	return nil
}

func resolveArtifactContent(contentValue, filePath string) (string, error) {
	if contentValue != "" {
		return contentValue, nil
	}
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("read content file: %w", err)
		}
		return string(data), nil
	}
	hasStdin, err := stdinHasData()
	if err != nil {
		return "", err
	}
	if !hasStdin {
		return "", errors.New("content must be provided with --content, --file, or stdin")
	}
	data, err := readPromptFromStdin()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func runArtifactList(class string, args []string) error {
	fs := flag.NewFlagSet(class+" list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	artifactType := fs.String("type", "", "artifact type filter")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return fmt.Errorf("usage: yanzi %s list [--type <type>]", class)
	}

	project, err := loadActiveProject()
	if err != nil {
		return err
	}
	if project == "" {
		return errors.New("no active project set")
	}

	artifacts, err := yanzilibrary.ListArtifacts(project, class, *artifactType)
	if err != nil {
		return err
	}

	fmt.Println("ID\tTYPE\tTITLE\tCREATED")
	for _, artifact := range artifacts {
		fmt.Printf("%s\t%s\t%s\t%s\n", artifact.ID, artifact.Type, artifact.Title, artifact.CreatedAt)
	}
	return nil
}
