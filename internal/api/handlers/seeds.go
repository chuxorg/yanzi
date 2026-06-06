package handlers

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/chuxorg/yanzi/internal/api/middleware"
	"github.com/chuxorg/yanzi/internal/api/responses"
	"github.com/chuxorg/yanzi/internal/auth"
	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/packs"
	"github.com/chuxorg/yanzi/internal/storage"
)

const seedsPath = "/v0/seeds"

// NewSeedsHandler returns the /v0/seeds handler.
func NewSeedsHandler(deps Dependencies) http.Handler {
	deps = deps.withDefaults()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg, err := deps.LoadConfig()
		if err != nil {
			responses.WriteError(w, http.StatusInternalServerError, "config_load_failed", err.Error())
			return
		}
		if cfg.Mode != config.ModeLocal {
			responses.WriteError(w, http.StatusBadRequest, "unsupported_mode", "seeds require local mode")
			return
		}
		if deps.PackStore == nil {
			responses.WriteError(w, http.StatusInternalServerError, "not_available", "pack store not initialized")
			return
		}

		switch {
		case r.URL.Path == seedsPath:
			handleSeedsCollection(w, r, deps, cfg)
		case strings.HasPrefix(r.URL.Path, seedsPath+"/"):
			handleSeedDetail(w, r, deps, cfg)
		default:
			http.NotFound(w, r)
		}
	})
}

func handleSeedsCollection(w http.ResponseWriter, r *http.Request, deps Dependencies, cfg config.Config) {
	switch r.Method {
	case http.MethodGet:
		filter := packs.SeedFilter{
			SeedType: r.URL.Query().Get("type"),
			Name:     r.URL.Query().Get("name"),
		}
		if rb := r.URL.Query().Get("role_bits"); rb != "" {
			if n, err := strconv.Atoi(rb); err == nil {
				filter.MinRoleBits = packs.RoleBits(n)
			}
		}
		seeds, err := deps.PackStore.ListSeeds(r.Context(), filter)
		if err != nil {
			log.Printf("seeds: list error: %v", err)
			responses.WriteError(w, http.StatusInternalServerError, "list_failed", "failed to list seeds")
			return
		}
		if seeds == nil {
			seeds = []packs.Seed{}
		}
		responses.WriteJSON(w, http.StatusOK, map[string]any{"seeds": seeds})

	case http.MethodPost:
		if !requireWriteScope(w, r, deps) {
			return
		}
		var seed packs.Seed
		if !decodePackBody(w, r, &seed) {
			return
		}
		if strings.TrimSpace(seed.Name) == "" {
			responses.WriteError(w, http.StatusBadRequest, "invalid_request", "name is required")
			return
		}
		if strings.TrimSpace(seed.SeedType) == "" {
			responses.WriteError(w, http.StatusBadRequest, "invalid_request", "seed_type is required")
			return
		}
		if len(seed.Content.Sections) == 0 {
			responses.WriteError(w, http.StatusBadRequest, "invalid_request", "content.sections is required")
			return
		}
		if seed.RoleAccessBits == 0 {
			seed.RoleAccessBits = packs.RoleObserver
		}
		seed.TokenEstimate = seed.Content.TokenEstimate()
		seed.CreatedAt = time.Now().UTC()

		artifactID, err := createBackingArtifact(r.Context(), deps, cfg, storage.CreateArtifactInput{
			Class: "seed",
			Type:  seed.SeedType,
			Title: seed.Name,
		})
		if err != nil {
			log.Printf("seeds: create backing artifact error: %v", err)
			responses.WriteError(w, http.StatusInternalServerError, "artifact_create_failed", "failed to create backing artifact")
			return
		}
		seed.ArtifactID = artifactID

		created, err := deps.PackStore.CreateSeed(r.Context(), seed)
		if err != nil {
			log.Printf("seeds: create error: %v", err)
			responses.WriteError(w, http.StatusInternalServerError, "create_failed", "failed to create seed")
			return
		}
		responses.WriteJSON(w, http.StatusCreated, created)

	default:
		responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	}
}

func handleSeedDetail(w http.ResponseWriter, r *http.Request, deps Dependencies, _ config.Config) {
	id := strings.TrimPrefix(r.URL.Path, seedsPath+"/")
	id = strings.TrimSpace(id)
	if id == "" {
		responses.WriteError(w, http.StatusBadRequest, "invalid_request", "seed id is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		seed, err := deps.PackStore.GetSeed(r.Context(), id)
		if err != nil {
			if errors.Is(err, packs.ErrNotFound) {
				responses.WriteError(w, http.StatusNotFound, "not_found", "seed not found")
				return
			}
			log.Printf("seeds: get error: %v", err)
			responses.WriteError(w, http.StatusInternalServerError, "get_failed", "failed to get seed")
			return
		}
		responses.WriteJSON(w, http.StatusOK, seed)

	case http.MethodDelete:
		if !requireAdminScope(w, r, deps) {
			return
		}
		if err := deps.PackStore.DeleteSeed(r.Context(), id); err != nil {
			if errors.Is(err, packs.ErrNotFound) {
				responses.WriteError(w, http.StatusNotFound, "not_found", "seed not found")
				return
			}
			log.Printf("seeds: delete error: %v", err)
			responses.WriteError(w, http.StatusInternalServerError, "delete_failed", "failed to delete seed")
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	}
}

// decodePackBody accepts application/json or application/yaml request bodies.
func decodePackBody(w http.ResponseWriter, r *http.Request, out any) bool {
	if r.Body == nil {
		responses.WriteError(w, http.StatusBadRequest, "invalid_request", "request body is required")
		return false
	}
	defer r.Body.Close()

	body, err := io.ReadAll(io.LimitReader(r.Body, maxJSONBodyBytes))
	if err != nil {
		responses.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to read request body")
		return false
	}

	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "yaml") {
		if err := yaml.Unmarshal(body, out); err != nil {
			responses.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid YAML request body")
			return false
		}
		return true
	}

	if err := decodeJSONBytes(w, body, out); err != nil {
		return false
	}
	return true
}

func requireWriteScope(w http.ResponseWriter, r *http.Request, deps Dependencies) bool {
	if !deps.AuthConfig.Enabled {
		return true
	}
	key, ok := middleware.APIKeyFromContext(r.Context())
	if !ok || !key.Scope.Allows(auth.ScopeWrite) {
		responses.WriteError(w, http.StatusForbidden, "forbidden", "write scope required")
		return false
	}
	return true
}

func requireAdminScope(w http.ResponseWriter, r *http.Request, deps Dependencies) bool {
	if !deps.AuthConfig.Enabled {
		return true
	}
	key, ok := middleware.APIKeyFromContext(r.Context())
	if !ok || !key.Scope.Allows(auth.ScopeAdmin) {
		responses.WriteError(w, http.StatusForbidden, "forbidden", "admin scope required")
		return false
	}
	return true
}
