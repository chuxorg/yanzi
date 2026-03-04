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

// RunVerify verifies the stored hash for a given intent id.
func RunVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: yanzi verify <intent-id>")
	}

	id := fs.Arg(0)
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	var resp verifyResult
	switch cfg.Mode {
	case config.ModeHTTP:
		cli := client.New(cfg.BaseURL)
		httpResp, err := cli.VerifyIntent(context.Background(), id)
		if err != nil {
			return fmt.Errorf("http request to %s failed: %w", cfg.BaseURL, err)
		}
		resp = verifyResult{
			ID:           httpResp.ID,
			Valid:        httpResp.Valid,
			StoredHash:   httpResp.StoredHash,
			ComputedHash: httpResp.ComputedHash,
			PrevHash:     httpResp.PrevHash,
			Error:        httpResp.Error,
		}
	case config.ModeLocal:
		ctx := context.Background()
		db, err := openLocalDB(cfg)
		if err != nil {
			return err
		}
		defer db.Close()

		localResp, err := verifyLocalIntent(ctx, db, id)
		if err != nil {
			return err
		}
		resp = localResp
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}

	status := "✖ INVALID"
	if resp.Valid {
		status = "✔ VALID"
	}
	fmt.Println(status)
	fmt.Printf("stored_hash: %s\n", resp.StoredHash)
	fmt.Printf("computed_hash: %s\n", resp.ComputedHash)
	if resp.Error != nil {
		fmt.Printf("error: %s\n", *resp.Error)
	}

	return nil
}
