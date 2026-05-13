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

// DefaultRehydrateFallbackLimit is the deterministic fallback window when no checkpoint exists.
const DefaultRehydrateFallbackLimit = 10

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

// RehydratePayload contains the checkpoint boundary or fallback state and the ordered intents to render.
type RehydratePayload struct {
	Project          string
	LatestCheckpoint *Checkpoint
	Intents          []Intent
	Fallback         bool
	FallbackReason   string
	FallbackLimit    int
}

// RehydrateProject loads the latest checkpoint and subsequent intents for a project.
//
// If no checkpoint exists, the payload falls back to the latest project intents
// so operational continuity remains available during recovery.
func RehydrateProject(project string) (*RehydratePayload, error) {
	return RehydrateProjectWithFallback(project, DefaultRehydrateFallbackLimit)
}

// RehydrateProjectWithFallback loads checkpoint-based rehydration data or a recent-capture fallback.
func RehydrateProjectWithFallback(project string, fallbackLimit int) (*RehydratePayload, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return nil, errors.New("project is required")
	}
	if fallbackLimit <= 0 {
		fallbackLimit = DefaultRehydrateFallbackLimit
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
		intents, err := recentProjectIntents(ctx, db, project, fallbackLimit)
		if err != nil {
			return nil, err
		}
		return &RehydratePayload{
			Project:        project,
			Intents:        intents,
			Fallback:       true,
			FallbackReason: ErrCheckpointNotFound.Error(),
			FallbackLimit:  fallbackLimit,
		}, nil
	}

	intents, err := intentsSinceCheckpoint(ctx, db, project, latest.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &RehydratePayload{
		Project:          project,
		LatestCheckpoint: latest,
		Intents:          intents,
		FallbackLimit:    fallbackLimit,
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

// intentsSinceCheckpoint returns project intents created strictly after the provided checkpoint timestamp.
func intentsSinceCheckpoint(ctx context.Context, db *sql.DB, project, checkpointCreatedAt string) ([]Intent, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash, metadata
		FROM intents
		WHERE created_at > ? AND source_type <> 'artifact'
		ORDER BY created_at ASC, id ASC`,
		checkpointCreatedAt,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProjectIntents(rows, project, 0)
}

// recentProjectIntents returns the most recent intents for the project in chronological order.
func recentProjectIntents(ctx context.Context, db *sql.DB, project string, limit int) ([]Intent, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash, metadata
		FROM intents
		WHERE source_type <> 'artifact'
		ORDER BY created_at DESC, id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProjectIntents(rows, project, limit)
}

func scanProjectIntents(rows *sql.Rows, project string, limit int) ([]Intent, error) {
	intents := make([]Intent, 0)
	for rows.Next() {
		intent, metaText, metadataText, err := scanRehydrateIntent(rows)
		if err != nil {
			return nil, err
		}

		combinedMeta, err := mergedRehydrateMeta(metaText, metadataText)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(combinedMeta["project"]) != project {
			continue
		}

		if len(combinedMeta) > 0 {
			metaJSON, err := json.Marshal(combinedMeta)
			if err != nil {
				return nil, fmt.Errorf("encode merged rehydrate meta: %w", err)
			}
			intent.Meta = metaJSON
		} else {
			intent.Meta = nil
		}

		intents = append(intents, intent)
		if limit > 0 && len(intents) == limit {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if limit > 0 {
		for left, right := 0, len(intents)-1; left < right; left, right = left+1, right-1 {
			intents[left], intents[right] = intents[right], intents[left]
		}
	}

	return intents, nil
}

func scanRehydrateIntent(rows *sql.Rows) (Intent, string, string, error) {
	var createdAtText string
	var meta sql.NullString
	var metadata sql.NullString
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
		&metadata,
	); err != nil {
		return Intent{}, "", "", err
	}

	createdAt, err := time.Parse(time.RFC3339Nano, createdAtText)
	if err != nil {
		return Intent{}, "", "", fmt.Errorf("parse intent created_at for %s: %w", intent.ID, err)
	}
	intent.CreatedAt = createdAt
	if title.Valid {
		intent.Title = title.String
	}
	if prevHash.Valid {
		intent.PrevHash = prevHash.String
	}

	return intent, meta.String, metadata.String, nil
}

func mergedRehydrateMeta(metaText, metadataText string) (map[string]string, error) {
	meta := map[string]string{}
	if strings.TrimSpace(metaText) != "" {
		decoded, err := decodeRehydrateMeta(metaText)
		if err != nil {
			return nil, err
		}
		for key, value := range decoded {
			meta[key] = value
		}
	}
	if strings.TrimSpace(metadataText) != "" {
		decoded, err := decodeRehydrateMeta(metadataText)
		if err != nil {
			return nil, err
		}
		for key, value := range decoded {
			meta[key] = value
		}
	}
	return meta, nil
}

func decodeRehydrateMeta(raw string) (map[string]string, error) {
	payload := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("decode rehydrate meta: %w", err)
	}
	meta := make(map[string]string, len(payload))
	for key, value := range payload {
		if text, ok := value.(string); ok {
			meta[key] = text
		}
	}
	return meta, nil
}
