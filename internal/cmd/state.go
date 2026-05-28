package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

type projectState struct {
	ActiveProject string `json:"active_project"`
}

func loadActiveProject() (string, error) {
	return yanzilibrary.LoadActiveProject()
}

func writeProjectBinding(name string) error {
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
	path, err := yanzilibrary.StatePath()
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
