package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/chuxorg/yanzi/internal/api/responses"
	"github.com/chuxorg/yanzi/internal/storage"
)

const maxJSONBodyBytes = 1 << 20

func decodeJSONBody(w http.ResponseWriter, r *http.Request, out any) bool {
	if r.Body == nil {
		responses.WriteError(w, http.StatusBadRequest, "invalid_request", "request body is required")
		return false
	}
	defer r.Body.Close()

	dec := json.NewDecoder(io.LimitReader(r.Body, maxJSONBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		responses.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid JSON request body")
		return false
	}
	if err := dec.Decode(new(struct{})); !errors.Is(err, io.EOF) {
		responses.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must contain a single JSON object")
		return false
	}
	return true
}

func writeOperationError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	message := err.Error()
	switch {
	case errors.Is(err, storage.ErrNotFound):
		responses.WriteError(w, http.StatusNotFound, "not_found", message)
	case strings.Contains(message, "already exists"):
		responses.WriteError(w, http.StatusConflict, "conflict", message)
	case strings.Contains(message, "not found"):
		responses.WriteError(w, http.StatusNotFound, "not_found", message)
	case strings.Contains(message, "required"),
		strings.Contains(message, "usage:"),
		strings.Contains(message, "invalid "),
		strings.Contains(message, "must match"),
		strings.Contains(message, "conflicts with"),
		strings.Contains(message, "no active project"):
		responses.WriteError(w, http.StatusBadRequest, "invalid_request", message)
	default:
		responses.WriteError(w, http.StatusInternalServerError, "internal_error", message)
	}
}
