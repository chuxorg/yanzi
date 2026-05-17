package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

var (
	intentTypeCatalog = []string{
		"change_request",
		"checkpoint",
		"decision",
		"note",
		"prompt",
		"task",
	}
	contextTypeCatalog = []string{
		"coding_standard",
		"note",
		"process_rule",
		"reference",
		"requirement",
	}
	contextTypeAliases = map[string]string{
		"governance": "process_rule",
	}
)

func normalizeContextType(value string) string {
	trimmed := strings.TrimSpace(value)
	if canonical, ok := contextTypeAliases[trimmed]; ok {
		return canonical
	}
	return trimmed
}

func artifactTypeCatalogJSON() ([]byte, error) {
	aliases := make(map[string]string, len(contextTypeAliases))
	for key, value := range contextTypeAliases {
		aliases[key] = value
	}
	sort.Strings(intentTypeCatalog)
	sort.Strings(contextTypeCatalog)
	payload := struct {
		SchemaVersion int               `json:"schema_version"`
		Kind          string            `json:"kind"`
		Intent        []string          `json:"intent"`
		Context       []string          `json:"context"`
		Aliases       map[string]string `json:"aliases"`
	}{
		SchemaVersion: machineContractSchemaVersion,
		Kind:          jsonKindArtifactTypes,
		Intent:        append([]string(nil), intentTypeCatalog...),
		Context:       append([]string(nil), contextTypeCatalog...),
		Aliases:       aliases,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode type catalog: %w", err)
	}
	return append(data, '\n'), nil
}
