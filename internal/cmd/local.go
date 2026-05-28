package cmd

import (
	"context"
	"database/sql"

	"github.com/chuxorg/yanzi/internal/config"
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
	return registry.Open(context.Background(), cfg, registry.Options{Migrations: yanzilibrary.MigrationsFS()})
}

func verifyLocalIntent(ctx context.Context, provider storage.Provider, id string) (verifyResult, error) {
	result, err := yanzilibrary.VerifyIntent(ctx, provider, id)
	if err != nil {
		return verifyResult{}, err
	}
	return verifyResult(result), nil
}

func chainLocalIntent(ctx context.Context, provider storage.Provider, id string) (chainResult, error) {
	result, err := yanzilibrary.ChainIntent(ctx, provider, id)
	if err != nil {
		return chainResult{}, err
	}
	return chainResult(result), nil
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
