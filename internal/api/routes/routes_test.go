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
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
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
		CreateProject: func(name, description string) (*yanzilibrary.Project, error) {
			return &yanzilibrary.Project{Name: name, Description: description}, nil
		},
		ListProjects: func() ([]yanzilibrary.Project, error) {
			return []yanzilibrary.Project{{Name: "alpha"}}, nil
		},
		ProjectExists: func(name string) (bool, error) {
			return name == "alpha", nil
		},
		LoadActiveProject: func() (string, error) {
			return "alpha", nil
		},
		SaveActiveProject: func(string) error {
			return nil
		},
		CreateCheckpoint: func(project, summary string, artifactIDs []string) (yanzilibrary.Checkpoint, error) {
			return yanzilibrary.Checkpoint{Project: project, Summary: summary, ArtifactIDs: artifactIDs}, nil
		},
		ListCheckpoints: func(string) ([]yanzilibrary.Checkpoint, error) {
			return []yanzilibrary.Checkpoint{}, nil
		},
		ListAllCheckpoints: func() ([]yanzilibrary.Checkpoint, error) {
			return []yanzilibrary.Checkpoint{}, nil
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
	if rehydrateRec.Code == http.StatusNotFound {
		t.Fatalf("expected rehydrate route to be registered, got 404")
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

	exportReq := httptest.NewRequest(http.MethodGet, "/v0/export/json", nil)
	exportRec := httptest.NewRecorder()
	handler.ServeHTTP(exportRec, exportReq)
	if exportRec.Code == http.StatusNotFound {
		t.Fatalf("expected export route to be registered, got 404")
	}

	projectsReq := httptest.NewRequest(http.MethodGet, "/v0/projects", nil)
	projectsRec := httptest.NewRecorder()
	handler.ServeHTTP(projectsRec, projectsReq)
	if projectsRec.Code == http.StatusNotFound {
		t.Fatalf("expected projects route to be registered, got 404")
	}

	projectCurrentReq := httptest.NewRequest(http.MethodGet, "/v0/projects/current", nil)
	projectCurrentRec := httptest.NewRecorder()
	handler.ServeHTTP(projectCurrentRec, projectCurrentReq)
	if projectCurrentRec.Code == http.StatusNotFound {
		t.Fatalf("expected current-project route to be registered, got 404")
	}

	checkpointsReq := httptest.NewRequest(http.MethodGet, "/v0/checkpoints", nil)
	checkpointsRec := httptest.NewRecorder()
	handler.ServeHTTP(checkpointsRec, checkpointsReq)
	if checkpointsRec.Code == http.StatusNotFound {
		t.Fatalf("expected checkpoints route to be registered, got 404")
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
