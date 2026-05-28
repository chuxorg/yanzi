package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/api/responses"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

const (
	verifyPath       = "/v0/verify/"
	chainPath        = "/v0/chain/"
	intentsPrefix    = "/v0/intents/"
	verifySuffixPath = "/verify"
	chainSuffixPath  = "/chain"
)

// NewVerifyHandler returns the verification read API handler.
func NewVerifyHandler(deps Dependencies) http.Handler {
	deps = deps.withDefaults()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			responses.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}

		id, mode, ok := parseVerificationRoute(r.URL.Path)
		if !ok {
			responses.WriteError(w, http.StatusNotFound, "intent_not_found", "intent not found")
			return
		}

		provider, err := openArtifactProvider(r.Context(), deps)
		if err != nil {
			responses.WriteError(w, http.StatusInternalServerError, "verification_unavailable", err.Error())
			return
		}
		defer func() {
			_ = provider.Close()
		}()

		switch mode {
		case "verify":
			result, err := yanzilibrary.VerifyIntent(r.Context(), provider, id)
			if err != nil {
				writeVerificationError(w, "verify", id, err)
				return
			}
			responses.WriteJSON(w, http.StatusOK, models.VerifyResponse{
				ID:           result.ID,
				Valid:        result.Valid,
				StoredHash:   result.StoredHash,
				ComputedHash: result.ComputedHash,
				PrevHash:     result.PrevHash,
				Error:        result.Error,
			})
		case "chain":
			result, err := yanzilibrary.ChainIntent(r.Context(), provider, id)
			if err != nil {
				writeVerificationError(w, "chain", id, err)
				return
			}
			intents := make([]models.ArtifactCaptureResponse, 0, len(result.Intents))
			for _, intent := range result.Intents {
				resp, err := artifactCaptureResponse(intent)
				if err != nil {
					responses.WriteError(w, http.StatusInternalServerError, "chain_response_failed", err.Error())
					return
				}
				intents = append(intents, resp)
			}
			responses.WriteJSON(w, http.StatusOK, models.ChainResponse{
				HeadID:       result.HeadID,
				Length:       result.Length,
				Intents:      intents,
				MissingLinks: result.MissingLinks,
			})
		default:
			responses.WriteError(w, http.StatusNotFound, "intent_not_found", "intent not found")
		}
	})
}

func parseVerificationRoute(path string) (id string, mode string, ok bool) {
	switch {
	case strings.HasPrefix(path, verifyPath):
		id = strings.TrimPrefix(path, verifyPath)
		mode = "verify"
	case strings.HasPrefix(path, chainPath):
		id = strings.TrimPrefix(path, chainPath)
		mode = "chain"
	case strings.HasPrefix(path, intentsPrefix) && strings.HasSuffix(path, verifySuffixPath):
		id = strings.TrimSuffix(strings.TrimPrefix(path, intentsPrefix), verifySuffixPath)
		mode = "verify"
	case strings.HasPrefix(path, intentsPrefix) && strings.HasSuffix(path, chainSuffixPath):
		id = strings.TrimSuffix(strings.TrimPrefix(path, intentsPrefix), chainSuffixPath)
		mode = "chain"
	default:
		return "", "", false
	}
	id = strings.TrimSpace(id)
	if id == "" || strings.Contains(id, "/") {
		return "", "", false
	}
	return id, mode, true
}

func writeVerificationError(w http.ResponseWriter, mode, id string, err error) {
	if strings.Contains(err.Error(), "not found") {
		responses.WriteError(w, http.StatusNotFound, "intent_not_found", fmt.Sprintf("intent not found: %s", id))
		return
	}
	code := "verification_failed"
	if mode == "chain" {
		code = "chain_failed"
	}
	responses.WriteError(w, http.StatusInternalServerError, code, err.Error())
}
