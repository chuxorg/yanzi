package auth

import "time"

// Scope defines the access level granted by an API key.
type Scope string

const (
	ScopeRead  Scope = "read"
	ScopeWrite Scope = "write"
	ScopeAdmin Scope = "admin"
)

// Allows reports whether scope s grants access to the required scope.
func (s Scope) Allows(required Scope) bool {
	switch s {
	case ScopeAdmin:
		return true
	case ScopeWrite:
		return required == ScopeRead || required == ScopeWrite
	case ScopeRead:
		return required == ScopeRead
	}
	return false
}

const (
	PrefixLive = "yk_live_"
	PrefixDev  = "yk_dev_"
)

// APIKey represents a stored API key record (never contains the plaintext key).
type APIKey struct {
	ID         string
	Name       string
	KeyHash    string     // SHA-256 hex of the full key
	KeyPrefix  string     // first 12 chars of full key + "..." for display
	Scope      Scope
	CreatedAt  time.Time
	LastUsedAt *time.Time
	RevokedAt  *time.Time
}
