package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chuxorg/yanzi/internal/config"
)

type apiProjectState struct {
	ActiveProject string `json:"active_project"`
}

func loadAPIActiveProject() (string, error) {
	project, err := loadAPIBoundProject()
	if err == nil && project != "" {
		return project, nil
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	path, err := apiStatePath()
	if err != nil {
		return "", err
	}

	project, err = readAPIActiveProject(path)
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
	project, err = readAPIActiveProject(fallback)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	return project, err
}

func loadAPIBoundProject() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working dir: %w", err)
	}
	return readAPIBoundProject(filepath.Join(wd, ".yanzi", "project"))
}

func readAPIBoundProject(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", os.ErrNotExist
		}
		return "", fmt.Errorf("read project binding: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func readAPIActiveProject(path string) (string, error) {
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

	var state apiProjectState
	if err := json.Unmarshal(data, &state); err != nil {
		return "", fmt.Errorf("invalid state file: %w", err)
	}
	return strings.TrimSpace(state.ActiveProject), nil
}

func apiStatePath() (string, error) {
	dir, err := config.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.json"), nil
}
