package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/config"
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
	_ = fs.String("api-key", "", "API key for HTTP mode authentication")
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
		captureArgs = append(captureArgs, "--profile", trimmed)
	}
	return RunCapture(captureArgs)
}

func runRulesList(args []string) error {
	fs := flag.NewFlagSet("rules list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	scope := fs.String("scope", "", "rule scope (global|project)")
	profile := fs.String("profile", "", "optional rule profile")
	includeDeleted := fs.Bool("include-deleted", false, "include tombstoned records")
	_ = fs.String("api-key", "", "API key for HTTP mode authentication")
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
		if trimmed != yanzilibrary.ContextScopeGlobal && trimmed != yanzilibrary.ContextScopeProject {
			return fmt.Errorf("invalid scope %q (expected global or project)", trimmed)
		}
		listArgs = append(listArgs, "--meta", "scope="+trimmed)
	}
	if trimmed := strings.TrimSpace(*profile); trimmed != "" {
		listArgs = append(listArgs, "--profile", trimmed)
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
	compose := fs.Bool("compose", false, "compose markdown or html output into system/profile sections")
	scope := fs.String("scope", "", "rule scope (global|project)")
	profile := fs.String("profile", "", "optional rule profile")
	includeDeleted := fs.Bool("include-deleted", false, "include tombstoned records")
	_ = fs.String("api-key", "", "API key for HTTP mode authentication")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("usage: yanzi rules export --format <markdown|json|html> [--scope <global|project>] [--profile <name>] [--compose]")
	}

	formatValue := strings.TrimSpace(*format)
	if *compose && (formatValue == "markdown" || formatValue == "html") {
		return runRulesComposeExport(cliVersion, formatValue, strings.TrimSpace(*scope), strings.TrimSpace(*profile), *includeDeleted)
	}

	exportArgs := []string{
		"--meta", "type=context",
		"--meta", "subtype=rules",
		"--format", formatValue,
	}
	if trimmed := strings.TrimSpace(*scope); trimmed != "" {
		if trimmed != yanzilibrary.ContextScopeGlobal && trimmed != yanzilibrary.ContextScopeProject {
			return fmt.Errorf("invalid scope %q (expected global or project)", trimmed)
		}
		exportArgs = append(exportArgs, "--meta", "scope="+trimmed)
	}
	if trimmed := strings.TrimSpace(*profile); trimmed != "" {
		exportArgs = append(exportArgs, "--profile", trimmed)
	}
	if *includeDeleted {
		exportArgs = append(exportArgs, "--include-deleted")
	}
	return RunExport(exportArgs, cliVersion)
}

func runRulesComposeExport(cliVersion, formatValue, scopeFilter, profileFilter string, includeDeleted bool) error {
	project, err := loadActiveProject()
	if err != nil {
		return err
	}
	if project == "" {
		return errors.New("no active project set")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.Mode != config.ModeLocal {
		return errors.New("export is only available in local mode")
	}

	provider, err := openLocalProvider(cfg)
	if err != nil {
		return err
	}
	defer func() {
		_ = provider.Close()
	}()

	items, _, err := loadExportItems(context.Background(), provider, project, map[string]string{
		"type":    "context",
		"subtype": "rules",
	}, includeDeleted)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	content := []byte(renderComposedRulesMarkdown(project, cliVersion, now, items, scopeFilter, profileFilter))
	path := filepath.Join(".", "YANZI_LOG.md")
	if formatValue == "html" {
		path = filepath.Join(".", "YANZI_LOG.html")
		content = []byte(renderComposedRulesHTML(project, cliVersion, now, items, scopeFilter, profileFilter))
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write export file: %w", err)
	}

	fmt.Printf("Exported %s\n", path)
	return nil
}

func renderComposedRulesMarkdown(project, cliVersion string, now time.Time, items []exportItem, scopeFilter, profileFilter string) string {
	systemRules := make([]string, 0)
	profileRules := make([]string, 0)

	for _, item := range items {
		if item.Kind != exportItemCapture {
			continue
		}

		scope := strings.TrimSpace(item.Metadata["scope"])
		profile := strings.TrimSpace(item.Metadata["profile"])

		if includeComposedSystemRule(scope, profile, scopeFilter) {
			systemRules = append(systemRules, strings.TrimRight(item.Prompt, "\n"))
			continue
		}
		if profileFilter != "" && includeComposedProfileRule(scope, profile, scopeFilter, profileFilter) {
			profileRules = append(profileRules, strings.TrimRight(item.Prompt, "\n"))
		}
	}

	var b strings.Builder
	b.WriteString("# Yanzi Rules Export\n\n")
	b.WriteString(fmt.Sprintf("Project: %s\n", project))
	b.WriteString(fmt.Sprintf("Exported: %s\n", now.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("Version: %s\n\n", cliVersion))

	b.WriteString("# SYSTEM RULES\n\n")
	if len(systemRules) == 0 {
		b.WriteString("No system rules.\n")
	} else {
		b.WriteString(strings.Join(systemRules, "\n\n---\n\n"))
		b.WriteString("\n")
	}

	if profileFilter != "" {
		b.WriteString("\n---\n\n")
		b.WriteString(fmt.Sprintf("# PROFILE: %s\n\n", profileFilter))
		if len(profileRules) == 0 {
			b.WriteString("No profile rules.\n")
		} else {
			b.WriteString(strings.Join(profileRules, "\n\n---\n\n"))
			b.WriteString("\n")
		}
	}

	return b.String()
}

type htmlRuleSection struct {
	Title string
	Items []exportItem
}

func renderComposedRulesHTML(project, cliVersion string, now time.Time, items []exportItem, scopeFilter, profileFilter string) string {
	sections := make([]htmlRuleSection, 0, 2)
	systemRules := make([]exportItem, 0)
	profileRules := make([]exportItem, 0)

	for _, item := range items {
		if item.Kind != exportItemCapture {
			continue
		}

		scope := strings.TrimSpace(item.Metadata["scope"])
		profile := strings.TrimSpace(item.Metadata["profile"])

		if includeComposedSystemRule(scope, profile, scopeFilter) {
			systemRules = append(systemRules, item)
			continue
		}
		if profileFilter != "" && includeComposedProfileRule(scope, profile, scopeFilter, profileFilter) {
			profileRules = append(profileRules, item)
		}
	}

	sections = append(sections, htmlRuleSection{
		Title: "SYSTEM RULES",
		Items: systemRules,
	})
	if profileFilter != "" {
		sections = append(sections, htmlRuleSection{
			Title: fmt.Sprintf("PROFILE: %s", profileFilter),
			Items: profileRules,
		})
	}

	return renderHTMLLog(project, cliVersion, now, flattenHTMLRuleSections(sections), sections...)
}

func flattenHTMLRuleSections(sections []htmlRuleSection) []exportItem {
	items := make([]exportItem, 0)
	for _, section := range sections {
		items = append(items, section.Items...)
	}
	return items
}

func includeComposedSystemRule(scope, profile, scopeFilter string) bool {
	if scopeFilter != "" && scope != scopeFilter {
		return false
	}
	return scope == yanzilibrary.ContextScopeGlobal || profile == ""
}

func includeComposedProfileRule(scope, profile, scopeFilter, profileFilter string) bool {
	if profile != profileFilter {
		return false
	}
	if scopeFilter != "" {
		return scope == scopeFilter
	}
	return scope == yanzilibrary.ContextScopeProject
}
