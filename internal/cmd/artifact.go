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
	return runIntentCommand(args)
}

func RunContext(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: yanzi context <add|list|show> [args]")
	}

	switch args[0] {
	case "add":
		return runContextAdd(args[1:])
	case "list":
		return runContextList(args[1:])
	case "show":
		return runContextShow(args[1:])
	default:
		return errors.New("usage: yanzi context <add|list|show> [args]")
	}
}

func runIntentCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: yanzi intent <add|list> [args]")
	}

	switch args[0] {
	case "add":
		return runIntentAdd(args[1:])
	case "list":
		return runIntentList(args[1:])
	default:
		return errors.New("usage: yanzi intent <add|list> [args]")
	}
}

func runIntentAdd(args []string) error {
	fs := flag.NewFlagSet("intent add", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	artifactType := fs.String("type", "note", "artifact type")
	title := fs.String("title", "", "artifact title")
	filePath := fs.String("file", "", "file path for content")
	contentFlag := fs.String("content", "", "artifact content")
	metadata := fs.String("metadata", "", "optional artifact metadata")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("usage: yanzi intent add --title <title> [--type <type>] [--content <text> | --file <path>]")
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

	artifact, err := yanzilibrary.CreateArtifact(project, yanzilibrary.ArtifactClassIntent, *artifactType, *title, content, *metadata)
	if err != nil {
		return err
	}

	fmt.Printf("ID\tTYPE\tTITLE\tCREATED\n")
	fmt.Printf("%s\t%s\t%s\t%s\n", artifact.ID, artifact.Type, artifact.Title, artifact.CreatedAt)
	return nil
}

func runIntentList(args []string) error {
	fs := flag.NewFlagSet("intent list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	artifactType := fs.String("type", "", "artifact type filter")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("usage: yanzi intent list [--type <type>]")
	}

	project, err := loadActiveProject()
	if err != nil {
		return err
	}
	if project == "" {
		return errors.New("no active project set")
	}

	artifacts, err := yanzilibrary.ListArtifacts(project, yanzilibrary.ArtifactClassIntent, *artifactType, false)
	if err != nil {
		return err
	}

	fmt.Println("ID\tTYPE\tTITLE\tCREATED")
	for _, artifact := range artifacts {
		fmt.Printf("%s\t%s\t%s\t%s\n", artifact.ID, artifact.Type, artifact.Title, artifact.CreatedAt)
	}
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

func runContextAdd(args []string) error {
	fs := flag.NewFlagSet("context add", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	artifactType := fs.String("type", "", "context type")
	title := fs.String("title", "", "context title")
	filePath := fs.String("file", "", "file path for content")
	contentFlag := fs.String("content", "", "context content")
	scope := fs.String("scope", "", "context scope (global|project)")
	metadata := fs.String("metadata", "", "optional context metadata")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("usage: yanzi context add --type <type> --title <title> [--scope <global|project>] [--content <text> | --file <path>]")
	}
	if strings.TrimSpace(*artifactType) == "" {
		return errors.New("--type is required")
	}
	*artifactType = normalizeContextType(*artifactType)
	if strings.TrimSpace(*title) == "" {
		return errors.New("--title is required")
	}

	content, err := resolveArtifactContent(*contentFlag, *filePath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(content) == "" {
		return errors.New("content is required")
	}

	activeProject, err := loadActiveProject()
	if err != nil {
		return err
	}

	scopeValue := strings.TrimSpace(*scope)
	if scopeValue == "" {
		if activeProject != "" {
			scopeValue = yanzilibrary.ContextScopeProject
		} else {
			scopeValue = yanzilibrary.ContextScopeGlobal
		}
	}

	project := ""
	if scopeValue == yanzilibrary.ContextScopeProject {
		project = activeProject
		if project == "" {
			return errors.New("project-scoped context requires an active project")
		}
	}

	artifact, err := yanzilibrary.CreateContextArtifact(project, *artifactType, scopeValue, *title, content, *metadata)
	if err != nil {
		return err
	}

	fmt.Printf("ID\tTYPE\tSCOPE\tPROJECT\tTITLE\tCREATED\n")
	fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", shortArtifactID(artifact.ID), artifact.Type, artifact.Scope, displayProject(artifact.Project), artifact.Title, artifact.CreatedAt)
	return nil
}

func runContextList(args []string) error {
	fs := flag.NewFlagSet("context list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	artifactType := fs.String("type", "", "context type filter")
	scope := fs.String("scope", "", "context scope filter")
	project := fs.String("project", "", "project filter")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("usage: yanzi context list [--type <type>] [--scope <global|project>] [--project <name>]")
	}
	*artifactType = normalizeContextType(*artifactType)

	activeProject, err := loadActiveProject()
	if err != nil {
		return err
	}

	artifacts, err := yanzilibrary.ListVisibleContextArtifacts(activeProject, *artifactType, *scope, *project, false)
	if err != nil {
		return err
	}

	fmt.Println("ID\tTYPE\tSCOPE\tPROJECT\tTITLE\tCREATED")
	for _, artifact := range artifacts {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", shortArtifactID(artifact.ID), artifact.Type, artifact.Scope, displayProject(artifact.Project), artifact.Title, artifact.CreatedAt)
	}
	return nil
}

func runContextShow(args []string) error {
	fs := flag.NewFlagSet("context show", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: yanzi context show <id>")
	}

	activeProject, err := loadActiveProject()
	if err != nil {
		return err
	}

	artifact, err := yanzilibrary.GetVisibleContextArtifact(fs.Arg(0), activeProject)
	if err != nil {
		return err
	}

	fmt.Printf("ID: %s\n", artifact.ID)
	fmt.Printf("Title: %s\n", artifact.Title)
	fmt.Printf("Type: %s\n", artifact.Type)
	fmt.Printf("Scope: %s\n", artifact.Scope)
	fmt.Printf("Project: %s\n", displayProject(artifact.Project))
	fmt.Printf("Created_At: %s\n", artifact.CreatedAt)
	if strings.TrimSpace(artifact.Metadata) != "" {
		fmt.Printf("Metadata: %s\n", artifact.Metadata)
	} else {
		fmt.Printf("Metadata: \n")
	}
	fmt.Println("--- Content ---")
	fmt.Println(artifact.Content)
	return nil
}

func shortArtifactID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}

func displayProject(project string) string {
	if strings.TrimSpace(project) == "" {
		return "-"
	}
	return project
}
