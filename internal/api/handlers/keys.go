package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/api/middleware"
	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
	"github.com/chuxorg/yanzi/internal/auth"
)

// NewCreateKeyHandler returns the POST /v0/keys handler.
func NewCreateKeyHandler(deps Dependencies) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !deps.AuthConfig.Enabled {
			responses.WriteError(w, http.StatusNotImplemented, "not_implemented", "auth is not enabled on this instance")
			return
		}

		callerKey, ok := middleware.APIKeyFromContext(r.Context())
		if !ok || !callerKey.Scope.Allows(auth.ScopeAdmin) {
			responses.WriteError(w, http.StatusForbidden, "forbidden", "admin scope required")
			return
		}

		var req models.APIKeyCreateRequest
		if !decodeJSONBody(w, r, &req) {
			return
		}

		name := strings.TrimSpace(req.Name)
		if name == "" {
			responses.WriteError(w, http.StatusBadRequest, "invalid_request", "name is required")
			return
		}
		if len(name) > 100 {
			responses.WriteError(w, http.StatusBadRequest, "invalid_request", "name must be 100 characters or fewer")
			return
		}

		scope := auth.Scope(strings.TrimSpace(req.Scope))
		switch scope {
		case auth.ScopeRead, auth.ScopeWrite, auth.ScopeAdmin:
		default:
			responses.WriteError(w, http.StatusBadRequest, "invalid_request", "scope must be read, write, or admin")
			return
		}

		key, fullKey, err := deps.APIKeyStore.CreateKey(r.Context(), name, scope, req.Dev)
		if err != nil {
			log.Printf("keys: create key error: %v", err)
			responses.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to create key")
			return
		}

		log.Printf("API key created: %s scope=%s", name, scope)
		responses.WriteJSON(w, http.StatusCreated, models.APIKeyCreateResponse{
			ID:        key.ID,
			Name:      key.Name,
			Key:       fullKey,
			KeyPrefix: key.KeyPrefix,
			Scope:     string(key.Scope),
			CreatedAt: key.CreatedAt.Format(time.RFC3339),
		})
	})
}

// NewListKeysHandler returns the GET /v0/keys handler.
func NewListKeysHandler(deps Dependencies) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !deps.AuthConfig.Enabled {
			responses.WriteError(w, http.StatusNotImplemented, "not_implemented", "auth is not enabled on this instance")
			return
		}

		callerKey, ok := middleware.APIKeyFromContext(r.Context())
		if !ok || !callerKey.Scope.Allows(auth.ScopeAdmin) {
			responses.WriteError(w, http.StatusForbidden, "forbidden", "admin scope required")
			return
		}

		keys, err := deps.APIKeyStore.ListKeys(r.Context())
		if err != nil {
			log.Printf("keys: list keys error: %v", err)
			responses.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list keys")
			return
		}

		summaries := make([]models.APIKeySummary, 0, len(keys))
		for _, k := range keys {
			s := models.APIKeySummary{
				ID:        k.ID,
				Name:      k.Name,
				KeyPrefix: k.KeyPrefix,
				Scope:     string(k.Scope),
				CreatedAt: k.CreatedAt.Format(time.RFC3339),
			}
			if k.LastUsedAt != nil {
				t := k.LastUsedAt.Format(time.RFC3339)
				s.LastUsedAt = &t
			}
			summaries = append(summaries, s)
		}

		responses.WriteJSON(w, http.StatusOK, models.APIKeyListResponse{Keys: summaries})
	})
}

// NewRevokeKeyHandler returns the DELETE /v0/keys/{id} handler.
func NewRevokeKeyHandler(deps Dependencies) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !deps.AuthConfig.Enabled {
			responses.WriteError(w, http.StatusNotImplemented, "not_implemented", "auth is not enabled on this instance")
			return
		}

		callerKey, ok := middleware.APIKeyFromContext(r.Context())
		if !ok || !callerKey.Scope.Allows(auth.ScopeAdmin) {
			responses.WriteError(w, http.StatusForbidden, "forbidden", "admin scope required")
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/v0/keys/")
		id = strings.TrimSpace(id)
		if id == "" {
			responses.WriteError(w, http.StatusBadRequest, "invalid_request", "key id is required")
			return
		}

		if err := deps.APIKeyStore.RevokeKey(r.Context(), id); err != nil {
			if errors.Is(err, auth.ErrKeyNotFound) {
				responses.WriteError(w, http.StatusNotFound, "not_found", "key not found")
				return
			}
			log.Printf("keys: revoke key error: %v", err)
			responses.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to revoke key")
			return
		}

		log.Printf("API key revoked: %s", id)
		w.WriteHeader(http.StatusNoContent)
	})
}
