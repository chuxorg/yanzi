package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

const rulesResponse = "Canonical rules"

func RunRules(args []string, cliVersion string) error {
	if len(args) == 0 {
		return errors.New("usage: yanzi rules <add|list|export> [args]")
	}

	switch args[0] {
	case "add":
		return runRulesAdd(args[1:])
	case "list":
		return runRulesList(args[1:])
	case "export":
		return runRulesExport(args[1:], cliVersion)
	default:
		return errors.New("usage: yanzi rules <add|list|export> [args]")
	}
}

func runRulesAdd(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: yanzi rules add <file> [--scope <global|project>] [--priority <value>] [--profile <name>]")
	}

	filePath := ""
	parseArgs := args
	if !strings.HasPrefix(args[0], "-") {
		filePath = args[0]
		parseArgs = args[1:]
	}

	fs := flag.NewFlagSet("rules add", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	scope := fs.String("scope", yanzilibrary.ContextScopeGlobal, "rule scope (global|project)")
	priority := fs.String("priority", "", "rule priority")
	profile := fs.String("profile", "", "optional rule profile")
	if err := fs.Parse(parseArgs); err != nil {
		return err
	}
	if filePath == "" {
		if fs.NArg() != 1 {
			return errors.New("usage: yanzi rules add <file> [--scope <global|project>] [--priority <value>] [--profile <name>]")
		}
		filePath = fs.Arg(0)
	} else if fs.NArg() != 0 {
		return errors.New("usage: yanzi rules add <file> [--scope <global|project>] [--priority <value>] [--profile <name>]")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read rules file: %w", err)
	}
	if len(strings.TrimSpace(string(content))) == 0 {
		return errors.New("rules file is empty")
	}

	scopeValue := strings.TrimSpace(*scope)
	if scopeValue != yanzilibrary.ContextScopeGlobal && scopeValue != yanzilibrary.ContextScopeProject {
		return fmt.Errorf("invalid scope %q (expected global or project)", scopeValue)
	}
	if scopeValue == yanzilibrary.ContextScopeProject {
		activeProject, err := loadActiveProject()
		if err != nil {
			return err
		}
		if strings.TrimSpace(activeProject) == "" {
			return errors.New("project-scoped rules require an active project")
		}
	}

	captureArgs := []string{
		"--author", "human",
		"--title", filepath.Base(filePath),
		"--prompt", string(content),
		"--response", rulesResponse,
		"--meta", "type=context",
		"--meta", "subtype=rules",
		"--meta", "scope=" + scopeValue,
	}
	if trimmed := strings.TrimSpace(*priority); trimmed != "" {
		captureArgs = append(captureArgs, "--meta", "priority="+trimmed)
	}
	if trimmed := strings.TrimSpace(*profile); trimmed != "" {
		captureArgs = append(captureArgs, "--meta", "profile="+trimmed)
	}
	return RunCapture(captureArgs)
}

func runRulesList(args []string) error {
	fs := flag.NewFlagSet("rules list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	scope := fs.String("scope", "", "rule scope (global|project)")
	profile := fs.String("profile", "", "optional rule profile")
	includeDeleted := fs.Bool("include-deleted", false, "include tombstoned records")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("usage: yanzi rules list [--scope <global|project>] [--profile <name>]")
	}

	listArgs := []string{
		"--meta", "type=context",
		"--meta", "subtype=rules",
	}
	if trimmed := strings.TrimSpace(*scope); trimmed != "" {
		listArgs = append(listArgs, "--meta", "scope="+trimmed)
	}
	if trimmed := strings.TrimSpace(*profile); trimmed != "" {
		listArgs = append(listArgs, "--meta", "profile="+trimmed)
	}
	if *includeDeleted {
		listArgs = append(listArgs, "--include-deleted")
	}
	return RunList(listArgs)
}

func runRulesExport(args []string, cliVersion string) error {
	fs := flag.NewFlagSet("rules export", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	format := fs.String("format", "", "export format (required: markdown|json|html)")
	scope := fs.String("scope", "", "rule scope (global|project)")
	profile := fs.String("profile", "", "optional rule profile")
	includeDeleted := fs.Bool("include-deleted", false, "include tombstoned records")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("usage: yanzi rules export --format <markdown|json|html> [--scope <global|project>] [--profile <name>]")
	}

	exportArgs := []string{
		"--meta", "type=context",
		"--meta", "subtype=rules",
		"--format", strings.TrimSpace(*format),
	}
	if trimmed := strings.TrimSpace(*scope); trimmed != "" {
		exportArgs = append(exportArgs, "--meta", "scope="+trimmed)
	}
	if trimmed := strings.TrimSpace(*profile); trimmed != "" {
		exportArgs = append(exportArgs, "--meta", "profile="+trimmed)
	}
	if *includeDeleted {
		exportArgs = append(exportArgs, "--include-deleted")
	}
	return RunExport(exportArgs, cliVersion)
}
