package sqlite_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/storage"
)

func TestProjectProviderContract(t *testing.T) {
	provider := openTestProvider(t)
	ctx := context.Background()

	project, err := provider.CreateProject(ctx, storage.CreateProjectInput{Name: " alpha ", Description: "first project"})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if project.Name != "alpha" || project.Description != "first project" || project.CreatedAt.IsZero() {
		t.Fatalf("unexpected project: %+v", project)
	}

	projects, err := provider.ListProjects(ctx)
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 1 || projects[0].Name != "alpha" || projects[0].Description != "first project" {
		t.Fatalf("unexpected projects: %+v", projects)
	}

	exists, err := provider.ProjectExists(ctx, "alpha")
	if err != nil {
		t.Fatalf("ProjectExists alpha: %v", err)
	}
	if !exists {
		t.Fatalf("expected alpha to exist")
	}
	exists, err = provider.ProjectExists(ctx, "missing")
	if err != nil {
		t.Fatalf("ProjectExists missing: %v", err)
	}
	if exists {
		t.Fatalf("did not expect missing project to exist")
	}

	_, err = provider.CreateProject(ctx, storage.CreateProjectInput{Name: "alpha"})
	if err == nil || !strings.Contains(err.Error(), "project already exists: alpha") {
		t.Fatalf("expected duplicate project error, got %v", err)
	}

	_, err = provider.CreateProject(ctx, storage.CreateProjectInput{Name: "   "})
	if err == nil || err.Error() != "project name is required" {
		t.Fatalf("expected required name error, got %v", err)
	}
}

func TestCheckpointProviderContract(t *testing.T) {
	provider := openTestProvider(t)
	ctx := context.Background()
	if _, err := provider.CreateProject(ctx, storage.CreateProjectInput{Name: "alpha"}); err != nil {
		t.Fatalf("CreateProject alpha: %v", err)
	}
	if _, err := provider.CreateProject(ctx, storage.CreateProjectInput{Name: "beta"}); err != nil {
		t.Fatalf("CreateProject beta: %v", err)
	}

	first, err := provider.CreateCheckpoint(ctx, storage.CreateCheckpointInput{Project: "alpha", Summary: " first ", ArtifactIDs: []string{"artifact-1"}})
	if err != nil {
		t.Fatalf("CreateCheckpoint first: %v", err)
	}
	if first.Project != "alpha" || first.Summary != "first" || len(first.ArtifactIDs) != 1 || first.Hash == "" || first.PreviousCheckpointID != "" {
		t.Fatalf("unexpected first checkpoint: %+v", first)
	}

	second, err := provider.CreateCheckpoint(ctx, storage.CreateCheckpointInput{Project: "alpha", Summary: "second"})
	if err != nil {
		t.Fatalf("CreateCheckpoint second: %v", err)
	}
	if second.PreviousCheckpointID != first.Hash {
		t.Fatalf("expected previous checkpoint %q, got %q", first.Hash, second.PreviousCheckpointID)
	}
	if second.Summary != "second" {
		t.Fatalf("unexpected summary: %q", second.Summary)
	}
	if second.ArtifactIDs == nil || len(second.ArtifactIDs) != 0 {
		t.Fatalf("expected empty stored artifact id list, got %#v", second.ArtifactIDs)
	}

	beta, err := provider.CreateCheckpoint(ctx, storage.CreateCheckpointInput{Project: "beta", Summary: "beta checkpoint"})
	if err != nil {
		t.Fatalf("CreateCheckpoint beta: %v", err)
	}
	if beta.Project != "beta" {
		t.Fatalf("unexpected beta checkpoint: %+v", beta)
	}

	alphaCheckpoints, err := provider.ListCheckpoints(ctx, "alpha")
	if err != nil {
		t.Fatalf("ListCheckpoints alpha: %v", err)
	}
	if len(alphaCheckpoints) != 2 {
		t.Fatalf("expected 2 alpha checkpoints, got %+v", alphaCheckpoints)
	}
	if alphaCheckpoints[0].Hash != second.Hash || alphaCheckpoints[1].Hash != first.Hash {
		t.Fatalf("expected newest-first alpha ordering, got %+v", alphaCheckpoints)
	}

	all, err := provider.ListAllCheckpoints(ctx)
	if err != nil {
		t.Fatalf("ListAllCheckpoints: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 checkpoints, got %+v", all)
	}
	if all[0].Project != "alpha" || all[0].Hash != second.Hash || all[1].Project != "alpha" || all[1].Hash != first.Hash || all[2].Project != "beta" {
		t.Fatalf("unexpected all-checkpoint ordering: %+v", all)
	}

	_, err = provider.CreateCheckpoint(ctx, storage.CreateCheckpointInput{Project: "missing", Summary: "nope"})
	if err == nil || !strings.Contains(err.Error(), "project not found: missing") {
		t.Fatalf("expected missing project error, got %v", err)
	}
	_, err = provider.ListCheckpoints(ctx, "missing")
	if err == nil || !strings.Contains(err.Error(), "project not found: missing") {
		t.Fatalf("expected missing project list error, got %v", err)
	}
	_, err = provider.CreateCheckpoint(ctx, storage.CreateCheckpointInput{Project: "alpha", Summary: "   "})
	if err == nil || err.Error() != "summary is required" {
		t.Fatalf("expected summary validation error, got %v", err)
	}
	_, err = provider.ListCheckpoints(ctx, "   ")
	if err == nil || err.Error() != "project is required" {
		t.Fatalf("expected project validation error, got %v", err)
	}
}
