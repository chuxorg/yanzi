package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/chuxorg/yanzi/internal/client"
	"github.com/chuxorg/yanzi/internal/config"
)

// RunList lists intent records.
func RunList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	author := fs.String("author", "", "author filter")
	source := fs.String("source", "", "source filter")
	profile := fs.String("profile", "", "profile filter")
	limit := fs.Int("limit", 20, "max records to return")
	allProjects := fs.Bool("all-projects", false, "list records across every project")
	includeDeleted := fs.Bool("include-deleted", false, "include tombstoned records")
	apiKey := fs.String("api-key", "", "API key for HTTP mode authentication")
	metaFilters := metaPairs{}
	fs.Var(&metaFilters, "meta", "meta filter key=value (repeatable; exact match; AND)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*profile) != "" {
		metaFilters["profile"] = strings.TrimSpace(*profile)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var intents []client.IntentRecord
	scopeLabel := "All projects"
	switch cfg.Mode {
	case config.ModeHTTP:
		cli := client.New(cfg.BaseURL, client.ResolveAuthHeader(cfg, *apiKey))
		resp, err := cli.ListIntents(context.Background(), *author, *source, *limit, map[string]string(metaFilters), *includeDeleted)
		if err != nil {
			return fmt.Errorf("http request to %s failed: %w", cfg.BaseURL, err)
		}
		intents = resp.Intents
	case config.ModeLocal:
		ctx := context.Background()
		db, err := openLocalDB(cfg)
		if err != nil {
			return err
		}
		defer db.Close()

		activeProject, err := loadActiveProject()
		if err != nil {
			return err
		}
		if !*allProjects {
			activeProject = strings.TrimSpace(activeProject)
			if activeProject == "" {
				return fmt.Errorf("no active project set")
			}
			if explicitProject, ok := metaFilters["project"]; ok && strings.TrimSpace(explicitProject) != activeProject {
				return fmt.Errorf("--meta project=%s conflicts with active project %s; use --all-projects for cross-project retrieval", strings.TrimSpace(explicitProject), activeProject)
			}
			metaFilters["project"] = activeProject
			scopeLabel = activeProject
		}

		localIntents, err := listLocalIntents(ctx, db, *author, *source, *limit, map[string]string(metaFilters), *includeDeleted)
		if err != nil {
			return err
		}
		intents = localIntents
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}

	sort.SliceStable(intents, func(i, j int) bool {
		if intents[i].CreatedAt == intents[j].CreatedAt {
			return intents[i].ID > intents[j].ID
		}
		return intents[i].CreatedAt > intents[j].CreatedAt
	})

	fmt.Printf("Project: %s\n\n", scopeLabel)
	if *allProjects {
		fmt.Println("ID\tCreated_At\tProject\tAuthor\tSource\tTitle\tMetadata")
	} else {
		fmt.Println("ID\tCreated_At\tAuthor\tSource\tTitle\tMetadata")
	}
	if len(intents) == 0 {
		fmt.Println("(none)")
		return nil
	}
	for _, intent := range intents {
		metadata := formatListMetadata(intent.Meta)
		if *allProjects {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", intent.ID, intent.CreatedAt, fallbackProject(intent.Meta), intent.Author, intent.SourceType, intent.Title, metadata)
			continue
		}
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", intent.ID, intent.CreatedAt, intent.Author, intent.SourceType, intent.Title, metadata)
	}

	return nil
}

func formatListMetadata(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	meta, err := decodeStringMeta(string(raw))
	if err != nil {
		return ""
	}
	meta = exportableMetadata(meta)
	if len(meta) == 0 {
		return ""
	}
	parts := make([]string, 0, len(meta))
	for _, key := range sortedMetaKeys(meta) {
		parts = append(parts, key+"="+meta[key])
	}
	return strings.Join(parts, "; ")
}

func fallbackProject(raw []byte) string {
	meta, err := decodeStringMeta(string(raw))
	if err != nil {
		return "-"
	}
	project := strings.TrimSpace(meta["project"])
	if project == "" {
		return "-"
	}
	return project
}

type metaPairs map[string]string

func (m *metaPairs) String() string {
	if m == nil || len(*m) == 0 {
		return ""
	}
	parts := make([]string, 0, len(*m))
	for key, value := range *m {
		parts = append(parts, key+"="+value)
	}
	return strings.Join(parts, ",")
}

func (m *metaPairs) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 || parts[0] == "" {
		return fmt.Errorf("invalid meta filter: %s (expected key=value)", value)
	}
	if *m == nil {
		*m = make(map[string]string)
	}
	(*m)[parts[0]] = parts[1]
	return nil
}
