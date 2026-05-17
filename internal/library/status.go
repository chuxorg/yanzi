package yanzilibrary

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

const DefaultStatusRecentLimit = 5

type ProjectStatus struct {
	Project                  string
	ProjectCreatedAt         string
	ContinuityMode           string
	ContinuityDepth          int
	TotalCaptures            int
	TotalProtocolAnnotations int
	TotalCheckpoints         int
	TotalIntentArtifacts     int
	VisibleContextArtifacts  int
	LastActivityAt           string
	LastCaptureAt            string
	LatestCheckpoint         *Checkpoint
	RecentActivity           []StatusActivity
	UnresolvedWork           []Artifact
}

type StatusActivity struct {
	Kind       string
	Timestamp  string
	ID         string
	Author     string
	Title      string
	Summary    string
	SourceType string
}

func LoadProjectStatus(project string, recentLimit int) (*ProjectStatus, error) {
	project = strings.TrimSpace(project)
	if project == "" {
		return nil, errors.New("project is required")
	}
	if recentLimit <= 0 {
		recentLimit = DefaultStatusRecentLimit
	}

	db, err := InitDB()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = db.Close()
	}()

	return loadProjectStatus(context.Background(), db, project, recentLimit)
}

func loadProjectStatus(ctx context.Context, db *sql.DB, project string, recentLimit int) (*ProjectStatus, error) {
	projectRow, err := loadProjectRow(ctx, db, project)
	if err != nil {
		return nil, err
	}

	latestCheckpoint, checkpointCount, err := loadCheckpointStatus(ctx, db, project)
	if err != nil {
		return nil, err
	}

	activities, totalCaptures, totalProtocolAnnotations, totalOperationalRecords, lastCaptureAt, err := loadProjectActivities(ctx, db, project, recentLimit)
	if err != nil {
		return nil, err
	}
	recentActivity := mergeRecentActivity(activities, latestCheckpoint, recentLimit)
	lastActivityAt := latestActivityTimestamp(recentActivity)

	intentArtifacts, err := ListArtifacts(project, ArtifactClassIntent, "", false)
	if err != nil {
		return nil, err
	}
	visibleContextArtifacts, err := ListVisibleContextArtifacts(project, "", "", "", false)
	if err != nil {
		return nil, err
	}

	unresolvedWork := unresolvedArtifacts(intentArtifacts)
	continuityMode := "fallback"
	continuityDepth := totalOperationalRecords
	if latestCheckpoint != nil {
		continuityMode = "checkpoint"
		if continuityDepth, err = countContinuityDepth(ctx, db, project, latestCheckpoint.CreatedAt); err != nil {
			return nil, err
		}
	} else if continuityDepth > DefaultRehydrateFallbackLimit {
		continuityDepth = DefaultRehydrateFallbackLimit
	}

	return &ProjectStatus{
		Project:                  project,
		ProjectCreatedAt:         projectRow.CreatedAt.Format(time.RFC3339Nano),
		ContinuityMode:           continuityMode,
		ContinuityDepth:          continuityDepth,
		TotalCaptures:            totalCaptures,
		TotalProtocolAnnotations: totalProtocolAnnotations,
		TotalCheckpoints:         checkpointCount,
		TotalIntentArtifacts:     len(intentArtifacts),
		VisibleContextArtifacts:  len(visibleContextArtifacts),
		LastActivityAt:           lastActivityAt,
		LastCaptureAt:            lastCaptureAt,
		LatestCheckpoint:         latestCheckpoint,
		RecentActivity:           recentActivity,
		UnresolvedWork:           unresolvedWork,
	}, nil
}

func loadProjectRow(ctx context.Context, db *sql.DB, project string) (Project, error) {
	var row Project
	var createdAtText string
	var description sql.NullString
	err := db.QueryRowContext(ctx, `SELECT name, description, created_at FROM projects WHERE name = ?`, project).Scan(&row.Name, &description, &createdAtText)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, ProjectNotFoundError{Name: project}
		}
		return Project{}, err
	}
	if description.Valid {
		row.Description = description.String
	}
	row.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAtText)
	if err != nil {
		return Project{}, fmt.Errorf("parse project created_at for %s: %w", project, err)
	}
	return row, nil
}

func loadCheckpointStatus(ctx context.Context, db *sql.DB, project string) (*Checkpoint, int, error) {
	rows, err := db.QueryContext(ctx, `SELECT hash, project, summary, created_at, artifact_ids, previous_checkpoint_id FROM checkpoints WHERE project = ? ORDER BY created_at DESC`, project)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	count := 0
	var latest *Checkpoint
	for rows.Next() {
		var checkpoint Checkpoint
		var artifactText string
		var prev sql.NullString
		if err := rows.Scan(&checkpoint.Hash, &checkpoint.Project, &checkpoint.Summary, &checkpoint.CreatedAt, &artifactText, &prev); err != nil {
			return nil, 0, err
		}
		if artifactText != "" {
			if err := json.Unmarshal([]byte(artifactText), &checkpoint.ArtifactIDs); err != nil {
				return nil, 0, fmt.Errorf("decode checkpoint artifact_ids: %w", err)
			}
		}
		if prev.Valid {
			checkpoint.PreviousCheckpointID = prev.String
		}
		if latest == nil {
			copied := checkpoint
			latest = &copied
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return latest, count, nil
}

func loadProjectActivities(ctx context.Context, db *sql.DB, project string, recentLimit int) ([]StatusActivity, int, int, int, string, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, created_at, author, source_type, title, prompt, response, meta, metadata
		FROM intents
		WHERE source_type <> 'artifact'
		ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, 0, 0, 0, "", err
	}
	defer rows.Close()

	activities := make([]StatusActivity, 0, recentLimit)
	totalCaptures := 0
	totalProtocolAnnotations := 0
	totalOperationalRecords := 0
	lastCaptureAt := ""
	for rows.Next() {
		var (
			id, createdAt, author, sourceType, prompt, response string
			title, metaText, metadataText                       sql.NullString
		)
		if err := rows.Scan(&id, &createdAt, &author, &sourceType, &title, &prompt, &response, &metaText, &metadataText); err != nil {
			return nil, 0, 0, 0, "", err
		}
		meta, err := mergeStatusMetadata(metaText.String, metadataText.String)
		if err != nil {
			return nil, 0, 0, 0, "", err
		}
		if strings.TrimSpace(meta["project"]) != project || isDeletedStatusMetadata(meta) {
			continue
		}
		totalOperationalRecords++

		kind := "capture"
		summary := truncateStatusLine(firstNonEmpty(strings.TrimSpace(title.String), prompt), 96)
		if isStatusProtocolSource(sourceType) {
			kind = "protocol_annotation"
			summary = truncateStatusLine(firstNonEmpty(prompt, response), 96)
			totalProtocolAnnotations++
		} else {
			totalCaptures++
			if lastCaptureAt == "" {
				lastCaptureAt = createdAt
			}
		}

		if len(activities) < recentLimit {
			activities = append(activities, StatusActivity{
				Kind:       kind,
				Timestamp:  createdAt,
				ID:         id,
				Author:     author,
				Title:      strings.TrimSpace(title.String),
				Summary:    summary,
				SourceType: sourceType,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, 0, 0, "", err
	}

	return activities, totalCaptures, totalProtocolAnnotations, totalOperationalRecords, lastCaptureAt, nil
}

func mergeRecentActivity(activities []StatusActivity, latestCheckpoint *Checkpoint, recentLimit int) []StatusActivity {
	merged := append([]StatusActivity(nil), activities...)
	if latestCheckpoint != nil {
		merged = append(merged, StatusActivity{
			Kind:      "checkpoint",
			Timestamp: latestCheckpoint.CreatedAt,
			ID:        latestCheckpoint.Hash,
			Summary:   strings.TrimSpace(latestCheckpoint.Summary),
		})
	}
	sort.SliceStable(merged, func(i, j int) bool {
		if merged[i].Timestamp == merged[j].Timestamp {
			return merged[i].ID > merged[j].ID
		}
		return merged[i].Timestamp > merged[j].Timestamp
	})
	if len(merged) > recentLimit {
		merged = merged[:recentLimit]
	}
	return merged
}

func countContinuityDepth(ctx context.Context, db *sql.DB, project, checkpointCreatedAt string) (int, error) {
	rows, err := db.QueryContext(ctx, `SELECT meta, metadata
		FROM intents
		WHERE source_type <> 'artifact' AND created_at > ?
		ORDER BY created_at ASC, id ASC`, checkpointCreatedAt)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var metaText, metadataText sql.NullString
		if err := rows.Scan(&metaText, &metadataText); err != nil {
			return 0, err
		}
		meta, err := mergeStatusMetadata(metaText.String, metadataText.String)
		if err != nil {
			return 0, err
		}
		if strings.TrimSpace(meta["project"]) != project || isDeletedStatusMetadata(meta) {
			continue
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

func unresolvedArtifacts(artifacts []Artifact) []Artifact {
	open := make([]Artifact, 0)
	for _, artifact := range artifacts {
		if artifact.Type != "task" && artifact.Type != "change_request" {
			continue
		}
		if artifactResolved(artifact) {
			continue
		}
		open = append(open, artifact)
	}
	sort.SliceStable(open, func(i, j int) bool {
		if open[i].CreatedAt == open[j].CreatedAt {
			return open[i].ID > open[j].ID
		}
		return open[i].CreatedAt > open[j].CreatedAt
	})
	return open
}

func artifactResolved(artifact Artifact) bool {
	if strings.TrimSpace(artifact.Metadata) == "" {
		return false
	}
	meta, err := decodeStatusStringMeta(artifact.Metadata)
	if err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(meta["status"])) {
	case "done", "completed", "complete", "closed", "resolved", "cancelled", "canceled", "rejected":
		return true
	default:
		return false
	}
}

func mergeStatusMetadata(metaText, metadataText string) (map[string]string, error) {
	meta := map[string]string{}
	for _, raw := range []string{metaText, metadataText} {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		decoded, err := decodeStatusStringMeta(raw)
		if err != nil {
			return nil, err
		}
		for key, value := range decoded {
			meta[key] = value
		}
	}
	return meta, nil
}

func decodeStatusStringMeta(raw string) (map[string]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var decoded map[string]string
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

func isDeletedStatusMetadata(meta map[string]string) bool {
	return strings.EqualFold(strings.TrimSpace(meta["deleted"]), "true")
}

func isStatusProtocolSource(sourceType string) bool {
	switch strings.ToLower(strings.TrimSpace(sourceType)) {
	case "meta-command", "meta_command", "event":
		return true
	default:
		return false
	}
}

func latestActivityTimestamp(activities []StatusActivity) string {
	if len(activities) == 0 {
		return ""
	}
	return activities[0].Timestamp
}

func truncateStatusLine(value string, limit int) string {
	normalized := strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if normalized == "" || limit <= 0 || len(normalized) <= limit {
		return normalized
	}
	return strings.TrimSpace(normalized[:limit]) + "..."
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
