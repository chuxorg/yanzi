package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/auth"
	"github.com/chuxorg/yanzi/internal/api/responses"
	"github.com/chuxorg/yanzi/internal/config"
)

type contextKey string

const contextKeyAPIKey contextKey = "api_key"

// Auth returns a middleware that enforces API key authentication when enabled.
// When cfg.Enabled is false all requests pass through unchanged, preserving
// existing behaviour completely.
func Auth(store auth.APIKeyStore, cfg config.AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// CORS preflight must never be blocked by auth.
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Health endpoint is always public.
			if r.Method == http.MethodGet && r.URL.Path == "/v0/health" {
				next.ServeHTTP(w, r)
				return
			}

			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				responses.WriteError(w, http.StatusUnauthorized, "unauthorized", "API key required")
				return
			}
			presented := strings.TrimPrefix(header, "Bearer ")

			if !strings.HasPrefix(presented, auth.PrefixLive) && !strings.HasPrefix(presented, auth.PrefixDev) {
				responses.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid API key format")
				return
			}
			if strings.HasPrefix(presented, auth.PrefixDev) && !cfg.DevKeysAllowed {
				responses.WriteError(w, http.StatusUnauthorized, "unauthorized", "development keys not accepted")
				return
			}

			hash := auth.HashKey(presented)
			key, err := store.GetKeyByHash(r.Context(), hash)
			if err != nil {
				if errors.Is(err, auth.ErrKeyNotFound) {
					responses.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid or revoked API key")
					return
				}
				log.Printf("auth: key lookup error: %v", err)
				responses.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
				return
			}

			required := scopeForMethod(r.Method)
			if !key.Scope.Allows(required) {
				responses.WriteError(w, http.StatusForbidden, "forbidden", "API key scope insufficient")
				return
			}

			go func() {
				_ = store.UpdateLastUsed(context.Background(), key.ID, time.Now())
			}()

			ctx := context.WithValue(r.Context(), contextKeyAPIKey, key)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// APIKeyFromContext retrieves the authenticated API key attached to ctx.
func APIKeyFromContext(ctx context.Context) (auth.APIKey, bool) {
	key, ok := ctx.Value(contextKeyAPIKey).(auth.APIKey)
	return key, ok
}

func scopeForMethod(method string) auth.Scope {
	switch method {
	case http.MethodGet, http.MethodHead:
		return auth.ScopeRead
	default:
		return auth.ScopeWrite
	}
}
