package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

const artifactsBasePath = "/v0/artifacts"

// NewArtifactHandler returns the read-only CAP-002 artifact handler.
func NewArtifactHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg, err := config.Load()
		if err != nil {
			responses.WriteError(w, http.StatusInternalServerError, "config_load_failed", err.Error())
			return
		}

		if cfg.Mode != config.ModeLocal {
			responses.WriteError(w, http.StatusBadRequest, "unsupported_mode", fmt.Sprintf("invalid mode: %s", cfg.Mode))
			return
		}

		dbPath, err := config.EffectiveLocalDBPath(cfg)
		if err != nil {
			responses.WriteError(w, http.StatusInternalServerError, "db_path_resolution_failed", err.Error())
			return
		}

		db, err := yanzilibrary.InitDBAtPath(dbPath)
		if err != nil {
			responses.WriteError(w, http.StatusInternalServerError, "db_open_failed", err.Error())
			return
		}
		defer db.Close()

		readStore := yanzilibrary.NewArtifactReadStore(db)
		trimmedPath := strings.TrimPrefix(r.URL.Path, artifactsBasePath)
		if trimmedPath == "" || trimmedPath == "/" {
			handleArtifactList(w, r, readStore)
			return
		}

		if strings.Count(strings.Trim(trimmedPath, "/"), "/") > 0 {
			responses.WriteError(w, http.StatusNotFound, "not_found", "not found")
			return
		}

		handleArtifactGet(w, r, readStore, strings.Trim(trimmedPath, "/"))
	})
}

func handleArtifactList(w http.ResponseWriter, r *http.Request, readStore *yanzilibrary.ArtifactReadStore) {
	query, err := artifactReadQueryFromRequest(r)
	if err != nil {
		responses.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	intents, err := readStore.ListIntentRecords(r.Context(), query)
	if err != nil {
		responses.WriteError(w, http.StatusInternalServerError, "artifact_list_failed", err.Error())
		return
	}

	resp := models.ArtifactListResponse{Artifacts: make([]models.ArtifactSummary, 0, len(intents))}
	for _, intent := range intents {
		resp.Artifacts = append(resp.Artifacts, models.ArtifactSummary{
			ID:        intent.ID,
			CreatedAt: intent.CreatedAt,
			Project:   artifactProject(intent.Meta),
			Author:    intent.Author,
			Source:    intent.SourceType,
			Title:     intent.Title,
			Metadata:  artifactListMetadata(intent.Meta),
		})
	}

	responses.WriteJSON(w, http.StatusOK, resp)
}

func handleArtifactGet(w http.ResponseWriter, r *http.Request, readStore *yanzilibrary.ArtifactReadStore, id string) {
	intent, err := readStore.GetIntentRecord(r.Context(), id)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			responses.WriteError(w, http.StatusNotFound, "artifact_not_found", err.Error())
			return
		}
		responses.WriteError(w, http.StatusInternalServerError, "artifact_read_failed", err.Error())
		return
	}

	responses.WriteJSON(w, http.StatusOK, models.ArtifactResponse{
		Artifact: models.Artifact{
			ID:        intent.ID,
			CreatedAt: intent.CreatedAt,
			Project:   artifactProject(intent.Meta),
			Author:    intent.Author,
			Source:    intent.SourceType,
			Title:     intent.Title,
			Prompt:    intent.Prompt,
			Response:  intent.Response,
			Metadata:  artifactDetailMetadata(intent.Meta),
			PrevHash:  intent.PrevHash,
			Hash:      intent.Hash,
		},
	})
}

func artifactReadQueryFromRequest(r *http.Request) (yanzilibrary.ArtifactReadQuery, error) {
	values := r.URL.Query()
	metaFilters := map[string]string{}
	for _, raw := range values["meta"] {
		parts := strings.SplitN(raw, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
			return yanzilibrary.ArtifactReadQuery{}, fmt.Errorf("invalid meta filter: %s (expected key=value)", raw)
		}
		metaFilters[parts[0]] = parts[1]
	}

	if profile := strings.TrimSpace(values.Get("profile")); profile != "" {
		metaFilters["profile"] = profile
	}

	limit := 20
	if rawLimit := strings.TrimSpace(values.Get("limit")); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			return yanzilibrary.ArtifactReadQuery{}, fmt.Errorf("invalid limit: %s", rawLimit)
		}
		limit = parsedLimit
	}

	allProjects, err := boolQueryParam(values.Get("all-projects"))
	if err != nil {
		return yanzilibrary.ArtifactReadQuery{}, fmt.Errorf("invalid all-projects value: %s", values.Get("all-projects"))
	}

	includeDeleted, err := boolQueryParam(values.Get("include-deleted"))
	if err != nil {
		return yanzilibrary.ArtifactReadQuery{}, fmt.Errorf("invalid include-deleted value: %s", values.Get("include-deleted"))
	}

	if !allProjects {
		activeProject, err := loadAPIActiveProject()
		if err != nil {
			return yanzilibrary.ArtifactReadQuery{}, err
		}
		activeProject = strings.TrimSpace(activeProject)
		if activeProject == "" {
			return yanzilibrary.ArtifactReadQuery{}, fmt.Errorf("no active project set")
		}
		if explicitProject, ok := metaFilters["project"]; ok && strings.TrimSpace(explicitProject) != activeProject {
			return yanzilibrary.ArtifactReadQuery{}, fmt.Errorf("--meta project=%s conflicts with active project %s; use --all-projects for cross-project retrieval", strings.TrimSpace(explicitProject), activeProject)
		}
		metaFilters["project"] = activeProject
	}

	return yanzilibrary.ArtifactReadQuery{
		Author:         strings.TrimSpace(values.Get("author")),
		Source:         strings.TrimSpace(values.Get("source")),
		Limit:          limit,
		MetaFilters:    metaFilters,
		IncludeDeleted: includeDeleted,
	}, nil
}

func boolQueryParam(raw string) (bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false, nil
	}
	return strconv.ParseBool(raw)
}

func artifactProject(raw []byte) string {
	meta, err := yanzilibrary.DecodeArtifactReadMetadata(string(raw))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(meta["project"])
}

func artifactListMetadata(raw []byte) map[string]string {
	meta, err := yanzilibrary.DecodeArtifactReadMetadata(string(raw))
	if err != nil || len(meta) == 0 {
		return nil
	}
	filtered := make(map[string]string, len(meta))
	for key, value := range meta {
		if strings.TrimSpace(key) == "project" {
			continue
		}
		filtered[key] = value
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func artifactDetailMetadata(raw []byte) map[string]string {
	meta, err := yanzilibrary.DecodeArtifactReadMetadata(string(raw))
	if err != nil || len(meta) == 0 {
		return nil
	}
	return meta
}
