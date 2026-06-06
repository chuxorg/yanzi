package client

import (
	"os"
	"strings"

	"github.com/chuxorg/yanzi/internal/config"
)

// ResolveAuthHeader returns the Authorization header value for CLI HTTP mode requests.
// Returns an empty string if no auth is configured (no header is sent).
//
// Resolution order:
//  1. flagAPIKey (from --api-key flag)
//  2. YANZI_API_KEY environment variable
//  3. YANZI_OIDC_TOKEN environment variable
//  4. cfg.Auth.APIKey (config.yaml auth.api_key)
func ResolveAuthHeader(cfg config.Config, flagAPIKey string) string {
	if v := strings.TrimSpace(flagAPIKey); v != "" {
		return "Bearer " + v
	}
	if v := strings.TrimSpace(os.Getenv("YANZI_API_KEY")); v != "" {
		return "Bearer " + v
	}
	if v := strings.TrimSpace(os.Getenv("YANZI_OIDC_TOKEN")); v != "" {
		return "Bearer " + v
	}
	if v := strings.TrimSpace(cfg.Auth.APIKey); v != "" {
		return "Bearer " + v
	}
	return ""
}
