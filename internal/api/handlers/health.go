package handlers

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
	"github.com/chuxorg/yanzi/internal/auth"
	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/core/model"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
	"github.com/chuxorg/yanzi/internal/projectstate"
	"github.com/chuxorg/yanzi/internal/storage"
	"github.com/chuxorg/yanzi/internal/storage/registry"
)

// ConfigLoadFunc loads the current Yanzi configuration for API handlers.
type ConfigLoadFunc func() (config.Config, error)

// ProviderOpenFunc opens the current storage provider for API handlers.
type ProviderOpenFunc func(context.Context, config.Config) (storage.Provider, error)

// ActiveProjectLoadFunc loads the current active project for API handlers.
type ActiveProjectLoadFunc func() (string, error)

// ArtifactReadStore exposes the current read-only behavior required by artifact handlers.
type ArtifactReadStore interface {
	ListIntentRecords(context.Context, yanzilibrary.ArtifactReadQuery) ([]model.IntentRecord, error)
	GetIntentRecord(context.Context, string) (model.IntentRecord, error)
}

// ArtifactReadOpenFunc opens the current artifact read boundary for API handlers.
type ArtifactReadOpenFunc func(context.Context, config.Config) (ArtifactReadStore, io.Closer, error)

// RuntimeStatusFunc reports the currently active runtime bootstrap visibility.
type RuntimeStatusFunc func() *models.RuntimeHealth

// Dependencies captures the lightweight handler dependencies used by the API foundation.
type Dependencies struct {
	Version               string
	LoadConfig            ConfigLoadFunc
	OpenProvider          ProviderOpenFunc
	CreateProject         func(string, string) (*yanzilibrary.Project, error)
	ListProjects          func() ([]yanzilibrary.Project, error)
	ProjectExists         func(string) (bool, error)
	LoadActiveProject     ActiveProjectLoadFunc
	SaveActiveProject     func(string) error
	CreateCheckpoint      func(string, string, []string) (yanzilibrary.Checkpoint, error)
	ListCheckpoints       func(string) ([]yanzilibrary.Checkpoint, error)
	ListAllCheckpoints    func() ([]yanzilibrary.Checkpoint, error)
	OpenArtifactReadStore ArtifactReadOpenFunc
	Now                   func() time.Time
	RuntimeStatus         RuntimeStatusFunc
	APIKeyStore           auth.APIKeyStore
	AuthConfig            config.AuthConfig
	OIDCValidator         *auth.OIDCValidator
}

// NewHealthHandler returns the minimal GET /v0/health handler.
func NewHealthHandler(deps Dependencies) http.Handler {
	deps = deps.withDefaults()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg, err := deps.LoadConfig()
		if err != nil {
			responses.WriteError(w, http.StatusInternalServerError, "config_load_failed", err.Error())
			return
		}

		resp := newHealthResponse(deps.Version, cfg.Mode, deps.RuntimeStatus, deps.AuthConfig)

		providerCfg := cfg
		providerCfg.Mode = config.ModeLocal
		provider, err := deps.OpenProvider(r.Context(), providerCfg)
		if err != nil {
			resp.Provider.Error = err.Error()
			responses.WriteJSON(w, http.StatusOK, resp)
			return
		}
		defer func() {
			_ = provider.Close()
		}()

		health := provider.Health(r.Context())
		if health.Provider != "" {
			resp.Provider.Name = string(health.Provider)
		}
		resp.Provider.Status = string(health.Status)
		resp.Provider.Error = health.Error

		responses.WriteJSON(w, http.StatusOK, resp)
	})
}

func newHealthResponse(version string, mode config.Mode, runtimeStatus RuntimeStatusFunc, authCfg config.AuthConfig) models.HealthResponse {
	resp := models.HealthResponse{
		Version: version,
		Mode:    string(mode),
		Provider: models.ProviderHealth{
			Name:   string(storage.ProviderSQLite),
			Status: string(storage.HealthUnavailable),
		},
		Auth: &models.AuthHealth{
			Enabled:        authCfg.Enabled,
			DevKeysAllowed: authCfg.DevKeysAllowed,
		},
	}
	if runtimeStatus != nil {
		resp.Runtime = runtimeStatus()
	}
	return resp
}

func (d Dependencies) withDefaults() Dependencies {
	if d.LoadConfig == nil {
		d.LoadConfig = config.Load
	}
	if d.OpenProvider == nil {
		d.OpenProvider = func(ctx context.Context, cfg config.Config) (storage.Provider, error) {
			return registry.Open(ctx, cfg, registry.Options{})
		}
	}
	if d.CreateProject == nil {
		d.CreateProject = yanzilibrary.CreateProject
	}
	if d.ListProjects == nil {
		d.ListProjects = yanzilibrary.ListProjects
	}
	if d.ProjectExists == nil {
		d.ProjectExists = yanzilibrary.ProjectExists
	}
	if d.LoadActiveProject == nil {
		d.LoadActiveProject = projectstate.LoadActiveProject
	}
	if d.SaveActiveProject == nil {
		d.SaveActiveProject = projectstate.SaveActiveProject
	}
	if d.CreateCheckpoint == nil {
		d.CreateCheckpoint = yanzilibrary.CreateProjectCheckpoint
	}
	if d.ListCheckpoints == nil {
		d.ListCheckpoints = yanzilibrary.ListProjectCheckpoints
	}
	if d.ListAllCheckpoints == nil {
		d.ListAllCheckpoints = yanzilibrary.ListAllProjectCheckpoints
	}
	if d.OpenArtifactReadStore == nil {
		d.OpenArtifactReadStore = func(ctx context.Context, cfg config.Config) (ArtifactReadStore, io.Closer, error) {
			dbPath, err := config.EffectiveLocalDBPath(cfg)
			if err != nil {
				return nil, nil, err
			}
			db, err := yanzilibrary.InitDBAtPath(dbPath)
			if err != nil {
				return nil, nil, err
			}
			return yanzilibrary.NewArtifactReadStore(db), db, nil
		}
	}
	if d.Now == nil {
		d.Now = time.Now
	}
	return d
}
