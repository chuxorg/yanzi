package routes

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/api/handlers"
	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/storage"
)

func TestNewHandlerRegistersOperationalRoutes(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})

	handler := NewHandler(handlers.Dependencies{
		Version: "v0.0.0-test",
		LoadConfig: func() (config.Config, error) {
			return config.Config{Mode: config.ModeLocal}, nil
		},
		OpenProvider: func(context.Context, config.Config) (storage.Provider, error) {
			return &routeStubProvider{
				health: storage.Health{Provider: storage.ProviderSQLite, Status: storage.HealthReady},
			}, nil
		},
	})

	healthReq := httptest.NewRequest(http.MethodGet, "/v0/health", nil)
	healthRec := httptest.NewRecorder()
	handler.ServeHTTP(healthRec, healthReq)
	if healthRec.Code != http.StatusOK || !strings.Contains(healthRec.Body.String(), "\"provider\"") {
		t.Fatalf("unexpected health response: code=%d body=%q", healthRec.Code, healthRec.Body.String())
	}

	rehydrateReq := httptest.NewRequest(http.MethodGet, "/v0/rehydrate", nil)
	rehydrateRec := httptest.NewRecorder()
	handler.ServeHTTP(rehydrateRec, rehydrateReq)
	if rehydrateRec.Code != http.StatusBadRequest || !strings.Contains(rehydrateRec.Body.String(), "\"active_project_not_set\"") {
		t.Fatalf("unexpected rehydrate response: code=%d body=%q", rehydrateRec.Code, rehydrateRec.Body.String())
	}

	artifactReq := httptest.NewRequest(http.MethodGet, "/v0/artifacts", nil)
	artifactRec := httptest.NewRecorder()
	handler.ServeHTTP(artifactRec, artifactReq)
	if artifactRec.Code == http.StatusNotFound {
		t.Fatalf("expected artifact route to be registered, got 404")
	}

	verifyReq := httptest.NewRequest(http.MethodGet, "/v0/verify/example", nil)
	verifyRec := httptest.NewRecorder()
	handler.ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code == http.StatusNotFound {
		t.Fatalf("expected verify route to be registered, got 404")
	}

	exportReq := httptest.NewRequest(http.MethodGet, "/v0/export/json?project=alpha", nil)
	exportRec := httptest.NewRecorder()
	handler.ServeHTTP(exportRec, exportReq)
	if exportRec.Code == http.StatusNotFound {
		t.Fatalf("expected export route to be registered, got 404")
	}

	deferredReq := httptest.NewRequest(http.MethodGet, "/v0/projects", nil)
	deferredRec := httptest.NewRecorder()
	handler.ServeHTTP(deferredRec, deferredReq)
	if deferredRec.Code != http.StatusNotImplemented || !strings.Contains(deferredRec.Body.String(), "\"status\":\"deferred\"") {
		t.Fatalf("unexpected deferred response: code=%d body=%q", deferredRec.Code, deferredRec.Body.String())
	}

	methodReq := httptest.NewRequest(http.MethodPost, "/v0/health", nil)
	methodRec := httptest.NewRecorder()
	handler.ServeHTTP(methodRec, methodReq)
	if methodRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", methodRec.Code)
	}
	if got := methodRec.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("unexpected allow header: %q", got)
	}
}

type routeStubProvider struct {
	health storage.Health
}

func (p *routeStubProvider) Name() storage.ProviderName { return p.health.Provider }
func (p *routeStubProvider) Health(context.Context) storage.Health {
	return p.health
}
func (p *routeStubProvider) SQLDB() *sql.DB     { return nil }
func (p *routeStubProvider) Close() error       { return nil }
func (p *routeStubProvider) Artifacts() bool    { return false }
func (p *routeStubProvider) Projects() bool     { return false }
func (p *routeStubProvider) Checkpoints() bool  { return false }
func (p *routeStubProvider) Verification() bool { return false }
func (p *routeStubProvider) ImportExport() bool { return false }
func (p *routeStubProvider) CreateArtifact(context.Context, storage.CreateArtifactInput) (storage.Artifact, error) {
	return storage.Artifact{}, nil
}
func (p *routeStubProvider) ListArtifacts(context.Context, storage.ArtifactQuery) ([]storage.Artifact, error) {
	return nil, nil
}
func (p *routeStubProvider) ListVisibleContextArtifacts(context.Context, storage.ContextArtifactQuery) ([]storage.Artifact, error) {
	return nil, nil
}
func (p *routeStubProvider) GetVisibleContextArtifact(context.Context, string, string) (storage.Artifact, error) {
	return storage.Artifact{}, nil
}
func (p *routeStubProvider) CreateProject(context.Context, storage.CreateProjectInput) (storage.Project, error) {
	return storage.Project{}, nil
}
func (p *routeStubProvider) ListProjects(context.Context) ([]storage.Project, error) {
	return nil, nil
}
func (p *routeStubProvider) ProjectExists(context.Context, string) (bool, error) {
	return false, nil
}
func (p *routeStubProvider) CreateCheckpoint(context.Context, storage.CreateCheckpointInput) (storage.Checkpoint, error) {
	return storage.Checkpoint{}, nil
}
func (p *routeStubProvider) ListCheckpoints(context.Context, string) ([]storage.Checkpoint, error) {
	return nil, nil
}
func (p *routeStubProvider) ListAllCheckpoints(context.Context) ([]storage.Checkpoint, error) {
	return nil, nil
}
func (p *routeStubProvider) GetVerificationIntent(context.Context, string) (storage.IntentRecord, error) {
	return storage.IntentRecord{}, nil
}
func (p *routeStubProvider) GetVerificationIntentByHash(context.Context, string) (storage.IntentRecord, error) {
	return storage.IntentRecord{}, nil
}
func (p *routeStubProvider) ListExportItems(context.Context, storage.ExportQuery) ([]storage.ExportItem, int, error) {
	return nil, 0, nil
}
