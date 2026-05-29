package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chuxorg/yanzi/internal/api/middleware"
)

func TestCORS_OptionsPreflightReturns204(t *testing.T) {
	handler := middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler must not be called for OPTIONS preflight")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/v0/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusNoContent)
	}
	assertCORSHeaders(t, rec)
}

func TestCORS_GETPassesThroughWithHeaders(t *testing.T) {
	const body = "ok"
	handler := middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))

	req := httptest.NewRequest(http.MethodGet, "/v0/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != body {
		t.Errorf("body: got %q, want %q", rec.Body.String(), body)
	}
	assertCORSHeaders(t, rec)
}

func TestCORS_POSTPassesThroughWithHeaders(t *testing.T) {
	const body = "created"
	handler := middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(body))
	}))

	req := httptest.NewRequest(http.MethodPost, "/v0/artifacts", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusCreated)
	}
	if rec.Body.String() != body {
		t.Errorf("body: got %q, want %q", rec.Body.String(), body)
	}
	assertCORSHeaders(t, rec)
}

func assertCORSHeaders(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	checks := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}
	for header, want := range checks {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s: got %q, want %q", header, got, want)
		}
	}
}
