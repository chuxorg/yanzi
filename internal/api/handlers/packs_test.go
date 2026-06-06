package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/packs"
	"github.com/chuxorg/yanzi/internal/storage"
)

func packsTestDeps(store *testPackStore) Dependencies {
	provider := &artifactProvider{stubProvider: &stubProvider{}, nextID: "art-pack-001"}
	return Dependencies{
		LoadConfig: func() (config.Config, error) {
			return config.Config{Mode: config.ModeLocal, DBPath: "/tmp/test.db"}, nil
		},
		OpenProvider: func(context.Context, config.Config) (storage.Provider, error) {
			return provider, nil
		},
		LoadActiveProject: func() (string, error) { return "test-project", nil },
		PackStore:         store,
	}
}

func TestCreatePack_Valid(t *testing.T) {
	store := newTestPackStore()
	// Pre-populate a seed so the seed reference is valid (we don't validate seeds exist on pack create).
	store.seeds["seed-art-1"] = packs.Seed{
		ID: "seed-art-1", ArtifactID: "seed-art-1", Name: "my-seed",
		SeedType: "process", CreatedAt: time.Now(),
	}
	deps := packsTestDeps(store)
	h := NewPacksHandler(deps)

	body := `{"name":"my-pack","role":1,"seeds":[{"name":"my-seed","artifact_id":"seed-art-1"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v0/packs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var pack packs.Pack
	if err := json.NewDecoder(rec.Body).Decode(&pack); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if pack.Name != "my-pack" {
		t.Errorf("Name = %q, want %q", pack.Name, "my-pack")
	}
	if pack.ArtifactID == "" {
		t.Error("ArtifactID should be set")
	}
}

func TestCreatePack_InvalidSeedReference(t *testing.T) {
	store := newTestPackStore()
	deps := packsTestDeps(store)
	h := NewPacksHandler(deps)

	// Seed reference missing artifact_id.
	body := `{"name":"bad-pack","role":1,"seeds":[{"name":"no-id"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v0/packs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestComposePack_Valid(t *testing.T) {
	store := newTestPackStore()
	seed := packs.Seed{
		ID: "s1", ArtifactID: "s1", Name: "yanzi-seed",
		SeedType: packs.SeedTypeYanzi, RoleAccessBits: packs.RoleObserver,
		Content: packs.SeedContent{
			Sections: []packs.ContentSection{{Section: "overview", Type: "instruction", Text: "Use yanzi"}},
		},
		TokenEstimate: 5, CreatedAt: time.Now(),
	}
	pack := packs.Pack{
		ID: "p1", ArtifactID: "p1", Name: "test-pack",
		Role: packs.RoleObserver, RoleLabel: "Observer",
		Seeds:     []packs.SeedReference{{Name: "yanzi-seed", ArtifactID: "s1"}},
		CreatedAt: time.Now(),
	}
	store.seeds["s1"] = seed
	store.packs["p1"] = pack

	deps := packsTestDeps(store)
	h := NewPacksHandler(deps)

	body := `{"pack_artifact_id":"p1","options":{"include_assembled_prompt":true}}`
	req := httptest.NewRequest(http.MethodPost, "/v0/packs/compose", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result packs.ComposeResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.Pack.Name != "test-pack" {
		t.Errorf("Pack.Name = %q, want %q", result.Pack.Name, "test-pack")
	}
}

func TestComposePack_WithTask(t *testing.T) {
	store := newTestPackStore()
	seed := packs.Seed{
		ID: "s2", ArtifactID: "s2", Name: "yanzi-seed2",
		SeedType: packs.SeedTypeYanzi, RoleAccessBits: packs.RoleObserver,
		Content: packs.SeedContent{
			Sections: []packs.ContentSection{{Section: "overview", Type: "instruction", Text: "Use yanzi"}},
		},
		TokenEstimate: 5, CreatedAt: time.Now(),
	}
	pack := packs.Pack{
		ID: "p2", ArtifactID: "p2", Name: "task-pack",
		Role: packs.RoleObserver, Seeds: []packs.SeedReference{{Name: "yanzi-seed2", ArtifactID: "s2"}},
		CreatedAt: time.Now(),
	}
	store.seeds["s2"] = seed
	store.packs["p2"] = pack

	deps := packsTestDeps(store)
	h := NewPacksHandler(deps)

	body := `{"pack_artifact_id":"p2","task_content":"implement the feature","options":{"include_assembled_prompt":true}}`
	req := httptest.NewRequest(http.MethodPost, "/v0/packs/compose", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result packs.ComposeResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !strings.Contains(result.AssembledPrompt, "implement the feature") {
		t.Error("assembled prompt missing task content")
	}
}

func TestComposePack_RecordsTokenUsage(t *testing.T) {
	store := newTestPackStore()
	seed := packs.Seed{
		ID: "s3", ArtifactID: "s3", Name: "yanzi-seed3",
		SeedType: packs.SeedTypeYanzi, RoleAccessBits: packs.RoleObserver,
		Content: packs.SeedContent{
			Sections: []packs.ContentSection{{Section: "overview", Type: "instruction", Text: "Use yanzi for tracking"}},
		},
		TokenEstimate: 10, CreatedAt: time.Now(),
	}
	pack := packs.Pack{
		ID: "p3", ArtifactID: "p3", Name: "token-pack",
		Role: packs.RoleObserver, Seeds: []packs.SeedReference{{Name: "yanzi-seed3", ArtifactID: "s3"}},
		CreatedAt: time.Now(),
	}
	store.seeds["s3"] = seed
	store.packs["p3"] = pack

	deps := packsTestDeps(store)
	h := NewPacksHandler(deps)

	body := `{"pack_artifact_id":"p3","options":{"include_assembled_prompt":false}}`
	req := httptest.NewRequest(http.MethodPost, "/v0/packs/compose?project=test-project", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(store.tokenRecords) == 0 {
		t.Error("expected token usage to be recorded")
	}
}
