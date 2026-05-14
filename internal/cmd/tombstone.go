package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
	"github.com/chuxorg/yanzi/internal/sqliteruntime"
)

const (
	tombstoneDeletedKey   = "deleted"
	tombstoneDeletedAtKey = "deleted_at"
)

type deleteTarget struct {
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

func mergedIntentMetadata(metaText, metadataText string) (map[string]string, error) {
	meta := map[string]string{}
	if strings.TrimSpace(metaText) != "" {
		decoded, err := decodeStringMeta(metaText)
		if err != nil {
			return nil, err
		}
		for key, value := range decoded {
			meta[key] = value
		}
	}
	if strings.TrimSpace(metadataText) != "" {
		decoded, err := decodeStringMeta(metadataText)
		if err != nil {
			return nil, err
		}
		for key, value := range decoded {
			meta[key] = value
		}
	}
	return meta, nil
}

func isDeletedMetadata(meta map[string]string) bool {
	return strings.EqualFold(strings.TrimSpace(meta[tombstoneDeletedKey]), "true")
}

func updatedTombstoneJSON(existingText string, deleted bool) (string, error) {
	payload := map[string]string{}
	if strings.TrimSpace(existingText) != "" {
		decoded, err := decodeStringMeta(existingText)
		if err != nil {
			return "", err
		}
		for key, value := range decoded {
			payload[key] = value
		}
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

func getDeleteTarget(ctx context.Context, db *sql.DB, id string) (deleteTarget, error) {
	var target deleteTarget
	var title sql.NullString
	var prevHash sql.NullString
	var meta sql.NullString
	var metadata sql.NullString
	row := db.QueryRowContext(ctx, `SELECT id, hash, prev_hash, source_type, title, meta, metadata FROM intents WHERE id = ?`, id)
	if err := row.Scan(&target.ID, &target.Hash, &prevHash, &target.SourceType, &title, &meta, &metadata); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return deleteTarget{}, fmt.Errorf("intent not found for ID %s", id)
		}
		return deleteTarget{}, err
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

func updateTombstone(ctx context.Context, db execContexter, target deleteTarget, deleted bool) error {
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
	query := fmt.Sprintf(`UPDATE intents SET %s = ? WHERE id = ?`, target.TombstoneCol)
	switch typed := db.(type) {
	case *sql.Tx:
		_, err = sqliteruntime.ExecTxContext(ctx, typed, yanzilibrary.ResolvedDBPath(), "update tombstone", query, value, target.ID)
	case *sql.DB:
		_, err = sqliteruntime.ExecContext(ctx, typed, yanzilibrary.ResolvedDBPath(), "update tombstone", query, value, target.ID)
	default:
		_, err = db.ExecContext(ctx, query, value, target.ID)
	}
	return err
}

func checkpointReferencesArtifact(ctx context.Context, db *sql.DB, artifactID string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT hash, artifact_ids FROM checkpoints`)
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

func collectCascadeTargets(ctx context.Context, db *sql.DB, head deleteTarget) ([]deleteTarget, error) {
	targets := []deleteTarget{head}
	queue := []deleteTarget{head}
	seen := map[string]struct{}{head.ID: {}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		rows, err := db.QueryContext(ctx, `SELECT id FROM intents WHERE prev_hash = ? ORDER BY created_at ASC, id ASC`, current.Hash)
		if err != nil {
			return nil, err
		}
		ids := make([]string, 0)
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				_ = rows.Close()
				return nil, err
			}
			ids = append(ids, id)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return nil, err
		}
		_ = rows.Close()

		for _, id := range ids {
			if _, ok := seen[id]; ok {
				continue
			}
			target, err := getDeleteTarget(ctx, db, id)
			if err != nil {
				return nil, err
			}
			seen[id] = struct{}{}
			targets = append(targets, target)
			queue = append(queue, target)
		}
	}

	return targets, nil
}

func performDelete(ctx context.Context, db *sql.DB, id string, cascade, force bool) ([]string, error) {
	head, err := getDeleteTarget(ctx, db, id)
	if err != nil {
		return nil, err
	}

	targets := []deleteTarget{head}
	if cascade {
		targets, err = collectCascadeTargets(ctx, db, head)
		if err != nil {
			return nil, err
		}
	}

	if !force {
		for _, target := range targets {
			refs, err := checkpointReferencesArtifact(ctx, db, target.ID)
			if err != nil {
				return nil, err
			}
			if len(refs) > 0 {
				return nil, fmt.Errorf("intent %s is referenced by checkpoint %s; use --force to tombstone it", target.ID, refs[0])
			}
		}
	}

	updatedIDs := make([]string, 0, len(targets))
	if err := sqliteruntime.RunTx(ctx, db, yanzilibrary.ResolvedDBPath(), "tombstone intents", func(tx *sql.Tx) error {
		updatedIDs = updatedIDs[:0]
		for _, target := range targets {
			if err := updateTombstone(ctx, tx, target, true); err != nil {
				return err
			}
			updatedIDs = append(updatedIDs, target.ID)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return updatedIDs, nil
}

func performRestore(ctx context.Context, db *sql.DB, id string) error {
	target, err := getDeleteTarget(ctx, db, id)
	if err != nil {
		return err
	}
	return updateTombstone(ctx, db, target, false)
}
