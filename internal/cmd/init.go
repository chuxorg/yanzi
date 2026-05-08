package cmd

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chuxorg/yanzi/internal/config"
)

// RunInit creates or reuses a project and binds it to the current directory.
//
// Problem:
// Local project setup is error-prone when project creation, activation, and
// directory binding are done as separate steps.
//
// Solution:
// RunInit ensures a project exists, writes .yanzi/project in the current
// directory, and makes that project active.
//
// Arguments:
//
//	args may contain one optional project name; otherwise the current directory
//	name is used.
//
// Example:
//
//	yanzi init demo
func RunInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 1 {
		return errors.New("usage: yanzi init [name]")
	}

	name := ""
	if fs.NArg() == 1 {
		name = strings.TrimSpace(fs.Arg(0))
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("resolve working dir: %w", err)
		}
		name = strings.TrimSpace(filepath.Base(wd))
	}
	if name == "" || name == "." || name == string(filepath.Separator) {
		return errors.New("project name is required")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.Mode != config.ModeLocal {
		return errors.New("init is only available in local mode")
	}

	db, err := openLocalDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	created, err := ensureProjectExists(context.Background(), db, name)
	if err != nil {
		return err
	}
	if err := writeProjectBinding(name); err != nil {
		return err
	}
	if err := saveActiveProject(name); err != nil {
		return err
	}

	if created {
		fmt.Printf("Project created: %s\n", name)
	} else {
		fmt.Printf("Project exists: %s\n", name)
	}
	fmt.Printf("Project bound: %s\n", filepath.Join(".yanzi", "project"))
	return nil
}

func ensureProjectExists(ctx context.Context, db *sql.DB, name string) (bool, error) {
	projects, err := listProjectsLocal(ctx, db)
	if err != nil {
		return false, err
	}
	for _, project := range projects {
		if project.Name == name {
			return false, nil
		}
	}
	if _, err := createProjectLocal(ctx, db, name); err != nil {
		return false, err
	}
	return true, nil
}
