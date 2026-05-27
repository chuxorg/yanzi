package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

// NewProjectsHandler returns the provider-backed project collection handler.
func NewProjectsHandler(deps Dependencies) http.Handler {
	deps = deps.withDefaults()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projects, err := deps.ListProjects()
			if err != nil {
				writeOperationError(w, err)
				return
			}
			resp := models.ProjectListResponse{Projects: make([]models.Project, 0, len(projects))}
			for _, project := range projects {
				resp.Projects = append(resp.Projects, projectModel(project))
			}
			responses.WriteJSON(w, http.StatusOK, resp)
		case http.MethodPost:
			var req models.ProjectCreateRequest
			if !decodeJSONBody(w, r, &req) {
				return
			}
			project, err := deps.CreateProject(req.Name, req.Description)
			if err != nil {
				writeOperationError(w, err)
				return
			}
			responses.WriteJSON(w, http.StatusCreated, projectModel(*project))
		default:
			responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		}
	})
}

// NewCurrentProjectHandler returns the active-project read/write handler.
func NewCurrentProjectHandler(deps Dependencies) http.Handler {
	deps = deps.withDefaults()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			active, err := deps.LoadActiveProject()
			if err != nil {
				writeOperationError(w, err)
				return
			}
			project, err := currentProject(active, deps.ListProjects)
			if err != nil {
				writeOperationError(w, err)
				return
			}
			responses.WriteJSON(w, http.StatusOK, models.CurrentProjectResponse{Project: project})
		case http.MethodPost:
			var req models.CurrentProjectRequest
			if !decodeJSONBody(w, r, &req) {
				return
			}
			name := strings.TrimSpace(req.Name)
			if name == "" {
				writeOperationError(w, errors.New("project name is required"))
				return
			}
			exists, err := deps.ProjectExists(name)
			if err != nil {
				writeOperationError(w, err)
				return
			}
			if !exists {
				responses.WriteError(w, http.StatusNotFound, "not_found", "project not found: "+name)
				return
			}
			if err := deps.SaveActiveProject(name); err != nil {
				writeOperationError(w, err)
				return
			}
			project, err := currentProject(name, deps.ListProjects)
			if err != nil {
				writeOperationError(w, err)
				return
			}
			responses.WriteJSON(w, http.StatusOK, models.CurrentProjectResponse{Project: project})
		default:
			responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		}
	})
}

func currentProject(active string, listProjects func() ([]yanzilibrary.Project, error)) (*models.Project, error) {
	active = strings.TrimSpace(active)
	if active == "" {
		return nil, nil
	}
	projects, err := listProjects()
	if err != nil {
		return nil, err
	}
	for _, project := range projects {
		if project.Name == active {
			value := projectModel(project)
			return &value, nil
		}
	}
	return &models.Project{Name: active}, nil
}

func projectModel(project yanzilibrary.Project) models.Project {
	return models.Project{
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt.Format(time.RFC3339Nano),
	}
}
