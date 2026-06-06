package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/chuxorg/yanzi/internal/auth"
)

func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// CreateKey generates a new API key, stores its hash, and returns the APIKey
// record plus the full plaintext key. The plaintext key is never stored.
func (p *Provider) CreateKey(ctx context.Context, name string, scope auth.Scope, dev bool) (auth.APIKey, string, error) {
	fullKey, prefix, hash, err := auth.GenerateKey(dev)
	if err != nil {
		return auth.APIKey{}, "", fmt.Errorf("generate key: %w", err)
	}

	id, err := generateID()
	if err != nil {
		return auth.APIKey{}, "", err
	}

	now := time.Now().UTC()
	_, err = p.db.ExecContext(ctx,
		`INSERT INTO api_keys (id, name, key_hash, key_prefix, scope, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		id, name, hash, prefix, string(scope), now,
	)
	if err != nil {
		return auth.APIKey{}, "", fmt.Errorf("insert api key: %w", err)
	}

	key := auth.APIKey{
		ID:        id,
		Name:      name,
		KeyHash:   hash,
		KeyPrefix: prefix,
		Scope:     scope,
		CreatedAt: now,
	}
	return key, fullKey, nil
}

// GetKeyByHash looks up an active (non-revoked) key by its SHA-256 hash.
func (p *Provider) GetKeyByHash(ctx context.Context, hash string) (auth.APIKey, error) {
	row := p.db.QueryRowContext(ctx,
		`SELECT id, name, key_hash, key_prefix, scope, created_at, last_used_at
		 FROM api_keys
		 WHERE key_hash = $1 AND revoked_at IS NULL`,
		hash,
	)
	return scanAPIKey(row)
}

// ListKeys returns all non-revoked keys ordered by creation time, newest first.
// key_hash is intentionally excluded.
func (p *Provider) ListKeys(ctx context.Context) ([]auth.APIKey, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, name, key_prefix, scope, created_at, last_used_at
		 FROM api_keys
		 WHERE revoked_at IS NULL
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	var keys []auth.APIKey
	for rows.Next() {
		var (
			id, name, prefix, scope string
			createdAt               time.Time
			lastUsedAt              sql.NullTime
		)
		if err := rows.Scan(&id, &name, &prefix, &scope, &createdAt, &lastUsedAt); err != nil {
			return nil, fmt.Errorf("scan api key row: %w", err)
		}
		k := auth.APIKey{
			ID:        id,
			Name:      name,
			KeyPrefix: prefix,
			Scope:     auth.Scope(scope),
			CreatedAt: createdAt.UTC(),
		}
		if lastUsedAt.Valid {
			t := lastUsedAt.Time.UTC()
			k.LastUsedAt = &t
		}
		keys = append(keys, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate api keys: %w", err)
	}
	return keys, nil
}

// RevokeKey soft-deletes a key by setting revoked_at.
func (p *Provider) RevokeKey(ctx context.Context, id string) error {
	_, err := p.db.ExecContext(ctx,
		`UPDATE api_keys SET revoked_at = $1 WHERE id = $2`,
		time.Now().UTC(), id,
	)
	if err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	return nil
}

// UpdateLastUsed records the time a key was last used for observability.
func (p *Provider) UpdateLastUsed(ctx context.Context, id string, at time.Time) error {
	_, err := p.db.ExecContext(ctx,
		`UPDATE api_keys SET last_used_at = $1 WHERE id = $2`,
		at.UTC(), id,
	)
	return err
}

func scanAPIKey(row *sql.Row) (auth.APIKey, error) {
	var (
		id, name, hash, prefix, scope string
		createdAt                     time.Time
		lastUsedAt                    sql.NullTime
	)
	if err := row.Scan(&id, &name, &hash, &prefix, &scope, &createdAt, &lastUsedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.APIKey{}, auth.ErrKeyNotFound
		}
		return auth.APIKey{}, fmt.Errorf("scan api key: %w", err)
	}
	k := auth.APIKey{
		ID:        id,
		Name:      name,
		KeyHash:   hash,
		KeyPrefix: prefix,
		Scope:     auth.Scope(scope),
		CreatedAt: createdAt.UTC(),
	}
	if lastUsedAt.Valid {
		t := lastUsedAt.Time.UTC()
		k.LastUsedAt = &t
	}
	return k, nil
}
