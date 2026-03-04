package cmd

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

func loadActiveProject() (string, error) {
	path, err := statePath()
	if err != nil {
		return "", err
	}

	project, err := readActiveProject(path)
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

func attachProjectMeta(meta json.RawMessage, project string) (json.RawMessage, error) {
	if strings.TrimSpace(project) == "" {
		return meta, nil
	}

	payload := map[string]string{}
	if len(meta) > 0 {
		if err := json.Unmarshal(meta, &payload); err != nil {
			return nil, fmt.Errorf("decode meta: %w", err)
		}
	}
	payload["project"] = strings.TrimSpace(project)

	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode meta: %w", err)
	}
	return json.RawMessage(encoded), nil
}

func saveActiveProject(name string) error {
	path, err := statePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	state := projectState{ActiveProject: strings.TrimSpace(name)}
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

func statePath() (string, error) {
	dir, err := config.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.json"), nil
}
