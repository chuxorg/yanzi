package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/chuxorg/yanzi/internal/api/responses"
	"github.com/chuxorg/yanzi/internal/packs"
)

// NewTokensHandler returns the /v0/tokens handler.
func NewTokensHandler(deps Dependencies) http.Handler {
	deps = deps.withDefaults()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if deps.PackStore == nil {
			responses.WriteError(w, http.StatusInternalServerError, "not_available", "pack store not initialized")
			return
		}
		if r.Method != http.MethodGet {
			responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}

		q := r.URL.Query()
		filter := packs.TokenFilter{
			Project: q.Get("project"),
			Phase:   q.Get("phase"),
			Task:    q.Get("task"),
		}
		if since := q.Get("since"); since != "" {
			if t, err := time.Parse(time.RFC3339, since); err == nil {
				filter.Since = t
			}
		}
		if filter.Project == "" {
			filter.Project, _ = deps.LoadActiveProject()
		}

		summary, err := deps.PackStore.GetTokenUsage(r.Context(), filter)
		if err != nil {
			log.Printf("tokens: get usage error: %v", err)
			responses.WriteError(w, http.StatusInternalServerError, "get_failed", "failed to get token usage")
			return
		}
		responses.WriteJSON(w, http.StatusOK, summary)
	})
}
