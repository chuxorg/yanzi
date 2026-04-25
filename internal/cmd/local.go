package cmd

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/core/hash"
	"github.com/chuxorg/yanzi/internal/core/model"
	"github.com/chuxorg/yanzi/internal/core/store"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

func openLocalDB(cfg config.Config) (*sql.DB, error) {
	if cfg.DBPath == "" {
		return nil, errors.New("db_path is required when mode=local")
	}
	if err := os.Setenv("YANZI_DB_PATH", cfg.DBPath); err != nil {
		return nil, fmt.Errorf("set YANZI_DB_PATH: %w", err)
	}
	return yanzilibrary.InitDB()
}

func buildLocalIntent(req createIntentInput) (model.IntentRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	id, err := newIntentID()
	if err != nil {
		return model.IntentRecord{}, err
	}

	record := model.IntentRecord{
		ID:         id,
		CreatedAt:  now,
		Author:     req.Author,
		SourceType: req.SourceType,
		Title:      req.Title,
		Prompt:     req.Prompt,
		Response:   req.Response,
		PrevHash:   req.PrevHash,
		Meta:       req.Meta,
	}
	sum, err := hash.HashIntent(record)
	if err != nil {
		return model.IntentRecord{}, err
	}
	record.Hash = sum
	return record, nil
}

func newIntentID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(buf[:]), nil
}

func createLocalIntent(ctx context.Context, db *sql.DB, record model.IntentRecord) error {
	var title any
	if record.Title != "" {
		title = record.Title
	}
	var meta any
	if len(record.Meta) > 0 {
		meta = string(record.Meta)
	}
	var prevHash any
	if record.PrevHash != "" {
		prevHash = record.PrevHash
	}

	_, err := db.ExecContext(
		ctx,
		`INSERT INTO intents (id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash, class, type, content, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ID,
		record.CreatedAt,
		record.Author,
		record.SourceType,
		title,
		record.Prompt,
		record.Response,
		meta,
		prevHash,
		record.Hash,
		"intent",
		"prompt",
		record.Prompt,
		meta,
	)
	return err
}

func verifyLocalIntent(ctx context.Context, db *sql.DB, id string) (verifyResult, error) {
	record, err := dbGetIntent(ctx, db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return verifyResult{}, fmt.Errorf("intent not found: %s", id)
		}
		return verifyResult{}, err
	}

	computed, err := hash.HashIntent(record)
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

func chainLocalIntent(ctx context.Context, db *sql.DB, id string) (chainResult, error) {
	head, err := dbGetIntent(ctx, db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return chainResult{}, fmt.Errorf("intent not found: %s", id)
		}
		return chainResult{}, err
	}

	intents := []model.IntentRecord{head}
	current := head
	var missing []string
	for current.PrevHash != "" {
		prev, err := dbGetIntentByHash(ctx, db, current.PrevHash)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				missing = append(missing, current.PrevHash)
				break
			}
			return chainResult{}, err
		}
		intents = append(intents, prev)
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

func listLocalIntents(ctx context.Context, db *sql.DB, author, source string, limit int, metaFilters map[string]string) ([]model.IntentRecord, error) {
	fetchLimit := limit
	if fetchLimit <= 0 {
		fetchLimit = 20
	}
	if author != "" || source != "" || len(metaFilters) > 0 {
		fetchLimit = fetchLimit * 5
		if fetchLimit < 100 {
			fetchLimit = 100
		}
	}

	intents, err := listLocalIntentsFromDB(ctx, db, fetchLimit)
	if err != nil {
		return nil, err
	}

	filtered := make([]model.IntentRecord, 0, len(intents))
	for _, intent := range intents {
		if author != "" && intent.Author != author {
			continue
		}
		if source != "" && intent.SourceType != source {
			continue
		}
		filtered = append(filtered, intent)
	}

	if len(metaFilters) > 0 {
		filtered, err = store.FilterIntentsByMeta(filtered, metaFilters)
		if err != nil {
			return nil, err
		}
	}

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}

func getLocalIntent(ctx context.Context, db *sql.DB, id string) (model.IntentRecord, error) {
	record, err := dbGetIntent(ctx, db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentRecord{}, fmt.Errorf("intent not found for ID %s", id)
		}
		return model.IntentRecord{}, err
	}
	return record, nil
}

func dbGetIntent(ctx context.Context, db *sql.DB, id string) (model.IntentRecord, error) {
	var record model.IntentRecord
	var title sql.NullString
	var meta sql.NullString
	var prevHash sql.NullString
	row := db.QueryRowContext(ctx, `SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash FROM intents WHERE id = ?`, id)
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

func listLocalIntentsFromDB(ctx context.Context, db *sql.DB, limit int) ([]model.IntentRecord, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := db.QueryContext(ctx, `SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash FROM intents WHERE source_type <> 'artifact' ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var intents []model.IntentRecord
	for rows.Next() {
		var record model.IntentRecord
		var title sql.NullString
		var meta sql.NullString
		var prevHash sql.NullString
		if err := rows.Scan(
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
			return nil, err
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
		intents = append(intents, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return intents, nil
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
