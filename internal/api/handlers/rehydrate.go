package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

const rehydrateSnippetLimit = 160

// NewRehydrateHandler returns the deterministic GET /v0/rehydrate handler.
func NewRehydrateHandler(deps Dependencies) http.Handler {
	deps = deps.withDefaults()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project, err := yanzilibrary.LoadActiveProject()
		if err != nil {
			responses.WriteError(w, http.StatusInternalServerError, "active_project_lookup_failed", err.Error())
			return
		}
		if strings.TrimSpace(project) == "" {
			responses.WriteError(w, http.StatusBadRequest, "active_project_not_set", "no active project set")
			return
		}

		cfg, err := deps.LoadConfig()
		if err != nil {
			responses.WriteError(w, http.StatusInternalServerError, "config_load_failed", err.Error())
			return
		}

		providerCfg := cfg
		providerCfg.Mode = config.ModeLocal
		provider, err := deps.OpenProvider(r.Context(), providerCfg)
		if err != nil {
			responses.WriteError(w, http.StatusInternalServerError, "rehydration_provider_open_failed", err.Error())
			return
		}
		defer func() {
			_ = provider.Close()
		}()

		db := provider.SQLDB()
		if db == nil {
			responses.WriteError(w, http.StatusInternalServerError, "rehydration_sql_db_unavailable", "rehydration requires a SQL database handle")
			return
		}

		payload, err := yanzilibrary.NewRehydrationService(db).RehydrateProjectWithFallback(r.Context(), project, yanzilibrary.DefaultRehydrateFallbackLimit)
		if err != nil {
			var notFound yanzilibrary.ProjectNotFoundError
			switch {
			case errors.As(err, &notFound):
				responses.WriteError(w, http.StatusNotFound, "project_not_found", err.Error())
			default:
				responses.WriteError(w, http.StatusInternalServerError, "rehydration_failed", err.Error())
			}
			return
		}

		responses.WriteJSON(w, http.StatusOK, models.RehydrateResponse{
			Project:        payload.Project,
			HasCheckpoint:  payload.LatestCheckpoint != nil,
			Fallback:       payload.Fallback,
			FallbackReason: payload.FallbackReason,
			FallbackLimit:  payload.FallbackLimit,
			Checkpoint:     checkpointToModel(payload.LatestCheckpoint),
			Intents:        intentsToModel(payload.Intents),
		})
	})
}

func checkpointToModel(checkpoint *yanzilibrary.Checkpoint) *models.RehydrateCheckpoint {
	if checkpoint == nil {
		return nil
	}
	artifactIDs := append([]string{}, checkpoint.ArtifactIDs...)
	return &models.RehydrateCheckpoint{
		Hash:                 checkpoint.Hash,
		Project:              checkpoint.Project,
		Summary:              checkpoint.Summary,
		CreatedAt:            checkpoint.CreatedAt,
		ArtifactIDs:          artifactIDs,
		PreviousCheckpointID: checkpoint.PreviousCheckpointID,
	}
}

func intentsToModel(intents []yanzilibrary.Intent) []models.RehydrateIntent {
	ordered := make([]models.RehydrateIntent, 0, len(intents))
	for _, intent := range intents {
		ordered = append(ordered, models.RehydrateIntent{
			ID:              intent.ID,
			Timestamp:       intent.CreatedAt.Format(time.RFC3339Nano),
			Author:          intent.Author,
			SourceType:      intent.SourceType,
			Title:           intent.Title,
			Prompt:          intent.Prompt,
			Response:        intent.Response,
			PromptSnippet:   truncateRehydrateSnippet(intent.Prompt),
			ResponseSnippet: truncateRehydrateSnippet(intent.Response),
			Metadata:        decodeRehydrateMetadata(intent.Meta),
			Hash:            intent.Hash,
			PrevHash:        intent.PrevHash,
		})
	}
	return ordered
}

func truncateRehydrateSnippet(value string) string {
	cleaned := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\n"), "\r", "\n"))
	if cleaned == "" {
		return "(empty)"
	}
	runes := []rune(cleaned)
	if len(runes) <= rehydrateSnippetLimit {
		return cleaned
	}
	return string(runes[:rehydrateSnippetLimit]) + "..."
}

func decodeRehydrateMetadata(raw json.RawMessage) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	metadata := map[string]string{}
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return nil
	}
	return metadata
}
