package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
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
	format := fs.String("format", "", "export format (required: markdown|json|html)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("usage: yanzi export --format <markdown|json|html>")
	}
	formatValue := strings.TrimSpace(*format)
	if formatValue != "markdown" && formatValue != "json" && formatValue != "html" {
		return errors.New("usage: yanzi export --format <markdown|json|html>")
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
	if formatValue == "html" {
		path = filepath.Join(".", "YANZI_LOG.html")
		content = []byte(renderHTMLLog(project, cliVersion, now, items))
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write export file: %w", err)
	}
	if err := exportArtifactDirectories(project); err != nil {
		return err
	}

	fmt.Printf("Exported %s\n", path)
	return nil
}

func loadExportItems(ctx context.Context, db *sql.DB, project string) ([]exportItem, int, error) {
	intents := make([]exportItem, 0)
	captureCount := 0

	intentRows, err := db.QueryContext(ctx, `SELECT rowid, id, created_at, author, source_type, prompt, response, hash, meta
		FROM intents
		WHERE source_type <> 'artifact'
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

func exportArtifactDirectories(project string) error {
	intentArtifacts, err := yanzilibrary.ListArtifacts(project, yanzilibrary.ArtifactClassIntent, "")
	if err != nil {
		return err
	}
	if err := writeArtifactDirectory("Intent", intentArtifacts); err != nil {
		return err
	}

	contextArtifacts, err := yanzilibrary.ListArtifacts(project, yanzilibrary.ArtifactClassContext, "")
	if err != nil {
		return err
	}
	return writeArtifactDirectory("Context", contextArtifacts)
}

func writeArtifactDirectory(dir string, artifacts []yanzilibrary.Artifact) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create artifact export directory: %w", err)
	}
	for _, artifact := range artifacts {
		path := filepath.Join(dir, artifactExportFilename(artifact))
		if err := os.WriteFile(path, []byte(renderArtifactMarkdown(artifact)), 0o644); err != nil {
			return fmt.Errorf("write artifact export file: %w", err)
		}
	}
	return nil
}

func artifactExportFilename(artifact yanzilibrary.Artifact) string {
	parsed, err := time.Parse(time.RFC3339Nano, artifact.CreatedAt)
	timestamp := artifact.CreatedAt
	if err == nil {
		timestamp = parsed.UTC().Format("20060102T150405Z")
	}
	return fmt.Sprintf("%s-%s.md", timestamp, slugify(artifact.Title))
}

func renderArtifactMarkdown(artifact yanzilibrary.Artifact) string {
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(artifact.Title)
	b.WriteString("\n\n")
	b.WriteString("Type: ")
	b.WriteString(artifact.Type)
	b.WriteString("\n")
	b.WriteString("Created: ")
	b.WriteString(artifact.CreatedAt)
	b.WriteString("\n\n")
	b.WriteString("## Content\n\n```text\n")
	b.WriteString(artifact.Content)
	if !strings.HasSuffix(artifact.Content, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("```\n")
	if strings.TrimSpace(artifact.Metadata) != "" {
		b.WriteString("\n## Metadata\n\n```text\n")
		b.WriteString(artifact.Metadata)
		if !strings.HasSuffix(artifact.Metadata, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("```\n")
	}
	return b.String()
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	prevDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			prevDash = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "artifact"
	}
	return slug
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
	keys := sortedMetaKeys(meta)
	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("  %s: %s", key, meta[key]))
	}
	return lines
}

func sortedMetaKeys(meta map[string]string) []string {
	keys := make([]string, 0, len(meta))
	for key := range meta {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
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

func renderHTMLLog(project, cliVersion string, now time.Time, items []exportItem) string {
	totalEvents := len(items)
	totalCaptures := 0
	totalCheckpoints := 0
	for _, item := range items {
		switch item.Kind {
		case exportItemCapture:
			totalCaptures++
		case exportItemCheckpoint:
			totalCheckpoints++
		}
	}

	var b strings.Builder
	b.WriteString("<!doctype html>\n")
	b.WriteString("<html lang=\"en\">\n")
	b.WriteString("<head>\n")
	b.WriteString("  <meta charset=\"utf-8\">\n")
	b.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString("  <title>Yanzi Log</title>\n")
	b.WriteString("  <style>\n")
	b.WriteString("    body{font-family:-apple-system,BlinkMacSystemFont,\"Segoe UI\",Roboto,Helvetica,Arial,sans-serif;margin:0;background:#f6f7f8;color:#1f2937;line-height:1.45}\n")
	b.WriteString("    main{max-width:980px;margin:0 auto;padding:24px 16px 40px}\n")
	b.WriteString("    header{background:#fff;border:1px solid #d1d5db;border-radius:8px;padding:16px;margin-bottom:16px}\n")
	b.WriteString("    h1{margin:0 0 8px;font-size:1.4rem}\n")
	b.WriteString("    .meta-line{margin:2px 0;color:#4b5563;font-size:.95rem}\n")
	b.WriteString("    .counts{display:flex;gap:12px;flex-wrap:wrap;margin-top:10px}\n")
	b.WriteString("    .count{background:#f3f4f6;border:1px solid #e5e7eb;border-radius:6px;padding:6px 10px;font-size:.9rem}\n")
	b.WriteString("    .timeline{display:flex;flex-direction:column;gap:12px}\n")
	b.WriteString("    .capture{background:#fff;border:1px solid #d1d5db;border-radius:8px;padding:12px}\n")
	b.WriteString("    .capture h3{margin:0 0 8px;font-size:1rem}\n")
	b.WriteString("    .checkpoint{background:#fff;border-radius:8px;padding:10px 12px}\n")
	b.WriteString("    .checkpoint h2{margin:0 0 6px;font-size:1.05rem}\n")
	b.WriteString("    .meta-event{background:#f9fafb;border:1px dashed #d1d5db;border-radius:8px;padding:10px 12px;color:#374151}\n")
	b.WriteString("    .label{font-weight:600}\n")
	b.WriteString("    table{border-collapse:collapse;margin:8px 0 6px;width:auto}\n")
	b.WriteString("    th,td{border:1px solid #e5e7eb;padding:4px 8px;font-size:.9rem;text-align:left}\n")
	b.WriteString("    pre{background:#111827;color:#e5e7eb;border-radius:6px;padding:10px;overflow:auto;white-space:pre-wrap}\n")
	b.WriteString("    hr{border:none;border-top:1px solid #d1d5db;margin:4px 0}\n")
	b.WriteString("  </style>\n")
	b.WriteString("</head>\n")
	b.WriteString("<body>\n")
	b.WriteString("<main>\n")
	b.WriteString("  <header>\n")
	b.WriteString("    <h1>Yanzi Agent Log</h1>\n")
	b.WriteString(fmt.Sprintf("    <div class=\"meta-line\"><span class=\"label\">Project:</span> %s</div>\n", html.EscapeString(project)))
	b.WriteString(fmt.Sprintf("    <div class=\"meta-line\"><span class=\"label\">Exported:</span> %s</div>\n", html.EscapeString(now.Format(time.RFC3339))))
	b.WriteString(fmt.Sprintf("    <div class=\"meta-line\"><span class=\"label\">Version:</span> %s</div>\n", html.EscapeString(cliVersion)))
	b.WriteString("    <div class=\"counts\">\n")
	b.WriteString(fmt.Sprintf("      <div class=\"count\">Total events: %d</div>\n", totalEvents))
	b.WriteString(fmt.Sprintf("      <div class=\"count\">Total captures: %d</div>\n", totalCaptures))
	b.WriteString(fmt.Sprintf("      <div class=\"count\">Total checkpoints: %d</div>\n", totalCheckpoints))
	b.WriteString("    </div>\n")
	b.WriteString("  </header>\n")
	b.WriteString("  <section class=\"timeline\">\n")

	for _, item := range items {
		switch item.Kind {
		case exportItemCheckpoint:
			b.WriteString("    <hr>\n")
			b.WriteString("    <section class=\"checkpoint\">\n")
			b.WriteString(fmt.Sprintf("      <h2>Checkpoint: %s</h2>\n", html.EscapeString(item.CheckpointID)))
			b.WriteString(fmt.Sprintf("      <div><span class=\"label\">Summary:</span> %s</div>\n", html.EscapeString(item.Summary)))
			b.WriteString(fmt.Sprintf("      <div><span class=\"label\">Timestamp:</span> %s</div>\n", html.EscapeString(item.Timestamp)))
			b.WriteString("    </section>\n")
		case exportItemMeta:
			b.WriteString("    <section class=\"meta-event\">\n")
			b.WriteString(fmt.Sprintf("      <div><span class=\"label\">Event:</span> %s</div>\n", html.EscapeString(item.Command)))
			if strings.TrimSpace(item.Value) != "" {
				b.WriteString(fmt.Sprintf("      <div><span class=\"label\">Value:</span> %s</div>\n", html.EscapeString(item.Value)))
			}
			b.WriteString(fmt.Sprintf("      <div><span class=\"label\">Timestamp:</span> %s</div>\n", html.EscapeString(item.Timestamp)))
			b.WriteString("    </section>\n")
		default:
			b.WriteString("    <section class=\"capture\">\n")
			b.WriteString(fmt.Sprintf("      <h3>Capture: %s</h3>\n", html.EscapeString(item.CaptureID)))
			b.WriteString(fmt.Sprintf("      <div><span class=\"label\">Role:</span> %s</div>\n", html.EscapeString(item.Role)))
			b.WriteString(fmt.Sprintf("      <div><span class=\"label\">Timestamp:</span> %s</div>\n", html.EscapeString(item.Timestamp)))
			b.WriteString(fmt.Sprintf("      <div><span class=\"label\">Hash:</span> %s</div>\n", html.EscapeString(item.Hash)))
			if len(item.Metadata) > 0 {
				keys := sortedMetaKeys(item.Metadata)
				b.WriteString("      <table>\n")
				b.WriteString("        <thead><tr><th>Metadata Key</th><th>Value</th></tr></thead>\n")
				b.WriteString("        <tbody>\n")
				for _, key := range keys {
					b.WriteString(fmt.Sprintf("          <tr><td>%s</td><td>%s</td></tr>\n", html.EscapeString(key), html.EscapeString(item.Metadata[key])))
				}
				b.WriteString("        </tbody>\n")
				b.WriteString("      </table>\n")
			}
			b.WriteString("      <div><span class=\"label\">Prompt:</span></div>\n")
			b.WriteString(fmt.Sprintf("      <pre>%s</pre>\n", html.EscapeString(item.Prompt)))
			b.WriteString("      <div><span class=\"label\">Response:</span></div>\n")
			b.WriteString(fmt.Sprintf("      <pre>%s</pre>\n", html.EscapeString(item.Response)))
			b.WriteString("    </section>\n")
		}
	}

	b.WriteString("  </section>\n")
	b.WriteString("</main>\n")
	b.WriteString("</body>\n")
	b.WriteString("</html>\n")
	return b.String()
}
