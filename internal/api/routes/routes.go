package routes

import (
	"net/http"

	"github.com/chuxorg/yanzi/internal/api/handlers"
	"github.com/chuxorg/yanzi/internal/api/middleware"
)

const (
	basePath        = "/v0"
	healthPath      = basePath + "/health"
	rehydratePath   = basePath + "/rehydrate"
	intentsPath     = basePath + "/intents"
	artifactsPath   = basePath + "/artifacts"
	verifyPath      = basePath + "/verify/"
	chainPath       = basePath + "/chain/"
	exportPath      = basePath + "/export/"
	projectsPath    = basePath + "/projects"
	checkpointsPath = basePath + "/checkpoints"
)

// NewHandler constructs the operational API route foundation.
func NewHandler(deps handlers.Dependencies) http.Handler {
	mux := http.NewServeMux()
	registerGet(mux, healthPath, handlers.NewHealthHandler(deps))
	registerGet(mux, rehydratePath, handlers.NewRehydrateHandler(deps))
	registerArtifacts(mux, handlers.NewArtifactHandler(deps))
	registerVerification(mux, handlers.NewVerifyHandler(deps))
	registerExport(mux, handlers.NewExportHandler(deps))
	registerDeferredGroup(mux, projectsPath, "projects")
	registerDeferredGroup(mux, checkpointsPath, "checkpoints")
	return mux
}

func registerGet(mux *http.ServeMux, path string, handler http.Handler) {
	mux.Handle(path, middleware.AllowMethods(handler, http.MethodGet))
}

func registerDeferredGroup(mux *http.ServeMux, path, group string) {
	handler := middleware.AllowMethods(handlers.NewDeferredRouteHandler(group), http.MethodGet)
	mux.Handle(path, handler)
	mux.Handle(path+"/", handler)
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
