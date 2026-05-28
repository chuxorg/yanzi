package routes

import (
	"net/http"

	"github.com/chuxorg/yanzi/internal/api/handlers"
	"github.com/chuxorg/yanzi/internal/api/middleware"
)

const (
	basePath        = "/v0"
	healthPath      = basePath + "/health"
	artifactsPath   = basePath + "/artifacts"
	projectsPath    = basePath + "/projects"
	checkpointsPath = basePath + "/checkpoints"
)

// NewHandler constructs the operational API route foundation.
func NewHandler(deps handlers.Dependencies) http.Handler {
	mux := http.NewServeMux()
	registerGet(mux, healthPath, handlers.NewHealthHandler(deps))
	registerGetPrefix(mux, artifactsPath, handlers.NewArtifactHandler(deps))
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

func registerGetPrefix(mux *http.ServeMux, path string, handler http.Handler) {
	allowed := middleware.AllowMethods(handler, http.MethodGet)
	mux.Handle(path, allowed)
	mux.Handle(path+"/", allowed)
}
