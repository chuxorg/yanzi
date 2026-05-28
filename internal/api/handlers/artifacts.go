package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/core/model"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
	"github.com/chuxorg/yanzi/internal/storage"
)

const artifactsPath = "/v0/artifacts"

// NewArtifactHandler returns the current artifact capture/read API handler.
func NewArtifactHandler(deps Dependencies) http.Handler {
	deps = deps.withDefaults()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg, err := deps.LoadConfig()
		if err != nil {
			responses.WriteError(w, http.StatusInternalServerError, "config_load_failed", err.Error())
			return
		}
		if cfg.Mode != config.ModeLocal {
			responses.WriteError(w, http.StatusBadRequest, "unsupported_mode", fmt.Sprintf("invalid mode: %s", cfg.Mode))
			return
		}

		switch {
		case r.URL.Path == artifactsPath:
			handleArtifactCollection(w, r, deps, cfg)
		case strings.HasPrefix(r.URL.Path, artifactsPath+"/"):
			handleArtifactDetail(w, r, deps, cfg)
		default:
			http.NotFound(w, r)
		}
	})
}

func handleArtifactCollection(w http.ResponseWriter, r *http.Request, deps Dependencies, cfg config.Config) {
	switch r.Method {
	case http.MethodGet:
		readStore, closer, err := deps.OpenArtifactReadStore(r.Context(), cfg)
		if err != nil {
			responses.WriteError(w, http.StatusInternalServerError, "artifact_read_open_failed", err.Error())
			return
		}
		defer func() {
			_ = closer.Close()
		}()

		query, err := artifactReadQueryFromRequest(r, deps.LoadActiveProject)
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
	case http.MethodPost:
		createArtifactCapture(w, r, deps, cfg)
	default:
		w.Header().Set("Allow", strings.Join([]string{http.MethodGet, http.MethodPost}, ", "))
		responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	}
}

func handleArtifactDetail(w http.ResponseWriter, r *http.Request, deps Dependencies, cfg config.Config) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, artifactsPath+"/")
	if strings.TrimSpace(id) == "" || strings.Contains(id, "/") {
		responses.WriteError(w, http.StatusNotFound, "artifact_not_found", "artifact not found")
		return
	}

	readStore, closer, err := deps.OpenArtifactReadStore(r.Context(), cfg)
	if err != nil {
		responses.WriteError(w, http.StatusInternalServerError, "artifact_read_open_failed", err.Error())
		return
	}
	defer func() {
		_ = closer.Close()
	}()

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

func createArtifactCapture(w http.ResponseWriter, r *http.Request, deps Dependencies, cfg config.Config) {
	var req models.ArtifactCaptureRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		responses.WriteError(w, http.StatusBadRequest, "malformed_json", "request body must be a valid artifact capture payload")
		return
	}
	if err := ensureJSONEOF(dec); err != nil {
		responses.WriteError(w, http.StatusBadRequest, "malformed_json", "request body must contain a single JSON object")
		return
	}
	if strings.TrimSpace(req.Author) == "" {
		responses.WriteError(w, http.StatusBadRequest, "validation_failed", "author is required")
		return
	}
	if req.Prompt == "" {
		responses.WriteError(w, http.StatusBadRequest, "validation_failed", "prompt is required")
		return
	}
	if req.Response == "" {
		responses.WriteError(w, http.StatusBadRequest, "validation_failed", "response is required")
		return
	}

	meta, err := encodeCaptureMetadata(req.Metadata)
	if err != nil {
		responses.WriteError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}
	sourceType := req.SourceType
	if strings.TrimSpace(sourceType) == "" {
		sourceType = "cli"
	}

	record, err := writeArtifactCapture(r.Context(), deps, yanzilibrary.CaptureWriteInput{
		Author:     req.Author,
		SourceType: sourceType,
		Title:      req.Title,
		Prompt:     req.Prompt,
		Response:   req.Response,
		Meta:       meta,
		Project:    req.Project,
		PrevHash:   req.PrevHash,
	})
	if err != nil {
		responses.WriteError(w, http.StatusInternalServerError, "artifact_capture_failed", err.Error())
		return
	}

	response, err := artifactCaptureResponse(record)
	if err != nil {
		responses.WriteError(w, http.StatusInternalServerError, "artifact_response_failed", err.Error())
		return
	}
	responses.WriteJSON(w, http.StatusCreated, response)
}

func writeArtifactCapture(ctx context.Context, deps Dependencies, input yanzilibrary.CaptureWriteInput) (model.IntentRecord, error) {
	provider, err := openArtifactProvider(ctx, deps)
	if err != nil {
		return model.IntentRecord{}, err
	}
	defer func() {
		_ = provider.Close()
	}()
	store := yanzilibrary.NewArtifactWriteStore(provider.SQLDB())
	return store.CreateCapture(ctx, input)
}

func openArtifactProvider(ctx context.Context, deps Dependencies) (storage.Provider, error) {
	cfg, err := deps.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	providerCfg := cfg
	providerCfg.Mode = config.ModeLocal
	provider, err := deps.OpenProvider(ctx, providerCfg)
	if err != nil {
		return nil, err
	}
	if provider.SQLDB() == nil {
		_ = provider.Close()
		return nil, errors.New("storage provider returned nil database")
	}
	return provider, nil
}

func encodeCaptureMetadata(metadata map[string]string) (json.RawMessage, error) {
	if len(metadata) == 0 {
		return nil, nil
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("encode metadata: %w", err)
	}
	return encoded, nil
}

func artifactCaptureResponse(record model.IntentRecord) (models.ArtifactCaptureResponse, error) {
	metadata, err := decodeCaptureMetadata(record.Meta)
	if err != nil {
		return models.ArtifactCaptureResponse{}, err
	}
	return models.ArtifactCaptureResponse{
		ID:         record.ID,
		CreatedAt:  record.CreatedAt,
		Author:     record.Author,
		SourceType: record.SourceType,
		Title:      record.Title,
		Prompt:     record.Prompt,
		Response:   record.Response,
		Metadata:   metadata,
		PrevHash:   record.PrevHash,
		Hash:       record.Hash,
	}, nil
}

func decodeCaptureMetadata(raw json.RawMessage) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var metadata map[string]string
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return nil, fmt.Errorf("decode metadata: %w", err)
	}
	if len(metadata) == 0 {
		return nil, nil
	}
	return metadata, nil
}

func ensureJSONEOF(dec *json.Decoder) error {
	var extra any
	if err := dec.Decode(&extra); err == nil {
		return errors.New("unexpected trailing JSON")
	} else if !errors.Is(err, io.EOF) {
		return errors.New("unexpected trailing JSON")
	}
	return nil
}

func artifactReadQueryFromRequest(r *http.Request, loadActiveProject ActiveProjectLoadFunc) (yanzilibrary.ArtifactReadQuery, error) {
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
		activeProject, err := loadActiveProject()
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
