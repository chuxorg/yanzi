package yanzilibrary

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chuxorg/yanzi/internal/config"
)

type projectState struct {
	ActiveProject string `json:"active_project"`
}

// LoadActiveProject resolves the current active project using the same local-state
// lookup order as the CLI.
func LoadActiveProject() (string, error) {
	project, err := loadBoundProject()
	if err == nil && project != "" {
		return project, nil
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	path, err := statePath()
	if err != nil {
		return "", err
	}

	project, err = readActiveProject(path)
	if err == nil {
		return project, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working dir: %w", err)
	}
	fallback := filepath.Join(wd, ".yanzi", "state.json")
	project, err = readActiveProject(fallback)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	return project, err
}

// StatePath resolves the canonical state.json path used for active project persistence.
func StatePath() (string, error) {
	dir, err := config.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.json"), nil
}

func loadBoundProject() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working dir: %w", err)
	}
	return readBoundProject(filepath.Join(wd, ".yanzi", "project"))
}

func readBoundProject(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", os.ErrNotExist
		}
		return "", fmt.Errorf("read project binding: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func readActiveProject(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", os.ErrNotExist
		}
		return "", fmt.Errorf("read state file: %w", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return "", nil
	}

	var state projectState
	if err := json.Unmarshal(data, &state); err != nil {
		return "", fmt.Errorf("invalid state file: %w", err)
	}
	return strings.TrimSpace(state.ActiveProject), nil
}

func statePath() (string, error) {
	dir, err := config.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.json"), nil
}
