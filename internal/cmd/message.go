package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/chuxorg/yanzi/internal/client"
	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/core/model"
)

const messageResponse = "Message note"

func RunMessage(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: yanzi message <send|list|pull> [args]")
	}

	switch args[0] {
	case "send":
		return runMessageSend(args[1:])
	case "list":
		return runMessageList(args[1:])
	case "pull":
		return runMessagePull(args[1:])
	default:
		return errors.New("usage: yanzi message <send|list|pull> [args]")
	}
}

func runMessageSend(args []string) error {
	fs := flag.NewFlagSet("message send", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	to := fs.String("to", "", "message recipient")
	from := fs.String("from", "", "message sender")
	channel := fs.String("channel", "", "optional message channel")
	title := fs.String("title", "", "optional message title")
	filePath := fs.String("file", "", "message file")
	content := fs.String("content", "", "inline message content")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("usage: yanzi message send --to <name> --from <name> [--channel <name>] [--title <title>] [--file <path> | --content <text>]")
	}
	if strings.TrimSpace(*to) == "" {
		return errors.New("--to is required")
	}
	if strings.TrimSpace(*from) == "" {
		return errors.New("--from is required")
	}

	messageContent, err := resolveArtifactContent(*content, *filePath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(messageContent) == "" {
		return errors.New("message content is required")
	}

	captureArgs := []string{
		"--author", strings.TrimSpace(*from),
		"--source", "message",
		"--prompt", messageContent,
		"--response", messageResponse,
		"--meta", "type=note",
		"--meta", "subtype=message",
		"--meta", "to=" + strings.TrimSpace(*to),
		"--meta", "from=" + strings.TrimSpace(*from),
	}
	if trimmed := strings.TrimSpace(*title); trimmed != "" {
		captureArgs = append(captureArgs, "--title", trimmed)
	}
	if trimmed := strings.TrimSpace(*channel); trimmed != "" {
		captureArgs = append(captureArgs, "--meta", "channel="+trimmed)
	}
	return RunCapture(captureArgs)
}

func runMessageList(args []string) error {
	fs := flag.NewFlagSet("message list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	to := fs.String("to", "", "message recipient")
	from := fs.String("from", "", "message sender")
	channel := fs.String("channel", "", "optional message channel")
	includeDeleted := fs.Bool("include-deleted", false, "include tombstoned records")
	limit := fs.Int("limit", 20, "max records to return")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("usage: yanzi message list [--to <name>] [--from <name>] [--channel <name>] [--include-deleted] [--limit <n>]")
	}

	listArgs := []string{
		"--source", "message",
		"--meta", "type=note",
		"--meta", "subtype=message",
		"--limit", fmt.Sprintf("%d", *limit),
	}
	if trimmed := strings.TrimSpace(*to); trimmed != "" {
		listArgs = append(listArgs, "--meta", "to="+trimmed)
	}
	if trimmed := strings.TrimSpace(*from); trimmed != "" {
		listArgs = append(listArgs, "--meta", "from="+trimmed)
	}
	if trimmed := strings.TrimSpace(*channel); trimmed != "" {
		listArgs = append(listArgs, "--meta", "channel="+trimmed)
	}
	if *includeDeleted {
		listArgs = append(listArgs, "--include-deleted")
	}
	return RunList(listArgs)
}

func runMessagePull(args []string) error {
	fs := flag.NewFlagSet("message pull", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	to := fs.String("to", "", "message recipient")
	from := fs.String("from", "", "message sender")
	channel := fs.String("channel", "", "optional message channel")
	includeDeleted := fs.Bool("include-deleted", false, "include tombstoned records")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("usage: yanzi message pull [--to <name>] [--from <name>] [--channel <name>] [--include-deleted]")
	}

	metaFilters := map[string]string{
		"type":    "note",
		"subtype": "message",
	}
	if trimmed := strings.TrimSpace(*to); trimmed != "" {
		metaFilters["to"] = trimmed
	}
	if trimmed := strings.TrimSpace(*from); trimmed != "" {
		metaFilters["from"] = trimmed
	}
	if trimmed := strings.TrimSpace(*channel); trimmed != "" {
		metaFilters["channel"] = trimmed
	}

	intents, err := loadMessageIntents(metaFilters, *includeDeleted)
	if err != nil {
		return err
	}
	if len(intents) == 0 {
		fmt.Println("No messages found.")
		return nil
	}

	sort.SliceStable(intents, func(i, j int) bool {
		if intents[i].CreatedAt == intents[j].CreatedAt {
			return intents[i].ID < intents[j].ID
		}
		return intents[i].CreatedAt < intents[j].CreatedAt
	})

	fmt.Println("# Messages")
	fmt.Println()
	for idx, intent := range intents {
		if idx > 0 {
			fmt.Println()
			fmt.Println("---")
			fmt.Println()
		}
		meta, _ := decodeStringMeta(string(intent.Meta))
		fmt.Printf("## %s -> %s\n", meta["from"], meta["to"])
		if channelValue := strings.TrimSpace(meta["channel"]); channelValue != "" {
			fmt.Printf("Channel: %s\n", channelValue)
		}
		if title := strings.TrimSpace(intent.Title); title != "" {
			fmt.Printf("Title: %s\n", title)
		}
		fmt.Printf("Created: %s\n\n", intent.CreatedAt)
		fmt.Println(strings.TrimSpace(intent.Prompt))
	}
	return nil
}

func loadMessageIntents(metaFilters map[string]string, includeDeleted bool) ([]model.IntentRecord, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	switch cfg.Mode {
	case config.ModeHTTP:
		cli := client.New(cfg.BaseURL)
		resp, err := cli.ListIntents(context.Background(), "", "message", 200, metaFilters, includeDeleted)
		if err != nil {
			return nil, fmt.Errorf("http request to %s failed: %w", cfg.BaseURL, err)
		}
		return resp.Intents, nil
	case config.ModeLocal:
		db, err := openLocalDB(cfg)
		if err != nil {
			return nil, err
		}
		defer db.Close()
		return listLocalIntents(context.Background(), db, "", "message", 200, metaFilters, includeDeleted)
	default:
		return nil, fmt.Errorf("invalid mode: %s", cfg.Mode)
	}
}
