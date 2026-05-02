package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestBuildURLResolves(t *testing.T) {
	cli := New("https://example.com/api")
	got, err := cli.buildURL("/v0/intents")
	if err != nil {
		t.Fatalf("buildURL error: %v", err)
	}
	want := "https://example.com/v0/intents"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestBuildURLInvalidBase(t *testing.T) {
	cli := New("http://[::1")
	_, err := cli.buildURL("/v0/intents")
	if err == nil {
		t.Fatal("expected error for invalid base url")
	}
	if !strings.Contains(err.Error(), "invalid base_url") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListIntentsQueryParams(t *testing.T) {
	var gotQuery url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"intents": []}`))
	}))
	t.Cleanup(srv.Close)

	cli := New(srv.URL)
	_, err := cli.ListIntents(context.Background(), "alice", "cli", 25, map[string]string{"team": "core"}, true)
	if err != nil {
		t.Fatalf("ListIntents error: %v", err)
	}

	if gotQuery.Get("author") != "alice" {
		t.Fatalf("expected author=alice, got %q", gotQuery.Get("author"))
	}
	if gotQuery.Get("source") != "cli" {
		t.Fatalf("expected source=cli, got %q", gotQuery.Get("source"))
	}
	if gotQuery.Get("limit") != "25" {
		t.Fatalf("expected limit=25, got %q", gotQuery.Get("limit"))
	}
	if gotQuery.Get("meta_team") != "core" {
		t.Fatalf("expected meta_team=core, got %q", gotQuery.Get("meta_team"))
	}
	if gotQuery.Get("include_deleted") != "true" {
		t.Fatalf("expected include_deleted=true, got %q", gotQuery.Get("include_deleted"))
	}
}

func TestDoJSONServerErrorUsesBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("nope"))
	}))
	t.Cleanup(srv.Close)

	cli := New(srv.URL)
	var out any
	err := cli.doJSON(context.Background(), http.MethodGet, "/v0/intents", nil, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateIntentSendsJSON(t *testing.T) {
	var contentType string
	var body CreateIntentRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"1"}`))
	}))
	t.Cleanup(srv.Close)

	cli := New(srv.URL)
	_, err := cli.CreateIntent(context.Background(), CreateIntentRequest{
		Author:     "alice",
		SourceType: "cli",
		Prompt:     "p",
		Response:   "r",
	})
	if err != nil {
		t.Fatalf("CreateIntent error: %v", err)
	}
	if contentType != "application/json" {
		t.Fatalf("expected content-type application/json, got %q", contentType)
	}
	if body.Author != "alice" || body.SourceType != "cli" {
		t.Fatalf("unexpected body: %+v", body)
	}
}
