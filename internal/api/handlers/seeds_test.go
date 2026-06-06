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

// testPackStore is a minimal in-memory PackStore for handler tests.
type testPackStore struct {
	seeds        map[string]packs.Seed
	packs        map[string]packs.Pack
	tokenRecords []packs.TokenUsage
}

func newTestPackStore() *testPackStore {
	return &testPackStore{
		seeds: map[string]packs.Seed{},
		packs: map[string]packs.Pack{},
	}
}

func (s *testPackStore) CreateSeed(_ context.Context, seed packs.Seed) (packs.Seed, error) {
	if seed.ID == "" {
		seed.ID = seed.ArtifactID
	}
	s.seeds[seed.ArtifactID] = seed
	return seed, nil
}
func (s *testPackStore) GetSeed(_ context.Context, id string) (packs.Seed, error) {
	seed, ok := s.seeds[id]
	if !ok {
		return packs.Seed{}, packs.ErrNotFound
	}
	return seed, nil
}
func (s *testPackStore) GetSeedByName(_ context.Context, name string) (packs.Seed, error) {
	for _, seed := range s.seeds {
		if seed.Name == name {
			return seed, nil
		}
	}
	return packs.Seed{}, packs.ErrNotFound
}
func (s *testPackStore) ListSeeds(_ context.Context, filter packs.SeedFilter) ([]packs.Seed, error) {
	var out []packs.Seed
	for _, seed := range s.seeds {
		if filter.SeedType != "" && seed.SeedType != filter.SeedType {
			continue
		}
		out = append(out, seed)
	}
	return out, nil
}
func (s *testPackStore) DeleteSeed(_ context.Context, id string) error {
	delete(s.seeds, id)
	return nil
}
func (s *testPackStore) CreatePack(_ context.Context, pack packs.Pack) (packs.Pack, error) {
	if pack.ID == "" {
		pack.ID = pack.ArtifactID
	}
	s.packs[pack.ArtifactID] = pack
	return pack, nil
}
func (s *testPackStore) GetPack(_ context.Context, id string) (packs.Pack, error) {
	p, ok := s.packs[id]
	if !ok {
		return packs.Pack{}, packs.ErrNotFound
	}
	return p, nil
}
func (s *testPackStore) GetPackByName(_ context.Context, name string) (packs.Pack, error) {
	for _, p := range s.packs {
		if p.Name == name {
			return p, nil
		}
	}
	return packs.Pack{}, packs.ErrNotFound
}
func (s *testPackStore) ListPacks(_ context.Context, _ packs.PackFilter) ([]packs.Pack, error) {
	var out []packs.Pack
	for _, p := range s.packs {
		out = append(out, p)
	}
	return out, nil
}
func (s *testPackStore) DeletePack(_ context.Context, id string) error {
	delete(s.packs, id)
	return nil
}
func (s *testPackStore) RecordTokenUsage(_ context.Context, u packs.TokenUsage) error {
	s.tokenRecords = append(s.tokenRecords, u)
	return nil
}
func (s *testPackStore) GetTokenUsage(_ context.Context, _ packs.TokenFilter) (packs.TokenUsageSummary, error) {
	return packs.TokenUsageSummary{}, nil
}

var _ packs.PackStore = (*testPackStore)(nil)

// artifactProvider embeds stubProvider but returns a valid artifact from CreateArtifact.
type artifactProvider struct {
	*stubProvider
	nextID string
}

func (p *artifactProvider) CreateArtifact(_ context.Context, in storage.CreateArtifactInput) (storage.Artifact, error) {
	id := p.nextID
	if id == "" {
		id = "test-artifact-id"
	}
	return storage.Artifact{ID: id, Class: in.Class, Type: in.Type, Title: in.Title}, nil
}

func seedsTestDeps(store *testPackStore) Dependencies {
	provider := &artifactProvider{stubProvider: &stubProvider{}, nextID: "art-001"}
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

func TestCreateSeed_Valid(t *testing.T) {
	store := newTestPackStore()
	deps := seedsTestDeps(store)
	h := NewSeedsHandler(deps)

	body := `{"name":"my-seed","seed_type":"process","role_access_bits":1,"content":{"sections":[{"section":"overview","type":"instruction","text":"do the thing"}]}}`
	req := httptest.NewRequest(http.MethodPost, "/v0/seeds", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var seed packs.Seed
	if err := json.NewDecoder(rec.Body).Decode(&seed); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if seed.Name != "my-seed" {
		t.Errorf("Name = %q, want %q", seed.Name, "my-seed")
	}
	if seed.ArtifactID == "" {
		t.Error("ArtifactID should be set")
	}
}

func TestCreateSeed_MissingName(t *testing.T) {
	store := newTestPackStore()
	deps := seedsTestDeps(store)
	h := NewSeedsHandler(deps)

	body := `{"seed_type":"process","content":{"sections":[{"section":"s","type":"instruction","text":"t"}]}}`
	req := httptest.NewRequest(http.MethodPost, "/v0/seeds", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateSeed_MissingSeedType(t *testing.T) {
	store := newTestPackStore()
	deps := seedsTestDeps(store)
	h := NewSeedsHandler(deps)

	body := `{"name":"my-seed","content":{"sections":[{"section":"s","type":"instruction","text":"t"}]}}`
	req := httptest.NewRequest(http.MethodPost, "/v0/seeds", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListSeeds_ByType(t *testing.T) {
	store := newTestPackStore()
	store.seeds["s1"] = packs.Seed{ID: "s1", ArtifactID: "s1", Name: "seed1", SeedType: "process", CreatedAt: time.Now()}
	store.seeds["s2"] = packs.Seed{ID: "s2", ArtifactID: "s2", Name: "seed2", SeedType: "skill", CreatedAt: time.Now()}
	deps := seedsTestDeps(store)
	h := NewSeedsHandler(deps)

	req := httptest.NewRequest(http.MethodGet, "/v0/seeds?type=process", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string][]packs.Seed
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	seeds := resp["seeds"]
	for _, s := range seeds {
		if s.SeedType != "process" {
			t.Errorf("unexpected seed type %q in filtered list", s.SeedType)
		}
	}
}

func TestGetSeed_NotFound(t *testing.T) {
	store := newTestPackStore()
	deps := seedsTestDeps(store)
	h := NewSeedsHandler(deps)

	req := httptest.NewRequest(http.MethodGet, "/v0/seeds/missing-id", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
