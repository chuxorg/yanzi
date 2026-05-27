package middleware

import (
	"net/http"
	"strings"

	"github.com/chuxorg/yanzi/internal/api/responses"
)

// AllowMethods restricts a handler to the provided HTTP methods.
func AllowMethods(next http.Handler, methods ...string) http.Handler {
	allowed := make(map[string]struct{}, len(methods))
	names := make([]string, 0, len(methods))
	for _, method := range methods {
		method = strings.TrimSpace(method)
		if method == "" {
			continue
		}
		allowed[method] = struct{}{}
		names = append(names, method)
	}
	allowHeader := strings.Join(names, ", ")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := allowed[r.Method]; !ok {
			if allowHeader != "" {
				w.Header().Set("Allow", allowHeader)
			}
			responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}
		next.ServeHTTP(w, r)
	})
}
