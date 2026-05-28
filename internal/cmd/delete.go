package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

// RunDelete tombstones an intent or artifact by id.
func RunDelete(args []string) error {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	cascade := fs.Bool("cascade", false, "tombstone dependent chain records")
	force := fs.Bool("force", false, "allow tombstoning records referenced by checkpoints")
	if err := fs.Parse(normalizeDeleteArgs(args)); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: yanzi delete <id> [--cascade] [--force]")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.Mode != config.ModeLocal {
		return errors.New("delete is only available in local mode")
	}

	db, err := openLocalDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	writeStore := yanzilibrary.NewArtifactWriteStore(db)
	updatedIDs, err := writeStore.Tombstone(context.Background(), fs.Arg(0), *cascade, *force)
	if err != nil {
		return err
	}

	fmt.Printf("tombstoned: %s\n", updatedIDs[0])
	if len(updatedIDs) > 1 {
		fmt.Printf("cascade_count: %d\n", len(updatedIDs)-1)
	}
	return nil
}

func normalizeDeleteArgs(args []string) []string {
	flags := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			continue
		}
		positionals = append(positionals, arg)
	}
	return append(flags, positionals...)
}

// RunRestore removes tombstone metadata from an intent or artifact by id.
func RunRestore(args []string) error {
	fs := flag.NewFlagSet("restore", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: yanzi restore <id>")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.Mode != config.ModeLocal {
		return errors.New("restore is only available in local mode")
	}

	db, err := openLocalDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	writeStore := yanzilibrary.NewArtifactWriteStore(db)
	if err := writeStore.Restore(context.Background(), fs.Arg(0)); err != nil {
		return err
	}

	fmt.Printf("restored: %s\n", fs.Arg(0))
	return nil
}
