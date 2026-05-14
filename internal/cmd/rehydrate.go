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
	rehydrateOpenIntentLimit      = 8
)

type rehydrateView struct {
	Payload     *yanzilibrary.RehydratePayload
	OpenIntents []rehydrateOpenIntentArtifact
}

type rehydrateOpenIntentArtifact struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Title     string            `json:"title"`
	Content   string            `json:"content"`
	Snippet   string            `json:"snippet"`
	CreatedAt string            `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type rehydrateJSONPayload struct {
	Project        string                        `json:"project"`
	ContinuityMode string                        `json:"continuity_mode"`
	HasCheckpoint  bool                          `json:"has_checkpoint"`
	Fallback       bool                          `json:"fallback"`
	FallbackReason string                        `json:"fallback_reason,omitempty"`
	FallbackLimit  int                           `json:"fallback_limit,omitempty"`
	Checkpoint     *rehydrateJSONCheckpoint      `json:"checkpoint,omitempty"`
	Intents        []rehydrateJSONIntent         `json:"intents"`
	OpenIntents    []rehydrateOpenIntentArtifact `json:"open_intents,omitempty"`
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
	ID              string                 `json:"id"`
	Timestamp       string                 `json:"timestamp"`
	Author          string                 `json:"author"`
	SourceType      string                 `json:"source_type"`
	Title           string                 `json:"title,omitempty"`
	Prompt          string                 `json:"prompt"`
	Response        string                 `json:"response"`
	PromptSnippet   string                 `json:"prompt_snippet"`
	ResponseSnippet string                 `json:"response_snippet"`
	Metadata        map[string]string      `json:"metadata,omitempty"`
	Hash            string                 `json:"hash"`
	PrevHash        string                 `json:"prev_hash,omitempty"`
	IsLatest        bool                   `json:"is_latest"`
	Protocol        *rehydrateJSONProtocol `json:"protocol,omitempty"`
}

type rehydrateJSONProtocol struct {
	Command    string `json:"command"`
	Kind       string `json:"kind"`
	Argument   string `json:"argument,omitempty"`
	Value      string `json:"value,omitempty"`
	Executable bool   `json:"executable"`
	Semantics  string `json:"semantics"`
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
	view, err := buildRehydrateView(payload)
	if err != nil {
		return err
	}

	switch strings.ToLower(strings.TrimSpace(*format)) {
	case "text":
		if *dryRun {
			renderRehydrateDryRun(view)
			return nil
		}
		renderRehydrateText(view)
		return nil
	case "json":
		if err := renderRehydrateJSON(view); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("invalid format: %s", *format)
	}
}

func buildRehydrateView(payload *yanzilibrary.RehydratePayload) (rehydrateView, error) {
	openIntents, err := loadOpenIntentArtifacts(payload.Project)
	if err != nil {
		return rehydrateView{}, err
	}
	return rehydrateView{
		Payload:     payload,
		OpenIntents: openIntents,
	}, nil
}

func renderRehydrateDryRun(view rehydrateView) {
	payload := view.Payload
	fmt.Printf("Project: %s\n", payload.Project)
	fmt.Printf("Continuity mode: %s\n", rehydrateMode(payload))
	if payload.LatestCheckpoint != nil {
		fmt.Printf("Checkpoints to load: %d\n", 1)
		fmt.Printf("Context count: %d\n", len(payload.LatestCheckpoint.ArtifactIDs))
		fmt.Printf("Last checkpoint summary: %s\n", payload.LatestCheckpoint.Summary)
		fmt.Printf("Intents to load: %d\n", len(payload.Intents))
		fmt.Printf("Open intent artifacts: %d\n", len(view.OpenIntents))
		return
	}

	fmt.Println("Checkpoint status: missing")
	fmt.Printf("Fallback captures to load: %d\n", len(payload.Intents))
	fmt.Printf("Fallback window: last %d captures\n", payload.FallbackLimit)
	fmt.Printf("Open intent artifacts: %d\n", len(view.OpenIntents))
}

func renderRehydrateText(view rehydrateView) {
	payload := view.Payload
	intents := sortedRehydrateIntents(payload.Intents)
	latestIntentID := ""
	if len(intents) > 0 {
		latestIntentID = intents[len(intents)-1].ID
	}

	fmt.Printf("Project: %s\n", payload.Project)
	fmt.Printf("Continuity Mode: %s\n", rehydrateMode(payload))
	if payload.LatestCheckpoint != nil {
		fmt.Printf("Checkpoint Boundary: %s\n", payload.LatestCheckpoint.CreatedAt)
	} else {
		fmt.Printf("Checkpoint Boundary: missing\n")
	}
	if latestIntentID != "" {
		fmt.Printf("Latest Continuity Point: %s\n", latestIntentID)
	}
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
	} else {
		for i, intent := range intents {
			if i > 0 {
				fmt.Println()
			}
			renderRehydrateIntentText(i+1, intent, intent.ID == latestIntentID)
		}
	}

	if len(view.OpenIntents) > 0 {
		fmt.Println()
		fmt.Println("Open Intent Artifacts")
		for _, artifact := range view.OpenIntents {
			fmt.Printf("- [%s] %s\n", artifact.Type, artifact.Title)
			fmt.Printf("  Created: %s\n", artifact.CreatedAt)
			fmt.Printf("  Detail: %s\n", artifact.Snippet)
			if metaSummary := summarizeStringMeta(artifact.Metadata); metaSummary != "" {
				fmt.Printf("  Meta: %s\n", metaSummary)
			}
		}
	}
}

func renderRehydrateIntentText(index int, intent yanzilibrary.Intent, latest bool) {
	fmt.Printf("[%d] %s\n", index, intent.CreatedAt.Format(time.RFC3339Nano))
	if latest {
		fmt.Println("Status: latest continuity point")
	}

	if protocol, ok := protocolAnnotationForIntent(intent); ok {
		fmt.Printf("Protocol Annotation: %s\n", protocol.Raw)
		fmt.Printf("Semantics: logging convention only\n")
		fmt.Printf("Annotation Type: %s\n", protocolKindLabel(protocol.Kind))
		if protocol.Argument != "" {
			fmt.Printf("Argument: %s\n", protocol.Argument)
		}
		if value := strings.TrimSpace(intent.Response); value != "" {
			fmt.Printf("Value: %s\n", value)
		}
		if metaSummary := summarizeIntentMeta(intent.Meta); metaSummary != "" {
			fmt.Println()
			fmt.Println("Meta:")
			fmt.Println(metaSummary)
		}
		return
	}

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

func renderRehydrateJSON(view rehydrateView) error {
	payload := view.Payload
	orderedIntents := sortedRehydrateIntents(payload.Intents)
	out := rehydrateJSONPayload{
		Project:        payload.Project,
		ContinuityMode: rehydrateMode(payload),
		HasCheckpoint:  payload.LatestCheckpoint != nil,
		Fallback:       payload.Fallback,
		FallbackLimit:  payload.FallbackLimit,
		Intents:        make([]rehydrateJSONIntent, 0, len(payload.Intents)),
		OpenIntents:    append([]rehydrateOpenIntentArtifact(nil), view.OpenIntents...),
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

	latestIntentID := ""
	if len(orderedIntents) > 0 {
		latestIntentID = orderedIntents[len(orderedIntents)-1].ID
	}

	for _, intent := range orderedIntents {
		record := rehydrateJSONIntent{
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
			IsLatest:        intent.ID == latestIntentID,
		}
		if protocol, ok := protocolAnnotationForIntent(intent); ok {
			record.Protocol = &rehydrateJSONProtocol{
				Command:    protocol.Raw,
				Kind:       protocolKindLabel(protocol.Kind),
				Argument:   protocol.Argument,
				Value:      strings.TrimSpace(intent.Response),
				Executable: protocol.Executable,
				Semantics:  protocol.Semantics,
			}
		}
		out.Intents = append(out.Intents, record)
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal rehydrate json: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func loadOpenIntentArtifacts(project string) ([]rehydrateOpenIntentArtifact, error) {
	artifacts, err := yanzilibrary.ListArtifacts(project, yanzilibrary.ArtifactClassIntent, "", false)
	if err != nil {
		return nil, err
	}

	open := make([]rehydrateOpenIntentArtifact, 0)
	for _, artifact := range artifacts {
		if artifact.Type != "task" && artifact.Type != "change_request" {
			continue
		}

		meta, err := decodeStringMeta(artifact.Metadata)
		if err != nil {
			continue
		}
		if closedArtifactStatus(meta) {
			continue
		}

		open = append(open, rehydrateOpenIntentArtifact{
			ID:        artifact.ID,
			Type:      artifact.Type,
			Title:     artifact.Title,
			Content:   artifact.Content,
			Snippet:   truncateSnippet(artifact.Content, rehydratePromptSnippetLimit),
			CreatedAt: artifact.CreatedAt,
			Metadata:  meta,
		})
	}

	sort.SliceStable(open, func(i, j int) bool {
		if open[i].CreatedAt == open[j].CreatedAt {
			return open[i].ID < open[j].ID
		}
		return open[i].CreatedAt < open[j].CreatedAt
	})
	if len(open) > rehydrateOpenIntentLimit {
		open = open[:rehydrateOpenIntentLimit]
	}
	return open, nil
}

func closedArtifactStatus(meta map[string]string) bool {
	if len(meta) == 0 {
		return false
	}
	status := strings.ToLower(strings.TrimSpace(meta["status"]))
	switch status {
	case "closed", "done", "resolved":
		return true
	default:
		return false
	}
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
	return summarizeStringMeta(decodeIntentMeta(raw))
}

func summarizeStringMeta(meta map[string]string) string {
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

func protocolAnnotationForIntent(intent yanzilibrary.Intent) (protocolAnnotation, bool) {
	if isMetaCommandSource(intent.SourceType) {
		return parseProtocolAnnotation(intent.Prompt)
	}
	return protocolAnnotation{}, false
}

func rehydrateMode(payload *yanzilibrary.RehydratePayload) string {
	if payload.Fallback {
		return "fallback"
	}
	return "checkpoint"
}

func fallbackText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
