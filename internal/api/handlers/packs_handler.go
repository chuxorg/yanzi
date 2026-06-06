package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/api/responses"
	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/packs"
	"github.com/chuxorg/yanzi/internal/storage"
)

const packsPath = "/v0/packs"

// NewPacksHandler returns the /v0/packs handler.
func NewPacksHandler(deps Dependencies) http.Handler {
	deps = deps.withDefaults()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg, err := deps.LoadConfig()
		if err != nil {
			responses.WriteError(w, http.StatusInternalServerError, "config_load_failed", err.Error())
			return
		}
		if cfg.Mode != config.ModeLocal {
			responses.WriteError(w, http.StatusBadRequest, "unsupported_mode", "packs require local mode")
			return
		}
		if deps.PackStore == nil {
			responses.WriteError(w, http.StatusInternalServerError, "not_available", "pack store not initialized")
			return
		}

		switch {
		case r.URL.Path == packsPath+"/compose":
			handlePackCompose(w, r, deps)
		case r.URL.Path == packsPath:
			handlePacksCollection(w, r, deps, cfg)
		case strings.HasPrefix(r.URL.Path, packsPath+"/"):
			handlePackDetail(w, r, deps, cfg)
		default:
			http.NotFound(w, r)
		}
	})
}

func handlePacksCollection(w http.ResponseWriter, r *http.Request, deps Dependencies, cfg config.Config) {
	switch r.Method {
	case http.MethodGet:
		filter := packs.PackFilter{
			Name: r.URL.Query().Get("name"),
		}
		result, err := deps.PackStore.ListPacks(r.Context(), filter)
		if err != nil {
			log.Printf("packs: list error: %v", err)
			responses.WriteError(w, http.StatusInternalServerError, "list_failed", "failed to list packs")
			return
		}
		if result == nil {
			result = []packs.Pack{}
		}
		responses.WriteJSON(w, http.StatusOK, map[string]any{"packs": result})

	case http.MethodPost:
		if !requireWriteScope(w, r, deps) {
			return
		}
		var pack packs.Pack
		if !decodePackBody(w, r, &pack) {
			return
		}
		if strings.TrimSpace(pack.Name) == "" {
			responses.WriteError(w, http.StatusBadRequest, "invalid_request", "name is required")
			return
		}
		if pack.Role == 0 {
			pack.Role = packs.RoleObserver
		}
		for _, ref := range pack.Seeds {
			if strings.TrimSpace(ref.ArtifactID) == "" {
				responses.WriteError(w, http.StatusBadRequest, "invalid_request", "each seed reference must have an artifact_id")
				return
			}
		}
		if pack.ExtendsID != "" {
			if _, err := deps.PackStore.GetPack(r.Context(), pack.ExtendsID); err != nil {
				responses.WriteError(w, http.StatusBadRequest, "invalid_request", "extends_id references unknown pack")
				return
			}
		}
		pack.CreatedAt = time.Now().UTC()

		artifactID, err := createBackingArtifact(r.Context(), deps, cfg, storage.CreateArtifactInput{
			Class: "pack",
			Type:  "composition",
			Title: pack.Name,
		})
		if err != nil {
			log.Printf("packs: create backing artifact error: %v", err)
			responses.WriteError(w, http.StatusInternalServerError, "artifact_create_failed", "failed to create backing artifact")
			return
		}
		pack.ArtifactID = artifactID

		created, err := deps.PackStore.CreatePack(r.Context(), pack)
		if err != nil {
			log.Printf("packs: create error: %v", err)
			responses.WriteError(w, http.StatusInternalServerError, "create_failed", "failed to create pack")
			return
		}
		responses.WriteJSON(w, http.StatusCreated, created)

	default:
		responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	}
}

func handlePackDetail(w http.ResponseWriter, r *http.Request, deps Dependencies, cfg config.Config) {
	_ = cfg
	id := strings.TrimPrefix(r.URL.Path, packsPath+"/")
	id = strings.TrimSpace(id)
	if id == "" {
		responses.WriteError(w, http.StatusBadRequest, "invalid_request", "pack id is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		pack, err := deps.PackStore.GetPack(r.Context(), id)
		if err != nil {
			if errors.Is(err, packs.ErrNotFound) {
				responses.WriteError(w, http.StatusNotFound, "not_found", "pack not found")
				return
			}
			log.Printf("packs: get error: %v", err)
			responses.WriteError(w, http.StatusInternalServerError, "get_failed", "failed to get pack")
			return
		}
		responses.WriteJSON(w, http.StatusOK, pack)

	case http.MethodDelete:
		if !requireAdminScope(w, r, deps) {
			return
		}
		if err := deps.PackStore.DeletePack(r.Context(), id); err != nil {
			if errors.Is(err, packs.ErrNotFound) {
				responses.WriteError(w, http.StatusNotFound, "not_found", "pack not found")
				return
			}
			log.Printf("packs: delete error: %v", err)
			responses.WriteError(w, http.StatusInternalServerError, "delete_failed", "failed to delete pack")
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	}
}

func handlePackCompose(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	if r.Method != http.MethodPost {
		responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	if !requireWriteScope(w, r, deps) {
		return
	}

	var req packs.ComposeRequest
	if !decodePackBody(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.PackArtifactID) == "" {
		responses.WriteError(w, http.StatusBadRequest, "invalid_request", "pack_artifact_id is required")
		return
	}

	composer := packs.NewComposer(deps.PackStore)
	result, err := composer.Compose(r.Context(), req)
	if err != nil {
		if errors.Is(err, packs.ErrNotFound) {
			responses.WriteError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		log.Printf("packs: compose error: %v", err)
		responses.WriteError(w, http.StatusInternalServerError, "compose_failed", err.Error())
		return
	}

	// Record token usage.
	project := strings.TrimSpace(r.URL.Query().Get("project"))
	if project == "" {
		project, _ = deps.LoadActiveProject()
	}
	if project != "" && result.TokenEstimate.Total > 0 {
		usage := packs.TokenUsage{
			Project:    project,
			Phase:      req.TaskContent,
			PackID:     req.PackArtifactID,
			TokenCount: result.TokenEstimate.Total,
			Approximate: true,
			ModelHint:  req.ModelHint,
		}
		if err := deps.PackStore.RecordTokenUsage(r.Context(), usage); err != nil {
			log.Printf("packs: record token usage error: %v", err)
		}
	}

	responses.WriteJSON(w, http.StatusOK, result)
}
