package storage

// ProviderName identifies a storage provider implementation.
type ProviderName string

const (
	// ProviderSQLite is the embedded local SQLite provider.
	ProviderSQLite ProviderName = "sqlite"
)

// HealthStatus reports internal provider readiness without exposing a CLI surface.
type HealthStatus string

const (
	HealthReady       HealthStatus = "ready"
	HealthUnavailable HealthStatus = "unavailable"
)

// Health describes the internal provider health state.
type Health struct {
	Provider ProviderName
	Status   HealthStatus
	Path     string
	Error    string
}

// ArtifactQuery captures the current artifact list dimensions.
type ArtifactQuery struct {
	Project        string
	Class          string
	Type           string
	IncludeDeleted bool
}

// ProjectQuery captures current project lookup dimensions.
type ProjectQuery struct {
	Name string
}

// CheckpointQuery captures current checkpoint list dimensions.
type CheckpointQuery struct {
	Project string
}

// ExportQuery captures current deterministic local export dimensions.
type ExportQuery struct {
	Project        string
	MetaFilters    map[string]string
	IncludeDeleted bool
}

// VerificationQuery captures current hash verification dimensions.
type VerificationQuery struct {
	ID string
}
