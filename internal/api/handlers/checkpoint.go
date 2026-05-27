package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

const queryAllProjects = "all_projects"

// NewCheckpointsHandler returns the provider-backed checkpoint collection handler.
func NewCheckpointsHandler(deps Dependencies) http.Handler {
	deps = deps.withDefaults()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			allProjects := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get(queryAllProjects)), "true")
			if allProjects {
				checkpoints, err := deps.ListAllCheckpoints()
				if err != nil {
					writeOperationError(w, err)
					return
				}
				responses.WriteJSON(w, http.StatusOK, models.CheckpointListResponse{
					Checkpoints: checkpointModels(checkpoints),
				})
				return
			}

			project, err := deps.LoadActiveProject()
			if err != nil {
				writeOperationError(w, err)
				return
			}
			project = strings.TrimSpace(project)
			if project == "" {
				writeOperationError(w, errors.New("no active project set"))
				return
			}

			checkpoints, err := deps.ListCheckpoints(project)
			if err != nil {
				writeOperationError(w, err)
				return
			}
			responses.WriteJSON(w, http.StatusOK, models.CheckpointListResponse{
				Checkpoints: checkpointModels(checkpoints),
			})
		case http.MethodPost:
			var req models.CheckpointCreateRequest
			if !decodeJSONBody(w, r, &req) {
				return
			}
			project, err := deps.LoadActiveProject()
			if err != nil {
				writeOperationError(w, err)
				return
			}
			project = strings.TrimSpace(project)
			if project == "" {
				writeOperationError(w, errors.New("no active project set"))
				return
			}
			if candidate := strings.TrimSpace(req.Project); candidate != "" && candidate != project {
				writeOperationError(w, errors.New("checkpoint project must match active project"))
				return
			}

			checkpoint, err := deps.CreateCheckpoint(project, req.Summary, req.ArtifactIDs)
			if err != nil {
				writeOperationError(w, err)
				return
			}
			responses.WriteJSON(w, http.StatusCreated, checkpointModel(checkpoint))
		default:
			responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		}
	})
}

func checkpointModels(checkpoints []yanzilibrary.Checkpoint) []models.Checkpoint {
	result := make([]models.Checkpoint, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		result = append(result, checkpointModel(checkpoint))
	}
	return result
}

func checkpointModel(checkpoint yanzilibrary.Checkpoint) models.Checkpoint {
	return models.Checkpoint{
		Hash:                 checkpoint.Hash,
		Project:              checkpoint.Project,
		Summary:              checkpoint.Summary,
		CreatedAt:            checkpoint.CreatedAt,
		ArtifactIDs:          checkpoint.ArtifactIDs,
		PreviousCheckpointID: checkpoint.PreviousCheckpointID,
	}
}
