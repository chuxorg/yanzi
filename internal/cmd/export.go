package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chuxorg/chux-yanzi-cli/internal/config"
)

type exportItemType string

const (
	exportItemCheckpoint exportItemType = "checkpoint"
	exportItemCapture    exportItemType = "capture"
	exportItemMeta       exportItemType = "meta"
)

type exportItem struct {
	Kind      exportItemType
	Timestamp string
	RowID     int64

	CheckpointID string
	Summary      string

	CaptureID string
	Role      string
	Hash      string
	Prompt    string
	Response  string
	Metadata  map[string]string

	Command string
	Value   string
}

// RunExport writes deterministic project history logs.
func RunExport(args []string, cliVersion string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	format := fs.String("format", "", "export format (required: markdown|json)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("usage: yanzi export --format <markdown|json>")
	}
	formatValue := strings.TrimSpace(*format)
	if formatValue != "markdown" && formatValue != "json" {
		return errors.New("usage: yanzi export --format <markdown|json>")
	}

	project, err := loadActiveProject()
	if err != nil {
		return err
	}
	if project == "" {
		return errors.New("no active project set")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.Mode != config.ModeLocal {
		return errors.New("export is only available in local mode")
	}

	db, err := openLocalDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	items, captureCount, err := loadExportItems(ctx, db, project)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	path := filepath.Join(".", "YANZI_LOG.md")
	content := []byte(renderMarkdownLog(project, cliVersion, now, items, captureCount))
	if formatValue == "json" {
		path = filepath.Join(".", "YANZI_LOG.json")
		content, err = renderJSONLog(project, cliVersion, now, items)
		if err != nil {
			return err
		}
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write export file: %w", err)
	}

	fmt.Printf("Exported %s\n", path)
	return nil
}

func loadExportItems(ctx context.Context, db *sql.DB, project string) ([]exportItem, int, error) {
	intents := make([]exportItem, 0)
	captureCount := 0

	intentRows, err := db.QueryContext(ctx, `SELECT rowid, id, created_at, author, source_type, prompt, response, hash, meta
		FROM intents
		ORDER BY created_at ASC, rowid ASC`)
	if err != nil {
		return nil, 0, err
	}
	defer intentRows.Close()

	for intentRows.Next() {
		var (
			rowID                                                          int64
			id, createdAt, author, sourceType, prompt, response, hashValue string
			metaText                                                       sql.NullString
		)
		if err := intentRows.Scan(&rowID, &id, &createdAt, &author, &sourceType, &prompt, &response, &hashValue, &metaText); err != nil {
			return nil, 0, err
		}
		meta, err := decodeStringMeta(metaText.String)
		if err != nil {
			continue
		}
		if strings.TrimSpace(meta["project"]) != project {
			continue
		}

		if isMetaCommandSource(sourceType) {
			intents = append(intents, exportItem{
				Kind:      exportItemMeta,
				Timestamp: createdAt,
				Command:   strings.TrimSpace(prompt),
				Value:     strings.TrimSpace(response),
				RowID:     rowID,
			})
			continue
		}

		captureCount++
		intents = append(intents, exportItem{
			Kind:      exportItemCapture,
			Timestamp: createdAt,
			CaptureID: id,
			Role:      author,
			Hash:      hashValue,
			Prompt:    prompt,
			Response:  response,
			Metadata:  exportableMetadata(meta),
			RowID:     rowID,
		})
	}
	if err := intentRows.Err(); err != nil {
		return nil, 0, err
	}

	checkpoints := make([]exportItem, 0)
	checkpointRows, err := db.QueryContext(ctx, `SELECT rowid, hash, summary, created_at
		FROM checkpoints
		WHERE project = ?
		ORDER BY created_at ASC, rowid ASC`, project)
	if err != nil {
		return nil, 0, err
	}
	defer checkpointRows.Close()

	for checkpointRows.Next() {
		var rowID int64
		var id, summary, createdAt string
		if err := checkpointRows.Scan(&rowID, &id, &summary, &createdAt); err != nil {
			return nil, 0, err
		}
		checkpoints = append(checkpoints, exportItem{
			Kind:         exportItemCheckpoint,
			Timestamp:    createdAt,
			CheckpointID: id,
			Summary:      summary,
			RowID:        rowID,
		})
	}
	if err := checkpointRows.Err(); err != nil {
		return nil, 0, err
	}

	return mergeChronological(intents, checkpoints), captureCount, nil
}

func decodeStringMeta(metaText string) (map[string]string, error) {
	if strings.TrimSpace(metaText) == "" {
		return nil, nil
	}

	var meta map[string]string
	if err := json.Unmarshal([]byte(metaText), &meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func sortedMetaPairs(meta map[string]string) []string {
	if len(meta) == 0 {
		return nil
	}
	keys := make([]string, 0, len(meta))
	for key := range meta {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("  %s: %s", key, meta[key]))
	}
	return lines
}

func exportableMetadata(meta map[string]string) map[string]string {
	if len(meta) == 0 {
		return nil
	}
	filtered := make(map[string]string, len(meta))
	for key, value := range meta {
		if strings.TrimSpace(key) == "project" {
			continue
		}
		filtered[key] = value
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func isMetaCommandSource(sourceType string) bool {
	value := strings.ToLower(strings.TrimSpace(sourceType))
	return value == "meta-command" || value == "meta_command" || value == "event"
}

func mergeChronological(intents, checkpoints []exportItem) []exportItem {
	merged := make([]exportItem, 0, len(intents)+len(checkpoints))
	i := 0
	j := 0
	for i < len(intents) && j < len(checkpoints) {
		if intents[i].Timestamp < checkpoints[j].Timestamp {
			merged = append(merged, intents[i])
			i++
			continue
		}
		if intents[i].Timestamp > checkpoints[j].Timestamp {
			merged = append(merged, checkpoints[j])
			j++
			continue
		}

		if intents[i].RowID <= checkpoints[j].RowID {
			merged = append(merged, intents[i])
			i++
		} else {
			merged = append(merged, checkpoints[j])
			j++
		}
	}
	for i < len(intents) {
		merged = append(merged, intents[i])
		i++
	}
	for j < len(checkpoints) {
		merged = append(merged, checkpoints[j])
		j++
	}
	return merged
}

func renderMarkdownLog(project, cliVersion string, now time.Time, items []exportItem, captureCount int) string {
	var b strings.Builder

	b.WriteString("# Yanzi Agent Log\n\n")
	b.WriteString(fmt.Sprintf("Project: %s\n", project))
	b.WriteString(fmt.Sprintf("Exported: %s\n", now.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("Version: %s\n\n", cliVersion))
	b.WriteString("---\n\n")

	if len(items) == 0 && captureCount == 0 {
		b.WriteString("No captures recorded.\n")
		return b.String()
	}

	for _, item := range items {
		switch item.Kind {
		case exportItemCheckpoint:
			b.WriteString(fmt.Sprintf("## Checkpoint: %s\n\n", item.CheckpointID))
			b.WriteString(fmt.Sprintf("Summary: %s\n", item.Summary))
			b.WriteString(fmt.Sprintf("Timestamp: %s\n", item.Timestamp))
			b.WriteString("----------------------\n\n")
		case exportItemMeta:
			b.WriteString(fmt.Sprintf("### Event: %s\n\n", item.Command))
			if strings.TrimSpace(item.Value) != "" {
				b.WriteString(fmt.Sprintf("Value: %s\n", item.Value))
			}
			b.WriteString(fmt.Sprintf("Timestamp: %s\n", item.Timestamp))
			b.WriteString("----------------------\n\n")
		default:
			b.WriteString(fmt.Sprintf("### Capture: %s\n\n", item.CaptureID))
			b.WriteString(fmt.Sprintf("Role: %s\n", item.Role))
			b.WriteString(fmt.Sprintf("Timestamp: %s\n", item.Timestamp))
			b.WriteString(fmt.Sprintf("Hash: %s\n\n", item.Hash))
			metaLines := sortedMetaPairs(item.Metadata)
			if len(metaLines) > 0 {
				b.WriteString("Metadata:\n")
				b.WriteString(strings.Join(metaLines, "\n"))
				b.WriteString("\n\n")
			}
			b.WriteString("**Prompt**\n")
			b.WriteString("```text\n")
			b.WriteString(item.Prompt)
			b.WriteString("\n```\n\n")
			b.WriteString("**Response**\n")
			b.WriteString("```text\n")
			b.WriteString(item.Response)
			b.WriteString("\n```\n\n")
			b.WriteString("---\n\n")
		}
	}

	return b.String()
}

type jsonExport struct {
	SchemaVersion int    `json:"schema_version"`
	Project       string `json:"project"`
	ExportedAt    string `json:"exported_at"`
	Version       string `json:"version"`
	Events        []any  `json:"events"`
}

type jsonCheckpointEvent struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Summary   string `json:"summary"`
	Timestamp string `json:"timestamp"`
}

type jsonCaptureEvent struct {
	Type      string            `json:"type"`
	ID        string            `json:"id"`
	Role      string            `json:"role"`
	Timestamp string            `json:"timestamp"`
	Hash      string            `json:"hash"`
	Prompt    string            `json:"prompt"`
	Response  string            `json:"response"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type jsonMetaEvent struct {
	Type      string `json:"type"`
	Command   string `json:"command"`
	Value     any    `json:"value"`
	Timestamp string `json:"timestamp"`
}

func renderJSONLog(project, cliVersion string, now time.Time, items []exportItem) ([]byte, error) {
	events := make([]any, 0, len(items))
	for _, item := range items {
		switch item.Kind {
		case exportItemCheckpoint:
			events = append(events, jsonCheckpointEvent{
				Type:      string(exportItemCheckpoint),
				ID:        item.CheckpointID,
				Summary:   item.Summary,
				Timestamp: item.Timestamp,
			})
		case exportItemMeta:
			var value any
			if strings.TrimSpace(item.Value) != "" {
				value = item.Value
			}
			events = append(events, jsonMetaEvent{
				Type:      string(exportItemMeta),
				Command:   item.Command,
				Value:     value,
				Timestamp: item.Timestamp,
			})
		default:
			events = append(events, jsonCaptureEvent{
				Type:      string(exportItemCapture),
				ID:        item.CaptureID,
				Role:      item.Role,
				Timestamp: item.Timestamp,
				Hash:      item.Hash,
				Prompt:    item.Prompt,
				Response:  item.Response,
				Metadata:  item.Metadata,
			})
		}
	}

	payload := jsonExport{
		SchemaVersion: 1,
		Project:       project,
		ExportedAt:    now.Format(time.RFC3339),
		Version:       cliVersion,
		Events:        events,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode json export: %w", err)
	}
	b = append(b, '\n')
	return b, nil
}
