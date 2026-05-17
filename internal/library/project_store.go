package yanzilibrary

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/sqliteruntime"
)

// CreateProject creates a unique project record and returns the created project.
func CreateProject(name string, description string) (*Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("project name is required")
	}

	db, err := InitDB()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	ctx := context.Background()
	exists, err := projectExists(ctx, db, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("project already exists: %s", name)
	}

	createdAt := time.Now().UTC()
	createdAtText := createdAt.Format(time.RFC3339Nano)
	hash := hashProjectRecord(name, description, createdAtText)

	if _, err := sqliteruntime.ExecContext(
		ctx,
		db,
		ResolvedDBPath(),
		"create project",
		`INSERT INTO projects (name, description, created_at, prev_hash, hash) VALUES (?, ?, ?, ?, ?)`,
		name,
		description,
		createdAtText,
		nil,
		hash,
	); err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("project already exists: %s", name)
		}
		return nil, err
	}

	return &Project{
		Name:        name,
		Description: description,
		CreatedAt:   createdAt,
	}, nil
}

// ListProjects returns all projects ordered by creation time, oldest first.
func ListProjects() ([]Project, error) {
	db, err := InitDB()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	rows, err := db.QueryContext(context.Background(), `SELECT name, description, created_at FROM projects ORDER BY created_at ASC, name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]Project, 0)
	for rows.Next() {
		var project Project
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

// hashProjectRecord computes a deterministic hash for the projects table hash column.
func hashProjectRecord(name, description, createdAt string) string {
	sum := sha256.Sum256([]byte(name + "\n" + description + "\n" + createdAt))
	return hex.EncodeToString(sum[:])
}

// isUniqueViolation reports whether the database error is a uniqueness constraint violation.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}
