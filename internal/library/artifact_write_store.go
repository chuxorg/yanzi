package yanzilibrary

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/core/hash"
	"github.com/chuxorg/yanzi/internal/core/model"
	"github.com/chuxorg/yanzi/internal/storage"
	storagesqlite "github.com/chuxorg/yanzi/internal/storage/sqlite"
)

const (
	tombstoneDeletedKey   = "deleted"
	tombstoneDeletedAtKey = "deleted_at"
)

// ArtifactWriteStore is the internal write boundary for current local artifact,
// capture, and tombstone writes.
type ArtifactWriteStore struct {
	db *sql.DB
}

// CaptureWriteInput captures the current local prompt/response write shape.
type CaptureWriteInput struct {
	Author     string
	SourceType string
	Title      string
	Prompt     string
	Response   string
	Meta       json.RawMessage
	PrevHash   string
}

type artifactDeleteTarget struct {
	ID           string
	Hash         string
	PrevHash     string
	SourceType   string
	Meta         string
	Metadata     string
	TombstoneCol string
}

type execContexter interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// NewArtifactWriteStore creates a local artifact write boundary over db.
func NewArtifactWriteStore(db *sql.DB) *ArtifactWriteStore {
	return &ArtifactWriteStore{db: db}
}

// CreateCapture stores a local capture using current intent ledger semantics.
func (s *ArtifactWriteStore) CreateCapture(ctx context.Context, input CaptureWriteInput) (model.IntentRecord, error) {
	if s == nil || s.db == nil {
		return model.IntentRecord{}, fmt.Errorf("artifact write store is not initialized")
	}
	record, err := buildCaptureRecord(input)
	if err != nil {
		return model.IntentRecord{}, err
	}
	if err := s.insertCapture(ctx, record); err != nil {
		return model.IntentRecord{}, err
	}
	return record, nil
}

// CreateArtifact stores an artifact through the provider-compatible SQLite path.
func (s *ArtifactWriteStore) CreateArtifact(ctx context.Context, input storage.CreateArtifactInput) (Artifact, error) {
	if s == nil || s.db == nil {
		return Artifact{}, fmt.Errorf("artifact write store is not initialized")
	}
	provider := storagesqlite.FromDB(s.db)
	artifact, err := provider.CreateArtifact(ctx, input)
	if err != nil {
		return Artifact{}, err
	}
	return artifactFromStorage(artifact), nil
}

// Tombstone marks an intent or artifact deleted using current metadata column rules.
func (s *ArtifactWriteStore) Tombstone(ctx context.Context, id string, cascade, force bool) ([]string, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("artifact write store is not initialized")
	}
	head, err := s.getDeleteTarget(ctx, id)
	if err != nil {
		return nil, err
	}

	targets := []artifactDeleteTarget{head}
	if cascade {
		targets, err = s.collectCascadeTargets(ctx, head)
		if err != nil {
			return nil, err
		}
	}

	if !force {
		for _, target := range targets {
			refs, err := s.checkpointReferencesArtifact(ctx, target.ID)
			if err != nil {
				return nil, err
			}
			if len(refs) > 0 {
				return nil, fmt.Errorf("intent %s is referenced by checkpoint %s; use --force to tombstone it", target.ID, refs[0])
			}
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	updatedIDs := make([]string, 0, len(targets))
	for _, target := range targets {
		if err := updateTombstone(ctx, tx, target, true); err != nil {
			return nil, err
		}
		updatedIDs = append(updatedIDs, target.ID)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return updatedIDs, nil
}

// Restore removes current tombstone metadata from an intent or artifact.
func (s *ArtifactWriteStore) Restore(ctx context.Context, id string) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("artifact write store is not initialized")
	}
	target, err := s.getDeleteTarget(ctx, id)
	if err != nil {
		return err
	}
	return updateTombstone(ctx, s.db, target, false)
}

func buildCaptureRecord(input CaptureWriteInput) (model.IntentRecord, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	id, err := newCaptureID()
	if err != nil {
		return model.IntentRecord{}, err
	}

	record := model.IntentRecord{
		ID:         id,
		CreatedAt:  now,
		Author:     input.Author,
		SourceType: input.SourceType,
		Title:      input.Title,
		Prompt:     input.Prompt,
		Response:   input.Response,
		PrevHash:   input.PrevHash,
		Meta:       input.Meta,
	}
	sum, err := hash.HashIntent(record)
	if err != nil {
		return model.IntentRecord{}, err
	}
	record.Hash = sum
	return record, nil
}

func newCaptureID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(buf[:]), nil
}

func (s *ArtifactWriteStore) insertCapture(ctx context.Context, record model.IntentRecord) error {
	var title any
	if record.Title != "" {
		title = record.Title
	}
	var meta any
	if len(record.Meta) > 0 {
		meta = string(record.Meta)
	}
	var prevHash any
	if record.PrevHash != "" {
		prevHash = record.PrevHash
	}

	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO intents (id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash, class, type, content, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ID,
		record.CreatedAt,
		record.Author,
		record.SourceType,
		title,
		record.Prompt,
		record.Response,
		meta,
		prevHash,
		record.Hash,
		"intent",
		"prompt",
		record.Prompt,
		meta,
	)
	return err
}

func decodeTombstoneMeta(metaText string) (map[string]string, error) {
	payload := map[string]string{}
	if strings.TrimSpace(metaText) == "" {
		return payload, nil
	}
	if err := json.Unmarshal([]byte(metaText), &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func updatedTombstoneJSON(existingText string, deleted bool) (string, error) {
	payload, err := decodeTombstoneMeta(existingText)
	if err != nil {
		return "", err
	}

	if deleted {
		payload[tombstoneDeletedKey] = "true"
		payload[tombstoneDeletedAtKey] = time.Now().UTC().Format(time.RFC3339Nano)
	} else {
		delete(payload, tombstoneDeletedKey)
		delete(payload, tombstoneDeletedAtKey)
	}

	if len(payload) == 0 {
		return "", nil
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode tombstone metadata: %w", err)
	}
	return string(encoded), nil
}

func (s *ArtifactWriteStore) getDeleteTarget(ctx context.Context, id string) (artifactDeleteTarget, error) {
	var target artifactDeleteTarget
	var title sql.NullString
	var prevHash sql.NullString
	var meta sql.NullString
	var metadata sql.NullString
	row := s.db.QueryRowContext(ctx, `SELECT id, hash, prev_hash, source_type, title, meta, metadata FROM intents WHERE id = ?`, id)
	if err := row.Scan(&target.ID, &target.Hash, &prevHash, &target.SourceType, &title, &meta, &metadata); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return artifactDeleteTarget{}, fmt.Errorf("intent not found for ID %s", id)
		}
		return artifactDeleteTarget{}, err
	}
	if prevHash.Valid {
		target.PrevHash = prevHash.String
	}
	target.Meta = meta.String
	target.Metadata = metadata.String
	if target.SourceType == "artifact" {
		target.TombstoneCol = "meta"
	} else {
		target.TombstoneCol = "metadata"
	}
	return target, nil
}

func updateTombstone(ctx context.Context, db execContexter, target artifactDeleteTarget, deleted bool) error {
	existingText := target.Metadata
	if target.TombstoneCol == "meta" {
		existingText = target.Meta
	}
	updated, err := updatedTombstoneJSON(existingText, deleted)
	if err != nil {
		return err
	}

	var value any
	if updated != "" {
		value = updated
	}
	_, err = db.ExecContext(ctx, fmt.Sprintf(`UPDATE intents SET %s = ? WHERE id = ?`, target.TombstoneCol), value, target.ID)
	return err
}

func (s *ArtifactWriteStore) checkpointReferencesArtifact(ctx context.Context, artifactID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT hash, artifact_ids FROM checkpoints`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hashes []string
	for rows.Next() {
		var hash string
		var raw string
		if err := rows.Scan(&hash, &raw); err != nil {
			return nil, err
		}
		if strings.TrimSpace(raw) == "" {
			continue
		}
		var ids []string
		if err := json.Unmarshal([]byte(raw), &ids); err != nil {
			return nil, fmt.Errorf("decode checkpoint artifact ids: %w", err)
		}
		for _, id := range ids {
			if id == artifactID {
				hashes = append(hashes, hash)
				break
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hashes, nil
}

func (s *ArtifactWriteStore) collectCascadeTargets(ctx context.Context, head artifactDeleteTarget) ([]artifactDeleteTarget, error) {
	targets := []artifactDeleteTarget{head}
	queue := []artifactDeleteTarget{head}
	seen := map[string]struct{}{head.ID: {}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		rows, err := s.db.QueryContext(ctx, `SELECT id FROM intents WHERE prev_hash = ? ORDER BY created_at ASC, id ASC`, current.Hash)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				_ = rows.Close()
				return nil, err
			}
			if _, ok := seen[id]; ok {
				continue
			}
			target, err := s.getDeleteTarget(ctx, id)
			if err != nil {
				_ = rows.Close()
				return nil, err
			}
			seen[id] = struct{}{}
			targets = append(targets, target)
			queue = append(queue, target)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return nil, err
		}
		_ = rows.Close()
	}

	return targets, nil
}
