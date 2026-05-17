package yanzilibrary

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/sqliteruntime"
)

// CheckpointStore provides persistence for checkpoints.
type CheckpointStore struct {
	db *sql.DB
}

// NewCheckpointStore constructs a CheckpointStore using the provided database handle.
func NewCheckpointStore(db *sql.DB) *CheckpointStore {
	return &CheckpointStore{db: db}
}

// CreateCheckpoint creates a new checkpoint artifact for a project.
func (s *CheckpointStore) CreateCheckpoint(ctx context.Context, project, summary string, artifactIDs []string) (Checkpoint, error) {
	if s == nil || s.db == nil {
		return Checkpoint{}, errors.New("checkpoint store is not initialized")
	}

	project = strings.TrimSpace(project)
	if project == "" {
		return Checkpoint{}, CheckpointValidationError{Field: "project", Message: "is required"}
	}
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return Checkpoint{}, CheckpointValidationError{Field: "summary", Message: "is required"}
	}

	exists, err := projectExists(ctx, s.db, project)
	if err != nil {
		return Checkpoint{}, err
	}
	if !exists {
		return Checkpoint{}, ProjectNotFoundError{Name: project}
	}

	createdAt := time.Now().UTC().Format(time.RFC3339Nano)
	previousID, err := s.latestCheckpointID(ctx, project)
	if err != nil {
		return Checkpoint{}, err
	}

	checkpoint := Checkpoint{
		Project:              project,
		Summary:              summary,
		CreatedAt:            createdAt,
		ArtifactIDs:          artifactIDs,
		PreviousCheckpointID: previousID,
	}
	checkpoint = checkpoint.Normalize()

	hashValue, err := HashCheckpoint(checkpoint)
	if err != nil {
		return Checkpoint{}, err
	}
	checkpoint.Hash = hashValue

	storedIDs := checkpoint.ArtifactIDs
	if storedIDs == nil {
		storedIDs = []string{}
	}
	artifactJSON, err := json.Marshal(storedIDs)
	if err != nil {
		return Checkpoint{}, err
	}

	var prev any
	if checkpoint.PreviousCheckpointID != "" {
		prev = checkpoint.PreviousCheckpointID
	}

	_, err = sqliteruntime.ExecContext(
		ctx,
		s.db,
		ResolvedDBPath(),
		"create checkpoint",
		`INSERT INTO checkpoints (hash, project, summary, created_at, artifact_ids, previous_checkpoint_id)
		VALUES (?, ?, ?, ?, ?, ?)`,
		checkpoint.Hash,
		checkpoint.Project,
		checkpoint.Summary,
		checkpoint.CreatedAt,
		string(artifactJSON),
		prev,
	)
	if err != nil {
		return Checkpoint{}, err
	}

	return checkpoint, nil
}

// ListCheckpoints returns checkpoints for a project ordered by creation time, newest first.
func (s *CheckpointStore) ListCheckpoints(ctx context.Context, project string) ([]Checkpoint, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("checkpoint store is not initialized")
	}

	project = strings.TrimSpace(project)
	if project == "" {
		return nil, CheckpointValidationError{Field: "project", Message: "is required"}
	}

	exists, err := projectExists(ctx, s.db, project)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ProjectNotFoundError{Name: project}
	}

	rows, err := s.db.QueryContext(ctx, `SELECT hash, project, summary, created_at, artifact_ids, previous_checkpoint_id FROM checkpoints WHERE project = ? ORDER BY created_at DESC`, project)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checkpoints := []Checkpoint{}
	for rows.Next() {
		var checkpoint Checkpoint
		var artifactText string
		var prev sql.NullString
		if err := rows.Scan(
			&checkpoint.Hash,
			&checkpoint.Project,
			&checkpoint.Summary,
			&checkpoint.CreatedAt,
			&artifactText,
			&prev,
		); err != nil {
			return nil, err
		}
		if artifactText != "" {
			if err := json.Unmarshal([]byte(artifactText), &checkpoint.ArtifactIDs); err != nil {
				return nil, err
			}
		}
		if prev.Valid {
			checkpoint.PreviousCheckpointID = prev.String
		}
		checkpoints = append(checkpoints, checkpoint)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return checkpoints, nil
}

// CreateCheckpoint is a convenience wrapper for CheckpointStore.CreateCheckpoint.
func CreateCheckpoint(ctx context.Context, db *sql.DB, project, summary string, artifactIDs []string) (Checkpoint, error) {
	return NewCheckpointStore(db).CreateCheckpoint(ctx, project, summary, artifactIDs)
}

// ListCheckpoints is a convenience wrapper for CheckpointStore.ListCheckpoints.
func ListCheckpoints(ctx context.Context, db *sql.DB, project string) ([]Checkpoint, error) {
	return NewCheckpointStore(db).ListCheckpoints(ctx, project)
}

// latestCheckpointID returns the most recent checkpoint hash for a project, or empty if none exist.
func (s *CheckpointStore) latestCheckpointID(ctx context.Context, project string) (string, error) {
	var hash string
	row := s.db.QueryRowContext(ctx, `SELECT hash FROM checkpoints WHERE project = ? ORDER BY created_at DESC LIMIT 1`, project)
	if err := row.Scan(&hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return hash, nil
}

// projectExists checks whether a project row exists for the provided project name.
func projectExists(ctx context.Context, db *sql.DB, project string) (bool, error) {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(1) FROM projects WHERE name = ?`, project).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}
