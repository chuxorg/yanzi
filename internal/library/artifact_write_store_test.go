package yanzilibrary

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/core/hash"
	"github.com/chuxorg/yanzi/internal/storage"
)

func TestArtifactWriteStoreCreateCapturePreservesCurrentColumns(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envDBPath, "")

	if _, err := Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	db, err := InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	store := NewArtifactWriteStore(db)
	record, err := store.CreateCapture(context.Background(), CaptureWriteInput{
		Author:     "Ada",
		SourceType: "agent",
		Title:      "Captured",
		Prompt:     "prompt",
		Response:   "response",
		Meta:       json.RawMessage(`{"project":"alpha"}`),
		PrevHash:   "previous",
	})
	if err != nil {
		t.Fatalf("CreateCapture: %v", err)
	}

	computed, err := hash.HashIntent(record)
	if err != nil {
		t.Fatalf("HashIntent: %v", err)
	}
	if computed != record.Hash {
		t.Fatalf("expected stored hash %q, got %q", record.Hash, computed)
	}

	var class string
	var artifactType string
	var content string
	var metadata string
	if err := db.QueryRow(`SELECT class, type, content, metadata FROM intents WHERE id = ?`, record.ID).Scan(&class, &artifactType, &content, &metadata); err != nil {
		t.Fatalf("query capture columns: %v", err)
	}
	if class != "intent" || artifactType != "prompt" || content != record.Prompt || metadata != string(record.Meta) {
		t.Fatalf("unexpected capture artifact columns: class=%q type=%q content=%q metadata=%q", class, artifactType, content, metadata)
	}
}

func TestArtifactWriteStoreCreateArtifactUsesProviderCompatibleWrite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envDBPath, "")

	if _, err := Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if _, err := CreateProject("alpha", ""); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	db, err := InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	store := NewArtifactWriteStore(db)
	artifact, err := store.CreateArtifact(context.Background(), storage.CreateArtifactInput{
		Project:  "alpha",
		Class:    ArtifactClassIntent,
		Type:     "decision",
		Title:    "Provider path",
		Content:  "artifact content",
		Metadata: `{"owner":"api"}`,
	})
	if err != nil {
		t.Fatalf("CreateArtifact: %v", err)
	}
	if artifact.Project != "alpha" || artifact.Type != "decision" || artifact.Metadata != `{"owner":"api"}` {
		t.Fatalf("unexpected artifact: %+v", artifact)
	}
}

func TestArtifactWriteStoreTombstoneAndRestorePreserveColumnRules(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envDBPath, "")

	if _, err := Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if _, err := CreateProject("alpha", ""); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	db, err := InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	store := NewArtifactWriteStore(db)
	capture, err := store.CreateCapture(context.Background(), CaptureWriteInput{
		Author:     "Ada",
		SourceType: "cli",
		Prompt:     "prompt",
		Response:   "response",
		Meta:       json.RawMessage(`{"project":"alpha"}`),
	})
	if err != nil {
		t.Fatalf("CreateCapture: %v", err)
	}
	artifact, err := store.CreateArtifact(context.Background(), storage.CreateArtifactInput{
		Project: "alpha",
		Class:   ArtifactClassIntent,
		Type:    "decision",
		Title:   "Artifact",
		Content: "content",
	})
	if err != nil {
		t.Fatalf("CreateArtifact: %v", err)
	}

	if _, err := store.Tombstone(context.Background(), capture.ID, false, false); err != nil {
		t.Fatalf("Tombstone capture: %v", err)
	}
	if _, err := store.Tombstone(context.Background(), artifact.ID, false, false); err != nil {
		t.Fatalf("Tombstone artifact: %v", err)
	}

	var captureMeta string
	var captureMetadata string
	if err := db.QueryRow(`SELECT meta, metadata FROM intents WHERE id = ?`, capture.ID).Scan(&captureMeta, &captureMetadata); err != nil {
		t.Fatalf("query capture tombstone: %v", err)
	}
	if strings.Contains(captureMeta, `"deleted"`) || !strings.Contains(captureMetadata, `"deleted":"true"`) {
		t.Fatalf("expected capture tombstone in metadata only, meta=%q metadata=%q", captureMeta, captureMetadata)
	}

	var artifactMeta string
	var artifactMetadata sql.NullString
	if err := db.QueryRow(`SELECT meta, metadata FROM intents WHERE id = ?`, artifact.ID).Scan(&artifactMeta, &artifactMetadata); err != nil {
		t.Fatalf("query artifact tombstone: %v", err)
	}
	if !strings.Contains(artifactMeta, `"deleted":"true"`) {
		t.Fatalf("expected artifact tombstone in meta, got %q", artifactMeta)
	}
	if artifactMetadata.Valid && strings.Contains(artifactMetadata.String, `"deleted"`) {
		t.Fatalf("did not expect artifact tombstone in metadata, got %q", artifactMetadata.String)
	}

	if err := store.Restore(context.Background(), artifact.ID); err != nil {
		t.Fatalf("Restore artifact: %v", err)
	}
	if err := db.QueryRow(`SELECT meta FROM intents WHERE id = ?`, artifact.ID).Scan(&artifactMeta); err != nil {
		t.Fatalf("query restored artifact: %v", err)
	}
	if strings.Contains(artifactMeta, `"deleted"`) {
		t.Fatalf("expected deleted metadata removed, got %q", artifactMeta)
	}
}
