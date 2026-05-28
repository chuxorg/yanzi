package sqlite_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/storage"
)

func TestArtifactProviderContractCreateListAndVisibility(t *testing.T) {
	provider := openTestProvider(t)
	ctx := context.Background()
	if _, err := provider.CreateProject(ctx, storage.CreateProjectInput{Name: "alpha"}); err != nil {
		t.Fatalf("CreateProject alpha: %v", err)
	}
	if _, err := provider.CreateProject(ctx, storage.CreateProjectInput{Name: "beta"}); err != nil {
		t.Fatalf("CreateProject beta: %v", err)
	}

	alphaIntent, err := provider.CreateArtifact(ctx, storage.CreateArtifactInput{
		Project:  "alpha",
		Class:    storage.ArtifactClassIntent,
		Type:     "decision",
		Title:    "Alpha decision",
		Content:  "Preserve alpha behavior.",
		Metadata: `{"owner":"core"}`,
	})
	if err != nil {
		t.Fatalf("CreateArtifact alpha intent: %v", err)
	}
	if alphaIntent.ID == "" || alphaIntent.Project != "alpha" || alphaIntent.Class != storage.ArtifactClassIntent || alphaIntent.Type != "decision" || alphaIntent.Scope != "" || alphaIntent.Metadata != `{"owner":"core"}` || alphaIntent.CreatedAt == "" {
		t.Fatalf("unexpected alpha intent: %+v", alphaIntent)
	}

	betaIntent, err := provider.CreateArtifact(ctx, storage.CreateArtifactInput{
		Project: "beta",
		Class:   storage.ArtifactClassIntent,
		Type:    "decision",
		Title:   "Beta decision",
		Content: "Preserve beta behavior.",
	})
	if err != nil {
		t.Fatalf("CreateArtifact beta intent: %v", err)
	}

	globalContext, err := provider.CreateArtifact(ctx, storage.CreateArtifactInput{
		Class:   storage.ArtifactClassContext,
		Type:    "reference",
		Scope:   storage.ContextScopeGlobal,
		Title:   "Global reference",
		Content: "Visible everywhere.",
	})
	if err != nil {
		t.Fatalf("CreateArtifact global context: %v", err)
	}
	alphaContext, err := provider.CreateArtifact(ctx, storage.CreateArtifactInput{
		Project: "alpha",
		Class:   storage.ArtifactClassContext,
		Type:    "reference",
		Scope:   storage.ContextScopeProject,
		Title:   "Alpha reference",
		Content: "Visible to alpha.",
	})
	if err != nil {
		t.Fatalf("CreateArtifact alpha context: %v", err)
	}
	betaContext, err := provider.CreateArtifact(ctx, storage.CreateArtifactInput{
		Project: "beta",
		Class:   storage.ArtifactClassContext,
		Type:    "reference",
		Scope:   storage.ContextScopeProject,
		Title:   "Beta reference",
		Content: "Visible to beta.",
	})
	if err != nil {
		t.Fatalf("CreateArtifact beta context: %v", err)
	}

	alphaIntents, err := provider.ListArtifacts(ctx, storage.ArtifactQuery{Project: "alpha", Class: storage.ArtifactClassIntent})
	if err != nil {
		t.Fatalf("ListArtifacts alpha: %v", err)
	}
	if len(alphaIntents) != 1 || alphaIntents[0].ID != alphaIntent.ID {
		t.Fatalf("unexpected alpha intents: %+v", alphaIntents)
	}

	allIntents, err := provider.ListArtifacts(ctx, storage.ArtifactQuery{Class: storage.ArtifactClassIntent, Type: "decision"})
	if err != nil {
		t.Fatalf("ListArtifacts all decisions: %v", err)
	}
	if len(allIntents) != 2 || allIntents[0].ID != betaIntent.ID || allIntents[1].ID != alphaIntent.ID {
		t.Fatalf("expected newest-first intent ordering, got %+v", allIntents)
	}

	visibleAlpha, err := provider.ListVisibleContextArtifacts(ctx, storage.ContextArtifactQuery{ActiveProject: "alpha", Type: "reference"})
	if err != nil {
		t.Fatalf("ListVisibleContextArtifacts alpha: %v", err)
	}
	if artifactIDs(visibleAlpha) != alphaContext.ID+","+globalContext.ID {
		t.Fatalf("unexpected alpha-visible context artifacts: %+v", visibleAlpha)
	}

	visibleAll, err := provider.ListVisibleContextArtifacts(ctx, storage.ContextArtifactQuery{AllProjects: true, Type: "reference", Scope: storage.ContextScopeProject})
	if err != nil {
		t.Fatalf("ListVisibleContextArtifacts all projects: %v", err)
	}
	if artifactIDs(visibleAll) != betaContext.ID+","+alphaContext.ID {
		t.Fatalf("unexpected all-project context artifacts: %+v", visibleAll)
	}

	resolved, err := provider.GetVisibleContextArtifact(ctx, alphaContext.ID[:12], "alpha")
	if err != nil {
		t.Fatalf("GetVisibleContextArtifact: %v", err)
	}
	if resolved.ID != alphaContext.ID {
		t.Fatalf("expected resolved alpha context, got %+v", resolved)
	}
}

func TestArtifactProviderContractErrorsAndDeletedVisibility(t *testing.T) {
	provider := openTestProvider(t)
	ctx := context.Background()
	if _, err := provider.CreateProject(ctx, storage.CreateProjectInput{Name: "alpha"}); err != nil {
		t.Fatalf("CreateProject alpha: %v", err)
	}

	_, err := provider.CreateArtifact(ctx, storage.CreateArtifactInput{
		Project: "missing",
		Class:   storage.ArtifactClassIntent,
		Type:    "decision",
		Title:   "Missing project",
		Content: "Should fail.",
	})
	if err == nil || !strings.Contains(err.Error(), "project not found: missing") {
		t.Fatalf("expected missing project error, got %v", err)
	}
	_, err = provider.CreateArtifact(ctx, storage.CreateArtifactInput{
		Project: "alpha",
		Class:   storage.ArtifactClassIntent,
		Type:    "invalid",
		Title:   "Invalid type",
		Content: "Should fail.",
	})
	if err == nil || !strings.Contains(err.Error(), "invalid intent type") {
		t.Fatalf("expected invalid type error, got %v", err)
	}
	_, err = provider.GetVisibleContextArtifact(ctx, "missing", "alpha")
	if err == nil || err.Error() != "context artifact not found: missing" {
		t.Fatalf("expected missing context artifact error, got %v", err)
	}

	artifact, err := provider.CreateArtifact(ctx, storage.CreateArtifactInput{
		Project: "alpha",
		Class:   storage.ArtifactClassContext,
		Type:    "reference",
		Scope:   storage.ContextScopeProject,
		Title:   "Deleted reference",
		Content: "Existing behavior hides deleted rows from lists.",
	})
	if err != nil {
		t.Fatalf("CreateArtifact deleted reference: %v", err)
	}
	_, err = provider.SQLDB().ExecContext(ctx, `UPDATE intents SET meta = ? WHERE id = ?`, `{"deleted":"true","deleted_at":"2026-01-01T00:00:00Z","project":"alpha","scope":"project"}`, artifact.ID)
	if err != nil {
		t.Fatalf("mark deleted: %v", err)
	}

	visible, err := provider.ListVisibleContextArtifacts(ctx, storage.ContextArtifactQuery{ActiveProject: "alpha"})
	if err != nil {
		t.Fatalf("ListVisibleContextArtifacts: %v", err)
	}
	if len(visible) != 0 {
		t.Fatalf("expected deleted context to be hidden, got %+v", visible)
	}

	withDeleted, err := provider.ListVisibleContextArtifacts(ctx, storage.ContextArtifactQuery{ActiveProject: "alpha", IncludeDeleted: true})
	if err != nil {
		t.Fatalf("ListVisibleContextArtifacts include deleted: %v", err)
	}
	if len(withDeleted) != 1 || withDeleted[0].ID != artifact.ID {
		t.Fatalf("expected deleted context when included, got %+v", withDeleted)
	}

	resolved, err := provider.GetVisibleContextArtifact(ctx, artifact.ID[:12], "alpha")
	if err != nil {
		t.Fatalf("GetVisibleContextArtifact deleted compatibility: %v", err)
	}
	if resolved.ID != artifact.ID {
		t.Fatalf("expected deleted artifact to resolve by current show semantics, got %+v", resolved)
	}
}

func artifactIDs(artifacts []storage.Artifact) string {
	ids := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		ids = append(ids, artifact.ID)
	}
	return strings.Join(ids, ",")
}
