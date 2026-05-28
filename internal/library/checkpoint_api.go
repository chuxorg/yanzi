package yanzilibrary

import (
	"context"

	"github.com/chuxorg/yanzi/internal/storage"
)

// CreateProjectCheckpoint creates a checkpoint for the provided project using the current local provider.
func CreateProjectCheckpoint(project, summary string, artifactIDs []string) (Checkpoint, error) {
	provider, err := openStorageProvider()
	if err != nil {
		return Checkpoint{}, err
	}
	defer func() {
		_ = provider.Close()
	}()

	checkpoint, err := provider.CreateCheckpoint(context.Background(), storage.CreateCheckpointInput{
		Project:     project,
		Summary:     summary,
		ArtifactIDs: artifactIDs,
	})
	if err != nil {
		return Checkpoint{}, err
	}
	return checkpointFromStorage(checkpoint), nil
}

// ListProjectCheckpoints lists checkpoints for a single project using the current local provider.
func ListProjectCheckpoints(project string) ([]Checkpoint, error) {
	provider, err := openStorageProvider()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = provider.Close()
	}()

	checkpoints, err := provider.ListCheckpoints(context.Background(), project)
	if err != nil {
		return nil, err
	}
	result := make([]Checkpoint, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		result = append(result, checkpointFromStorage(checkpoint))
	}
	return result, nil
}

// ListAllProjectCheckpoints lists checkpoints across all projects using the current local provider.
func ListAllProjectCheckpoints() ([]Checkpoint, error) {
	provider, err := openStorageProvider()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = provider.Close()
	}()

	checkpoints, err := provider.ListAllCheckpoints(context.Background())
	if err != nil {
		return nil, err
	}
	result := make([]Checkpoint, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		result = append(result, checkpointFromStorage(checkpoint))
	}
	return result, nil
}
