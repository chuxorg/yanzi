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

// RunShow prints full intent details by id.
func RunShow(args []string) error {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: yanzi show <intent-id>")
	}

	id := fs.Arg(0)
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var intent client.IntentRecord
	switch cfg.Mode {
	case config.ModeHTTP:
		cli := client.New(cfg.BaseURL)
		record, err := cli.GetIntent(context.Background(), id)
		if err != nil {
			if isNotFoundError(err) {
				return fmt.Errorf("Intent not found for ID %s", id)
			}
			return fmt.Errorf("http request to %s failed: %w", cfg.BaseURL, err)
		}
		intent = record
	case config.ModeLocal:
		ctx := context.Background()
		db, err := openLocalDB(cfg)
		if err != nil {
			return err
		}
		defer db.Close()

		record, err := getLocalIntent(ctx, db, id)
		if err != nil {
			return err
		}
		intent = record
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}

	fmt.Printf("ID: %s\n", intent.ID)
	fmt.Printf("Created_At: %s\n", intent.CreatedAt)
	fmt.Printf("Author: %s\n", intent.Author)
	fmt.Printf("Source: %s\n", intent.SourceType)
	fmt.Printf("Title: %s\n", intent.Title)
	fmt.Printf("Prev_Hash: %s\n", intent.PrevHash)
	fmt.Printf("Hash: %s\n", intent.Hash)
	if len(intent.Meta) > 0 {
		fmt.Printf("Meta: %s\n", string(intent.Meta))
	} else {
		fmt.Printf("Meta: \n")
	}
	fmt.Println("--- Prompt ---")
	fmt.Println(intent.Prompt)
	fmt.Println("--- Response ---")
	fmt.Println(intent.Response)

	return nil
}

func isNotFoundError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "404") || strings.Contains(msg, "not found")
}
