package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
	"github.com/chuxorg/yanzi/internal/storage"
	"github.com/chuxorg/yanzi/internal/storage/registry"
)

// ConfigLoadFunc loads the current Yanzi configuration for API handlers.
type ConfigLoadFunc func() (config.Config, error)

// ProviderOpenFunc opens the current storage provider for API handlers.
type ProviderOpenFunc func(context.Context, config.Config) (storage.Provider, error)

// RuntimeStatusFunc reports the currently active runtime bootstrap visibility.
type RuntimeStatusFunc func() *models.RuntimeHealth

// Dependencies captures the lightweight handler dependencies used by the API foundation.
type Dependencies struct {
	Version       string
	LoadConfig    ConfigLoadFunc
	OpenProvider  ProviderOpenFunc
	RuntimeStatus RuntimeStatusFunc
	Now           func() time.Time
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

		resp := newHealthResponse(deps.Version, cfg.Mode, deps.RuntimeStatus)

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

func newHealthResponse(version string, mode config.Mode, runtimeStatus RuntimeStatusFunc) models.HealthResponse {
	resp := models.HealthResponse{
		Version: version,
		Mode:    string(mode),
		Provider: models.ProviderHealth{
			Name:   string(storage.ProviderSQLite),
			Status: string(storage.HealthUnavailable),
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
			return registry.Open(ctx, cfg, registry.Options{Migrations: yanzilibrary.MigrationsFS()})
		}
	}
	if d.Now == nil {
		d.Now = time.Now
	}
	return d
}
