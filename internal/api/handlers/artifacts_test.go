package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/api/handlers"
	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
	cmdpkg "github.com/chuxorg/yanzi/internal/cmd"
	"github.com/chuxorg/yanzi/internal/core/hash"
	"github.com/chuxorg/yanzi/internal/core/model"
	_ "modernc.org/sqlite"
)

func TestArtifactHandlerListsScopedArtifactsWithDeterministicOrder(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withCwd(t, home)
	writeAPIHandlerTestConfig(t, home)

	if _, err := captureStdout(t, func() error { return cmdpkg.RunProject([]string{"create", "alpha"}) }); err != nil {
		t.Fatalf("create alpha: %v", err)
	}
	if _, err := captureStdout(t, func() error { return cmdpkg.RunProject([]string{"create", "beta"}) }); err != nil {
		t.Fatalf("create beta: %v", err)
	}
	if _, err := captureStdout(t, func() error { return cmdpkg.RunProject([]string{"use", "alpha"}) }); err != nil {
		t.Fatalf("use alpha: %v", err)
	}

	if _, err := captureStdout(t, func() error {
		return cmdpkg.RunCapture([]string{
			"--author", "alice",
			"--title", "Alpha Older",
			"--prompt", "older prompt",
			"--response", "older response",
			"--meta", "kind=note",
			"--meta", "profile=engineer",
		})
	}); err != nil {
		t.Fatalf("capture alpha older: %v", err)
	}
	if _, err := captureStdout(t, func() error {
		return cmdpkg.RunCapture([]string{
			"--author", "alice",
			"--title", "Alpha Newer",
			"--prompt", "newer prompt",
			"--response", "newer response",
			"--meta", "kind=note",
			"--meta", "profile=engineer",
		})
	}); err != nil {
		t.Fatalf("capture alpha newer: %v", err)
	}

	if _, err := captureStdout(t, func() error { return cmdpkg.RunProject([]string{"use", "beta"}) }); err != nil {
		t.Fatalf("use beta: %v", err)
	}
	if _, err := captureStdout(t, func() error {
		return cmdpkg.RunCapture([]string{
			"--author", "alice",
			"--title", "Beta Note",
			"--prompt", "beta prompt",
			"--response", "beta response",
			"--meta", "kind=note",
		})
	}); err != nil {
		t.Fatalf("capture beta: %v", err)
	}

	if _, err := captureStdout(t, func() error { return cmdpkg.RunProject([]string{"use", "alpha"}) }); err != nil {
		t.Fatalf("use alpha again: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v0/artifacts?author=alice&profile=engineer&meta=kind=note", nil)
	rec := httptest.NewRecorder()
	handlers.NewArtifactHandler(handlers.Dependencies{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", rec.Code, rec.Body.String())
	}

	var resp models.ArtifactListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(resp.Artifacts) != 2 {
		t.Fatalf("expected 2 alpha artifacts, got %d: %+v", len(resp.Artifacts), resp.Artifacts)
	}
	if resp.Artifacts[0].Title != "Alpha Newer" || resp.Artifacts[1].Title != "Alpha Older" {
		t.Fatalf("unexpected ordering: %+v", resp.Artifacts)
	}
	for _, artifact := range resp.Artifacts {
		if artifact.Project != "alpha" {
			t.Fatalf("expected scoped project alpha, got %+v", artifact)
		}
		if artifact.Metadata["project"] != "" {
			t.Fatalf("did not expect project in list metadata: %+v", artifact.Metadata)
		}
		if artifact.Metadata["profile"] != "engineer" || artifact.Metadata["kind"] != "note" {
			t.Fatalf("unexpected metadata: %+v", artifact.Metadata)
		}
	}
}

func TestArtifactHandlerGetsArtifactByID(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withCwd(t, home)
	writeAPIHandlerTestConfig(t, home)

	if _, err := captureStdout(t, func() error { return cmdpkg.RunProject([]string{"create", "alpha"}) }); err != nil {
		t.Fatalf("create alpha: %v", err)
	}
	if _, err := captureStdout(t, func() error { return cmdpkg.RunProject([]string{"use", "alpha"}) }); err != nil {
		t.Fatalf("use alpha: %v", err)
	}

	output, err := captureStdout(t, func() error {
		return cmdpkg.RunCapture([]string{
			"--author", "alice",
			"--title", "Artifact Detail",
			"--prompt", "detail prompt",
			"--response", "detail response",
			"--meta", "kind=note",
			"--meta", "profile=engineer",
		})
	})
	if err != nil {
		t.Fatalf("capture detail: %v", err)
	}
	id := parseCapturedID(t, output)

	req := httptest.NewRequest(http.MethodGet, "/v0/artifacts/"+id, nil)
	rec := httptest.NewRecorder()
	handlers.NewArtifactHandler(handlers.Dependencies{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", rec.Code, rec.Body.String())
	}

	var resp models.ArtifactResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode detail response: %v", err)
	}
	if resp.Artifact.ID != id || resp.Artifact.Title != "Artifact Detail" {
		t.Fatalf("unexpected artifact: %+v", resp.Artifact)
	}
	if resp.Artifact.Project != "alpha" || resp.Artifact.Author != "alice" || resp.Artifact.Source != "cli" {
		t.Fatalf("unexpected artifact fields: %+v", resp.Artifact)
	}
	if resp.Artifact.Prompt != "detail prompt" || resp.Artifact.Response != "detail response" || resp.Artifact.Hash == "" {
		t.Fatalf("unexpected prompt/response/hash: %+v", resp.Artifact)
	}
	if resp.Artifact.Metadata["project"] != "alpha" || resp.Artifact.Metadata["kind"] != "note" || resp.Artifact.Metadata["profile"] != "engineer" {
		t.Fatalf("unexpected metadata: %+v", resp.Artifact.Metadata)
	}
}

func TestArtifactHandlerReturnsNotFoundForMissingArtifact(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withCwd(t, home)
	writeAPIHandlerTestConfig(t, home)

	if _, err := captureStdout(t, func() error { return cmdpkg.RunProject([]string{"create", "alpha"}) }); err != nil {
		t.Fatalf("create alpha: %v", err)
	}
	if _, err := captureStdout(t, func() error { return cmdpkg.RunProject([]string{"use", "alpha"}) }); err != nil {
		t.Fatalf("use alpha: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v0/artifacts/missing-id", nil)
	rec := httptest.NewRecorder()
	handlers.NewArtifactHandler(handlers.Dependencies{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%q", rec.Code, rec.Body.String())
	}

	var resp responses.ErrorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error.Code != "artifact_not_found" || resp.Error.Message != "intent not found for ID missing-id" {
		t.Fatalf("unexpected error response: %+v", resp.Error)
	}
}

func TestArtifactCaptureEndpointCreatesReadableCapture(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeAPIHandlerTestConfig(t, home)

	handler := handlers.NewArtifactHandler(handlers.Dependencies{})
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

	db, err := sql.Open("sqlite", artifactDBPath(home))
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
	var fetched models.ArtifactResponse
	if err := json.Unmarshal(getRec.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("decode fetched response: %v", err)
	}
	if fetched.Artifact.ID != created.ID || fetched.Artifact.CreatedAt != created.CreatedAt {
		t.Fatalf("read-after-write id mismatch:\ncreated=%+v\nfetched=%+v", created, fetched)
	}
	if fetched.Artifact.Author != created.Author || fetched.Artifact.Source != created.SourceType || fetched.Artifact.Title != created.Title {
		t.Fatalf("read-after-write identity mismatch:\ncreated=%+v\nfetched=%+v", created, fetched)
	}
	if fetched.Artifact.Prompt != created.Prompt || fetched.Artifact.Response != created.Response || fetched.Artifact.PrevHash != created.PrevHash {
		t.Fatalf("read-after-write payload mismatch:\ncreated=%+v\nfetched=%+v", created, fetched)
	}
	if !reflect.DeepEqual(fetched.Artifact.Metadata, created.Metadata) {
		t.Fatalf("read-after-write metadata mismatch:\ncreated=%+v\nfetched=%+v", created, fetched)
	}
	if fetched.Artifact.Hash != created.Hash {
		t.Fatalf("read-after-write hash mismatch:\ncreated=%+v\nfetched=%+v", created, fetched)
	}

	record := model.IntentRecord{
		ID:         fetched.Artifact.ID,
		CreatedAt:  fetched.Artifact.CreatedAt,
		Author:     fetched.Artifact.Author,
		SourceType: fetched.Artifact.Source,
		Title:      fetched.Artifact.Title,
		Prompt:     fetched.Artifact.Prompt,
		Response:   fetched.Artifact.Response,
		PrevHash:   fetched.Artifact.PrevHash,
		Meta:       mustMarshalMetadata(t, fetched.Artifact.Metadata),
	}
	computedHash, err := hash.HashIntent(record)
	if err != nil {
		t.Fatalf("HashIntent: %v", err)
	}
	if computedHash != fetched.Artifact.Hash {
		t.Fatalf("expected deterministic hash %q, got %q", fetched.Artifact.Hash, computedHash)
	}
}

func TestArtifactCaptureEndpointRejectsMalformedPayload(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeAPIHandlerTestConfig(t, home)

	handler := handlers.NewArtifactHandler(handlers.Dependencies{})
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
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeAPIHandlerTestConfig(t, home)

	handler := handlers.NewArtifactHandler(handlers.Dependencies{})
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

func TestArtifactCaptureEndpointRejectsMutationMethods(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeAPIHandlerTestConfig(t, home)

	handler := handlers.NewArtifactHandler(handlers.Dependencies{})
	deleteRec := httptest.NewRecorder()
	handler.ServeHTTP(deleteRec, httptest.NewRequest(http.MethodDelete, "/v0/artifacts/example", nil))
	if deleteRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected delete to remain unavailable, got %d", deleteRec.Code)
	}
	if got := deleteRec.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("unexpected Allow header: %q", got)
	}
}

func TestArtifactCaptureResponseShapeIsStable(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeAPIHandlerTestConfig(t, home)

	handler := handlers.NewArtifactHandler(handlers.Dependencies{})
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

func writeAPIHandlerTestConfig(t *testing.T, home string) {
	t.Helper()

	stateDir := filepath.Join(home, ".yanzi")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatalf("create state dir: %v", err)
	}
	dbPath := filepath.Join(stateDir, "yanzi.db")
	configPath := filepath.Join(stateDir, "config.yaml")
	content := []byte("mode: local\ndb_path: " + dbPath + "\n")
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func withCwd(t *testing.T, dir string) {
	t.Helper()

	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prev)
	})
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	stdout := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = stdout
	}()

	runErr := fn()
	_ = writer.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, reader)
	_ = reader.Close()
	return buf.String(), runErr
}

func parseCapturedID(t *testing.T, output string) string {
	t.Helper()

	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "id: ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "id: "))
		}
	}
	t.Fatalf("capture output missing id: %q", output)
	return ""
}

func artifactDBPath(home string) string {
	return filepath.Join(home, ".yanzi", "yanzi.db")
}

func mustMarshalMetadata(t *testing.T, metadata map[string]string) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}
	return raw
}
