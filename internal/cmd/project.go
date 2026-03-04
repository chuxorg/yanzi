package cmd

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

// RunProject handles project subcommands.
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
		db, err := openLocalDB(cfg)
		if err != nil {
			return err
		}
		defer db.Close()

		project, err := createProjectLocal(context.Background(), db, name)
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
		db, err := openLocalDB(cfg)
		if err != nil {
			return err
		}
		defer db.Close()

		projects, err := listProjectsLocal(context.Background(), db)
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
		db, err := openLocalDB(cfg)
		if err != nil {
			return err
		}
		defer db.Close()

		projects, err := listProjectsLocal(context.Background(), db)
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

func createProjectLocal(ctx context.Context, db *sql.DB, name string) (yanzilibrary.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return yanzilibrary.Project{}, errors.New("project name is required")
	}

	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(1) FROM projects WHERE name = ?`, name).Scan(&count); err != nil {
		return yanzilibrary.Project{}, err
	}
	if count > 0 {
		return yanzilibrary.Project{}, fmt.Errorf("project already exists: %s", name)
	}

	createdAt := time.Now().UTC()
	createdAtText := createdAt.Format(time.RFC3339Nano)
	description := ""
	hash := hashProjectRecord(name, description, createdAtText)
	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO projects (name, description, created_at, prev_hash, hash) VALUES (?, ?, ?, ?, ?)`,
		name,
		description,
		createdAtText,
		nil,
		hash,
	); err != nil {
		if isProjectUniqueViolation(err) {
			return yanzilibrary.Project{}, fmt.Errorf("project already exists: %s", name)
		}
		return yanzilibrary.Project{}, err
	}
	return yanzilibrary.Project{
		Name:        name,
		Description: description,
		CreatedAt:   createdAt,
	}, nil
}

func listProjectsLocal(ctx context.Context, db *sql.DB) ([]yanzilibrary.Project, error) {
	rows, err := db.QueryContext(ctx, `SELECT name, description, created_at FROM projects ORDER BY created_at ASC, name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]yanzilibrary.Project, 0)
	for rows.Next() {
		var project yanzilibrary.Project
		var description sql.NullString
		var createdAtText string
		if err := rows.Scan(&project.Name, &description, &createdAtText); err != nil {
			return nil, err
		}
		if description.Valid {
			project.Description = description.String
		}
		project.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAtText)
		if err != nil {
			return nil, fmt.Errorf("parse project created_at for %s: %w", project.Name, err)
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return projects, nil
}

func hashProjectRecord(name, description, createdAt string) string {
	sum := sha256.Sum256([]byte(name + "\n" + description + "\n" + createdAt))
	return hex.EncodeToString(sum[:])
}

func isProjectUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}
