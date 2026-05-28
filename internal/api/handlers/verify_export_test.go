package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
	"github.com/chuxorg/yanzi/internal/storage"
	"github.com/chuxorg/yanzi/internal/storage/registry"
)

func TestVerifyEndpointReturnsDeterministicVerification(t *testing.T) {
	deps, dbPath := newVerifyExportTestDeps(t)
	verifyHandler := NewVerifyHandler(deps)
	record := seedCaptureForVerifyExport(t, dbPath, yanzilibrary.CaptureWriteInput{
		Author:   "Ada",
		Prompt:   "verify prompt",
		Response: "verify response",
		Project:  "alpha",
	})

	rec := httptest.NewRecorder()
	verifyHandler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v0/verify/"+record.ID, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", rec.Code, rec.Body.String())
	}

	var body models.VerifyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode verify response: %v", err)
	}
	if !body.Valid || body.ID != record.ID || body.StoredHash != record.Hash || body.ComputedHash != record.Hash {
		t.Fatalf("unexpected verify response: %+v", body)
	}
}

func TestVerifyEndpointSupportsIntentAliasAndMissingBehavior(t *testing.T) {
	deps, dbPath := newVerifyExportTestDeps(t)
	verifyHandler := NewVerifyHandler(deps)
	record := seedCaptureForVerifyExport(t, dbPath, yanzilibrary.CaptureWriteInput{
		Author:   "Ada",
		Prompt:   "alias prompt",
		Response: "alias response",
		Project:  "alpha",
	})

	aliasRec := httptest.NewRecorder()
	verifyHandler.ServeHTTP(aliasRec, httptest.NewRequest(http.MethodGet, "/v0/intents/"+record.ID+"/verify", nil))
	if aliasRec.Code != http.StatusOK {
		t.Fatalf("expected alias verify 200, got %d body=%q", aliasRec.Code, aliasRec.Body.String())
	}

	missingRec := httptest.NewRecorder()
	verifyHandler.ServeHTTP(missingRec, httptest.NewRequest(http.MethodGet, "/v0/verify/missing", nil))
	if missingRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%q", missingRec.Code, missingRec.Body.String())
	}
	var errBody responses.ErrorBody
	if err := json.Unmarshal(missingRec.Body.Bytes(), &errBody); err != nil {
		t.Fatalf("decode missing error: %v", err)
	}
	if errBody.Error.Code != "intent_not_found" {
		t.Fatalf("unexpected missing error: %+v", errBody)
	}
}

func TestChainEndpointPreservesLineageOrder(t *testing.T) {
	deps, dbPath := newVerifyExportTestDeps(t)
	chainHandler := NewVerifyHandler(deps)
	root := seedCaptureForVerifyExport(t, dbPath, yanzilibrary.CaptureWriteInput{
		Author:   "Ada",
		Prompt:   "root prompt",
		Response: "root response",
		Project:  "alpha",
	})
	child := seedCaptureForVerifyExport(t, dbPath, yanzilibrary.CaptureWriteInput{
		Author:   "Ada",
		Prompt:   "child prompt",
		Response: "child response",
		Project:  "alpha",
		PrevHash: root.Hash,
	})

	rec := httptest.NewRecorder()
	chainHandler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v0/chain/"+child.ID, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", rec.Code, rec.Body.String())
	}

	var body models.ChainResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode chain response: %v", err)
	}
	if body.HeadID != child.ID || body.Length != 2 {
		t.Fatalf("unexpected chain header: %+v", body)
	}
	if len(body.Intents) != 2 || body.Intents[0].ID != root.ID || body.Intents[1].ID != child.ID {
		t.Fatalf("unexpected chain ordering: %+v", body.Intents)
	}
}

func TestExportEndpointsReturnDeterministicProjectScopedFormats(t *testing.T) {
	deps, dbPath := newVerifyExportTestDeps(t)
	exportHandler := NewExportHandler(deps)
	alphaOne := seedCaptureForVerifyExport(t, dbPath, yanzilibrary.CaptureWriteInput{
		Author:   "Ada",
		Prompt:   "alpha prompt one",
		Response: "alpha response one",
		Project:  "alpha",
		Meta:     mustMarshalMetadata(t, map[string]string{"area": "auth"}),
	})
	_ = alphaOne
	seedCheckpointForVerifyExport(t, dbPath, "alpha", "checkpoint one")
	seedCaptureForVerifyExport(t, dbPath, yanzilibrary.CaptureWriteInput{
		Author:   "Ada",
		Prompt:   "alpha prompt two",
		Response: "alpha response two",
		Project:  "alpha",
		Meta:     mustMarshalMetadata(t, map[string]string{"area": "ops"}),
	})
	seedCaptureForVerifyExport(t, dbPath, yanzilibrary.CaptureWriteInput{
		Author:   "Bea",
		Prompt:   "beta prompt",
		Response: "beta response",
		Project:  "beta",
	})

	markdownRec := httptest.NewRecorder()
	exportHandler.ServeHTTP(markdownRec, httptest.NewRequest(http.MethodGet, "/v0/export/markdown?project=alpha", nil))
	if markdownRec.Code != http.StatusOK || !strings.Contains(markdownRec.Body.String(), "# Yanzi Agent Log") || !strings.Contains(markdownRec.Body.String(), "checkpoint one") || strings.Contains(markdownRec.Body.String(), "beta prompt") {
		t.Fatalf("unexpected markdown export: code=%d body=%q", markdownRec.Code, markdownRec.Body.String())
	}

	jsonReq := httptest.NewRequest(http.MethodGet, "/v0/export/json?project=alpha&meta_area=auth", nil)
	jsonRec := httptest.NewRecorder()
	exportHandler.ServeHTTP(jsonRec, jsonReq)
	if jsonRec.Code != http.StatusOK {
		t.Fatalf("expected json export 200, got %d body=%q", jsonRec.Code, jsonRec.Body.String())
	}
	var jsonBody map[string]any
	if err := json.Unmarshal(jsonRec.Body.Bytes(), &jsonBody); err != nil {
		t.Fatalf("decode json export: %v", err)
	}
	events, ok := jsonBody["events"].([]any)
	if !ok || len(events) != 1 {
		t.Fatalf("expected one filtered event, got %#v", jsonBody["events"])
	}
	if got := jsonBody["project"]; got != "alpha" {
		t.Fatalf("unexpected project scope: %#v", got)
	}

	jsonRecRepeat := httptest.NewRecorder()
	exportHandler.ServeHTTP(jsonRecRepeat, jsonReq)
	if jsonRec.Body.String() != jsonRecRepeat.Body.String() {
		t.Fatalf("expected deterministic json export bodies\nfirst=%q\nsecond=%q", jsonRec.Body.String(), jsonRecRepeat.Body.String())
	}

	htmlRec := httptest.NewRecorder()
	exportHandler.ServeHTTP(htmlRec, httptest.NewRequest(http.MethodGet, "/v0/export/html?project=alpha", nil))
	if htmlRec.Code != http.StatusOK || !strings.Contains(htmlRec.Body.String(), "<!doctype html>") || !strings.Contains(htmlRec.Body.String(), "Yanzi Agent Log") {
		t.Fatalf("unexpected html export: code=%d body=%q", htmlRec.Code, htmlRec.Body.String())
	}
}

func TestExportEndpointValidation(t *testing.T) {
	deps, _ := newVerifyExportTestDeps(t)
	exportHandler := NewExportHandler(deps)

	missingProject := httptest.NewRecorder()
	exportHandler.ServeHTTP(missingProject, httptest.NewRequest(http.MethodGet, "/v0/export/json", nil))
	if missingProject.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%q", missingProject.Code, missingProject.Body.String())
	}

	checkpointRec := httptest.NewRecorder()
	exportHandler.ServeHTTP(checkpointRec, httptest.NewRequest(http.MethodGet, "/v0/export/json?project=alpha&checkpoint=latest", nil))
	if checkpointRec.Code != http.StatusBadRequest {
		t.Fatalf("expected checkpoint validation 400, got %d body=%q", checkpointRec.Code, checkpointRec.Body.String())
	}
}

func newVerifyExportTestDeps(t *testing.T) (Dependencies, string) {
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
			return time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
		},
	}
	return deps, dbPath
}

func seedCaptureForVerifyExport(t *testing.T, dbPath string, input yanzilibrary.CaptureWriteInput) models.ArtifactCaptureResponse {
	t.Helper()
	if strings.TrimSpace(input.SourceType) == "" {
		input.SourceType = "cli"
	}
	provider, _, err := registry.OpenAtPath(context.Background(), dbPath, registry.Options{Migrations: yanzilibrary.MigrationsFS()})
	if err != nil {
		t.Fatalf("open provider: %v", err)
	}
	defer func() {
		_ = provider.Close()
	}()

	ensureProjectForVerifyExport(t, provider, input.Project)
	record, err := yanzilibrary.NewArtifactWriteStore(provider.SQLDB()).CreateCapture(context.Background(), input)
	if err != nil {
		t.Fatalf("CreateCapture: %v", err)
	}
	response, err := artifactCaptureResponse(record)
	if err != nil {
		t.Fatalf("artifactCaptureResponse: %v", err)
	}
	return response
}

func seedCheckpointForVerifyExport(t *testing.T, dbPath, project, summary string) {
	t.Helper()
	provider, _, err := registry.OpenAtPath(context.Background(), dbPath, registry.Options{Migrations: yanzilibrary.MigrationsFS()})
	if err != nil {
		t.Fatalf("open provider: %v", err)
	}
	defer func() {
		_ = provider.Close()
	}()

	ensureProjectForVerifyExport(t, provider, project)
	if _, err := provider.CreateCheckpoint(context.Background(), storage.CreateCheckpointInput{
		Project:     project,
		Summary:     summary,
		ArtifactIDs: []string{},
	}); err != nil {
		t.Fatalf("CreateCheckpoint: %v", err)
	}
}

func ensureProjectForVerifyExport(t *testing.T, provider storage.Provider, project string) {
	t.Helper()
	project = strings.TrimSpace(project)
	if project == "" {
		return
	}
	exists, err := provider.ProjectExists(context.Background(), project)
	if err != nil {
		t.Fatalf("ProjectExists: %v", err)
	}
	if exists {
		return
	}
	if _, err := provider.CreateProject(context.Background(), storage.CreateProjectInput{Name: project}); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
}
