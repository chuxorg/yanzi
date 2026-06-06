package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/chuxorg/yanzi/internal/api/middleware"
	"github.com/chuxorg/yanzi/internal/auth"
	"github.com/chuxorg/yanzi/internal/auth/testutil"
	"github.com/chuxorg/yanzi/internal/config"
)

// memKeyStore is a minimal in-memory APIKeyStore for middleware tests.
type memKeyStore struct {
	mu   sync.Mutex
	keys map[string]auth.APIKey // keyed by hash
	raw  map[string]string      // id → plaintext key (for test setup)
}

func newMemKeyStore() *memKeyStore {
	return &memKeyStore{
		keys: make(map[string]auth.APIKey),
		raw:  make(map[string]string),
	}
}

func (s *memKeyStore) CreateKey(_ context.Context, name string, scope auth.Scope, dev bool) (auth.APIKey, string, error) {
	fullKey, prefix, hash, err := auth.GenerateKey(dev)
	if err != nil {
		return auth.APIKey{}, "", err
	}
	id := fmt.Sprintf("key-%d", len(s.keys)+1)
	key := auth.APIKey{
		ID:        id,
		Name:      name,
		KeyHash:   hash,
		KeyPrefix: prefix,
		Scope:     scope,
		CreatedAt: time.Now(),
	}
	s.mu.Lock()
	s.keys[hash] = key
	s.raw[id] = fullKey
	s.mu.Unlock()
	return key, fullKey, nil
}

func (s *memKeyStore) GetKeyByHash(_ context.Context, hash string) (auth.APIKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k, ok := s.keys[hash]
	if !ok {
		return auth.APIKey{}, auth.ErrKeyNotFound
	}
	if k.RevokedAt != nil {
		return auth.APIKey{}, auth.ErrKeyNotFound
	}
	return k, nil
}

func (s *memKeyStore) ListKeys(_ context.Context) ([]auth.APIKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]auth.APIKey, 0, len(s.keys))
	for _, k := range s.keys {
		out = append(out, k)
	}
	return out, nil
}

func (s *memKeyStore) RevokeKey(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for hash, k := range s.keys {
		if k.ID == id {
			now := time.Now()
			k.RevokedAt = &now
			s.keys[hash] = k
			return nil
		}
	}
	return auth.ErrKeyNotFound
}

func (s *memKeyStore) UpdateLastUsed(_ context.Context, id string, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for hash, k := range s.keys {
		if k.ID == id {
			k.LastUsedAt = &at
			s.keys[hash] = k
			return nil
		}
	}
	return nil
}

// buildOIDCValidator creates a real OIDCValidator backed by the mock provider.
func buildOIDCValidator(t *testing.T, mock *testutil.MockOIDCProvider, issuer string) *auth.OIDCValidator {
	t.Helper()
	cfg := config.OIDCConfig{
		Enabled:      true,
		IssuerURL:    issuer,
		ScopeClaim:   "yanzi_scope",
		ScopeDefault: "read",
	}
	v, err := auth.NewOIDCValidator(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewOIDCValidator: %v", err)
	}
	return v
}

func oidcEnabledAuthCfg() config.AuthConfig {
	return config.AuthConfig{
		Enabled:        true,
		DevKeysAllowed: true,
	}
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestAuth_OIDCToken_Valid(t *testing.T) {
	mock, issuer := testutil.NewMockOIDCProvider(t)
	v := buildOIDCValidator(t, mock, issuer)

	token := mock.SignToken(map[string]interface{}{
		"sub":         "user-1",
		"yanzi_scope": "read",
	}, time.Hour)

	m := middleware.Auth(nil, v, oidcEnabledAuthCfg())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v0/artifacts", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	m(okHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestAuth_OIDCToken_InsufficientScope(t *testing.T) {
	mock, issuer := testutil.NewMockOIDCProvider(t)
	v := buildOIDCValidator(t, mock, issuer)

	// Read-scoped token trying a POST (requires write).
	token := mock.SignToken(map[string]interface{}{
		"sub":         "user-1",
		"yanzi_scope": "read",
	}, time.Hour)

	m := middleware.Auth(nil, v, oidcEnabledAuthCfg())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v0/artifacts", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	m(okHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestAuth_OIDCToken_NotConfigured(t *testing.T) {
	// oidcValidator is nil — instance has no OIDC configured.
	m := middleware.Auth(nil, nil, oidcEnabledAuthCfg())

	// Present something that isn't an API key prefix.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v0/artifacts", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.fake.token")

	m(okHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuth_OIDCToken_Expired(t *testing.T) {
	mock, issuer := testutil.NewMockOIDCProvider(t)
	v := buildOIDCValidator(t, mock, issuer)

	token := mock.SignToken(map[string]interface{}{"sub": "user-1"}, -time.Minute)

	m := middleware.Auth(nil, v, oidcEnabledAuthCfg())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v0/artifacts", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	m(okHandler()).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuth_APIKeyAndOIDC_Coexist(t *testing.T) {
	mock, issuer := testutil.NewMockOIDCProvider(t)
	oidcV := buildOIDCValidator(t, mock, issuer)

	// Build a real key store with one key.
	store := newMemKeyStore()
	_, rawKey, err := store.CreateKey(context.Background(), "test", auth.ScopeRead, true)
	if err != nil {
		t.Fatalf("CreateKey: %v", err)
	}

	cfg := config.AuthConfig{
		Enabled:        true,
		DevKeysAllowed: true,
	}
	m := middleware.Auth(store, oidcV, cfg)

	// API key works.
	t.Run("api_key", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v0/artifacts", nil)
		req.Header.Set("Authorization", "Bearer "+rawKey)
		m(okHandler()).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("api key: status = %d, want 200", rec.Code)
		}
	})

	// OIDC token works simultaneously.
	t.Run("oidc_token", func(t *testing.T) {
		token := mock.SignToken(map[string]interface{}{
			"sub":         "user-2",
			"yanzi_scope": "read",
		}, time.Hour)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v0/artifacts", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		m(okHandler()).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("oidc token: status = %d, want 200", rec.Code)
		}
	})
}
