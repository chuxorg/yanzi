package cmd

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

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
	allProjects := fs.Bool("all-projects", false, "list checkpoints across every project")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("usage: yanzi checkpoint list [--all-projects]")
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

		if *allProjects {
			fmt.Println("Project: All projects")
			fmt.Println()
			checkpoints, err := listAllCheckpointsLocal(ctx, db)
			if err != nil {
				return err
			}
			fmt.Println("Project\tCreatedAt\tSummary")
			if len(checkpoints) == 0 {
				fmt.Println("(none)")
				return nil
			}
			for _, checkpoint := range checkpoints {
				fmt.Printf("%s\t%s\t%s\n", checkpoint.Project, checkpoint.CreatedAt, checkpoint.Summary)
			}
			return nil
		}

		project, err := loadActiveProject()
		if err != nil {
			return err
		}
		project = strings.TrimSpace(project)
		if project == "" {
			return errors.New("no active project set")
		}
		fmt.Printf("Project: %s\n\n", project)

		checkpoints, err := yanzilibrary.ListCheckpoints(ctx, db, project)
		if err != nil {
			return err
		}

		fmt.Println("CreatedAt\tSummary")
		if len(checkpoints) == 0 {
			fmt.Println("(none)")
			return nil
		}
		for _, checkpoint := range checkpoints {
			fmt.Printf("%s\t%s\n", checkpoint.CreatedAt, checkpoint.Summary)
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

func listAllCheckpointsLocal(ctx context.Context, db *sql.DB) ([]yanzilibrary.Checkpoint, error) {
	return yanzilibrary.ListAllCheckpoints(ctx, db)
}
