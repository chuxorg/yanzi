package routes

import (
	"net/http"

	"github.com/chuxorg/yanzi/internal/api/handlers"
	"github.com/chuxorg/yanzi/internal/api/middleware"
)

const (
	basePath           = "/v0"
	healthPath         = basePath + "/health"
	rehydratePath      = basePath + "/rehydrate"
	intentsPath        = basePath + "/intents"
	artifactsPath      = basePath + "/artifacts"
	verifyPath         = basePath + "/verify/"
	chainPath          = basePath + "/chain/"
	exportPath         = basePath + "/export/"
	projectsPath       = basePath + "/projects"
	projectCurrentPath = projectsPath + "/current"
	checkpointsPath    = basePath + "/checkpoints"
)

// NewHandler constructs the operational API route foundation.
func NewHandler(deps handlers.Dependencies) http.Handler {
	mux := http.NewServeMux()
	registerGet(mux, healthPath, handlers.NewHealthHandler(deps))
	registerGet(mux, rehydratePath, handlers.NewRehydrateHandler(deps))
	registerArtifacts(mux, handlers.NewArtifactHandler(deps))
	registerVerification(mux, handlers.NewVerifyHandler(deps))
	registerExport(mux, handlers.NewExportHandler(deps))
	registerMethods(mux, projectsPath, handlers.NewProjectsHandler(deps), http.MethodGet, http.MethodPost)
	registerMethods(mux, projectCurrentPath, handlers.NewCurrentProjectHandler(deps), http.MethodGet, http.MethodPost)
	registerMethods(mux, checkpointsPath, handlers.NewCheckpointsHandler(deps), http.MethodGet, http.MethodPost)
	return mux
}

func registerGet(mux *http.ServeMux, path string, handler http.Handler) {
	mux.Handle(path, middleware.AllowMethods(handler, http.MethodGet))
}

func registerMethods(mux *http.ServeMux, path string, handler http.Handler, methods ...string) {
	mux.Handle(path, middleware.AllowMethods(handler, methods...))
}

func registerArtifacts(mux *http.ServeMux, handler http.Handler) {
	mux.Handle(artifactsPath, handler)
	mux.Handle(artifactsPath+"/", handler)
}

func registerVerification(mux *http.ServeMux, handler http.Handler) {
	mux.Handle(verifyPath, handler)
	mux.Handle(chainPath, handler)
	mux.Handle(intentsPath+"/", handler)
}

func registerExport(mux *http.ServeMux, handler http.Handler) {
	mux.Handle(exportPath, handler)
}
