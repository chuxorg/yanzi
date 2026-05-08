package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

// RunCheckpoint manages checkpoint creation and listing for the active project.
//
// Problem:
// Reloading a project from the beginning is unnecessary when a stable boundary
// already exists.
//
// Solution:
// RunCheckpoint provides `create` and `list` subcommands for append-only
// project checkpoint records.
//
// Arguments:
//
//	args starts with `create` or `list` followed by that subcommand's flags.
//
// Example:
//
//	yanzi checkpoint create --summary "Initial project state"
func RunCheckpoint(args []string) error {
	if len(args) == 0 {
		return checkpointUsageError()
	}

	switch args[0] {
	case "create":
		return runCheckpointCreate(args[1:])
	case "list":
		return runCheckpointList(args[1:])
	default:
		return checkpointUsageError()
	}
}

func runCheckpointCreate(args []string) error {
	fs := flag.NewFlagSet("checkpoint create", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	summary := fs.String("summary", "", "checkpoint summary")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("usage: yanzi checkpoint create --summary \"...\"")
	}
	if *summary == "" {
		return errors.New("summary is required")
	}

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

	switch cfg.Mode {
	case config.ModeLocal:
		ctx := context.Background()
		db, err := openLocalDB(cfg)
		if err != nil {
			return err
		}
		defer db.Close()

		checkpoint, err := yanzilibrary.CreateCheckpoint(ctx, db, project, *summary, []string{})
		if err != nil {
			return err
		}

		fmt.Printf("id: %s\n", checkpoint.Hash)
		fmt.Printf("summary: %s\n", checkpoint.Summary)
		return nil
	case config.ModeHTTP:
		return errors.New("checkpoint commands are not available in http mode")
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}
}

func runCheckpointList(args []string) error {
	fs := flag.NewFlagSet("checkpoint list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("usage: yanzi checkpoint list")
	}

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

	switch cfg.Mode {
	case config.ModeLocal:
		ctx := context.Background()
		db, err := openLocalDB(cfg)
		if err != nil {
			return err
		}
		defer db.Close()

		checkpoints, err := yanzilibrary.ListCheckpoints(ctx, db, project)
		if err != nil {
			return err
		}

		fmt.Println("Index\tCreatedAt\tSummary")
		for i, checkpoint := range checkpoints {
			fmt.Printf("%d\t%s\t%s\n", i+1, checkpoint.CreatedAt, checkpoint.Summary)
		}
		return nil
	case config.ModeHTTP:
		return errors.New("checkpoint commands are not available in http mode")
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}
}

func checkpointUsageError() error {
	return errors.New("usage: yanzi checkpoint <create|list>")
}
