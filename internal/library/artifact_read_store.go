package yanzilibrary

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/chuxorg/yanzi/internal/core/model"
	"github.com/chuxorg/yanzi/internal/core/store"
)

// ArtifactReadQuery describes the current local list/show read behavior.
type ArtifactReadQuery struct {
	Author         string
	Source         string
	Limit          int
	MetaFilters    map[string]string
	IncludeDeleted bool
}

// ArtifactReadStore isolates the legacy SQL-backed list/show read path.
type ArtifactReadStore struct {
	db *sql.DB
}

// NewArtifactReadStore constructs a read store using the provided database handle.
func NewArtifactReadStore(db *sql.DB) *ArtifactReadStore {
	return &ArtifactReadStore{db: db}
}

// ListIntentRecords returns current list/show records with existing filter semantics preserved.
func (s *ArtifactReadStore) ListIntentRecords(ctx context.Context, query ArtifactReadQuery) ([]model.IntentRecord, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("artifact read store is not initialized")
	}

	fetchLimit := query.Limit
	if fetchLimit <= 0 {
		fetchLimit = 20
	}
	if query.Author != "" || query.Source != "" || len(query.MetaFilters) > 0 {
		fetchLimit = fetchLimit * 5
		if fetchLimit < 100 {
			fetchLimit = 100
		}
	}

	intents, err := s.listIntentRecordsFromDB(ctx, fetchLimit, query.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	filtered := make([]model.IntentRecord, 0, len(intents))
	for _, intent := range intents {
		if query.Author != "" && intent.Author != query.Author {
			continue
		}
		if query.Source != "" && intent.SourceType != query.Source {
			continue
		}
		filtered = append(filtered, intent)
	}

	if len(query.MetaFilters) > 0 {
		filtered, err = store.FilterIntentsByMeta(filtered, query.MetaFilters)
		if err != nil {
			return nil, err
		}
	}

	if query.Limit > 0 && len(filtered) > query.Limit {
		filtered = filtered[:query.Limit]
	}
	return filtered, nil
}

// GetIntentRecord returns the current local show record with existing not-found behavior preserved.
func (s *ArtifactReadStore) GetIntentRecord(ctx context.Context, id string) (model.IntentRecord, error) {
	if s == nil || s.db == nil {
		return model.IntentRecord{}, errors.New("artifact read store is not initialized")
	}

	record, err := s.getIntentRecordFromDB(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentRecord{}, fmt.Errorf("intent not found for ID %s", id)
		}
		return model.IntentRecord{}, err
	}
	return record, nil
}

func (s *ArtifactReadStore) getIntentRecordFromDB(ctx context.Context, id string) (model.IntentRecord, error) {
	var record model.IntentRecord
	var title sql.NullString
	var meta sql.NullString
	var prevHash sql.NullString
	row := s.db.QueryRowContext(ctx, `SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash FROM intents WHERE id = ?`, id)
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

func (s *ArtifactReadStore) listIntentRecordsFromDB(ctx context.Context, limit int, includeDeleted bool) ([]model.IntentRecord, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.QueryContext(ctx, `SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash, metadata FROM intents WHERE source_type <> 'artifact' ORDER BY created_at DESC, id DESC LIMIT ?`, limit)
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
		var metadata sql.NullString
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
			&metadata,
		); err != nil {
			return nil, err
		}
		if title.Valid {
			record.Title = title.String
		}
		if meta.Valid || metadata.Valid {
			combinedMeta, err := mergedArtifactReadMetadata(meta.String, metadata.String)
			if err != nil {
				return nil, err
			}
			if !includeDeleted && isDeletedArtifactReadMetadata(combinedMeta) {
				continue
			}
			if len(combinedMeta) > 0 {
				encodedMeta, err := json.Marshal(combinedMeta)
				if err != nil {
					return nil, fmt.Errorf("encode merged meta: %w", err)
				}
				record.Meta = encodedMeta
			}
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

func mergedArtifactReadMetadata(metaText, metadataText string) (map[string]string, error) {
	merged := map[string]string{}
	for _, raw := range []string{metaText, metadataText} {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		decoded, err := decodeArtifactReadMetadata(raw)
		if err != nil {
			return nil, err
		}
		for key, value := range decoded {
			merged[key] = value
		}
	}
	return merged, nil
}

func isDeletedArtifactReadMetadata(meta map[string]string) bool {
	return strings.EqualFold(strings.TrimSpace(meta["deleted"]), "true")
}

func decodeArtifactReadMetadata(raw string) (map[string]string, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]string{}, nil
	}

	var decoded map[string]string
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return nil, err
	}
	if decoded == nil {
		return map[string]string{}, nil
	}
	return decoded, nil
}
