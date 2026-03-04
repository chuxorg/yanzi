// Package hash provides deterministic intent hashing logic.
package hash

import (
	"encoding/json"
	"testing"

	"github.com/chuxorg/yanzi/internal/core/model"
)

func TestHashIntentCanonicalization(t *testing.T) {
	base := model.IntentRecord{
		ID:         "01HZYFQ7T9ZV54X2G4A8M4J2C1",
		CreatedAt:  "2026-02-09T10:00:00Z",
		Author:     "alice",
		SourceType: "cli",
		Title:      "",
		Prompt:     "line1\nline2",
		Response:   "resp\nline2",
		Meta:       json.RawMessage(`{"b":2,"a":1}`),
		PrevHash:   "",
	}

	hash1, err := HashIntent(base)
	if err != nil {
		t.Fatalf("hash base: %v", err)
	}
	hash2, err := HashIntent(base)
	if err != nil {
		t.Fatalf("hash base repeat: %v", err)
	}
	if hash1 != hash2 {
		t.Fatalf("expected stable hash, got %s and %s", hash1, hash2)
	}

	metaReordered := base
	metaReordered.Meta = json.RawMessage(`{"a":1,"b":2}`)
	hash3, err := HashIntent(metaReordered)
	if err != nil {
		t.Fatalf("hash meta reorder: %v", err)
	}
	if hash1 != hash3 {
		t.Fatalf("expected identical hash for reordered meta, got %s and %s", hash1, hash3)
	}

	newlineVariant := base
	newlineVariant.Prompt = "line1\r\nline2"
	newlineVariant.Response = "resp\rline2"
	hash4, err := HashIntent(newlineVariant)
	if err != nil {
		t.Fatalf("hash newline variant: %v", err)
	}
	if hash1 != hash4 {
		t.Fatalf("expected identical hash for newline variants, got %s and %s", hash1, hash4)
	}
}
