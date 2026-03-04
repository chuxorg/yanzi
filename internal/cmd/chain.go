package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/chuxorg/yanzi/internal/client"
	"github.com/chuxorg/yanzi/internal/config"
)

// RunChain prints the intent chain from oldest to newest.
func RunChain(args []string) error {
	fs := flag.NewFlagSet("chain", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: yanzi chain <intent-id>")
	}

	id := fs.Arg(0)
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	var resp chainResult
	switch cfg.Mode {
	case config.ModeHTTP:
		cli := client.New(cfg.BaseURL)
		httpResp, err := cli.ChainIntent(context.Background(), id)
		if err != nil {
			return fmt.Errorf("http request to %s failed: %w", cfg.BaseURL, err)
		}
		resp = chainResult{
			HeadID:       httpResp.HeadID,
			Length:       httpResp.Length,
			Intents:      httpResp.Intents,
			MissingLinks: httpResp.MissingLinks,
		}
	case config.ModeLocal:
		ctx := context.Background()
		db, err := openLocalDB(cfg)
		if err != nil {
			return err
		}
		defer db.Close()

		localResp, err := chainLocalIntent(ctx, db, id)
		if err != nil {
			return err
		}
		resp = localResp
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}

	fmt.Printf("chain head: %s\n", resp.HeadID)
	for i, intent := range resp.Intents {
		fmt.Printf("%d\t%s\t%s\t%s\t%s\n", i+1, intent.CreatedAt, intent.Title, intent.Author, intent.Hash)
	}
	if len(resp.MissingLinks) > 0 {
		fmt.Printf("missing_links: %s\n", joinComma(resp.MissingLinks))
	}

	return nil
}

func joinComma(values []string) string {
	if len(values) == 0 {
		return ""
	}
	out := values[0]
	for i := 1; i < len(values); i++ {
		out += "," + values[i]
	}
	return out
}
