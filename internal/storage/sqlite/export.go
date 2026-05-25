package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/chuxorg/yanzi/internal/storage"
)

// ListExportItems returns the current SQLite-backed export timeline source data.
func (p *Provider) ListExportItems(ctx context.Context, query storage.ExportQuery) ([]storage.ExportItem, int, error) {
	if p == nil || p.db == nil {
		return nil, 0, storage.ErrProviderUnavailable
	}

	captures, captureCount, err := p.listExportCaptures(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	if len(query.MetaFilters) > 0 {
		return mergeExportItems(captures, nil), captureCount, nil
	}

	checkpoints, err := p.listExportCheckpoints(ctx, strings.TrimSpace(query.Project))
	if err != nil {
		return nil, 0, err
	}
	return mergeExportItems(captures, checkpoints), captureCount, nil
}

func (p *Provider) listExportCaptures(ctx context.Context, query storage.ExportQuery) ([]storage.ExportItem, int, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT rowid, id, created_at, author, source_type, title, prompt, response, hash, meta, metadata
		FROM intents
		WHERE source_type <> 'artifact'
		ORDER BY created_at ASC, rowid ASC`)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	project := strings.TrimSpace(query.Project)
	items := make([]storage.ExportItem, 0)
	captureCount := 0
	for rows.Next() {
		var (
			rowID int64
			item  storage.ExportItem
			meta  sql.NullString
			extra sql.NullString
			title sql.NullString
		)
		item.Kind = storage.ExportItemCapture
		if err := rows.Scan(
			&rowID,
			&item.Capture.ID,
			&item.Capture.CreatedAt,
			&item.Capture.Author,
			&item.Capture.Source,
			&title,
			&item.Capture.Prompt,
			&item.Capture.Response,
			&item.Capture.Hash,
			&meta,
			&extra,
		); err != nil {
			return nil, 0, err
		}
		item.Timestamp = item.Capture.CreatedAt
		item.RowID = rowID
		if title.Valid {
			item.Capture.Title = title.String
		}

		metadata, err := mergedStringMetadata(meta.String, extra.String)
		if err != nil {
			continue
		}
		if strings.TrimSpace(metadata["project"]) != project {
			continue
		}
		if !query.IncludeDeleted && metadataDeleted(metadata) {
			continue
		}
		if len(query.MetaFilters) > 0 && !stringMetadataMatchesAll(metadata, query.MetaFilters) {
			continue
		}
		item.Capture.Metadata = metadata

		if exportMetaSource(item.Capture.Source) {
			if len(query.MetaFilters) > 0 {
				continue
			}
			item.Kind = storage.ExportItemMeta
			item.Meta = storage.ExportMeta{
				CreatedAt: item.Capture.CreatedAt,
				Command:   strings.TrimSpace(item.Capture.Prompt),
				Value:     strings.TrimSpace(item.Capture.Response),
			}
			item.Capture = storage.ExportCapture{}
			items = append(items, item)
			continue
		}

		captureCount++
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, captureCount, nil
}

func (p *Provider) listExportCheckpoints(ctx context.Context, project string) ([]storage.ExportItem, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT rowid, hash, summary, created_at
		FROM checkpoints
		WHERE project = ?
		ORDER BY created_at ASC, rowid ASC`, project)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]storage.ExportItem, 0)
	for rows.Next() {
		var item storage.ExportItem
		item.Kind = storage.ExportItemCheckpoint
		if err := rows.Scan(&item.RowID, &item.Checkpoint.Hash, &item.Checkpoint.Summary, &item.Checkpoint.CreatedAt); err != nil {
			return nil, err
		}
		item.Timestamp = item.Checkpoint.CreatedAt
		item.Checkpoint.Project = project
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func mergeExportItems(captures, checkpoints []storage.ExportItem) []storage.ExportItem {
	merged := make([]storage.ExportItem, 0, len(captures)+len(checkpoints))
	i := 0
	j := 0
	for i < len(captures) && j < len(checkpoints) {
		if captures[i].Timestamp < checkpoints[j].Timestamp {
			merged = append(merged, captures[i])
			i++
			continue
		}
		if captures[i].Timestamp > checkpoints[j].Timestamp {
			merged = append(merged, checkpoints[j])
			j++
			continue
		}
		if captures[i].RowID <= checkpoints[j].RowID {
			merged = append(merged, captures[i])
			i++
			continue
		}
		merged = append(merged, checkpoints[j])
		j++
	}
	merged = append(merged, captures[i:]...)
	merged = append(merged, checkpoints[j:]...)
	return merged
}

func mergedStringMetadata(primary, secondary string) (map[string]string, error) {
	metadata := map[string]string{}
	for _, raw := range []string{primary, secondary} {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		decoded, err := decodeStringMetadata(raw)
		if err != nil {
			return nil, err
		}
		for key, value := range decoded {
			metadata[key] = value
		}
	}
	return metadata, nil
}

func decodeStringMetadata(raw string) (map[string]string, error) {
	var metadata map[string]string
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

func metadataDeleted(metadata map[string]string) bool {
	return strings.EqualFold(strings.TrimSpace(metadata["deleted"]), "true")
}

func stringMetadataMatchesAll(metadata, filters map[string]string) bool {
	for key, value := range filters {
		if metadata[key] != value {
			return false
		}
	}
	return true
}

func exportMetaSource(source string) bool {
	value := strings.ToLower(strings.TrimSpace(source))
	return value == "meta-command" || value == "meta_command" || value == "event"
}
