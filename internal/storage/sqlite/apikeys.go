package sqlite

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
		 VALUES (?, ?, ?, ?, ?, ?)`,
		id, name, hash, prefix, string(scope), now.Format(time.RFC3339Nano),
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
		 WHERE key_hash = ? AND revoked_at IS NULL`,
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
			createdAt               string
			lastUsedAt              sql.NullString
		)
		if err := rows.Scan(&id, &name, &prefix, &scope, &createdAt, &lastUsedAt); err != nil {
			return nil, fmt.Errorf("scan api key row: %w", err)
		}
		k, err := buildAPIKey(id, name, "", prefix, scope, createdAt, lastUsedAt)
		if err != nil {
			return nil, err
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
		`UPDATE api_keys SET revoked_at = ? WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339Nano), id,
	)
	if err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	return nil
}

// UpdateLastUsed records the time a key was last used for observability.
func (p *Provider) UpdateLastUsed(ctx context.Context, id string, at time.Time) error {
	_, err := p.db.ExecContext(ctx,
		`UPDATE api_keys SET last_used_at = ? WHERE id = ?`,
		at.UTC().Format(time.RFC3339Nano), id,
	)
	return err
}

func scanAPIKey(row *sql.Row) (auth.APIKey, error) {
	var (
		id, name, hash, prefix, scope string
		createdAt                     string
		lastUsedAt                    sql.NullString
	)
	if err := row.Scan(&id, &name, &hash, &prefix, &scope, &createdAt, &lastUsedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.APIKey{}, auth.ErrKeyNotFound
		}
		return auth.APIKey{}, fmt.Errorf("scan api key: %w", err)
	}
	return buildAPIKey(id, name, hash, prefix, scope, createdAt, lastUsedAt)
}

func buildAPIKey(id, name, hash, prefix, scope, createdAtStr string, lastUsedAtStr sql.NullString) (auth.APIKey, error) {
	createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
	if err != nil {
		return auth.APIKey{}, fmt.Errorf("parse created_at: %w", err)
	}
	k := auth.APIKey{
		ID:        id,
		Name:      name,
		KeyHash:   hash,
		KeyPrefix: prefix,
		Scope:     auth.Scope(scope),
		CreatedAt: createdAt,
	}
	if lastUsedAtStr.Valid {
		t, err := time.Parse(time.RFC3339Nano, lastUsedAtStr.String)
		if err != nil {
			return auth.APIKey{}, fmt.Errorf("parse last_used_at: %w", err)
		}
		k.LastUsedAt = &t
	}
	return k, nil
}
