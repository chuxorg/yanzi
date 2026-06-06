package packs

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a requested seed or pack does not exist.
var ErrNotFound = errors.New("packs: record not found")

// PackStore persists and retrieves Seeds, Packs, and TokenUsage records.
type PackStore interface {
	// Seed operations
	CreateSeed(ctx context.Context, seed Seed) (Seed, error)
	GetSeed(ctx context.Context, artifactID string) (Seed, error)
	GetSeedByName(ctx context.Context, name string) (Seed, error)
	ListSeeds(ctx context.Context, filter SeedFilter) ([]Seed, error)
	DeleteSeed(ctx context.Context, artifactID string) error

	// Pack operations
	CreatePack(ctx context.Context, pack Pack) (Pack, error)
	GetPack(ctx context.Context, artifactID string) (Pack, error)
	GetPackByName(ctx context.Context, name string) (Pack, error)
	ListPacks(ctx context.Context, filter PackFilter) ([]Pack, error)
	DeletePack(ctx context.Context, artifactID string) error

	// Token usage
	RecordTokenUsage(ctx context.Context, usage TokenUsage) error
	GetTokenUsage(ctx context.Context, filter TokenFilter) (TokenUsageSummary, error)
}

// SeedFilter constrains seed list queries.
type SeedFilter struct {
	SeedType    string
	MinRoleBits RoleBits
	Tags        []string
	Name        string
}

// PackFilter constrains pack list queries.
type PackFilter struct {
	Role RoleBits
	Tags []string
	Name string
}

// TokenFilter constrains token usage queries.
type TokenFilter struct {
	Project string
	Phase   string
	Task    string
	Since   time.Time
}

// TokenUsage records a single token consumption event.
type TokenUsage struct {
	ID         string
	Project    string
	Phase      string
	Task       string
	ArtifactID string
	PackID     string
	TokenCount int
	Approximate bool
	ModelHint  string
	RecordedAt time.Time
}

// TokenUsageSummary aggregates token usage by project.
type TokenUsageSummary struct {
	Project     string
	TotalTokens int
	ByPhase     map[string]int
	ByTask      map[string]int
	Approximate bool
}
