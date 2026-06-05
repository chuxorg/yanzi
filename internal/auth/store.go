package auth

import (
	"context"
	"time"
)

// APIKeyStore defines key lifecycle operations implemented by each storage provider.
type APIKeyStore interface {
	CreateKey(ctx context.Context, name string, scope Scope, dev bool) (APIKey, string, error)
	GetKeyByHash(ctx context.Context, hash string) (APIKey, error)
	ListKeys(ctx context.Context) ([]APIKey, error)
	RevokeKey(ctx context.Context, id string) error
	UpdateLastUsed(ctx context.Context, id string, at time.Time) error
}
