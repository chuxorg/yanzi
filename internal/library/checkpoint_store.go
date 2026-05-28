package yanzilibrary

import (
	"context"
	"database/sql"
	"errors"

	"github.com/chuxorg/yanzi/internal/storage"
	storagesqlite "github.com/chuxorg/yanzi/internal/storage/sqlite"
)

// CheckpointStore provides persistence for checkpoints.
type CheckpointStore struct {
	provider storage.Provider
}

// NewCheckpointStore constructs a CheckpointStore using the provided database handle.
func NewCheckpointStore(db *sql.DB) *CheckpointStore {
	return &CheckpointStore{provider: storagesqlite.FromDB(db)}
}

// CreateCheckpoint creates a new checkpoint artifact for a project.
func (s *CheckpointStore) CreateCheckpoint(ctx context.Context, project, summary string, artifactIDs []string) (Checkpoint, error) {
	if s == nil || s.provider == nil {
		return Checkpoint{}, errors.New("checkpoint store is not initialized")
	}
	checkpoint, err := s.provider.CreateCheckpoint(ctx, storage.CreateCheckpointInput{
		Project:     project,
		Summary:     summary,
		ArtifactIDs: artifactIDs,
	})
	if err != nil {
		return Checkpoint{}, err
	}
	return checkpointFromStorage(checkpoint), nil
}

// ListCheckpoints returns checkpoints for a project ordered by creation time, newest first.
func (s *CheckpointStore) ListCheckpoints(ctx context.Context, project string) ([]Checkpoint, error) {
	if s == nil || s.provider == nil {
		return nil, errors.New("checkpoint store is not initialized")
	}
	checkpoints, err := s.provider.ListCheckpoints(ctx, project)
	if err != nil {
		return nil, err
	}
	result := make([]Checkpoint, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		result = append(result, checkpointFromStorage(checkpoint))
	}
	return result, nil
}

// CreateCheckpoint is a convenience wrapper for CheckpointStore.CreateCheckpoint.
func CreateCheckpoint(ctx context.Context, db *sql.DB, project, summary string, artifactIDs []string) (Checkpoint, error) {
	return NewCheckpointStore(db).CreateCheckpoint(ctx, project, summary, artifactIDs)
}

// ListCheckpoints is a convenience wrapper for CheckpointStore.ListCheckpoints.
func ListCheckpoints(ctx context.Context, db *sql.DB, project string) ([]Checkpoint, error) {
	return NewCheckpointStore(db).ListCheckpoints(ctx, project)
}

// ListAllCheckpoints is a convenience wrapper for all-project checkpoint listing.
func ListAllCheckpoints(ctx context.Context, db *sql.DB) ([]Checkpoint, error) {
	provider := storagesqlite.FromDB(db)
	checkpoints, err := provider.ListAllCheckpoints(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Checkpoint, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		result = append(result, checkpointFromStorage(checkpoint))
	}
	return result, nil
}

func checkpointFromStorage(checkpoint storage.Checkpoint) Checkpoint {
	return Checkpoint{
		Project:              checkpoint.Project,
		Summary:              checkpoint.Summary,
		CreatedAt:            checkpoint.CreatedAt,
		ArtifactIDs:          checkpoint.ArtifactIDs,
		PreviousCheckpointID: checkpoint.PreviousCheckpointID,
		Hash:                 checkpoint.Hash,
	}
}

// projectExists checks whether a project row exists for the provided project name.
func projectExists(ctx context.Context, db *sql.DB, project string) (bool, error) {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(1) FROM projects WHERE name = ?`, project).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}
