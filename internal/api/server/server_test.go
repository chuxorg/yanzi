package server

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/api/handlers"
	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/storage"
)

func TestNewLocalWiresRouteFoundation(t *testing.T) {
	srv := NewLocal(LocalOptions{
		Addr:    "127.0.0.1:0",
		Version: "v0.0.0-test",
		Dependencies: handlers.Dependencies{
			LoadConfig: func() (config.Config, error) {
				return config.Config{Mode: config.ModeLocal}, nil
			},
			OpenProvider: func(context.Context, config.Config) (storage.Provider, error) {
				return &serverStubProvider{
					health: storage.Health{Provider: storage.ProviderSQLite, Status: storage.HealthReady},
				}, nil
			},
		},
	})

	if srv == nil || srv.Handler() == nil || srv.HTTPServer() == nil {
		t.Fatalf("expected constructed server, got %#v", srv)
	}
	if srv.HTTPServer().Addr != "127.0.0.1:0" {
		t.Fatalf("unexpected addr: %q", srv.HTTPServer().Addr)
	}

	req := httptest.NewRequest(http.MethodGet, "/v0/checkpoints/example", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented || !strings.Contains(rec.Body.String(), "checkpoints endpoints are deferred") {
		t.Fatalf("unexpected routed response: code=%d body=%q", rec.Code, rec.Body.String())
	}
}

type serverStubProvider struct {
	health storage.Health
}

func (p *serverStubProvider) Name() storage.ProviderName { return p.health.Provider }
func (p *serverStubProvider) Health(context.Context) storage.Health {
	return p.health
}
func (p *serverStubProvider) SQLDB() *sql.DB     { return nil }
func (p *serverStubProvider) Close() error       { return nil }
func (p *serverStubProvider) Artifacts() bool    { return false }
func (p *serverStubProvider) Projects() bool     { return false }
func (p *serverStubProvider) Checkpoints() bool  { return false }
func (p *serverStubProvider) Verification() bool { return false }
func (p *serverStubProvider) ImportExport() bool { return false }
func (p *serverStubProvider) CreateArtifact(context.Context, storage.CreateArtifactInput) (storage.Artifact, error) {
	return storage.Artifact{}, nil
}
func (p *serverStubProvider) ListArtifacts(context.Context, storage.ArtifactQuery) ([]storage.Artifact, error) {
	return nil, nil
}
func (p *serverStubProvider) ListVisibleContextArtifacts(context.Context, storage.ContextArtifactQuery) ([]storage.Artifact, error) {
	return nil, nil
}
func (p *serverStubProvider) GetVisibleContextArtifact(context.Context, string, string) (storage.Artifact, error) {
	return storage.Artifact{}, nil
}
func (p *serverStubProvider) CreateProject(context.Context, storage.CreateProjectInput) (storage.Project, error) {
	return storage.Project{}, nil
}
func (p *serverStubProvider) ListProjects(context.Context) ([]storage.Project, error) {
	return nil, nil
}
func (p *serverStubProvider) ProjectExists(context.Context, string) (bool, error) {
	return false, nil
}
func (p *serverStubProvider) CreateCheckpoint(context.Context, storage.CreateCheckpointInput) (storage.Checkpoint, error) {
	return storage.Checkpoint{}, nil
}
func (p *serverStubProvider) ListCheckpoints(context.Context, string) ([]storage.Checkpoint, error) {
	return nil, nil
}
func (p *serverStubProvider) ListAllCheckpoints(context.Context) ([]storage.Checkpoint, error) {
	return nil, nil
}
func (p *serverStubProvider) GetVerificationIntent(context.Context, string) (storage.IntentRecord, error) {
	return storage.IntentRecord{}, nil
}
func (p *serverStubProvider) GetVerificationIntentByHash(context.Context, string) (storage.IntentRecord, error) {
	return storage.IntentRecord{}, nil
}
func (p *serverStubProvider) ListExportItems(context.Context, storage.ExportQuery) ([]storage.ExportItem, int, error) {
	return nil, 0, nil
}
