package store

import (
	"encoding/json"
	"testing"

	"github.com/chuxorg/yanzi/internal/core/model"
)

func TestFilterIntentsByMeta(t *testing.T) {
	intents := []model.IntentRecord{
		{
			ID:   "a",
			Meta: json.RawMessage(`{"env":"prod","owner":"alice"}`),
		},
		{
			ID:   "b",
			Meta: json.RawMessage(`{"env":"staging","owner":"bob"}`),
		},
		{
			ID:   "c",
			Meta: json.RawMessage(`{"env":"prod","count":2}`),
		},
	}

	filtered, err := FilterIntentsByMeta(intents, map[string]string{"env": "prod"})
	if err != nil {
		t.Fatalf("filter env=prod: %v", err)
	}
	if len(filtered) != 2 || filtered[0].ID != "a" || filtered[1].ID != "c" {
		t.Fatalf("unexpected filtered result: %+v", filtered)
	}

	filtered, err = FilterIntentsByMeta(intents, map[string]string{"env": "prod", "owner": "alice"})
	if err != nil {
		t.Fatalf("filter env+owner: %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != "a" {
		t.Fatalf("unexpected filtered result: %+v", filtered)
	}

	filtered, err = FilterIntentsByMeta(intents, map[string]string{"missing": "value"})
	if err != nil {
		t.Fatalf("filter missing key: %v", err)
	}
	if len(filtered) != 0 {
		t.Fatalf("expected empty filtered result, got %+v", filtered)
	}
}

func TestFilterIntentsByMetaEmptyFilters(t *testing.T) {
	intents := []model.IntentRecord{
		{ID: "a"},
		{ID: "b"},
	}

	filtered, err := FilterIntentsByMeta(intents, nil)
	if err != nil {
		t.Fatalf("filter nil: %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("expected full result, got %+v", filtered)
	}
}

func TestFilterIntentsByMetaInvalidJSON(t *testing.T) {
	intents := []model.IntentRecord{
		{ID: "a", Meta: json.RawMessage(`{"env":`)},
	}

	_, err := FilterIntentsByMeta(intents, map[string]string{"env": "prod"})
	if err == nil {
		t.Fatalf("expected error for invalid meta JSON")
	}
}
