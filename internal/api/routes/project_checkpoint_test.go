package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/api/handlers"
	"github.com/chuxorg/yanzi/internal/api/models"
)

func TestProjectEndpointsCreateListAndCurrentFlow(t *testing.T) {
	home := apiTestHome(t)
	handler := NewHandler(handlers.Dependencies{Version: "v0.0.0-test"})

	current := apiJSONRequest(t, handler, http.MethodGet, "/v0/projects/current", nil)
	if current.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", current.Code)
	}
	var currentResp models.CurrentProjectResponse
	decodeAPIResponse(t, current.Body.Bytes(), &currentResp)
	if currentResp.Project != nil {
		t.Fatalf("expected no active project, got %+v", currentResp.Project)
	}

	create := apiJSONRequest(t, handler, http.MethodPost, "/v0/projects", models.ProjectCreateRequest{Name: "alpha"})
	if create.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%q", create.Code, create.Body.String())
	}
	var project models.Project
	decodeAPIResponse(t, create.Body.Bytes(), &project)
	if project.Name != "alpha" || project.CreatedAt == "" {
		t.Fatalf("unexpected created project: %+v", project)
	}

	list := apiJSONRequest(t, handler, http.MethodGet, "/v0/projects", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", list.Code)
	}
	var listResp models.ProjectListResponse
	decodeAPIResponse(t, list.Body.Bytes(), &listResp)
	if len(listResp.Projects) != 1 || listResp.Projects[0].Name != "alpha" {
		t.Fatalf("unexpected project list: %+v", listResp.Projects)
	}

	setCurrent := apiJSONRequest(t, handler, http.MethodPost, "/v0/projects/current", models.CurrentProjectRequest{Name: "alpha"})
	if setCurrent.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", setCurrent.Code, setCurrent.Body.String())
	}
	decodeAPIResponse(t, setCurrent.Body.Bytes(), &currentResp)
	if currentResp.Project == nil || currentResp.Project.Name != "alpha" {
		t.Fatalf("unexpected current project set response: %+v", currentResp.Project)
	}

	stateData, err := os.ReadFile(filepath.Join(home, ".yanzi", "state.json"))
	if err != nil {
		t.Fatalf("read state file: %v", err)
	}
	if !strings.Contains(string(stateData), "\"active_project\": \"alpha\"") {
		t.Fatalf("unexpected state file content: %q", string(stateData))
	}

	current = apiJSONRequest(t, handler, http.MethodGet, "/v0/projects/current", nil)
	if current.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", current.Code)
	}
	decodeAPIResponse(t, current.Body.Bytes(), &currentResp)
	if currentResp.Project == nil || currentResp.Project.Name != "alpha" {
		t.Fatalf("unexpected current project response: %+v", currentResp.Project)
	}
}

func TestProjectEndpointErrorsAreDeterministic(t *testing.T) {
	apiTestHome(t)
	handler := NewHandler(handlers.Dependencies{Version: "v0.0.0-test"})

	first := apiJSONRequest(t, handler, http.MethodPost, "/v0/projects", models.ProjectCreateRequest{Name: "alpha"})
	if first.Code != http.StatusCreated {
		t.Fatalf("expected initial create success, got %d", first.Code)
	}

	duplicate := apiJSONRequest(t, handler, http.MethodPost, "/v0/projects", models.ProjectCreateRequest{Name: "alpha"})
	if duplicate.Code != http.StatusConflict || !strings.Contains(duplicate.Body.String(), "\"code\":\"conflict\"") {
		t.Fatalf("unexpected duplicate response: code=%d body=%q", duplicate.Code, duplicate.Body.String())
	}

	invalid := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v0/projects", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(invalid, req)
	if invalid.Code != http.StatusBadRequest || !strings.Contains(invalid.Body.String(), "\"code\":\"invalid_request\"") {
		t.Fatalf("unexpected invalid JSON response: code=%d body=%q", invalid.Code, invalid.Body.String())
	}

	missing := apiJSONRequest(t, handler, http.MethodPost, "/v0/projects/current", models.CurrentProjectRequest{Name: "missing"})
	if missing.Code != http.StatusNotFound || !strings.Contains(missing.Body.String(), "project not found: missing") {
		t.Fatalf("unexpected missing project response: code=%d body=%q", missing.Code, missing.Body.String())
	}
}

func TestCheckpointEndpointsCreateListAndAllProjectsFlow(t *testing.T) {
	apiTestHome(t)
	handler := NewHandler(handlers.Dependencies{Version: "v0.0.0-test"})

	noActive := apiJSONRequest(t, handler, http.MethodGet, "/v0/checkpoints", nil)
	if noActive.Code != http.StatusBadRequest || !strings.Contains(noActive.Body.String(), "no active project set") {
		t.Fatalf("unexpected no-active checkpoint response: code=%d body=%q", noActive.Code, noActive.Body.String())
	}

	apiJSONRequest(t, handler, http.MethodPost, "/v0/projects", models.ProjectCreateRequest{Name: "alpha"})
	apiJSONRequest(t, handler, http.MethodPost, "/v0/projects/current", models.CurrentProjectRequest{Name: "alpha"})

	create := apiJSONRequest(t, handler, http.MethodPost, "/v0/checkpoints", models.CheckpointCreateRequest{Summary: "first checkpoint"})
	if create.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%q", create.Code, create.Body.String())
	}
	var checkpoint models.Checkpoint
	decodeAPIResponse(t, create.Body.Bytes(), &checkpoint)
	if checkpoint.Project != "alpha" || checkpoint.Summary != "first checkpoint" || checkpoint.Hash == "" {
		t.Fatalf("unexpected checkpoint response: %+v", checkpoint)
	}

	list := apiJSONRequest(t, handler, http.MethodGet, "/v0/checkpoints", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", list.Code)
	}
	var listResp models.CheckpointListResponse
	decodeAPIResponse(t, list.Body.Bytes(), &listResp)
	if len(listResp.Checkpoints) != 1 || listResp.Checkpoints[0].Hash != checkpoint.Hash {
		t.Fatalf("unexpected checkpoint list: %+v", listResp.Checkpoints)
	}

	apiJSONRequest(t, handler, http.MethodPost, "/v0/projects", models.ProjectCreateRequest{Name: "beta"})
	apiJSONRequest(t, handler, http.MethodPost, "/v0/projects/current", models.CurrentProjectRequest{Name: "beta"})
	apiJSONRequest(t, handler, http.MethodPost, "/v0/checkpoints", models.CheckpointCreateRequest{Summary: "beta checkpoint"})

	all := apiJSONRequest(t, handler, http.MethodGet, "/v0/checkpoints?all_projects=true", nil)
	if all.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", all.Code)
	}
	decodeAPIResponse(t, all.Body.Bytes(), &listResp)
	if len(listResp.Checkpoints) != 2 {
		t.Fatalf("expected 2 checkpoints across projects, got %+v", listResp.Checkpoints)
	}
	if listResp.Checkpoints[0].Project == "" || listResp.Checkpoints[1].Project == "" {
		t.Fatalf("expected project association on all-project checkpoints, got %+v", listResp.Checkpoints)
	}
}

func TestCheckpointEndpointErrorsAreDeterministic(t *testing.T) {
	apiTestHome(t)
	handler := NewHandler(handlers.Dependencies{Version: "v0.0.0-test"})

	apiJSONRequest(t, handler, http.MethodPost, "/v0/projects", models.ProjectCreateRequest{Name: "alpha"})
	apiJSONRequest(t, handler, http.MethodPost, "/v0/projects/current", models.CurrentProjectRequest{Name: "alpha"})

	missingSummary := apiJSONRequest(t, handler, http.MethodPost, "/v0/checkpoints", models.CheckpointCreateRequest{})
	if missingSummary.Code != http.StatusBadRequest || !strings.Contains(missingSummary.Body.String(), "\"code\":\"invalid_request\"") {
		t.Fatalf("unexpected missing summary response: code=%d body=%q", missingSummary.Code, missingSummary.Body.String())
	}

	mismatch := apiJSONRequest(t, handler, http.MethodPost, "/v0/checkpoints", models.CheckpointCreateRequest{
		Project: "beta",
		Summary: "bad checkpoint",
	})
	if mismatch.Code != http.StatusBadRequest || !strings.Contains(mismatch.Body.String(), "checkpoint project must match active project") {
		t.Fatalf("unexpected mismatched project response: code=%d body=%q", mismatch.Code, mismatch.Body.String())
	}
}

func apiTestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)

	configDir := filepath.Join(home, ".yanzi")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	dbPath := filepath.Join(home, "state", "yanzi.db")
	configData := []byte("mode: local\ndb_path: " + dbPath + "\n")
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), configData, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return home
}

func apiJSONRequest(t *testing.T, handler http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func decodeAPIResponse(t *testing.T, body []byte, out any) {
	t.Helper()
	if err := json.Unmarshal(body, out); err != nil {
		t.Fatalf("decode response %q: %v", string(body), err)
	}
}
