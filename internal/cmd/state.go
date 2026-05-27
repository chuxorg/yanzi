package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chuxorg/yanzi/internal/projectstate"
)

func loadActiveProject() (string, error) {
	return projectstate.LoadActiveProject()
}

func writeProjectBinding(name string) error {
	return projectstate.WriteProjectBinding(name)
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
	return projectstate.SaveActiveProject(name)
}
