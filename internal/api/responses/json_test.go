package responses

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteErrorUsesDeterministicEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusBadRequest, "bad_request", "bad input")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("unexpected content type: %q", got)
	}
	if got := rec.Body.String(); got != "{\"error\":{\"code\":\"bad_request\",\"message\":\"bad input\"}}\n" {
		t.Fatalf("unexpected error body: %q", got)
	}
}
