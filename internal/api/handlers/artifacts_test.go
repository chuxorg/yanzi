package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/core/hash"
	"github.com/chuxorg/yanzi/internal/core/model"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
	"github.com/chuxorg/yanzi/internal/storage"
	"github.com/chuxorg/yanzi/internal/storage/registry"
)

func TestArtifactCaptureEndpointCreatesReadableCapture(t *testing.T) {
	handler, dbPath := newArtifactTestHandler(t)
	payload := `{
		"author":"Ada",
		"source_type":"agent",
		"title":"Capture endpoint",
		"prompt":"What changed?",
		"response":"Added POST /v0/artifacts.",
		"metadata":{"area":"api"},
		"project":"alpha",
		"prev_hash":"previous"
	}`

	postRec := httptest.NewRecorder()
	handler.ServeHTTP(postRec, httptest.NewRequest(http.MethodPost, "/v0/artifacts", strings.NewReader(payload)))
	if postRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%q", postRec.Code, postRec.Body.String())
	}

	var created models.ArtifactCaptureResponse
	if err := json.Unmarshal(postRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created response: %v", err)
	}
	if created.ID == "" || created.Hash == "" || created.CreatedAt == "" {
		t.Fatalf("expected deterministic identity fields, got %+v", created)
	}
	if created.Author != "Ada" || created.SourceType != "agent" || created.Title != "Capture endpoint" {
		t.Fatalf("unexpected capture identity: %+v", created)
	}
	if created.Prompt != "What changed?" || created.Response != "Added POST /v0/artifacts." || created.PrevHash != "previous" {
		t.Fatalf("unexpected capture payload: %+v", created)
	}
	if created.Metadata["area"] != "api" || created.Metadata["project"] != "alpha" {
		t.Fatalf("unexpected metadata: %#v", created.Metadata)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	var class string
	var artifactType string
	var content string
	var metadataText string
	if err := db.QueryRow(`SELECT class, type, content, metadata FROM intents WHERE id = ?`, created.ID).Scan(&class, &artifactType, &content, &metadataText); err != nil {
		t.Fatalf("query capture: %v", err)
	}
	if class != "intent" || artifactType != "prompt" || content != created.Prompt {
		t.Fatalf("unexpected artifact-compatible columns: class=%q type=%q content=%q", class, artifactType, content)
	}
	if !strings.Contains(metadataText, `"project":"alpha"`) || !strings.Contains(metadataText, `"area":"api"`) {
		t.Fatalf("unexpected persisted metadata: %q", metadataText)
	}

	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/v0/artifacts/"+created.ID, nil))
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", getRec.Code, getRec.Body.String())
	}
	var fetched models.ArtifactCaptureResponse
	if err := json.Unmarshal(getRec.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("decode fetched response: %v", err)
	}
	if !reflect.DeepEqual(fetched, created) {
		t.Fatalf("read-after-write mismatch:\ncreated=%+v\nfetched=%+v", created, fetched)
	}

	record := model.IntentRecord{
		ID:         fetched.ID,
		CreatedAt:  fetched.CreatedAt,
		Author:     fetched.Author,
		SourceType: fetched.SourceType,
		Title:      fetched.Title,
		Prompt:     fetched.Prompt,
		Response:   fetched.Response,
		PrevHash:   fetched.PrevHash,
		Meta:       mustMarshalMetadata(t, fetched.Metadata),
	}
	computedHash, err := hash.HashIntent(record)
	if err != nil {
		t.Fatalf("HashIntent: %v", err)
	}
	if computedHash != fetched.Hash {
		t.Fatalf("expected deterministic hash %q, got %q", fetched.Hash, computedHash)
	}
}

func TestArtifactCaptureEndpointRejectsMalformedPayload(t *testing.T) {
	handler, _ := newArtifactTestHandler(t)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v0/artifacts", strings.NewReader(`{"author":`)))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "{\"error\":{\"code\":\"malformed_json\",\"message\":\"request body must be a valid artifact capture payload\"}}\n" {
		t.Fatalf("unexpected error body: %q", got)
	}
}

func TestArtifactCaptureEndpointRejectsValidationErrors(t *testing.T) {
	handler, _ := newArtifactTestHandler(t)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v0/artifacts", strings.NewReader(`{"prompt":"p","response":"r"}`)))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	var body responses.ErrorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body.Error.Code != "validation_failed" || body.Error.Message != "author is required" {
		t.Fatalf("unexpected validation error: %+v", body)
	}
}

func TestArtifactEndpointKeepsDeferredAndMutationBoundaries(t *testing.T) {
	handler, _ := newArtifactTestHandler(t)

	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, httptest.NewRequest(http.MethodGet, "/v0/artifacts", nil))
	if listRec.Code != http.StatusNotImplemented || !strings.Contains(listRec.Body.String(), "artifact list endpoints are deferred") {
		t.Fatalf("unexpected collection GET response: code=%d body=%q", listRec.Code, listRec.Body.String())
	}

	deleteRec := httptest.NewRecorder()
	handler.ServeHTTP(deleteRec, httptest.NewRequest(http.MethodDelete, "/v0/artifacts/example", nil))
	if deleteRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected delete to remain unavailable, got %d", deleteRec.Code)
	}
	if got := deleteRec.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("unexpected Allow header: %q", got)
	}
}

func newArtifactTestHandler(t *testing.T) (http.Handler, string) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "yanzi.db")
	deps := Dependencies{
		Version: "v0.0.0-test",
		LoadConfig: func() (config.Config, error) {
			return config.Config{Mode: config.ModeLocal, DBPath: dbPath}, nil
		},
		OpenProvider: func(ctx context.Context, cfg config.Config) (storage.Provider, error) {
			provider, _, err := registry.OpenAtPath(ctx, dbPath, registry.Options{Migrations: yanzilibrary.MigrationsFS()})
			return provider, err
		},
		Now: func() time.Time {
			return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		},
	}
	return NewArtifactHandler(deps), dbPath
}

func mustMarshalMetadata(t *testing.T, metadata map[string]string) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}
	return raw
}

func TestArtifactCaptureResponseShapeIsStable(t *testing.T) {
	handler, _ := newArtifactTestHandler(t)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v0/artifacts", bytes.NewBufferString(`{"author":"Ada","prompt":"p","response":"r"}`)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%q", rec.Code, rec.Body.String())
	}

	var fields map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &fields); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	for _, key := range []string{"id", "created_at", "author", "source_type", "prompt", "response", "hash"} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("missing response field %q in %s", key, rec.Body.String())
		}
	}
	if _, ok := fields["title"]; ok {
		t.Fatalf("did not expect empty title field in %s", rec.Body.String())
	}
	if fields["source_type"] != "cli" {
		t.Fatalf("expected default source_type cli, got %#v", fields["source_type"])
	}
}
