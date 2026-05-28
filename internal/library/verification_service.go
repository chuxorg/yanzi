package yanzilibrary

import (
	"context"
	"errors"
	"fmt"

	"github.com/chuxorg/yanzi/internal/core/hash"
	"github.com/chuxorg/yanzi/internal/core/model"
	"github.com/chuxorg/yanzi/internal/storage"
)

// VerifyResult captures current deterministic verification output.
type VerifyResult struct {
	ID           string
	Valid        bool
	StoredHash   string
	ComputedHash string
	PrevHash     string
	Error        *string
}

// ChainResult captures current deterministic chain traversal output.
type ChainResult struct {
	HeadID       string
	Length       int
	Intents      []model.IntentRecord
	MissingLinks []string
}

// VerifyIntent preserves current provider-backed verification semantics.
func VerifyIntent(ctx context.Context, provider storage.Provider, id string) (VerifyResult, error) {
	record, err := provider.GetVerificationIntent(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return VerifyResult{}, fmt.Errorf("intent not found: %s", id)
		}
		return VerifyResult{}, err
	}

	computed, err := hash.HashIntent(modelIntentFromStorage(record))
	result := VerifyResult{
		ID:           record.ID,
		StoredHash:   record.Hash,
		ComputedHash: computed,
		PrevHash:     record.PrevHash,
		Valid:        err == nil && computed == record.Hash,
	}
	if err != nil {
		msg := err.Error()
		result.Error = &msg
	}
	return result, nil
}

// ChainIntent preserves current provider-backed chain traversal semantics.
func ChainIntent(ctx context.Context, provider storage.Provider, id string) (ChainResult, error) {
	head, err := provider.GetVerificationIntent(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ChainResult{}, fmt.Errorf("intent not found: %s", id)
		}
		return ChainResult{}, err
	}

	headIntent := modelIntentFromStorage(head)
	intents := []model.IntentRecord{headIntent}
	current := head
	var missing []string
	for current.PrevHash != "" {
		prev, err := provider.GetVerificationIntentByHash(ctx, current.PrevHash)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				missing = append(missing, current.PrevHash)
				break
			}
			return ChainResult{}, err
		}
		intents = append(intents, modelIntentFromStorage(prev))
		current = prev
	}

	for i, j := 0, len(intents)-1; i < j; i, j = i+1, j-1 {
		intents[i], intents[j] = intents[j], intents[i]
	}

	return ChainResult{
		HeadID:       head.ID,
		Length:       len(intents),
		Intents:      intents,
		MissingLinks: missing,
	}, nil
}

func modelIntentFromStorage(record storage.IntentRecord) model.IntentRecord {
	return model.IntentRecord{
		ID:         record.ID,
		CreatedAt:  record.CreatedAt,
		Author:     record.Author,
		SourceType: record.SourceType,
		Title:      record.Title,
		Prompt:     record.Prompt,
		Response:   record.Response,
		Meta:       record.Meta,
		PrevHash:   record.PrevHash,
		Hash:       record.Hash,
	}
}
