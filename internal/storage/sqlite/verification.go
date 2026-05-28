package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/chuxorg/yanzi/internal/storage"
)

// GetVerificationIntent loads an intent by ID using current verification semantics.
func (p *Provider) GetVerificationIntent(ctx context.Context, id string) (storage.IntentRecord, error) {
	if p == nil || p.db == nil {
		return storage.IntentRecord{}, storage.ErrProviderUnavailable
	}
	record, err := p.getVerificationIntent(ctx, `SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash FROM intents WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.IntentRecord{}, fmt.Errorf("%w: intent not found: %s", storage.ErrNotFound, id)
		}
		return storage.IntentRecord{}, err
	}
	return record, nil
}

// GetVerificationIntentByHash loads an intent by hash using current chain traversal semantics.
func (p *Provider) GetVerificationIntentByHash(ctx context.Context, intentHash string) (storage.IntentRecord, error) {
	if p == nil || p.db == nil {
		return storage.IntentRecord{}, storage.ErrProviderUnavailable
	}
	record, err := p.getVerificationIntent(ctx, `SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash FROM intents WHERE hash = ?`, intentHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.IntentRecord{}, fmt.Errorf("%w: intent hash not found: %s", storage.ErrNotFound, intentHash)
		}
		return storage.IntentRecord{}, err
	}
	return record, nil
}

func (p *Provider) getVerificationIntent(ctx context.Context, query string, arg string) (storage.IntentRecord, error) {
	var record storage.IntentRecord
	var title sql.NullString
	var meta sql.NullString
	var prevHash sql.NullString
	row := p.db.QueryRowContext(ctx, query, arg)
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
		return storage.IntentRecord{}, err
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
