package yanzilibrary

import (
	"context"
	"errors"
	"strings"

	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/storage"
	"github.com/chuxorg/yanzi/internal/storage/registry"
)

// CreateProject creates a unique project record and returns the created project.
func CreateProject(name string, description string) (*Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("project name is required")
	}

	provider, err := openProjectProvider()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = provider.Close()
	}()

	project, err := provider.CreateProject(context.Background(), storage.CreateProjectInput{Name: name, Description: description})
	if err != nil {
		return nil, err
	}
	return &Project{
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt,
	}, nil
}

// ListProjects returns all projects ordered by creation time, oldest first.
func ListProjects() ([]Project, error) {
	provider, err := openProjectProvider()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = provider.Close()
	}()

	records, err := provider.ListProjects(context.Background())
	if err != nil {
		return nil, err
	}

	projects := make([]Project, 0, len(records))
	for _, record := range records {
		projects = append(projects, Project{
			Name:        record.Name,
			Description: record.Description,
			CreatedAt:   record.CreatedAt,
		})
	}
	return projects, nil
}

func openProjectProvider() (storage.Provider, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	provider, err := registry.Open(context.Background(), cfg, registry.Options{Migrations: MigrationsFS()})
	if err != nil {
		return nil, err
	}
	if health := provider.Health(context.Background()); health.Path != "" {
		setResolvedDBPath(health.Path)
	}
	return provider, nil
}

// isUniqueViolation reports whether the database error is a uniqueness constraint violation.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}
