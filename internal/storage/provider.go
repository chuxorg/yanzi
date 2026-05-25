package storage

import (
	"context"
	"database/sql"
)

// Provider is the current internal storage boundary.
//
// The SQLDB method intentionally preserves existing SQLite-backed call sites for
// CAP-001 Phase 1. Future phases can move operations behind narrower methods
// without changing CLI contracts.
type Provider interface {
	ArtifactOperations
	ProjectOperations
	CheckpointOperations
	VerificationOperations
	ImportExportOperations

	Name() ProviderName
	Health(context.Context) Health
	SQLDB() *sql.DB
	Close() error
}

// ArtifactOperations represents artifact persistence and retrieval capability.
type ArtifactOperations interface {
	Artifacts() bool
	CreateArtifact(context.Context, CreateArtifactInput) (Artifact, error)
	ListArtifacts(context.Context, ArtifactQuery) ([]Artifact, error)
	ListVisibleContextArtifacts(context.Context, ContextArtifactQuery) ([]Artifact, error)
	GetVisibleContextArtifact(context.Context, string, string) (Artifact, error)
}

// ProjectOperations represents project persistence and retrieval capability.
type ProjectOperations interface {
	Projects() bool
	CreateProject(context.Context, CreateProjectInput) (Project, error)
	ListProjects(context.Context) ([]Project, error)
	ProjectExists(context.Context, string) (bool, error)
}

// CheckpointOperations represents checkpoint persistence and retrieval capability.
type CheckpointOperations interface {
	Checkpoints() bool
	CreateCheckpoint(context.Context, CreateCheckpointInput) (Checkpoint, error)
	ListCheckpoints(context.Context, string) ([]Checkpoint, error)
	ListAllCheckpoints(context.Context) ([]Checkpoint, error)
}

// VerificationOperations represents local digest verification capability.
type VerificationOperations interface {
	Verification() bool
	GetVerificationIntent(context.Context, string) (IntentRecord, error)
	GetVerificationIntentByHash(context.Context, string) (IntentRecord, error)
}

// ImportExportOperations represents deterministic local import/export capability.
type ImportExportOperations interface {
	ImportExport() bool
	ListExportItems(context.Context, ExportQuery) ([]ExportItem, int, error)
}
