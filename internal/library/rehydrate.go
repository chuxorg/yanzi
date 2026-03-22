package yanzilibrary

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrCheckpointNotFound indicates that no checkpoint exists for the requested project.
var ErrCheckpointNotFound = errors.New("checkpoint not found")

// Intent represents an intent artifact loaded from the intents table for rehydration.
type Intent struct {
	ID         string
	CreatedAt  time.Time
	Author     string
	SourceType string
	Title      string
	Prompt     string
	Response   string
	Meta       json.RawMessage
	PrevHash   string
	Hash       string
}

// RehydratePayload contains the latest checkpoint and the intents created after it.
type RehydratePayload struct {
	Project          string
	LatestCheckpoint Checkpoint
	IntentsSince     []Intent
}

// RehydrateProject loads the latest checkpoint and subsequent intents for a project.
func RehydrateProject(project string) (*RehydratePayload, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return nil, errors.New("project is required")
	}

	db, err := InitDB()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	ctx := context.Background()
	exists, err := projectExists(ctx, db, project)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ProjectNotFoundError{Name: project}
	}

	latest, err := latestCheckpointByProject(ctx, db, project)
	if err != nil {
		return nil, err
	}
	if latest == nil {
		return nil, ErrCheckpointNotFound
	}

	intents, err := intentsSinceCheckpoint(ctx, db, latest.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &RehydratePayload{
		Project:          project,
		LatestCheckpoint: *latest,
		IntentsSince:     intents,
	}, nil
}

// latestCheckpointByProject returns the latest checkpoint for a project by created_at descending.
func latestCheckpointByProject(ctx context.Context, db *sql.DB, project string) (*Checkpoint, error) {
	row := db.QueryRowContext(ctx, `SELECT hash, project, summary, created_at, artifact_ids, previous_checkpoint_id FROM checkpoints WHERE project = ? ORDER BY created_at DESC LIMIT 1`, project)

	var checkpoint Checkpoint
	var artifactText string
	var prev sql.NullString
	if err := row.Scan(
		&checkpoint.Hash,
		&checkpoint.Project,
		&checkpoint.Summary,
		&checkpoint.CreatedAt,
		&artifactText,
		&prev,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if artifactText != "" {
		if err := json.Unmarshal([]byte(artifactText), &checkpoint.ArtifactIDs); err != nil {
			return nil, fmt.Errorf("decode checkpoint artifact_ids: %w", err)
		}
	}
	if prev.Valid {
		checkpoint.PreviousCheckpointID = prev.String
	}

	return &checkpoint, nil
}

// intentsSinceCheckpoint returns intents created strictly after the provided checkpoint timestamp.
func intentsSinceCheckpoint(ctx context.Context, db *sql.DB, checkpointCreatedAt string) ([]Intent, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash
		FROM intents
		WHERE created_at > ? AND source_type <> 'artifact'
		ORDER BY created_at ASC, id ASC`,
		checkpointCreatedAt,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	intents := make([]Intent, 0)
	for rows.Next() {
		var createdAtText string
		var meta sql.NullString
		var title sql.NullString
		var prevHash sql.NullString
		var intent Intent
		if err := rows.Scan(
			&intent.ID,
			&createdAtText,
			&intent.Author,
			&intent.SourceType,
			&title,
			&intent.Prompt,
			&intent.Response,
			&meta,
			&prevHash,
			&intent.Hash,
		); err != nil {
			return nil, err
		}
		createdAt, err := time.Parse(time.RFC3339Nano, createdAtText)
		if err != nil {
			return nil, fmt.Errorf("parse intent created_at for %s: %w", intent.ID, err)
		}
		intent.CreatedAt = createdAt
		if title.Valid {
			intent.Title = title.String
		}
		if meta.Valid {
			intent.Meta = json.RawMessage(meta.String)
		}
		if prevHash.Valid {
			intent.PrevHash = prevHash.String
		}
		intents = append(intents, intent)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return intents, nil
}
