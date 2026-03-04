package store

import (
	"encoding/json"
	"fmt"

	"github.com/chuxorg/yanzi/internal/core/model"
)

// FilterIntentsByMeta returns intents that match all meta filters (AND semantics).
func FilterIntentsByMeta(intents []model.IntentRecord, filters map[string]string) ([]model.IntentRecord, error) {
	if len(filters) == 0 {
		return intents, nil
	}

	filtered := make([]model.IntentRecord, 0, len(intents))
	for _, intent := range intents {
		match, err := matchesMetaFilters(intent.Meta, filters)
		if err != nil {
			return nil, err
		}
		if match {
			filtered = append(filtered, intent)
		}
	}

	return filtered, nil
}

func matchesMetaFilters(raw json.RawMessage, filters map[string]string) (bool, error) {
	if len(filters) == 0 {
		return true, nil
	}
	if len(raw) == 0 {
		return false, nil
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return false, fmt.Errorf("decode meta: %w", err)
	}

	meta := make(map[string]string, len(payload))
	for key, value := range payload {
		if s, ok := value.(string); ok {
			meta[key] = s
		}
	}

	for key, want := range filters {
		have, ok := meta[key]
		if !ok || have != want {
			return false, nil
		}
	}

	return true, nil
}
