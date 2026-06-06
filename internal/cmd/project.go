package cmd

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

// RunProject manages project creation, selection, and inspection.
//
// Arguments:
//
//	args starts with `create`, `use`, `current`, or `list`.
//
// Example:
//
//	yanzi project use demo
func RunProject(args []string) error {
	if len(args) == 0 {
		return projectUsageError()
	}

	switch args[0] {
	case "create":
		return runProjectCreate(args[1:])
	case "list":
		return runProjectList(args[1:])
	case "use":
		return runProjectUse(args[1:])
	case "current":
		return runProjectCurrent(args[1:])
	default:
		return projectUsageError()
	}
}

func runProjectCreate(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: yanzi project create <name>")
	}
	name := strings.TrimSpace(args[0])
	if name == "" {
		return errors.New("usage: yanzi project create <name>")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch cfg.Mode {
	case config.ModeLocal:
		project, err := yanzilibrary.CreateProject(name, "")
		if err != nil {
			return err
		}

		fmt.Printf("Project created: %s\n", project.Name)
		return nil
	case config.ModeHTTP:
		return errors.New("project commands are not available in http mode")
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}
}

func runProjectList(args []string) error {
	fs := flag.NewFlagSet("project list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	_ = fs.String("api-key", "", "API key for HTTP mode authentication")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("usage: yanzi project list")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch cfg.Mode {
	case config.ModeLocal:
		projects, err := yanzilibrary.ListProjects()
		if err != nil {
			return err
		}

		fmt.Println("Name\tCreatedAt\tDescription")
		for _, project := range projects {
			fmt.Printf("%s\t%s\t%s\n", project.Name, project.CreatedAt.Format(time.RFC3339Nano), project.Description)
		}
		return nil
	case config.ModeHTTP:
		return errors.New("project commands are not available in http mode")
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}
}

func runProjectUse(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: yanzi project use <name>")
	}
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch cfg.Mode {
	case config.ModeLocal:
		projects, err := yanzilibrary.ListProjects()
		if err != nil {
			return err
		}

		found := false
		for _, project := range projects {
			if project.Name == name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("project not found: %s", name)
		}

		if err := saveActiveProject(name); err != nil {
			return err
		}

		fmt.Printf("Active project set to %s.\n", name)
		return nil
	case config.ModeHTTP:
		return errors.New("project commands are not available in http mode")
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}
}

func runProjectCurrent(args []string) error {
	if len(args) != 0 {
		return errors.New("usage: yanzi project current")
	}

	active, err := loadActiveProject()
	if err != nil {
		return err
	}
	if active == "" {
		fmt.Println("No active project")
		return nil
	}

	fmt.Printf("Active project: %s\n", active)
	return nil
}

func projectUsageError() error {
	return errors.New("usage: yanzi project <create|list|use|current>")
}

func createProjectLocal(_ context.Context, _ *sql.DB, name string) (yanzilibrary.Project, error) {
	project, err := yanzilibrary.CreateProject(name, "")
	if err != nil {
		return yanzilibrary.Project{}, err
	}
	return *project, nil
}

func listProjectsLocal(_ context.Context, _ *sql.DB) ([]yanzilibrary.Project, error) {
	return yanzilibrary.ListProjects()
}
