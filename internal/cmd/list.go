package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
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
	limit := fs.Int("limit", 20, "max records to return")
	metaFilters := metaPairs{}
	fs.Var(&metaFilters, "meta", "meta filter key=value (repeatable; exact match; AND)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var intents []client.IntentRecord
	switch cfg.Mode {
	case config.ModeHTTP:
		cli := client.New(cfg.BaseURL)
		resp, err := cli.ListIntents(context.Background(), *author, *source, *limit, map[string]string(metaFilters))
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

		localIntents, err := listLocalIntents(ctx, db, *author, *source, *limit, map[string]string(metaFilters))
		if err != nil {
			return err
		}
		intents = localIntents
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}

	fmt.Println("ID\tCreated_At\tAuthor\tSource\tTitle")
	for _, intent := range intents {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\n", intent.ID, intent.CreatedAt, intent.Author, intent.SourceType, intent.Title)
	}

	return nil
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
