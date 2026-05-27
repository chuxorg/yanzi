package routes

import (
	"net/http"

	"github.com/chuxorg/yanzi/internal/api/handlers"
	"github.com/chuxorg/yanzi/internal/api/middleware"
)

const (
	basePath           = "/v0"
	healthPath         = basePath + "/health"
	artifactsPath      = basePath + "/artifacts"
	projectsPath       = basePath + "/projects"
	projectCurrentPath = projectsPath + "/current"
	checkpointsPath    = basePath + "/checkpoints"
)

// NewHandler constructs the operational API route foundation.
func NewHandler(deps handlers.Dependencies) http.Handler {
	mux := http.NewServeMux()
	registerGet(mux, healthPath, handlers.NewHealthHandler(deps))
	registerDeferredGroup(mux, artifactsPath, "artifacts")
	registerMethods(mux, projectsPath, handlers.NewProjectsHandler(deps), http.MethodGet, http.MethodPost)
	registerMethods(mux, projectCurrentPath, handlers.NewCurrentProjectHandler(deps), http.MethodGet, http.MethodPost)
	registerDeferredGroup(mux, checkpointsPath, "checkpoints")
	return mux
}

func registerGet(mux *http.ServeMux, path string, handler http.Handler) {
	mux.Handle(path, middleware.AllowMethods(handler, http.MethodGet))
}

func registerMethods(mux *http.ServeMux, path string, handler http.Handler, methods ...string) {
	mux.Handle(path, middleware.AllowMethods(handler, methods...))
}

func registerDeferredGroup(mux *http.ServeMux, path, group string) {
	handler := middleware.AllowMethods(handlers.NewDeferredRouteHandler(group), http.MethodGet)
	mux.Handle(path, handler)
	mux.Handle(path+"/", handler)
}
