package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/core/hash"
	"github.com/chuxorg/yanzi/internal/core/model"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
	"github.com/chuxorg/yanzi/internal/storage"
	"github.com/chuxorg/yanzi/internal/storage/registry"
)

func openLocalDB(cfg config.Config) (*sql.DB, error) {
	path, err := config.EffectiveLocalDBPath(cfg)
	if err != nil {
		return nil, err
	}
	return yanzilibrary.InitDBAtPath(path)
}

func openLocalProvider(cfg config.Config) (storage.Provider, error) {
	return registry.Open(context.Background(), cfg, registry.Options{})
}

func verifyLocalIntent(ctx context.Context, provider storage.Provider, id string) (verifyResult, error) {
	record, err := provider.GetVerificationIntent(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return verifyResult{}, fmt.Errorf("intent not found: %s", id)
		}
		return verifyResult{}, err
	}

	computed, err := hash.HashIntent(modelIntentFromStorage(record))
	result := verifyResult{
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

func chainLocalIntent(ctx context.Context, provider storage.Provider, id string) (chainResult, error) {
	head, err := provider.GetVerificationIntent(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return chainResult{}, fmt.Errorf("intent not found: %s", id)
		}
		return chainResult{}, err
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
			return chainResult{}, err
		}
		intents = append(intents, modelIntentFromStorage(prev))
		current = prev
	}

	for i, j := 0, len(intents)-1; i < j; i, j = i+1, j-1 {
		intents[i], intents[j] = intents[j], intents[i]
	}

	return chainResult{
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

func listLocalIntents(ctx context.Context, db *sql.DB, author, source string, limit int, metaFilters map[string]string, includeDeleted bool) ([]model.IntentRecord, error) {
	readStore := yanzilibrary.NewArtifactReadStore(db)
	return readStore.ListIntentRecords(ctx, yanzilibrary.ArtifactReadQuery{
		Author:         author,
		Source:         source,
		Limit:          limit,
		MetaFilters:    metaFilters,
		IncludeDeleted: includeDeleted,
	})
}

func getLocalIntent(ctx context.Context, db *sql.DB, id string) (model.IntentRecord, error) {
	readStore := yanzilibrary.NewArtifactReadStore(db)
	return readStore.GetIntentRecord(ctx, id)
}

func dbGetIntentByHash(ctx context.Context, db *sql.DB, intentHash string) (model.IntentRecord, error) {
	var record model.IntentRecord
	var title sql.NullString
	var meta sql.NullString
	var prevHash sql.NullString
	row := db.QueryRowContext(ctx, `SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash FROM intents WHERE hash = ?`, intentHash)
	if err := row.Scan(
		&record.ID,
		&record.CreatedAt,
		&record.Author,
		&record.SourceType,
		&title,
		&record.Prompt,
		&record.Response,
		&meta,
		&prevHash,
		&record.Hash,
	); err != nil {
		return model.IntentRecord{}, err
	}
	if title.Valid {
		record.Title = title.String
	}
	if meta.Valid && meta.String != "" {
		record.Meta = []byte(meta.String)
	}
	if prevHash.Valid {
		record.PrevHash = prevHash.String
	}
	return record, nil
}

type createIntentInput struct {
	Author     string
	SourceType string
	Title      string
	Prompt     string
	Response   string
	Meta       []byte
	PrevHash   string
}

type verifyResult struct {
	ID           string
	Valid        bool
	StoredHash   string
	ComputedHash string
	PrevHash     string
	Error        *string
}

type chainResult struct {
	HeadID       string
	Length       int
	Intents      []model.IntentRecord
	MissingLinks []string
}
