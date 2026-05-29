package middleware

import "net/http"

// TODO(CAP-004): tighten Access-Control-Allow-Origin to explicit allowed origins
// configurable via config.yaml once the identity and transport baseline is implemented.

// CORS adds permissive cross-origin headers to every response so that browser-based
// UI clients served from any local origin can reach yanzi serve without being blocked
// by the same-origin policy.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
