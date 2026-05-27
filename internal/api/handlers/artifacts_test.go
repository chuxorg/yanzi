package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
	cmdpkg "github.com/chuxorg/yanzi/internal/cmd"
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
	NewArtifactHandler(Dependencies{}).ServeHTTP(rec, req)

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
	NewArtifactHandler(Dependencies{}).ServeHTTP(rec, req)

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
	NewArtifactHandler(Dependencies{}).ServeHTTP(rec, req)

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
