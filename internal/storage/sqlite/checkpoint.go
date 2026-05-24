package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/storage"
)

// CreateCheckpoint creates a checkpoint using current SQLite checkpoint semantics.
func (p *Provider) CreateCheckpoint(ctx context.Context, input storage.CreateCheckpointInput) (storage.Checkpoint, error) {
	if p == nil || p.db == nil {
		return storage.Checkpoint{}, storage.ErrProviderUnavailable
	}
	project := strings.TrimSpace(input.Project)
	if project == "" {
		return storage.Checkpoint{}, errors.New("project is required")
	}
	summary := strings.TrimSpace(input.Summary)
	if summary == "" {
		return storage.Checkpoint{}, errors.New("summary is required")
	}
	exists, err := p.ProjectExists(ctx, project)
	if err != nil {
		return storage.Checkpoint{}, err
	}
	if !exists {
		return storage.Checkpoint{}, fmt.Errorf("project not found: %s", project)
	}

	createdAt := time.Now().UTC().Format(time.RFC3339Nano)
	previousID, err := p.latestCheckpointID(ctx, project)
	if err != nil {
		return storage.Checkpoint{}, err
	}
	checkpoint := normalizeCheckpoint(storage.Checkpoint{
		Project:              project,
		Summary:              summary,
		CreatedAt:            createdAt,
		ArtifactIDs:          input.ArtifactIDs,
		PreviousCheckpointID: previousID,
	})
	hashValue, err := hashCheckpoint(checkpoint)
	if err != nil {
		return storage.Checkpoint{}, err
	}
	checkpoint.Hash = hashValue

	storedIDs := checkpoint.ArtifactIDs
	if storedIDs == nil {
		storedIDs = []string{}
	}
	artifactJSON, err := json.Marshal(storedIDs)
	if err != nil {
		return storage.Checkpoint{}, err
	}
	var prev any
	if checkpoint.PreviousCheckpointID != "" {
		prev = checkpoint.PreviousCheckpointID
	}

	_, err = p.db.ExecContext(
		ctx,
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
		return storage.Checkpoint{}, err
	}
	return checkpoint, nil
}

// ListCheckpoints returns project checkpoints ordered newest first.
func (p *Provider) ListCheckpoints(ctx context.Context, project string) ([]storage.Checkpoint, error) {
	if p == nil || p.db == nil {
		return nil, storage.ErrProviderUnavailable
	}
	project = strings.TrimSpace(project)
	if project == "" {
		return nil, errors.New("project is required")
	}
	exists, err := p.ProjectExists(ctx, project)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("project not found: %s", project)
	}
	rows, err := p.db.QueryContext(ctx, `SELECT hash, project, summary, created_at, artifact_ids, previous_checkpoint_id FROM checkpoints WHERE project = ? ORDER BY created_at DESC`, project)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCheckpoints(rows)
}

// ListAllCheckpoints returns all checkpoints using current CLI all-project ordering.
func (p *Provider) ListAllCheckpoints(ctx context.Context) ([]storage.Checkpoint, error) {
	if p == nil || p.db == nil {
		return nil, storage.ErrProviderUnavailable
	}
	rows, err := p.db.QueryContext(ctx, `SELECT hash, project, summary, created_at, artifact_ids, previous_checkpoint_id FROM checkpoints`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	checkpoints, err := scanCheckpoints(rows)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(checkpoints, func(i, j int) bool {
		if checkpoints[i].Project == checkpoints[j].Project {
			if checkpoints[i].CreatedAt == checkpoints[j].CreatedAt {
				return checkpoints[i].Hash > checkpoints[j].Hash
			}
			return checkpoints[i].CreatedAt > checkpoints[j].CreatedAt
		}
		return checkpoints[i].Project < checkpoints[j].Project
	})
	return checkpoints, nil
}

func scanCheckpoints(rows *sql.Rows) ([]storage.Checkpoint, error) {
	checkpoints := make([]storage.Checkpoint, 0)
	for rows.Next() {
		var checkpoint storage.Checkpoint
		var artifactText string
		var prev sql.NullString
		if err := rows.Scan(&checkpoint.Hash, &checkpoint.Project, &checkpoint.Summary, &checkpoint.CreatedAt, &artifactText, &prev); err != nil {
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

func (p *Provider) latestCheckpointID(ctx context.Context, project string) (string, error) {
	var hash string
	row := p.db.QueryRowContext(ctx, `SELECT hash FROM checkpoints WHERE project = ? ORDER BY created_at DESC LIMIT 1`, project)
	if err := row.Scan(&hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return hash, nil
}

func hashCheckpoint(checkpoint storage.Checkpoint) (string, error) {
	preimage, err := canonicalCheckpointPreimage(normalizeCheckpoint(checkpoint))
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(preimage)
	return hex.EncodeToString(sum[:]), nil
}

func normalizeCheckpoint(checkpoint storage.Checkpoint) storage.Checkpoint {
	out := checkpoint
	out.Project = normalizeNewlines(strings.TrimSpace(checkpoint.Project))
	out.Summary = normalizeNewlines(strings.TrimSpace(checkpoint.Summary))
	out.PreviousCheckpointID = normalizeNewlines(checkpoint.PreviousCheckpointID)
	if len(out.ArtifactIDs) > 0 {
		ids := make([]string, len(out.ArtifactIDs))
		for i, id := range out.ArtifactIDs {
			ids[i] = normalizeNewlines(id)
		}
		out.ArtifactIDs = ids
	}
	return out
}

func canonicalCheckpointPreimage(checkpoint storage.Checkpoint) ([]byte, error) {
	if strings.TrimSpace(checkpoint.Project) == "" {
		return nil, errors.New("project is required for hashing")
	}
	if strings.TrimSpace(checkpoint.Summary) == "" {
		return nil, errors.New("summary is required for hashing")
	}
	if checkpoint.CreatedAt == "" {
		return nil, errors.New("created_at is required for hashing")
	}
	createdAt, err := normalizeRFC3339(checkpoint.CreatedAt)
	if err != nil {
		return nil, errors.New("created_at must be RFC3339")
	}
	artifactIDs := checkpoint.ArtifactIDs
	if artifactIDs == nil {
		artifactIDs = []string{}
	}
	artifactJSON, err := json.Marshal(artifactIDs)
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	b.WriteByte('{')
	first := true
	addStringField(&b, &first, "project", checkpoint.Project)
	addStringField(&b, &first, "created_at", createdAt)
	addStringField(&b, &first, "summary", checkpoint.Summary)
	addRawField(&b, &first, "artifact_ids", artifactJSON)
	if checkpoint.PreviousCheckpointID != "" {
		addStringField(&b, &first, "previous_checkpoint_id", checkpoint.PreviousCheckpointID)
	}
	b.WriteByte('}')
	return []byte(b.String()), nil
}

func normalizeRFC3339(value string) (string, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return "", err
	}
	return parsed.UTC().Format(time.RFC3339Nano), nil
}

func addStringField(b *strings.Builder, first *bool, name string, value string) {
	if !*first {
		b.WriteByte(',')
	}
	*first = false
	b.WriteByte('"')
	b.WriteString(name)
	b.WriteString(`":`)
	encoded, _ := json.Marshal(value)
	b.Write(encoded)
}

func addRawField(b *strings.Builder, first *bool, name string, raw json.RawMessage) {
	if !*first {
		b.WriteByte(',')
	}
	*first = false
	b.WriteByte('"')
	b.WriteString(name)
	b.WriteString(`":`)
	b.Write(raw)
}

func normalizeNewlines(value string) string {
	if value == "" {
		return value
	}
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return value
}
