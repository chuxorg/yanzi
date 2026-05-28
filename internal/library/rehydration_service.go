package yanzilibrary

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// RehydrationService isolates the current SQL-backed rehydration flow.
type RehydrationService struct {
	db *sql.DB
}

// NewRehydrationService constructs a rehydration service around the provided database handle.
func NewRehydrationService(db *sql.DB) *RehydrationService {
	return &RehydrationService{db: db}
}

// RehydrateProject loads the latest checkpoint and subsequent intents for a project.
func (s *RehydrationService) RehydrateProject(ctx context.Context, project string) (*RehydratePayload, error) {
	return s.RehydrateProjectWithFallback(ctx, project, DefaultRehydrateFallbackLimit)
}

// RehydrateProjectWithFallback loads checkpoint-based rehydration data or a recent-capture fallback.
func (s *RehydrationService) RehydrateProjectWithFallback(ctx context.Context, project string, fallbackLimit int) (*RehydratePayload, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("rehydration service is not initialized")
	}

	project = strings.TrimSpace(project)
	if project == "" {
		return nil, errors.New("project is required")
	}
	if fallbackLimit <= 0 {
		fallbackLimit = DefaultRehydrateFallbackLimit
	}

	exists, err := projectExists(ctx, s.db, project)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ProjectNotFoundError{Name: project}
	}

	latest, err := latestCheckpointByProject(ctx, s.db, project)
	if err != nil {
		return nil, err
	}
	if latest == nil {
		intents, err := recentProjectIntents(ctx, s.db, project, fallbackLimit)
		if err != nil {
			return nil, err
		}
		return &RehydratePayload{
			Project:        project,
			Intents:        intents,
			Fallback:       true,
			FallbackReason: ErrCheckpointNotFound.Error(),
			FallbackLimit:  fallbackLimit,
		}, nil
	}

	intents, err := intentsSinceCheckpoint(ctx, s.db, project, latest.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &RehydratePayload{
		Project:          project,
		LatestCheckpoint: latest,
		Intents:          intents,
		FallbackLimit:    fallbackLimit,
	}, nil
}
