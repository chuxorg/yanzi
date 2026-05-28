package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/storage"
)

func TestHealthHandlerReportsProviderHealth(t *testing.T) {
	provider := &stubProvider{
		health: storage.Health{
			Provider: storage.ProviderSQLite,
			Status:   storage.HealthReady,
		},
	}
	handler := NewHealthHandler(Dependencies{
		Version: "v0.0.0-test",
		LoadConfig: func() (config.Config, error) {
			return config.Config{Mode: config.ModeLocal}, nil
		},
		OpenProvider: func(context.Context, config.Config) (storage.Provider, error) {
			return provider, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/v0/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("unexpected content type: %q", got)
	}

	var resp models.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if resp.Version != "v0.0.0-test" || resp.Mode != string(config.ModeLocal) {
		t.Fatalf("unexpected version/mode: %+v", resp)
	}
	if resp.Provider.Name != string(storage.ProviderSQLite) || resp.Provider.Status != string(storage.HealthReady) || resp.Provider.Error != "" {
		t.Fatalf("unexpected provider health: %+v", resp.Provider)
	}
	if !provider.closed {
		t.Fatal("expected provider to be closed")
	}
}

func TestHealthHandlerReturnsConfigLoadError(t *testing.T) {
	handler := NewHealthHandler(Dependencies{
		LoadConfig: func() (config.Config, error) {
			return config.Config{}, errors.New("bad config")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/v0/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var resp responses.ErrorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error.Code != "config_load_failed" || resp.Error.Message != "bad config" {
		t.Fatalf("unexpected error response: %+v", resp)
	}
}

type stubProvider struct {
	health storage.Health
	closed bool
	db     *sql.DB
}

func (p *stubProvider) Name() storage.ProviderName { return p.health.Provider }
func (p *stubProvider) Health(context.Context) storage.Health {
	return p.health
}
func (p *stubProvider) SQLDB() *sql.DB { return p.db }
func (p *stubProvider) Close() error {
	p.closed = true
	return nil
}
func (p *stubProvider) Artifacts() bool    { return false }
func (p *stubProvider) Projects() bool     { return false }
func (p *stubProvider) Checkpoints() bool  { return false }
func (p *stubProvider) Verification() bool { return false }
func (p *stubProvider) ImportExport() bool { return false }
func (p *stubProvider) CreateArtifact(context.Context, storage.CreateArtifactInput) (storage.Artifact, error) {
	return storage.Artifact{}, errors.New("not implemented")
}
func (p *stubProvider) ListArtifacts(context.Context, storage.ArtifactQuery) ([]storage.Artifact, error) {
	return nil, errors.New("not implemented")
}
func (p *stubProvider) ListVisibleContextArtifacts(context.Context, storage.ContextArtifactQuery) ([]storage.Artifact, error) {
	return nil, errors.New("not implemented")
}
func (p *stubProvider) GetVisibleContextArtifact(context.Context, string, string) (storage.Artifact, error) {
	return storage.Artifact{}, errors.New("not implemented")
}
func (p *stubProvider) CreateProject(context.Context, storage.CreateProjectInput) (storage.Project, error) {
	return storage.Project{}, errors.New("not implemented")
}
func (p *stubProvider) ListProjects(context.Context) ([]storage.Project, error) {
	return nil, errors.New("not implemented")
}
func (p *stubProvider) ProjectExists(context.Context, string) (bool, error) {
	return false, errors.New("not implemented")
}
func (p *stubProvider) CreateCheckpoint(context.Context, storage.CreateCheckpointInput) (storage.Checkpoint, error) {
	return storage.Checkpoint{}, errors.New("not implemented")
}
func (p *stubProvider) ListCheckpoints(context.Context, string) ([]storage.Checkpoint, error) {
	return nil, errors.New("not implemented")
}
func (p *stubProvider) ListAllCheckpoints(context.Context) ([]storage.Checkpoint, error) {
	return nil, errors.New("not implemented")
}
func (p *stubProvider) GetVerificationIntent(context.Context, string) (storage.IntentRecord, error) {
	return storage.IntentRecord{}, errors.New("not implemented")
}
func (p *stubProvider) GetVerificationIntentByHash(context.Context, string) (storage.IntentRecord, error) {
	return storage.IntentRecord{}, errors.New("not implemented")
}
func (p *stubProvider) ListExportItems(context.Context, storage.ExportQuery) ([]storage.ExportItem, int, error) {
	return nil, 0, errors.New("not implemented")
}
