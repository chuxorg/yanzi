package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

var openExportInBrowser = openBrowser

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
	Source    string
	Hash      string
	Prompt    string
	Response  string
	Metadata  map[string]string

	Command string
	Value   string
}

type contextExportQuery struct {
	TypeFilters map[string]struct{}
	MetaFilters map[string]string
	Fields      []string
	FieldSet    map[string]struct{}
	Order       string
	Limit       int
}

// RunExport writes deterministic project exports for history or filtered context.
//
// Problem:
// Callers sometimes need a full timeline export and sometimes need a filtered
// retrieval of stored context only.
//
// Solution:
// RunExport supports log formats and context retrieval formats, with explicit
// filters for type, metadata, field selection, ordering, and limits.
//
// Arguments:
//
//	args contains export flags such as format, filters, and output behavior;
//	cliVersion is written into the rendered export headers.
//
// Example:
//
//	yanzi export --type process_rule --meta role=engineer --fields title,content
func RunExport(args []string, cliVersion string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	format := fs.String("format", "", "export format (markdown|json|html|claude-context)")
	open := fs.Bool("open", false, "open generated html export in the default browser")
	profile := fs.String("profile", "", "profile filter")
	typeFlag := fs.String("type", "", "context type filter (comma-separated)")
	fieldsFlag := fs.String("fields", "", "fields to include (comma-separated)")
	orderFlag := fs.String("order", "", "order field (created_at|updated_at)")
	limit := fs.Int("limit", 0, "max records to return after filtering")
	includeDeleted := fs.Bool("include-deleted", false, "include tombstoned records")
	metaFilters := metaPairs{}
	fs.Var(&metaFilters, "meta", "meta filter key=value (repeatable; exact match; AND)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return errors.New("usage: yanzi export [--format <markdown|json|html|claude-context>] [--type <type[,type...]>] [--meta key=value ...] [--fields <field[,field...]>] [--order <created_at|updated_at>] [--limit <n>] [--open]")
	}
	if strings.TrimSpace(*profile) != "" {
		metaFilters["profile"] = strings.TrimSpace(*profile)
	}
	formatValue := strings.TrimSpace(*format)
	if formatValue == "" {
		formatValue = "claude-context"
	}
	if formatValue != "markdown" && formatValue != "json" && formatValue != "html" && formatValue != "claude-context" {
		return errors.New("usage: yanzi export [--format <markdown|json|html|claude-context>] [--type <type[,type...]>] [--meta key=value ...] [--fields <field[,field...]>] [--order <created_at|updated_at>] [--limit <n>] [--open]")
	}
	if *open && formatValue != "html" {
		return errors.New("--open is only supported with --format html")
	}
	query, err := buildContextExportQuery(*typeFlag, map[string]string(metaFilters), *fieldsFlag, *orderFlag, *limit)
	if err != nil {
		return err
	}
	retrievalMode := formatValue == "claude-context" || contextExportQueryEnabled(query)

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

	now := time.Now().UTC()
	if retrievalMode {
		artifacts, err := loadFilteredContextArtifacts(project, query, *includeDeleted)
		if err != nil {
			return err
		}
		path := filepath.Join(".", "CLAUDE_CONTEXT.md")
		var content []byte
		switch formatValue {
		case "markdown":
			path = filepath.Join(".", "YANZI_LOG.md")
			content = []byte(renderContextMarkdownExport("Yanzi Context Export", project, cliVersion, now, artifacts, query))
		case "json":
			path = filepath.Join(".", "YANZI_LOG.json")
			content, err = renderContextJSONExport(project, cliVersion, now, artifacts, query)
			if err != nil {
				return err
			}
		case "html":
			path = filepath.Join(".", "YANZI_LOG.html")
			content = []byte(renderContextHTMLExport(project, cliVersion, now, artifacts, query))
		default:
			content = []byte(renderContextMarkdownExport("Claude Context", project, cliVersion, now, artifacts, query))
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return fmt.Errorf("write export file: %w", err)
		}
		if err := exportFilteredContextArtifacts(artifacts); err != nil {
			return err
		}
		fmt.Printf("Exported %s\n", path)
		if *open {
			if err := openExportInBrowser(path); err != nil {
				return fmt.Errorf("open export in browser: %w", err)
			}
		}
		return nil
	}

	ctx := context.Background()
	items, captureCount, err := loadExportItems(ctx, db, project, map[string]string(metaFilters), *includeDeleted)
	if err != nil {
		return err
	}

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
	if err := exportArtifactDirectories(project, map[string]string(metaFilters), *includeDeleted); err != nil {
		return err
	}

	fmt.Printf("Exported %s\n", path)
	if *open {
		if err := openExportInBrowser(path); err != nil {
			return fmt.Errorf("open export in browser: %w", err)
		}
	}
	return nil
}

func buildContextExportQuery(typeValue string, metaFilters map[string]string, fieldsValue, orderValue string, limit int) (contextExportQuery, error) {
	typeFilters, err := parseContextTypeFilters(typeValue)
	if err != nil {
		return contextExportQuery{}, err
	}
	fields, fieldSet, err := parseContextExportFields(fieldsValue)
	if err != nil {
		return contextExportQuery{}, err
	}
	order := strings.TrimSpace(orderValue)
	switch order {
	case "", "created_at", "updated_at":
	default:
		return contextExportQuery{}, errors.New("invalid --order value (expected created_at or updated_at)")
	}
	if limit < 0 {
		return contextExportQuery{}, errors.New("--limit must be >= 0")
	}
	return contextExportQuery{
		TypeFilters: typeFilters,
		MetaFilters: metaFilters,
		Fields:      fields,
		FieldSet:    fieldSet,
		Order:       order,
		Limit:       limit,
	}, nil
}

func contextExportQueryEnabled(query contextExportQuery) bool {
	return len(query.TypeFilters) > 0 || len(query.Fields) > 0 || strings.TrimSpace(query.Order) != "" || query.Limit > 0
}

func parseContextTypeFilters(value string) (map[string]struct{}, error) {
	result := map[string]struct{}{}
	for _, raw := range strings.Split(strings.TrimSpace(value), ",") {
		trimmed := normalizeContextType(strings.TrimSpace(raw))
		if trimmed == "" {
			continue
		}
		if !isSupportedContextType(trimmed) {
			return nil, fmt.Errorf("invalid context type %q: allowed values are requirement, process_rule, coding_standard, reference, note", trimmed)
		}
		result[trimmed] = struct{}{}
	}
	return result, nil
}

func parseContextExportFields(value string) ([]string, map[string]struct{}, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil, nil
	}
	allowed := map[string]struct{}{
		"id":         {},
		"title":      {},
		"type":       {},
		"scope":      {},
		"project":    {},
		"content":    {},
		"metadata":   {},
		"created_at": {},
		"updated_at": {},
	}
	fields := make([]string, 0)
	fieldSet := make(map[string]struct{})
	for _, raw := range strings.Split(value, ",") {
		field := strings.TrimSpace(raw)
		if field == "" {
			continue
		}
		if _, ok := allowed[field]; !ok {
			return nil, nil, fmt.Errorf("invalid --fields value %q", field)
		}
		if _, ok := fieldSet[field]; ok {
			continue
		}
		fieldSet[field] = struct{}{}
		fields = append(fields, field)
	}
	if len(fields) == 0 {
		return nil, nil, errors.New("--fields requires at least one field")
	}
	return fields, fieldSet, nil
}

func isSupportedContextType(value string) bool {
	for _, candidate := range contextTypeCatalog {
		if candidate == value {
			return true
		}
	}
	return false
}

func loadFilteredContextArtifacts(project string, query contextExportQuery, includeDeleted bool) ([]yanzilibrary.Artifact, error) {
	artifacts, err := yanzilibrary.ListVisibleContextArtifacts(project, "", "", "", includeDeleted)
	if err != nil {
		return nil, err
	}

	filtered := make([]yanzilibrary.Artifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		if len(query.TypeFilters) > 0 {
			if _, ok := query.TypeFilters[strings.TrimSpace(artifact.Type)]; !ok {
				continue
			}
		}
		metadata, err := decodeStringMeta(artifact.Metadata)
		if err != nil {
			continue
		}
		if len(query.MetaFilters) > 0 && !metadataMatchesAll(metadata, query.MetaFilters) {
			continue
		}
		filtered = append(filtered, artifact)
	}

	sortContextArtifacts(filtered, query.Order)
	if query.Limit > 0 && len(filtered) > query.Limit {
		filtered = filtered[:query.Limit]
	}
	return filtered, nil
}

func sortContextArtifacts(artifacts []yanzilibrary.Artifact, order string) {
	if len(artifacts) < 2 {
		return
	}
	sort.SliceStable(artifacts, func(i, j int) bool {
		left := contextArtifactOrderValue(artifacts[i], order)
		right := contextArtifactOrderValue(artifacts[j], order)
		if left != right {
			return left < right
		}
		if artifacts[i].Title != artifacts[j].Title {
			return artifacts[i].Title < artifacts[j].Title
		}
		return artifacts[i].ID < artifacts[j].ID
	})
}

func contextArtifactOrderValue(artifact yanzilibrary.Artifact, order string) string {
	switch strings.TrimSpace(order) {
	case "", "created_at", "updated_at":
		return artifact.CreatedAt
	default:
		return artifact.CreatedAt
	}
}

func exportFilteredContextArtifacts(artifacts []yanzilibrary.Artifact) error {
	if err := writeArtifactDirectory("Intent", nil); err != nil {
		return err
	}
	return writeArtifactDirectory("Context", artifacts)
}

func openBrowser(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve export path: %w", err)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", absPath)
	case "linux":
		cmd = exec.Command("xdg-open", absPath)
	case "windows":
		fileURL := (&url.URL{Scheme: "file", Path: filepath.ToSlash(absPath)}).String()
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", fileURL)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

func loadExportItems(ctx context.Context, db *sql.DB, project string, metaFilters map[string]string, includeDeleted bool) ([]exportItem, int, error) {
	intents := make([]exportItem, 0)
	captureCount := 0

	intentRows, err := db.QueryContext(ctx, `SELECT rowid, id, created_at, author, source_type, prompt, response, hash, meta, metadata
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
			metadataText                                                   sql.NullString
		)
		if err := intentRows.Scan(&rowID, &id, &createdAt, &author, &sourceType, &prompt, &response, &hashValue, &metaText, &metadataText); err != nil {
			return nil, 0, err
		}
		meta, err := mergedIntentMetadata(metaText.String, metadataText.String)
		if err != nil {
			continue
		}
		if strings.TrimSpace(meta["project"]) != project {
			continue
		}
		if !includeDeleted && isDeletedMetadata(meta) {
			continue
		}
		if len(metaFilters) > 0 {
			if !metadataMatchesAll(meta, metaFilters) {
				continue
			}
		}

		if isMetaCommandSource(sourceType) {
			if len(metaFilters) > 0 {
				continue
			}
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
			Source:    sourceType,
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
	if len(metaFilters) > 0 {
		return mergeChronological(intents, checkpoints), captureCount, nil
	}
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

func exportArtifactDirectories(project string, metaFilters map[string]string, includeDeleted bool) error {
	if len(metaFilters) > 0 {
		if err := writeArtifactDirectory("Intent", nil); err != nil {
			return err
		}
		return writeArtifactDirectory("Context", nil)
	}

	intentArtifacts, err := yanzilibrary.ListArtifacts(project, yanzilibrary.ArtifactClassIntent, "", includeDeleted)
	if err != nil {
		return err
	}
	if err := writeArtifactDirectory("Intent", intentArtifacts); err != nil {
		return err
	}

	contextArtifacts, err := yanzilibrary.ListArtifacts(project, yanzilibrary.ArtifactClassContext, "", includeDeleted)
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

func metadataMatchesAll(meta, filters map[string]string) bool {
	if len(filters) == 0 {
		return true
	}
	for key, value := range filters {
		if meta[key] != value {
			return false
		}
	}
	return true
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

func renderContextMarkdownExport(title, project, cliVersion string, now time.Time, artifacts []yanzilibrary.Artifact, query contextExportQuery) string {
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Project: %s\n", project))
	b.WriteString(fmt.Sprintf("Exported: %s\n", now.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("Version: %s\n", cliVersion))
	if len(artifacts) == 0 {
		b.WriteString("\n\n_No context artifacts available._\n")
		return b.String()
	}
	if title == "Claude Context" && len(query.Fields) == 0 {
		return renderDefaultClaudeContextMarkdown(b.String(), artifacts)
	}
	for i, artifact := range artifacts {
		b.WriteString("\n\n")
		b.WriteString(contextArtifactHeading(i, artifact, query.FieldSet))
		b.WriteString("\n")
		for _, field := range contextFieldsForRender(query) {
			block := renderContextFieldMarkdown(field, artifact)
			if block == "" {
				continue
			}
			b.WriteString("\n")
			b.WriteString(block)
			b.WriteString("\n")
		}
	}
	return b.String()
}

func renderDefaultClaudeContextMarkdown(header string, artifacts []yanzilibrary.Artifact) string {
	grouped := make(map[string][]yanzilibrary.Artifact)
	orderedTypes := make([]string, 0)
	for _, artifact := range artifacts {
		artifactType := strings.TrimSpace(artifact.Type)
		if _, ok := grouped[artifactType]; !ok {
			orderedTypes = append(orderedTypes, artifactType)
		}
		grouped[artifactType] = append(grouped[artifactType], artifact)
	}
	sort.Strings(orderedTypes)

	var b strings.Builder
	b.WriteString(header)
	for _, artifactType := range orderedTypes {
		b.WriteString(fmt.Sprintf("\n\n## %s\n", artifactType))
		for _, artifact := range grouped[artifactType] {
			b.WriteString(fmt.Sprintf("\n### %s\n", artifact.Title))
			b.WriteString(fmt.Sprintf("- Type: %s\n", artifact.Type))
			b.WriteString(fmt.Sprintf("- Scope: %s\n", artifact.Scope))
			if metadata := strings.TrimSpace(artifact.Metadata); metadata != "" {
				b.WriteString(fmt.Sprintf("- Metadata: %s\n", metadata))
			}
			b.WriteString("\n")
			b.WriteString(strings.TrimRight(artifact.Content, "\n"))
			b.WriteString("\n")
		}
	}
	return b.String()
}

func contextArtifactHeading(index int, artifact yanzilibrary.Artifact, fieldSet map[string]struct{}) string {
	if len(fieldSet) == 0 {
		return fmt.Sprintf("## %s", artifact.Title)
	}
	if _, ok := fieldSet["title"]; ok && strings.TrimSpace(artifact.Title) != "" {
		return fmt.Sprintf("## %s", artifact.Title)
	}
	return fmt.Sprintf("## Artifact %d", index+1)
}

func contextFieldsForRender(query contextExportQuery) []string {
	if len(query.Fields) > 0 {
		return query.Fields
	}
	return []string{"title", "type", "scope", "project", "metadata", "content", "created_at"}
}

func renderContextFieldMarkdown(field string, artifact yanzilibrary.Artifact) string {
	switch field {
	case "id":
		return fmt.Sprintf("- ID: %s", artifact.ID)
	case "title":
		return fmt.Sprintf("- Title: %s", artifact.Title)
	case "type":
		return fmt.Sprintf("- Type: %s", artifact.Type)
	case "scope":
		return fmt.Sprintf("- Scope: %s", artifact.Scope)
	case "project":
		return fmt.Sprintf("- Project: %s", displayProject(artifact.Project))
	case "created_at":
		return fmt.Sprintf("- Created At: %s", artifact.CreatedAt)
	case "updated_at":
		return fmt.Sprintf("- Updated At: %s", artifact.CreatedAt)
	case "metadata":
		metadata := renderArtifactMetadataMarkdown(artifact.Metadata)
		if metadata == "" {
			return ""
		}
		return metadata
	case "content":
		return renderArtifactContentMarkdown(artifact.Content)
	default:
		return ""
	}
}

func renderArtifactMetadataMarkdown(raw string) string {
	decoded, err := decodeStringMeta(raw)
	if err != nil || len(decoded) == 0 {
		return ""
	}
	lines := []string{"### Metadata"}
	for _, key := range sortedMetaKeys(decoded) {
		lines = append(lines, fmt.Sprintf("- %s: %s", key, decoded[key]))
	}
	return strings.Join(lines, "\n")
}

func renderArtifactContentMarkdown(content string) string {
	var b strings.Builder
	b.WriteString("### Content\n")
	b.WriteString("```text\n")
	b.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("```")
	return b.String()
}

type jsonContextExport struct {
	SchemaVersion int              `json:"schema_version"`
	Project       string           `json:"project"`
	ExportedAt    string           `json:"exported_at"`
	Version       string           `json:"version"`
	Artifacts     []map[string]any `json:"artifacts"`
}

func renderContextJSONExport(project, cliVersion string, now time.Time, artifacts []yanzilibrary.Artifact, query contextExportQuery) ([]byte, error) {
	payload := jsonContextExport{
		SchemaVersion: 1,
		Project:       project,
		ExportedAt:    now.Format(time.RFC3339),
		Version:       cliVersion,
		Artifacts:     make([]map[string]any, 0, len(artifacts)),
	}
	fields := contextFieldsForRender(query)
	for _, artifact := range artifacts {
		record := make(map[string]any, len(fields))
		for _, field := range fields {
			record[field] = contextArtifactFieldValue(field, artifact)
		}
		payload.Artifacts = append(payload.Artifacts, record)
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode context json export: %w", err)
	}
	return append(data, '\n'), nil
}

func contextArtifactFieldValue(field string, artifact yanzilibrary.Artifact) any {
	switch field {
	case "id":
		return artifact.ID
	case "title":
		return artifact.Title
	case "type":
		return artifact.Type
	case "scope":
		return artifact.Scope
	case "project":
		return artifact.Project
	case "content":
		return artifact.Content
	case "metadata":
		decoded, err := decodeStringMeta(artifact.Metadata)
		if err != nil || len(decoded) == 0 {
			return nil
		}
		return decoded
	case "created_at":
		return artifact.CreatedAt
	case "updated_at":
		return artifact.CreatedAt
	default:
		return nil
	}
}

func renderContextHTMLExport(project, cliVersion string, now time.Time, artifacts []yanzilibrary.Artifact, query contextExportQuery) string {
	var b strings.Builder
	b.WriteString("<!doctype html>\n<html lang=\"en\">\n<head>\n")
	b.WriteString("  <meta charset=\"utf-8\">\n")
	b.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString("  <title>Yanzi Context Export</title>\n")
	b.WriteString("  <style>body{font-family:-apple-system,BlinkMacSystemFont,\"Segoe UI\",sans-serif;margin:0;background:#f6f8fb;color:#172033}main{max-width:960px;margin:0 auto;padding:24px 16px 40px}header,article{background:#fff;border:1px solid #d7dee8;border-radius:14px;box-shadow:0 8px 24px rgba(15,23,42,.06)}header{padding:18px;margin-bottom:16px}article{padding:16px;margin-bottom:14px}h1,h2,h3{margin:0 0 10px}pre{background:#111827;color:#e5e7eb;border-radius:10px;padding:12px;overflow:auto;white-space:pre-wrap}ul{margin:8px 0;padding-left:20px}.muted{color:#526075}</style>\n")
	b.WriteString("</head>\n<body>\n<main>\n")
	b.WriteString("  <header>\n")
	b.WriteString("    <h1>Yanzi Context Export</h1>\n")
	b.WriteString(fmt.Sprintf("    <div class=\"muted\">Project: %s</div>\n", html.EscapeString(project)))
	b.WriteString(fmt.Sprintf("    <div class=\"muted\">Exported: %s</div>\n", html.EscapeString(now.Format(time.RFC3339))))
	b.WriteString(fmt.Sprintf("    <div class=\"muted\">Version: %s</div>\n", html.EscapeString(cliVersion)))
	b.WriteString("  </header>\n")
	if len(artifacts) == 0 {
		b.WriteString("  <article><p class=\"muted\">No context artifacts available.</p></article>\n")
	} else {
		for i, artifact := range artifacts {
			b.WriteString("  <article>\n")
			b.WriteString(fmt.Sprintf("    <h2>%s</h2>\n", html.EscapeString(strings.TrimPrefix(contextArtifactHeading(i, artifact, query.FieldSet), "## "))))
			for _, field := range contextFieldsForRender(query) {
				b.WriteString(renderContextFieldHTML(field, artifact))
			}
			b.WriteString("  </article>\n")
		}
	}
	b.WriteString("</main>\n</body>\n</html>\n")
	return b.String()
}

func renderContextFieldHTML(field string, artifact yanzilibrary.Artifact) string {
	switch field {
	case "content":
		return fmt.Sprintf("    <h3>Content</h3>\n    <pre>%s</pre>\n", html.EscapeString(artifact.Content))
	case "metadata":
		decoded, err := decodeStringMeta(artifact.Metadata)
		if err != nil || len(decoded) == 0 {
			return ""
		}
		var b strings.Builder
		b.WriteString("    <h3>Metadata</h3>\n    <ul>\n")
		for _, key := range sortedMetaKeys(decoded) {
			b.WriteString(fmt.Sprintf("      <li><strong>%s:</strong> %s</li>\n", html.EscapeString(key), html.EscapeString(decoded[key])))
		}
		b.WriteString("    </ul>\n")
		return b.String()
	case "title":
		return fmt.Sprintf("    <p><strong>Title:</strong> %s</p>\n", html.EscapeString(artifact.Title))
	case "id":
		return fmt.Sprintf("    <p><strong>ID:</strong> %s</p>\n", html.EscapeString(artifact.ID))
	case "type":
		return fmt.Sprintf("    <p><strong>Type:</strong> %s</p>\n", html.EscapeString(artifact.Type))
	case "scope":
		return fmt.Sprintf("    <p><strong>Scope:</strong> %s</p>\n", html.EscapeString(artifact.Scope))
	case "project":
		return fmt.Sprintf("    <p><strong>Project:</strong> %s</p>\n", html.EscapeString(displayProject(artifact.Project)))
	case "created_at":
		return fmt.Sprintf("    <p><strong>Created At:</strong> %s</p>\n", html.EscapeString(artifact.CreatedAt))
	case "updated_at":
		return fmt.Sprintf("    <p><strong>Updated At:</strong> %s</p>\n", html.EscapeString(artifact.CreatedAt))
	default:
		return ""
	}
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

func renderHTMLLog(project, cliVersion string, now time.Time, items []exportItem, sections ...htmlRuleSection) string {
	totalEvents := len(items)
	totalCaptures := 0
	totalCheckpoints := 0
	latestCheckpointID := ""
	latestCheckpointSummary := ""
	for _, item := range items {
		switch item.Kind {
		case exportItemCapture:
			totalCaptures++
		case exportItemCheckpoint:
			totalCheckpoints++
			latestCheckpointID = item.CheckpointID
			latestCheckpointSummary = item.Summary
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
	b.WriteString("    :root{color-scheme:light;--bg:#eef2f7;--surface:#ffffff;--surface-muted:#f8fafc;--text:#172033;--muted:#526075;--border:#ced6e0;--border-strong:#90a0b8;--accent:#0f766e;--accent-soft:#dff7f2;--checkpoint-bg:#10233d;--checkpoint-border:#f2c14e;--checkpoint-text:#f8fafc;--meta-bg:#f7f8fb;--shadow:0 10px 30px rgba(15,23,42,.08)}\n")
	b.WriteString("    *{box-sizing:border-box}\n")
	b.WriteString("    body{font-family:-apple-system,BlinkMacSystemFont,\"Segoe UI\",Roboto,Helvetica,Arial,sans-serif;margin:0;background:linear-gradient(180deg,#f8fafc 0%,var(--bg) 100%);color:var(--text);line-height:1.45}\n")
	b.WriteString("    main{max-width:1080px;margin:0 auto;padding:24px 16px 40px}\n")
	b.WriteString("    header{position:sticky;top:0;z-index:20;background:rgba(255,255,255,.94);backdrop-filter:blur(10px);border:1px solid var(--border);border-radius:14px;padding:16px;margin-bottom:16px;box-shadow:var(--shadow)}\n")
	b.WriteString("    .header-top{display:flex;justify-content:space-between;gap:16px;align-items:flex-start;flex-wrap:wrap}\n")
	b.WriteString("    h1{margin:0 0 8px;font-size:1.4rem}\n")
	b.WriteString("    .meta-line{margin:2px 0;color:var(--muted);font-size:.95rem}\n")
	b.WriteString("    .counts{display:flex;gap:12px;flex-wrap:wrap;margin-top:10px}\n")
	b.WriteString("    .count{background:var(--surface-muted);border:1px solid var(--border);border-radius:999px;padding:6px 10px;font-size:.9rem}\n")
	b.WriteString("    .toolbar{display:flex;gap:12px;flex-wrap:wrap;align-items:end;margin-top:14px;padding-top:14px;border-top:1px solid #e5e7eb}\n")
	b.WriteString("    .search-group{flex:1 1 320px}\n")
	b.WriteString("    .search-label{display:block;font-size:.88rem;font-weight:600;color:var(--muted);margin-bottom:6px}\n")
	b.WriteString("    .search-input{width:100%;padding:10px 12px;border:1px solid var(--border-strong);border-radius:10px;font:inherit;background:#fff;color:var(--text)}\n")
	b.WriteString("    .match-count{font-size:.92rem;color:var(--muted);white-space:nowrap}\n")
	b.WriteString("    .timeline{position:relative;display:flex;flex-direction:column;gap:18px;padding:8px 0 8px 92px}\n")
	b.WriteString("    .timeline::before{content:\"\";position:absolute;left:38px;top:0;bottom:0;width:4px;border-radius:999px;background:linear-gradient(180deg,#c8d3e3 0%,#93a7c5 45%,#c8d3e3 100%)}\n")
	b.WriteString("    .timeline-entry{position:relative}\n")
	b.WriteString("    .timeline-entry[hidden]{display:none !important}\n")
	b.WriteString("    .timeline-marker{position:absolute;left:-97px;top:20px;width:22px;height:22px;border-radius:999px;border:3px solid #fff;background:#9fb0c7;box-shadow:0 0 0 2px rgba(144,160,184,.38),0 6px 14px rgba(15,23,42,.1)}\n")
	b.WriteString("    .timeline-card,.capture,.checkpoint,.meta-event{box-shadow:var(--shadow)}\n")
	b.WriteString("    .capture{background:var(--surface);border:1px solid var(--border);border-radius:14px;padding:14px}\n")
	b.WriteString("    .capture h3{margin:0;font-size:1rem}\n")
	b.WriteString("    .event-header{display:flex;gap:10px;justify-content:space-between;align-items:flex-start;flex-wrap:wrap;margin-bottom:10px}\n")
	b.WriteString("    .event-main{min-width:0;flex:1 1 320px}\n")
	b.WriteString("    .event-actions{display:flex;gap:8px;flex-wrap:wrap;align-items:center}\n")
	b.WriteString("    .checkpoint{background:linear-gradient(135deg,#0f172a 0%,var(--checkpoint-bg) 100%);border:1px solid var(--checkpoint-border);border-radius:18px;padding:18px 18px 18px 20px;color:var(--checkpoint-text);position:relative;overflow:hidden}\n")
	b.WriteString("    .checkpoint::before{content:\"CHECKPOINT\";position:absolute;top:10px;right:12px;font-size:.72rem;letter-spacing:.18em;color:rgba(248,250,252,.55)}\n")
	b.WriteString("    .checkpoint h2{margin:0 0 6px;font-size:1.05rem;padding-right:120px}\n")
	b.WriteString("    .checkpoint .label,.checkpoint .meta-line,.checkpoint .mono-inline{color:inherit}\n")
	b.WriteString("    .meta-event{background:var(--meta-bg);border:1px dashed var(--border-strong);border-radius:14px;padding:12px 14px;color:#374151}\n")
	b.WriteString("    .timeline-entry-checkpoint .timeline-marker{left:-101px;top:14px;width:30px;height:30px;background:radial-gradient(circle at 30% 30%,#fff4c3 0%,#f2c14e 45%,#8a6610 100%);box-shadow:0 0 0 3px rgba(242,193,78,.24),0 10px 18px rgba(15,23,42,.15)}\n")
	b.WriteString("    .timeline-entry-meta .timeline-marker{background:#d8dee8;box-shadow:0 0 0 2px rgba(144,160,184,.32),0 5px 10px rgba(15,23,42,.08)}\n")
	b.WriteString("    .label{font-weight:600}\n")
	b.WriteString("    .badge-row{display:flex;gap:8px;flex-wrap:wrap;margin:0 0 10px}\n")
	b.WriteString("    .badge{display:inline-flex;align-items:center;gap:6px;padding:5px 9px;border-radius:999px;border:1px solid var(--border);background:var(--surface-muted);font-size:.78rem;font-weight:600;letter-spacing:.01em;color:var(--muted)}\n")
	b.WriteString("    .badge-strong{border-color:rgba(242,193,78,.45);background:rgba(242,193,78,.16);color:#f8e6ac}\n")
	b.WriteString("    .badge-accent{border-color:rgba(15,118,110,.22);background:rgba(15,118,110,.09);color:#0f5c56}\n")
	b.WriteString("    .badge-muted{background:#f3f6fa}\n")
	b.WriteString("    .checkpoint .badge{border-color:rgba(255,255,255,.18);background:rgba(255,255,255,.1);color:#f8fafc}\n")
	b.WriteString("    table{border-collapse:collapse;margin:8px 0 6px;width:auto}\n")
	b.WriteString("    th,td{border:1px solid #e5e7eb;padding:4px 8px;font-size:.9rem;text-align:left}\n")
	b.WriteString("    .field-row{display:flex;gap:8px;align-items:center;flex-wrap:wrap;margin:4px 0}\n")
	b.WriteString("    .mono-inline{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;background:rgba(15,23,42,.05);border-radius:6px;padding:2px 6px;max-width:100%;overflow-wrap:anywhere}\n")
	b.WriteString("    .event-actions{gap:6px}\n")
	b.WriteString("    .copy-btn,.toggle-btn{appearance:none;display:inline-flex;align-items:center;justify-content:center;min-width:110px;height:34px;border:1px solid var(--border);background:var(--surface-muted);border-radius:999px;padding:0 12px;font:inherit;font-size:.82rem;font-weight:600;color:var(--text);cursor:pointer;white-space:nowrap}\n")
	b.WriteString("    .copy-btn:hover,.toggle-btn:hover,.search-input:focus{border-color:var(--accent)}\n")
	b.WriteString("    .copy-btn:focus,.toggle-btn:focus,.search-input:focus{outline:2px solid rgba(15,118,110,.2);outline-offset:2px}\n")
	b.WriteString("    .copy-btn[data-copied=\"true\"]{background:var(--accent-soft);border-color:var(--accent);color:#0f5132}\n")
	b.WriteString("    .artifact-shell{display:flex;flex-direction:column;gap:12px}\n")
	b.WriteString("    .artifact-summary{display:flex;justify-content:space-between;gap:12px;align-items:flex-start;flex-wrap:wrap}\n")
	b.WriteString("    .artifact-title-row{display:flex;align-items:center;gap:10px;flex-wrap:wrap;margin-bottom:8px}\n")
	b.WriteString("    .artifact-preview{margin:0;color:var(--muted);font-size:.94rem;white-space:pre-wrap}\n")
	b.WriteString("    .artifact-body{padding-top:2px}\n")
	b.WriteString("    .toggle-row{display:flex;justify-content:space-between;gap:8px;align-items:center;flex-wrap:wrap;margin:10px 0 6px}\n")
	b.WriteString("    .block-label{font-weight:600;font-size:.95rem}\n")
	b.WriteString("    .preview-text{margin:0 0 10px;padding:10px 12px;border:1px dashed #d7dde7;border-radius:10px;background:#fbfcfe;color:var(--muted);font-size:.92rem;white-space:pre-wrap}\n")
	b.WriteString("    .content-block[hidden]{display:none !important}\n")
	b.WriteString("    .artifact-body[hidden]{display:none !important}\n")
	b.WriteString("    pre{margin:0;background:#111827;color:#e5e7eb;border-radius:10px;padding:12px;overflow:auto;white-space:pre-wrap;font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace}\n")
	b.WriteString("    .timeline-divider{position:relative;height:18px;margin:0}\n")
	b.WriteString("    .timeline-divider::before{content:\"\";position:absolute;left:-54px;right:0;top:8px;border-top:1px dashed rgba(144,160,184,.55)}\n")
	b.WriteString("    .timeline-section{position:relative;margin:2px 0 4px}\n")
	b.WriteString("    .timeline-section-label{display:inline-flex;align-items:center;padding:6px 10px;border-radius:999px;border:1px solid var(--border-strong);background:rgba(255,255,255,.9);color:var(--muted);font-size:.8rem;font-weight:700;letter-spacing:.04em;text-transform:uppercase}\n")
	b.WriteString("    .checkpoint-divider{margin:0 0 10px;padding:8px 12px;border-left:4px solid var(--checkpoint-border);border-radius:12px;background:rgba(242,193,78,.08);color:#5b6472;font-size:.9rem}\n")
	b.WriteString("    .timeline-time{font-weight:700;white-space:nowrap}\n")
	b.WriteString("    .empty-state{padding:18px;border:1px dashed var(--border-strong);border-radius:14px;text-align:center;color:var(--muted);background:rgba(255,255,255,.7)}\n")
	b.WriteString("    @media (max-width:760px){.timeline{padding-left:76px}.timeline::before{left:28px}.timeline-marker{left:-79px;width:20px;height:20px}.timeline-entry-checkpoint .timeline-marker{left:-82px;width:26px;height:26px}.timeline-divider::before{left:-44px}}\n")
	b.WriteString("    @media (max-width:640px){main{padding:16px 12px 32px}header{padding:14px}.checkpoint h2{padding-right:0}.event-actions{width:100%}.artifact-summary{flex-direction:column}.timeline{padding-left:0}.timeline::before{left:12px}.timeline-marker{left:3px;top:-6px;width:16px;height:16px;border-width:2px}.timeline-entry-checkpoint .timeline-marker{left:0;top:-10px;width:22px;height:22px}.timeline-divider::before{left:12px}}\n")
	b.WriteString("  </style>\n")
	b.WriteString("</head>\n")
	b.WriteString("<body>\n")
	b.WriteString("<main>\n")
	b.WriteString("  <header>\n")
	b.WriteString("    <div class=\"header-top\">\n")
	b.WriteString("      <div>\n")
	b.WriteString("        <h1>Yanzi Agent Log</h1>\n")
	b.WriteString(fmt.Sprintf("        <div class=\"meta-line\"><span class=\"label\">Project:</span> %s</div>\n", html.EscapeString(project)))
	if latestCheckpointID != "" {
		b.WriteString(fmt.Sprintf("        <div class=\"meta-line\"><span class=\"label\">Checkpoint:</span> <span class=\"mono-inline\">%s</span>", html.EscapeString(latestCheckpointID)))
		if strings.TrimSpace(latestCheckpointSummary) != "" {
			b.WriteString(fmt.Sprintf(" - %s", html.EscapeString(latestCheckpointSummary)))
		}
		b.WriteString("</div>\n")
	}
	b.WriteString(fmt.Sprintf("        <div class=\"meta-line\"><span class=\"label\">Exported:</span> %s</div>\n", html.EscapeString(now.Format(time.RFC3339))))
	b.WriteString(fmt.Sprintf("        <div class=\"meta-line\"><span class=\"label\">Version:</span> %s</div>\n", html.EscapeString(cliVersion)))
	b.WriteString("      </div>\n")
	b.WriteString("      <div class=\"counts\">\n")
	b.WriteString(fmt.Sprintf("        <div class=\"count\">Total artifacts: %d</div>\n", totalCaptures))
	b.WriteString(fmt.Sprintf("        <div class=\"count\">Total events: %d</div>\n", totalEvents))
	b.WriteString(fmt.Sprintf("        <div class=\"count\">Checkpoints: %d</div>\n", totalCheckpoints))
	b.WriteString("      </div>\n")
	b.WriteString("    </div>\n")
	b.WriteString("    <div class=\"toolbar\">\n")
	b.WriteString("      <div class=\"search-group\">\n")
	b.WriteString("        <label class=\"search-label\" for=\"event-search\">Search artifacts</label>\n")
	b.WriteString("        <input id=\"event-search\" class=\"search-input\" type=\"search\" placeholder=\"Search prompt, response, metadata, summary, or timestamp\">\n")
	b.WriteString("      </div>\n")
	b.WriteString(fmt.Sprintf("      <div id=\"match-count\" class=\"match-count\" aria-live=\"polite\">Showing %d of %d events</div>\n", totalEvents, totalEvents))
	b.WriteString("    </div>\n")
	b.WriteString("  </header>\n")
	b.WriteString("  <section class=\"timeline\">\n")
	if totalEvents == 0 {
		b.WriteString("    <div class=\"empty-state\">No events recorded for this export.</div>\n")
	}
	if len(sections) == 0 {
		sections = []htmlRuleSection{{Items: items}}
	}

	idx := 0
	for _, section := range sections {
		if strings.TrimSpace(section.Title) != "" {
			b.WriteString("    <div class=\"timeline-section\">")
			b.WriteString(fmt.Sprintf("<div class=\"timeline-section-label\">%s</div>", html.EscapeString(section.Title)))
			b.WriteString("</div>\n")
		}
		for _, item := range section.Items {
			searchText := html.EscapeString(exportSearchText(item))
			entryClass := "timeline-entry"
			if item.Kind == exportItemCheckpoint {
				entryClass += " timeline-entry-checkpoint"
			} else if item.Kind == exportItemMeta {
				entryClass += " timeline-entry-meta"
			}
			b.WriteString(fmt.Sprintf("    <div class=\"%s event-card\" data-search=\"%s\">\n", entryClass, searchText))
			b.WriteString("      <div class=\"timeline-marker\" aria-hidden=\"true\"></div>\n")
			switch item.Kind {
			case exportItemCheckpoint:
				b.WriteString("      <div class=\"timeline-divider\" aria-hidden=\"true\"></div>\n")
				b.WriteString(fmt.Sprintf("      <div class=\"checkpoint-divider\">Checkpoint boundary%s</div>\n", checkpointDividerSuffix(item.Summary)))
				b.WriteString("      <section class=\"checkpoint timeline-card\">\n")
				b.WriteString("      <div class=\"event-header\">\n")
				b.WriteString("        <div class=\"event-main\">\n")
				b.WriteString("          <div class=\"badge-row\">\n")
				for _, badge := range checkpointBadges(item) {
					b.WriteString(fmt.Sprintf("            <span class=\"badge badge-strong\">%s</span>\n", html.EscapeString(badge)))
				}
				b.WriteString("          </div>\n")
				b.WriteString(fmt.Sprintf("          <h2>Checkpoint: <span class=\"mono-inline\">%s</span></h2>\n", html.EscapeString(item.CheckpointID)))
				b.WriteString(fmt.Sprintf("          <div><span class=\"label\">Summary:</span> %s</div>\n", html.EscapeString(item.Summary)))
				b.WriteString(fmt.Sprintf("          <div class=\"meta-line\"><span class=\"label\">Timestamp:</span> <span class=\"js-timestamp timeline-time\" data-timestamp=\"%s\" title=\"%s\">%s</span></div>\n", html.EscapeString(item.Timestamp), html.EscapeString(item.Timestamp), html.EscapeString(item.Timestamp)))
				b.WriteString("        </div>\n")
				b.WriteString("        <div class=\"event-actions\">\n")
				b.WriteString(fmt.Sprintf("          <button type=\"button\" class=\"copy-btn\" data-copy-text=\"%s\" title=\"Copy checkpoint ID\" aria-label=\"Copy checkpoint ID\">Copy checkpoint ID</button>\n", html.EscapeString(item.CheckpointID)))
				b.WriteString(fmt.Sprintf("          <button type=\"button\" class=\"copy-btn\" data-copy-text=\"%s\" title=\"Copy checkpoint hash\" aria-label=\"Copy checkpoint hash\">Copy hash</button>\n", html.EscapeString(item.CheckpointID)))
				b.WriteString("        </div>\n")
				b.WriteString("      </div>\n")
				b.WriteString("      </section>\n")
			case exportItemMeta:
				b.WriteString("      <section class=\"meta-event timeline-card\">\n")
				b.WriteString(fmt.Sprintf("      <div><span class=\"label\">Event:</span> %s</div>\n", html.EscapeString(item.Command)))
				if strings.TrimSpace(item.Value) != "" {
					b.WriteString(fmt.Sprintf("      <div><span class=\"label\">Value:</span> %s</div>\n", html.EscapeString(item.Value)))
				}
				b.WriteString(fmt.Sprintf("      <div><span class=\"label\">Timestamp:</span> <span class=\"js-timestamp timeline-time\" data-timestamp=\"%s\" title=\"%s\">%s</span></div>\n", html.EscapeString(item.Timestamp), html.EscapeString(item.Timestamp), html.EscapeString(item.Timestamp)))
				b.WriteString("      </section>\n")
			default:
				promptID := fmt.Sprintf("prompt-%d", idx)
				responseID := fmt.Sprintf("response-%d", idx)
				bodyID := fmt.Sprintf("artifact-body-%d", idx)
				b.WriteString("      <section class=\"capture timeline-card\">\n")
				b.WriteString("      <div class=\"artifact-shell\">\n")
				b.WriteString("      <div class=\"artifact-summary\">\n")
				b.WriteString("        <div class=\"event-main\">\n")
				b.WriteString("          <div class=\"artifact-title-row\">\n")
				b.WriteString(fmt.Sprintf("            <h3>Artifact: <span class=\"mono-inline\">%s</span></h3>\n", html.EscapeString(item.CaptureID)))
				b.WriteString(fmt.Sprintf("            <span class=\"js-timestamp timeline-time\" data-timestamp=\"%s\" title=\"%s\">%s</span>\n", html.EscapeString(item.Timestamp), html.EscapeString(item.Timestamp), html.EscapeString(item.Timestamp)))
				b.WriteString("          </div>\n")
				b.WriteString(fmt.Sprintf("          <p class=\"artifact-preview\">%s</p>\n", html.EscapeString(previewText(item.Prompt, 120))))
				b.WriteString("          <div class=\"badge-row\">\n")
				for _, badge := range captureBadges(item) {
					className := "badge badge-muted"
					if badge == "SYSTEM RULE" || strings.HasPrefix(badge, "PROFILE:") || strings.HasPrefix(badge, "Role:") || strings.HasPrefix(badge, "Source:") {
						className = "badge badge-accent"
					}
					b.WriteString(fmt.Sprintf("            <span class=\"%s\">%s</span>\n", className, html.EscapeString(badge)))
				}
				b.WriteString("          </div>\n")
				b.WriteString("        </div>\n")
				b.WriteString("        <div class=\"event-actions\">\n")
				b.WriteString(fmt.Sprintf("          <button type=\"button\" class=\"toggle-btn artifact-toggle\" data-target=\"%s\" aria-expanded=\"false\" title=\"Show full artifact\" aria-label=\"Show full artifact\">Show details</button>\n", bodyID))
				b.WriteString("        </div>\n")
				b.WriteString("      </div>\n")
				b.WriteString(fmt.Sprintf("      <div class=\"artifact-body\" id=\"%s\" hidden>\n", bodyID))
				b.WriteString("      <div class=\"event-header\">\n")
				b.WriteString("        <div class=\"event-main\">\n")
				b.WriteString(fmt.Sprintf("          <div class=\"field-row\"><span class=\"label\">Role:</span> <span>%s</span></div>\n", html.EscapeString(item.Role)))
				b.WriteString(fmt.Sprintf("          <div class=\"field-row\"><span class=\"label\">Hash:</span> <code class=\"mono-inline\">%s</code></div>\n", html.EscapeString(item.Hash)))
				b.WriteString("        </div>\n")
				b.WriteString("        <div class=\"event-actions\">\n")
				b.WriteString(fmt.Sprintf("          <button type=\"button\" class=\"copy-btn\" data-copy-text=\"%s\" title=\"Copy capture ID\" aria-label=\"Copy capture ID\">Copy capture ID</button>\n", html.EscapeString(item.CaptureID)))
				b.WriteString(fmt.Sprintf("          <button type=\"button\" class=\"copy-btn\" data-copy-text=\"%s\" title=\"Copy capture hash\" aria-label=\"Copy capture hash\">Copy hash</button>\n", html.EscapeString(item.Hash)))
				b.WriteString(fmt.Sprintf("          <button type=\"button\" class=\"copy-btn\" data-copy-source=\"%s\" title=\"Copy prompt\" aria-label=\"Copy prompt\">Copy prompt</button>\n", promptID))
				b.WriteString(fmt.Sprintf("          <button type=\"button\" class=\"copy-btn\" data-copy-source=\"%s\" title=\"Copy response\" aria-label=\"Copy response\">Copy response</button>\n", responseID))
				b.WriteString("        </div>\n")
				b.WriteString("      </div>\n")
				if len(item.Metadata) > 0 && !isRuleArtifact(item) {
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
				b.WriteString("      <div class=\"toggle-row\">\n")
				b.WriteString("        <div class=\"block-label\">Prompt</div>\n")
				b.WriteString("      </div>\n")
				b.WriteString(fmt.Sprintf("      <div id=\"%s\" class=\"content-block\">\n", promptID))
				b.WriteString(fmt.Sprintf("        <pre>%s</pre>\n", html.EscapeString(item.Prompt)))
				b.WriteString("      </div>\n")
				b.WriteString("      <div class=\"toggle-row\">\n")
				b.WriteString("        <div class=\"block-label\">Response</div>\n")
				b.WriteString("      </div>\n")
				b.WriteString(fmt.Sprintf("      <div id=\"%s\" class=\"content-block\">\n", responseID))
				b.WriteString(fmt.Sprintf("        <pre>%s</pre>\n", html.EscapeString(item.Response)))
				b.WriteString("      </div>\n")
				b.WriteString("      </div>\n")
				b.WriteString("      </div>\n")
				b.WriteString("      </section>\n")
			}
			b.WriteString("    </div>\n")
			idx++
		}
	}

	b.WriteString("  </section>\n")
	b.WriteString("  <script>\n")
	b.WriteString("    (function(){\n")
	b.WriteString("      const cards=Array.from(document.querySelectorAll('.event-card'));\n")
	b.WriteString("      const searchInput=document.getElementById('event-search');\n")
	b.WriteString("      const matchCount=document.getElementById('match-count');\n")
	b.WriteString("      const timestampNodes=Array.from(document.querySelectorAll('.js-timestamp'));\n")
	b.WriteString("      function updateCount(){\n")
	b.WriteString("        const visible=cards.filter(card=>!card.hidden).length;\n")
	b.WriteString("        if(matchCount){matchCount.textContent='Showing '+visible+' of '+cards.length+' events';}\n")
	b.WriteString("      }\n")
	b.WriteString("      function formatTimestamps(){\n")
	b.WriteString("        const formatter=new Intl.DateTimeFormat(undefined,{month:'short',day:'numeric',year:'numeric',hour:'numeric',minute:'2-digit'});\n")
	b.WriteString("        timestampNodes.forEach(function(node){const raw=node.getAttribute('data-timestamp');if(!raw){return;}const date=new Date(raw);if(Number.isNaN(date.getTime())){return;}node.textContent=formatter.format(date);node.setAttribute('title',raw);});\n")
	b.WriteString("      }\n")
	b.WriteString("      function applyFilter(){\n")
	b.WriteString("        const query=(searchInput&&searchInput.value||'').trim().toLowerCase();\n")
	b.WriteString("        cards.forEach(card=>{const haystack=(card.getAttribute('data-search')||'').toLowerCase();card.hidden=query!==''&&!haystack.includes(query);});\n")
	b.WriteString("        updateCount();\n")
	b.WriteString("      }\n")
	b.WriteString("      function copyText(text){\n")
	b.WriteString("        if(navigator.clipboard&&window.isSecureContext){return navigator.clipboard.writeText(text);}\n")
	b.WriteString("        return new Promise(function(resolve,reject){\n")
	b.WriteString("          const input=document.createElement('textarea');\n")
	b.WriteString("          input.value=text;input.setAttribute('readonly','');input.style.position='fixed';input.style.opacity='0';document.body.appendChild(input);input.select();\n")
	b.WriteString("          try{if(document.execCommand('copy')){resolve();}else{reject(new Error('copy failed'));}}catch(err){reject(err);}finally{document.body.removeChild(input);}\n")
	b.WriteString("        });\n")
	b.WriteString("      }\n")
	b.WriteString("      document.addEventListener('click',function(event){\n")
	b.WriteString("        const toggle=event.target.closest('.toggle-btn');\n")
	b.WriteString("        if(toggle){\n")
	b.WriteString("          const target=document.getElementById(toggle.getAttribute('data-target'));\n")
	b.WriteString("          if(!target){return;}\n")
	b.WriteString("          const expanded=toggle.getAttribute('aria-expanded')==='true';\n")
	b.WriteString("          target.hidden=expanded;\n")
	b.WriteString("          toggle.setAttribute('aria-expanded',String(!expanded));\n")
	b.WriteString("          const showLabel=toggle.classList.contains('artifact-toggle')?'Show details':'Show';\n")
	b.WriteString("          const hideLabel=toggle.classList.contains('artifact-toggle')?'Hide details':'Hide';\n")
	b.WriteString("          toggle.textContent=expanded?showLabel:hideLabel;\n")
	b.WriteString("          toggle.setAttribute('title',expanded?'Show full artifact':'Hide full artifact');\n")
	b.WriteString("          toggle.setAttribute('aria-label',expanded?'Show full artifact':'Hide full artifact');\n")
	b.WriteString("          return;\n")
	b.WriteString("        }\n")
	b.WriteString("        const copyButton=event.target.closest('.copy-btn');\n")
	b.WriteString("        if(copyButton){\n")
	b.WriteString("          const sourceId=copyButton.getAttribute('data-copy-source');\n")
	b.WriteString("          const text=sourceId?(document.getElementById(sourceId)||{}).textContent||'':copyButton.getAttribute('data-copy-text')||'';\n")
	b.WriteString("          const original=copyButton.getAttribute('data-label')||copyButton.textContent;\n")
	b.WriteString("          copyButton.setAttribute('data-label',original);\n")
	b.WriteString("          copyText(text).then(function(){copyButton.textContent='Copied';copyButton.setAttribute('data-copied','true');setTimeout(function(){copyButton.textContent=original;copyButton.removeAttribute('data-copied');},1200);}).catch(function(){copyButton.textContent='Press Ctrl+C';copyButton.setAttribute('data-copied','true');setTimeout(function(){copyButton.textContent=original;copyButton.removeAttribute('data-copied');},1600);});\n")
	b.WriteString("        }\n")
	b.WriteString("      });\n")
	b.WriteString("      if(searchInput){searchInput.addEventListener('input',applyFilter);}\n")
	b.WriteString("      formatTimestamps();\n")
	b.WriteString("      updateCount();\n")
	b.WriteString("    })();\n")
	b.WriteString("  </script>\n")
	b.WriteString("</main>\n")
	b.WriteString("</body>\n")
	b.WriteString("</html>\n")
	return b.String()
}

func exportSearchText(item exportItem) string {
	parts := []string{string(item.Kind), item.Timestamp}
	switch item.Kind {
	case exportItemCheckpoint:
		parts = append(parts, item.CheckpointID, item.Summary)
		parts = append(parts, checkpointBadges(item)...)
	case exportItemMeta:
		parts = append(parts, item.Command, item.Value)
	default:
		parts = append(parts, item.CaptureID, item.Role, item.Source, item.Hash, item.Prompt, item.Response)
		parts = append(parts, captureBadges(item)...)
		for _, key := range sortedMetaKeys(item.Metadata) {
			parts = append(parts, key, item.Metadata[key])
		}
	}
	return strings.Join(parts, " ")
}

func previewText(value string, limit int) string {
	normalized := strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if normalized == "" {
		return ""
	}
	if len(normalized) <= limit {
		return normalized
	}
	return strings.TrimSpace(normalized[:limit]) + "..."
}

func checkpointDividerSuffix(summary string) string {
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return ""
	}
	return ": " + summary
}

func checkpointBadges(item exportItem) []string {
	return []string{"Checkpoint", "Boundary", "Rehydration Anchor", "Hash"}
}

func captureBadges(item exportItem) []string {
	badges := []string{"Capture", "Prompt", "Response", "Hash"}
	if isRuleArtifact(item) {
		profile := strings.TrimSpace(item.Metadata["profile"])
		if profile == "" {
			badges = append([]string{"SYSTEM RULE"}, badges...)
		} else {
			badges = append([]string{fmt.Sprintf("PROFILE: %s", profile)}, badges...)
		}
	}
	if strings.TrimSpace(item.Role) != "" {
		badges = append(badges, fmt.Sprintf("Role: %s", item.Role))
	}
	if strings.TrimSpace(item.Source) != "" {
		badges = append(badges, fmt.Sprintf("Source: %s", item.Source))
	}
	if len(item.Metadata) > 0 && !isRuleArtifact(item) {
		badges = append(badges, "Metadata")
	}
	return badges
}

func isRuleArtifact(item exportItem) bool {
	return strings.TrimSpace(item.Metadata["type"]) == "context" && strings.TrimSpace(item.Metadata["subtype"]) == "rules"
}
