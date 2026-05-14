package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

const (
	rehydratePromptSnippetLimit   = 160
	rehydrateResponseSnippetLimit = 160
)

type rehydrateJSONPayload struct {
	Project        string                   `json:"project"`
	HasCheckpoint  bool                     `json:"has_checkpoint"`
	Fallback       bool                     `json:"fallback"`
	FallbackReason string                   `json:"fallback_reason,omitempty"`
	FallbackLimit  int                      `json:"fallback_limit,omitempty"`
	Checkpoint     *rehydrateJSONCheckpoint `json:"checkpoint,omitempty"`
	Intents        []rehydrateJSONIntent    `json:"intents"`
}

type rehydrateJSONCheckpoint struct {
	Hash                 string   `json:"hash"`
	Project              string   `json:"project"`
	Summary              string   `json:"summary"`
	CreatedAt            string   `json:"created_at"`
	ArtifactIDs          []string `json:"artifact_ids"`
	PreviousCheckpointID string   `json:"previous_checkpoint_id,omitempty"`
}

type rehydrateJSONIntent struct {
	ID              string            `json:"id"`
	Timestamp       string            `json:"timestamp"`
	Author          string            `json:"author"`
	SourceType      string            `json:"source_type"`
	Title           string            `json:"title,omitempty"`
	Prompt          string            `json:"prompt"`
	Response        string            `json:"response"`
	PromptSnippet   string            `json:"prompt_snippet"`
	ResponseSnippet string            `json:"response_snippet"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Hash            string            `json:"hash"`
	PrevHash        string            `json:"prev_hash,omitempty"`
}

// RunRehydrate prints the latest checkpoint and the intent records after it.
//
// Problem:
// Agents need current project state without reconstructing it manually from all
// earlier records.
//
// Solution:
// RunRehydrate loads the active project, resolves the latest checkpoint, and
// renders either a dry-run summary or the ordered records since that boundary.
//
// Arguments:
//
//	args supports `--dry-run`, `--format text|json`, and no positional arguments.
//
// Example:
//
//	yanzi rehydrate --dry-run
func RunRehydrate(args []string) error {
	fs := flag.NewFlagSet("rehydrate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dryRun := fs.Bool("dry-run", false, "preview what rehydrate would load")
	format := fs.String("format", "text", "output format: text or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("usage: yanzi rehydrate [--dry-run] [--format text|json]")
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

	switch cfg.Mode {
	case config.ModeLocal:
	case config.ModeHTTP:
		return errors.New("rehydrate is not available in http mode")
	default:
		return fmt.Errorf("invalid mode: %s", cfg.Mode)
	}

	payload, err := yanzilibrary.RehydrateProject(project)
	if err != nil {
		return err
	}
	status, err := yanzilibrary.LoadProjectStatus(project, yanzilibrary.DefaultStatusRecentLimit)
	if err != nil {
		return err
	}

	switch strings.ToLower(strings.TrimSpace(*format)) {
	case "text":
		if *dryRun {
			renderRehydrateDryRun(payload, status)
			return nil
		}
		renderRehydrateText(payload, status)
		return nil
	case "json":
		if err := renderRehydrateJSON(payload); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("invalid format: %s", *format)
	}
}

func renderRehydrateDryRun(payload *yanzilibrary.RehydratePayload, status *yanzilibrary.ProjectStatus) {
	fmt.Printf("Project: %s\n", payload.Project)
	fmt.Printf("Continuity mode: %s\n", status.ContinuityMode)
	fmt.Printf("Continuity depth: %d\n", status.ContinuityDepth)
	fmt.Printf("Last activity: %s\n", fallbackText(status.LastActivityAt, "(none)"))
	fmt.Printf("Open work: %d\n", len(status.UnresolvedWork))
	if payload.LatestCheckpoint != nil {
		fmt.Printf("Checkpoints to load: %d\n", 1)
		fmt.Printf("Context count: %d\n", len(payload.LatestCheckpoint.ArtifactIDs))
		fmt.Printf("Last checkpoint summary: %s\n", payload.LatestCheckpoint.Summary)
		fmt.Printf("Intents to load: %d\n", len(payload.Intents))
		return
	}

	fmt.Println("Checkpoint status: missing")
	fmt.Printf("Fallback captures to load: %d\n", len(payload.Intents))
	fmt.Printf("Fallback window: last %d captures\n", payload.FallbackLimit)
}

func renderRehydrateText(payload *yanzilibrary.RehydratePayload, status *yanzilibrary.ProjectStatus) {
	intents := sortedRehydrateIntents(payload.Intents)

	fmt.Printf("Project: %s\n\n", payload.Project)
	fmt.Println("Continuity Summary")
	fmt.Printf("Mode: %s\n", status.ContinuityMode)
	fmt.Printf("Depth: %d\n", status.ContinuityDepth)
	fmt.Printf("Last activity: %s\n", fallbackText(status.LastActivityAt, "(none)"))
	fmt.Printf("Open work: %d\n", len(status.UnresolvedWork))
	fmt.Println()
	if payload.LatestCheckpoint != nil {
		fmt.Println("Checkpoint")
		fmt.Printf("Timestamp: %s\n", payload.LatestCheckpoint.CreatedAt)
		fmt.Printf("Summary: %s\n", payload.LatestCheckpoint.Summary)
		if payload.LatestCheckpoint.Hash != "" {
			fmt.Printf("ID: %s\n", payload.LatestCheckpoint.Hash)
		}
		fmt.Println()
		fmt.Println("Post-Checkpoint Continuity")
	} else {
		fmt.Println("Warning: No checkpoint found for active project.")
		fmt.Printf("Showing last %d captures instead.\n\n", payload.FallbackLimit)
		fmt.Println("Recent Continuity")
	}

	if len(intents) == 0 {
		fmt.Println("(none)")
		return
	}

	for i, intent := range intents {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("[%d] %s\n", i+1, intent.CreatedAt.Format(time.RFC3339Nano))
		fmt.Printf("Author: %s\n", fallbackText(intent.Author, "(unknown)"))
		if title := strings.TrimSpace(intent.Title); title != "" {
			fmt.Printf("Title: %s\n", title)
		}
		fmt.Println()
		fmt.Println("Prompt:")
		fmt.Println(truncateSnippet(intent.Prompt, rehydratePromptSnippetLimit))
		fmt.Println()
		fmt.Println("Response:")
		fmt.Println(truncateSnippet(intent.Response, rehydrateResponseSnippetLimit))
		if metaSummary := summarizeIntentMeta(intent.Meta); metaSummary != "" {
			fmt.Println()
			fmt.Println("Meta:")
			fmt.Println(metaSummary)
		}
	}
}

func renderRehydrateJSON(payload *yanzilibrary.RehydratePayload) error {
	out := rehydrateJSONPayload{
		Project:       payload.Project,
		HasCheckpoint: payload.LatestCheckpoint != nil,
		Fallback:      payload.Fallback,
		FallbackLimit: payload.FallbackLimit,
		Intents:       make([]rehydrateJSONIntent, 0, len(payload.Intents)),
	}
	if payload.FallbackReason != "" {
		out.FallbackReason = payload.FallbackReason
	}
	if payload.LatestCheckpoint != nil {
		artifactIDs := append([]string{}, payload.LatestCheckpoint.ArtifactIDs...)
		out.Checkpoint = &rehydrateJSONCheckpoint{
			Hash:                 payload.LatestCheckpoint.Hash,
			Project:              payload.LatestCheckpoint.Project,
			Summary:              payload.LatestCheckpoint.Summary,
			CreatedAt:            payload.LatestCheckpoint.CreatedAt,
			ArtifactIDs:          artifactIDs,
			PreviousCheckpointID: payload.LatestCheckpoint.PreviousCheckpointID,
		}
	}

	for _, intent := range sortedRehydrateIntents(payload.Intents) {
		out.Intents = append(out.Intents, rehydrateJSONIntent{
			ID:              intent.ID,
			Timestamp:       intent.CreatedAt.Format(time.RFC3339Nano),
			Author:          intent.Author,
			SourceType:      intent.SourceType,
			Title:           intent.Title,
			Prompt:          intent.Prompt,
			Response:        intent.Response,
			PromptSnippet:   truncateSnippet(intent.Prompt, rehydratePromptSnippetLimit),
			ResponseSnippet: truncateSnippet(intent.Response, rehydrateResponseSnippetLimit),
			Metadata:        decodeIntentMeta(intent.Meta),
			Hash:            intent.Hash,
			PrevHash:        intent.PrevHash,
		})
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal rehydrate json: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func sortedRehydrateIntents(intents []yanzilibrary.Intent) []yanzilibrary.Intent {
	ordered := append([]yanzilibrary.Intent(nil), intents...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].CreatedAt.Equal(ordered[j].CreatedAt) {
			return ordered[i].ID < ordered[j].ID
		}
		return ordered[i].CreatedAt.Before(ordered[j].CreatedAt)
	})
	return ordered
}

func truncateSnippet(value string, limit int) string {
	cleaned := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\n"), "\r", "\n"))
	if cleaned == "" {
		return "(empty)"
	}
	if limit <= 0 {
		return cleaned
	}

	runes := []rune(cleaned)
	if len(runes) <= limit {
		return cleaned
	}

	cutoff := limit - 3
	if cutoff < 1 {
		cutoff = 1
	}
	snippet := strings.TrimSpace(string(runes[:cutoff]))
	return snippet + "..."
}

func summarizeIntentMeta(raw json.RawMessage) string {
	meta := decodeIntentMeta(raw)
	if len(meta) == 0 {
		return ""
	}
	keys := sortedMetaKeys(meta)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+meta[key])
	}
	return strings.Join(parts, " ")
}

func decodeIntentMeta(raw json.RawMessage) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	meta, err := decodeStringMeta(string(raw))
	if err != nil {
		return nil
	}
	if len(meta) == 0 {
		return nil
	}
	return meta
}

func fallbackText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
