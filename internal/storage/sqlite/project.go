package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/storage"
)

// CreateProject creates a project using current SQLite project semantics.
func (p *Provider) CreateProject(ctx context.Context, input storage.CreateProjectInput) (storage.Project, error) {
	if p == nil || p.db == nil {
		return storage.Project{}, storage.ErrProviderUnavailable
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return storage.Project{}, errors.New("project name is required")
	}
	description := input.Description

	exists, err := p.ProjectExists(ctx, name)
	if err != nil {
		return storage.Project{}, err
	}
	if exists {
		return storage.Project{}, fmt.Errorf("project already exists: %s", name)
	}

	createdAt := time.Now().UTC()
	createdAtText := createdAt.Format(time.RFC3339Nano)
	hash := hashProjectRecord(name, description, createdAtText)
	if _, err := p.db.ExecContext(
		ctx,
		`INSERT INTO projects (name, description, created_at, prev_hash, hash) VALUES (?, ?, ?, ?, ?)`,
		name,
		description,
		createdAtText,
		nil,
		hash,
	); err != nil {
		if isUniqueViolation(err) {
			return storage.Project{}, fmt.Errorf("project already exists: %s", name)
		}
		return storage.Project{}, err
	}
	return storage.Project{Name: name, Description: description, CreatedAt: createdAt}, nil
}

// ListProjects returns projects ordered by creation time, oldest first.
func (p *Provider) ListProjects(ctx context.Context) ([]storage.Project, error) {
	if p == nil || p.db == nil {
		return nil, storage.ErrProviderUnavailable
	}
	rows, err := p.db.QueryContext(ctx, `SELECT name, description, created_at FROM projects ORDER BY created_at ASC, name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]storage.Project, 0)
	for rows.Next() {
		var project storage.Project
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

// ProjectExists checks whether a project row exists for the provided name.
func (p *Provider) ProjectExists(ctx context.Context, name string) (bool, error) {
	if p == nil || p.db == nil {
		return false, storage.ErrProviderUnavailable
	}
	var count int
	if err := p.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM projects WHERE name = ?`, name).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func hashProjectRecord(name, description, createdAt string) string {
	sum := sha256.Sum256([]byte(name + "\n" + description + "\n" + createdAt))
	return hex.EncodeToString(sum[:])
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}
