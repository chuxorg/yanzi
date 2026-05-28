package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/chuxorg/yanzi/internal/api/responses"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

const exportPath = "/v0/export/"

// NewExportHandler returns the deterministic export read API handler.
func NewExportHandler(deps Dependencies) http.Handler {
	deps = deps.withDefaults()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}

		format, ok := parseExportFormat(r.URL.Path)
		if !ok {
			responses.WriteError(w, http.StatusNotFound, "export_not_found", "export endpoint not found")
			return
		}

		project := strings.TrimSpace(r.URL.Query().Get("project"))
		if project == "" {
			responses.WriteError(w, http.StatusBadRequest, "validation_failed", "project is required")
			return
		}
		if strings.TrimSpace(r.URL.Query().Get("checkpoint")) != "" {
			responses.WriteError(w, http.StatusBadRequest, "validation_failed", "checkpoint filter is not supported")
			return
		}

		includeDeleted, err := parseBoolQuery(r.URL.Query().Get("include_deleted"))
		if err != nil {
			responses.WriteError(w, http.StatusBadRequest, "validation_failed", "include_deleted must be true or false")
			return
		}
		metaFilters := exportMetaFilters(r)

		provider, err := openArtifactProvider(r.Context(), deps)
		if err != nil {
			responses.WriteError(w, http.StatusInternalServerError, "export_unavailable", err.Error())
			return
		}
		defer func() {
			_ = provider.Close()
		}()

		content, contentType, err := yanzilibrary.RenderOperationalExportLog(r.Context(), provider, project, deps.Version, deps.Now().UTC(), format, metaFilters, includeDeleted)
		if err != nil {
			responses.WriteError(w, http.StatusInternalServerError, "export_failed", err.Error())
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	})
}

func parseExportFormat(path string) (yanzilibrary.ExportLogFormat, bool) {
	switch strings.TrimPrefix(path, exportPath) {
	case "markdown":
		return yanzilibrary.ExportLogFormatMarkdown, true
	case "json":
		return yanzilibrary.ExportLogFormatJSON, true
	case "html":
		return yanzilibrary.ExportLogFormatHTML, true
	default:
		return "", false
	}
}

func exportMetaFilters(r *http.Request) map[string]string {
	metaFilters := map[string]string{}
	for key, values := range r.URL.Query() {
		if !strings.HasPrefix(key, "meta_") || len(values) == 0 {
			continue
		}
		metaKey := strings.TrimPrefix(key, "meta_")
		if strings.TrimSpace(metaKey) == "" {
			continue
		}
		metaFilters[metaKey] = values[len(values)-1]
	}
	if profile := strings.TrimSpace(r.URL.Query().Get("profile")); profile != "" {
		metaFilters["profile"] = profile
	}
	if len(metaFilters) == 0 {
		return nil
	}
	return metaFilters
}

func parseBoolQuery(raw string) (bool, error) {
	if strings.TrimSpace(raw) == "" {
		return false, nil
	}
	return strconv.ParseBool(raw)
}
