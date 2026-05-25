package storage

import (
	"encoding/json"
	"time"
)

// ProviderName identifies a storage provider implementation.
type ProviderName string

const (
	// ProviderSQLite is the embedded local SQLite provider.
	ProviderSQLite ProviderName = "sqlite"
)

const (
	// ArtifactClassIntent is the current intent artifact class.
	ArtifactClassIntent = "intent"
	// ArtifactClassContext is the current context artifact class.
	ArtifactClassContext = "context"
	// ContextScopeGlobal is the current globally visible context scope.
	ContextScopeGlobal = "global"
	// ContextScopeProject is the current project-visible context scope.
	ContextScopeProject = "project"
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

// ContextArtifactQuery captures current context visibility dimensions.
type ContextArtifactQuery struct {
	ActiveProject  string
	Type           string
	Scope          string
	Project        string
	IncludeDeleted bool
	AllProjects    bool
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

// ExportItemKind identifies the current export timeline item category.
type ExportItemKind string

const (
	// ExportItemCheckpoint is a checkpoint boundary in a deterministic export.
	ExportItemCheckpoint ExportItemKind = "checkpoint"
	// ExportItemCapture is a captured prompt/response record in a deterministic export.
	ExportItemCapture ExportItemKind = "capture"
	// ExportItemMeta is a current meta-command/event record in a deterministic export.
	ExportItemMeta ExportItemKind = "meta"
)

// IntentRecord is the provider-level form of the current intent record used by verification reads.
type IntentRecord struct {
	ID         string
	CreatedAt  string
	Author     string
	SourceType string
	Title      string
	Prompt     string
	Response   string
	Meta       json.RawMessage
	PrevHash   string
	Hash       string
}

// ExportCapture is the provider-level capture payload used by current export renderers.
type ExportCapture struct {
	ID        string
	CreatedAt string
	Author    string
	Source    string
	Title     string
	Hash      string
	Prompt    string
	Response  string
	Metadata  map[string]string
}

// ExportMeta is the provider-level meta event payload used by current export renderers.
type ExportMeta struct {
	CreatedAt string
	Command   string
	Value     string
}

// ExportItem is the provider-level event used by current deterministic exports.
type ExportItem struct {
	Kind       ExportItemKind
	Timestamp  string
	RowID      int64
	Capture    ExportCapture
	Checkpoint Checkpoint
	Meta       ExportMeta
}

// CreateArtifactInput captures current artifact creation inputs.
type CreateArtifactInput struct {
	Project  string
	Class    string
	Type     string
	Scope    string
	Title    string
	Content  string
	Metadata string
}

// Artifact is the provider-level artifact record used by current storage behavior.
type Artifact struct {
	ID        string
	Class     string
	Type      string
	Scope     string
	Project   string
	Title     string
	Content   string
	Metadata  string
	CreatedAt string
}

// CreateProjectInput captures current project creation inputs.
type CreateProjectInput struct {
	Name        string
	Description string
}

// Project is the provider-level project record used by current storage behavior.
type Project struct {
	Name        string
	Description string
	CreatedAt   time.Time
}

// CreateCheckpointInput captures current checkpoint creation inputs.
type CreateCheckpointInput struct {
	Project     string
	Summary     string
	ArtifactIDs []string
}

// Checkpoint is the provider-level checkpoint record used by current storage behavior.
type Checkpoint struct {
	Project              string
	Summary              string
	CreatedAt            string
	ArtifactIDs          []string
	PreviousCheckpointID string
	Hash                 string
}
