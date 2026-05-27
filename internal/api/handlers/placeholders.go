package handlers

import (
	"fmt"
	"net/http"

	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
)

// NewDeferredRouteHandler returns a deterministic placeholder for deferred route groups.
func NewDeferredRouteHandler(group string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responses.WriteJSON(w, http.StatusNotImplemented, models.StatusResponse{
			Status:  "deferred",
			Message: fmt.Sprintf("%s endpoints are deferred beyond CAP-002 Phase 1", group),
		})
	})
}
