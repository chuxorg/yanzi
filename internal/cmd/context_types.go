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
	payload := map[string]any{
		"intent":  intentTypeCatalog,
		"context": contextTypeCatalog,
		"aliases": aliases,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode type catalog: %w", err)
	}
	return append(data, '\n'), nil
}
