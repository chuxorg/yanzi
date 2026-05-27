package projectstate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chuxorg/yanzi/internal/config"
)

type stateFile struct {
	ActiveProject string `json:"active_project"`
}

// LoadActiveProject resolves the active project using the current binding and state-file precedence.
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

// SaveActiveProject persists the active project in the current user state directory.
func SaveActiveProject(name string) error {
	path, err := statePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	state := stateFile{ActiveProject: strings.TrimSpace(name)}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(dir, "state-*.json")
	if err != nil {
		return fmt.Errorf("create state file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("write state file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("close state file: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("persist state file: %w", err)
	}
	return nil
}

// WriteProjectBinding persists the current working-directory binding used by init and project workflows.
func WriteProjectBinding(name string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working dir: %w", err)
	}

	dir := filepath.Join(wd, ".yanzi")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create binding dir: %w", err)
	}
	path := filepath.Join(dir, "project")
	return os.WriteFile(path, []byte(strings.TrimSpace(name)+"\n"), 0o644)
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

	var state stateFile
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
