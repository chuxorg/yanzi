package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
		switch {
		case r.URL.Path == artifactsPath:
			handleArtifactCollection(w, r, deps)
		case strings.HasPrefix(r.URL.Path, artifactsPath+"/"):
			handleArtifactDetail(w, r, deps)
		default:
			http.NotFound(w, r)
		}
	})
}

func handleArtifactCollection(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	switch r.Method {
	case http.MethodPost:
		createArtifactCapture(w, r, deps)
	case http.MethodGet:
		responses.WriteJSON(w, http.StatusNotImplemented, models.StatusResponse{
			Status:  "deferred",
			Message: "artifact list endpoints are deferred beyond CAP-002 Phase 6",
		})
	default:
		w.Header().Set("Allow", strings.Join([]string{http.MethodGet, http.MethodPost}, ", "))
		responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	}
}

func handleArtifactDetail(w http.ResponseWriter, r *http.Request, deps Dependencies) {
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

	record, err := readArtifactCapture(r.Context(), deps, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			responses.WriteError(w, http.StatusNotFound, "artifact_not_found", fmt.Sprintf("artifact not found: %s", id))
			return
		}
		responses.WriteError(w, http.StatusInternalServerError, "artifact_read_failed", err.Error())
		return
	}
	response, err := artifactCaptureResponse(record)
	if err != nil {
		responses.WriteError(w, http.StatusInternalServerError, "artifact_response_failed", err.Error())
		return
	}
	responses.WriteJSON(w, http.StatusOK, response)
}

func createArtifactCapture(w http.ResponseWriter, r *http.Request, deps Dependencies) {
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

func readArtifactCapture(ctx context.Context, deps Dependencies, id string) (model.IntentRecord, error) {
	provider, err := openArtifactProvider(ctx, deps)
	if err != nil {
		return model.IntentRecord{}, err
	}
	defer func() {
		_ = provider.Close()
	}()
	store := yanzilibrary.NewArtifactReadStore(provider.SQLDB())
	return store.GetIntentRecord(ctx, id)
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
