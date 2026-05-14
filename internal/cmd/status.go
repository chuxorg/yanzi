package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

type statusJSONPayload struct {
	Project                  string                `json:"project"`
	ProjectCreatedAt         string                `json:"project_created_at"`
	ContinuityMode           string                `json:"continuity_mode"`
	ContinuityDepth          int                   `json:"continuity_depth"`
	TotalCaptures            int                   `json:"total_captures"`
	TotalProtocolAnnotations int                   `json:"total_protocol_annotations"`
	TotalCheckpoints         int                   `json:"total_checkpoints"`
	TotalIntentArtifacts     int                   `json:"total_intent_artifacts"`
	VisibleContextArtifacts  int                   `json:"visible_context_artifacts"`
	LastActivityAt           string                `json:"last_activity_at,omitempty"`
	LastCaptureAt            string                `json:"last_capture_at,omitempty"`
	LatestCheckpoint         *statusJSONCheckpoint `json:"latest_checkpoint,omitempty"`
	RecentActivity           []statusJSONActivity  `json:"recent_activity"`
	UnresolvedWork           []statusJSONArtifact  `json:"unresolved_work"`
}

type statusJSONCheckpoint struct {
	Hash      string   `json:"hash"`
	Summary   string   `json:"summary"`
	CreatedAt string   `json:"created_at"`
	Artifacts []string `json:"artifact_ids"`
}

type statusJSONActivity struct {
	Kind       string `json:"kind"`
	Timestamp  string `json:"timestamp"`
	ID         string `json:"id,omitempty"`
	Author     string `json:"author,omitempty"`
	Title      string `json:"title,omitempty"`
	Summary    string `json:"summary,omitempty"`
	SourceType string `json:"source_type,omitempty"`
}

type statusJSONArtifact struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Title     string            `json:"title"`
	CreatedAt string            `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

func RunStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	format := fs.String("format", "text", "output format: text or json")
	recent := fs.Int("recent", yanzilibrary.DefaultStatusRecentLimit, "number of recent activity entries to show")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("usage: yanzi status [--format text|json] [--recent <n>]")
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
		return errors.New("status is only available in local mode")
	}

	status, err := yanzilibrary.LoadProjectStatus(project, *recent)
	if err != nil {
		return err
	}

	switch strings.ToLower(strings.TrimSpace(*format)) {
	case "text":
		renderStatusText(status)
		return nil
	case "json":
		return renderStatusJSON(status)
	default:
		return fmt.Errorf("invalid format: %s", *format)
	}
}

func renderStatusText(status *yanzilibrary.ProjectStatus) {
	fmt.Printf("Project: %s\n\n", status.Project)
	fmt.Println("Continuity Status")
	fmt.Printf("Mode: %s\n", status.ContinuityMode)
	fmt.Printf("Project created: %s\n", fallbackText(status.ProjectCreatedAt, "(unknown)"))
	fmt.Printf("Last activity: %s\n", fallbackText(status.LastActivityAt, "(none)"))
	if status.LatestCheckpoint != nil {
		fmt.Printf("Latest checkpoint: %s (%s)\n", fallbackText(strings.TrimSpace(status.LatestCheckpoint.Summary), status.LatestCheckpoint.Hash), status.LatestCheckpoint.CreatedAt)
	} else {
		fmt.Println("Latest checkpoint: (none)")
	}
	fmt.Printf("Continuity depth: %d\n", status.ContinuityDepth)
	fmt.Printf("Open work: %d\n", len(status.UnresolvedWork))
	fmt.Println()
	fmt.Println("Operational Metrics")
	fmt.Printf("Captures: %d\n", status.TotalCaptures)
	fmt.Printf("Protocol annotations: %d\n", status.TotalProtocolAnnotations)
	fmt.Printf("Checkpoints: %d\n", status.TotalCheckpoints)
	fmt.Printf("Intent artifacts: %d\n", status.TotalIntentArtifacts)
	fmt.Printf("Visible context artifacts: %d\n", status.VisibleContextArtifacts)

	fmt.Println()
	fmt.Println("Recent Activity")
	if len(status.RecentActivity) == 0 {
		fmt.Println("(none)")
	} else {
		for i, item := range status.RecentActivity {
			fmt.Printf("[%d] %s  %s\n", i+1, item.Timestamp, statusActivityLabel(item))
		}
	}

	fmt.Println()
	fmt.Println("Unresolved Work")
	if len(status.UnresolvedWork) == 0 {
		fmt.Println("(none)")
		return
	}
	for i, artifact := range status.UnresolvedWork {
		fmt.Printf("[%d] %s  %s: %s\n", i+1, artifact.CreatedAt, artifact.Type, artifact.Title)
		if meta := summarizeArtifactStatusMeta(artifact.Metadata); meta != "" {
			fmt.Printf("    %s\n", meta)
		}
	}
}

func renderStatusJSON(status *yanzilibrary.ProjectStatus) error {
	payload := statusJSONPayload{
		Project:                  status.Project,
		ProjectCreatedAt:         status.ProjectCreatedAt,
		ContinuityMode:           status.ContinuityMode,
		ContinuityDepth:          status.ContinuityDepth,
		TotalCaptures:            status.TotalCaptures,
		TotalProtocolAnnotations: status.TotalProtocolAnnotations,
		TotalCheckpoints:         status.TotalCheckpoints,
		TotalIntentArtifacts:     status.TotalIntentArtifacts,
		VisibleContextArtifacts:  status.VisibleContextArtifacts,
		LastActivityAt:           status.LastActivityAt,
		LastCaptureAt:            status.LastCaptureAt,
		RecentActivity:           make([]statusJSONActivity, 0, len(status.RecentActivity)),
		UnresolvedWork:           make([]statusJSONArtifact, 0, len(status.UnresolvedWork)),
	}
	if status.LatestCheckpoint != nil {
		payload.LatestCheckpoint = &statusJSONCheckpoint{
			Hash:      status.LatestCheckpoint.Hash,
			Summary:   status.LatestCheckpoint.Summary,
			CreatedAt: status.LatestCheckpoint.CreatedAt,
			Artifacts: append([]string{}, status.LatestCheckpoint.ArtifactIDs...),
		}
	}
	for _, item := range status.RecentActivity {
		payload.RecentActivity = append(payload.RecentActivity, statusJSONActivity{
			Kind:       item.Kind,
			Timestamp:  item.Timestamp,
			ID:         item.ID,
			Author:     item.Author,
			Title:      item.Title,
			Summary:    item.Summary,
			SourceType: item.SourceType,
		})
	}
	for _, artifact := range status.UnresolvedWork {
		payload.UnresolvedWork = append(payload.UnresolvedWork, statusJSONArtifact{
			ID:        artifact.ID,
			Type:      artifact.Type,
			Title:     artifact.Title,
			CreatedAt: artifact.CreatedAt,
			Metadata:  decodeMetadataForStatusJSON(artifact.Metadata),
		})
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal status json: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func statusActivityLabel(item yanzilibrary.StatusActivity) string {
	switch item.Kind {
	case "checkpoint":
		return "checkpoint: " + fallbackText(item.Summary, item.ID)
	case "protocol_annotation":
		return "protocol: " + fallbackText(item.Summary, item.ID)
	default:
		label := fallbackText(item.Summary, item.ID)
		if strings.TrimSpace(item.Author) != "" {
			return fmt.Sprintf("capture by %s: %s", item.Author, label)
		}
		return "capture: " + label
	}
}

func summarizeArtifactStatusMeta(raw string) string {
	meta, err := decodeStringMeta(raw)
	if err != nil || len(meta) == 0 {
		return ""
	}
	lines := make([]string, 0, len(meta))
	for _, key := range sortedMetaKeys(meta) {
		lines = append(lines, key+"="+meta[key])
	}
	return strings.Join(lines, " ")
}

func decodeMetadataForStatusJSON(raw string) map[string]string {
	meta, err := decodeStringMeta(raw)
	if err != nil || len(meta) == 0 {
		return nil
	}
	return meta
}
