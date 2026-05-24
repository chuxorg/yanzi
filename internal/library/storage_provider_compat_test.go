package yanzilibrary

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/chuxorg/yanzi/internal/config"
)

func TestStorageProviderBoundaryPreservesLibraryWorkflow(t *testing.T) {
	home := t.TempDir()
	dbPath := filepath.Join(home, "state", "yanzi.db")
	t.Setenv("HOME", home)
	t.Setenv(config.LocalDBPathEnvVar, dbPath)

	initialized, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if !initialized {
		t.Fatalf("expected first initialization")
	}

	if _, err := CreateProject("alpha", "provider compatibility"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	artifact, err := CreateArtifact("alpha", ArtifactClassIntent, "decision", "Provider seam", "Keep CLI behavior stable.", `{"phase":"cap-001"}`)
	if err != nil {
		t.Fatalf("CreateArtifact: %v", err)
	}

	db, err := InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	checkpoint, err := CreateCheckpoint(context.Background(), db, "alpha", "provider boundary", []string{artifact.ID})
	if err != nil {
		t.Fatalf("CreateCheckpoint: %v", err)
	}
	if checkpoint.Project != "alpha" || checkpoint.Hash == "" {
		t.Fatalf("unexpected checkpoint: %+v", checkpoint)
	}

	listed, err := ListArtifacts("alpha", ArtifactClassIntent, "decision", false)
	if err != nil {
		t.Fatalf("ListArtifacts: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != artifact.ID {
		t.Fatalf("unexpected artifacts: %+v", listed)
	}

	payload, err := RehydrateProject("alpha")
	if err != nil {
		t.Fatalf("RehydrateProject: %v", err)
	}
	if payload.Project != "alpha" || payload.LatestCheckpoint == nil || payload.LatestCheckpoint.Hash != checkpoint.Hash {
		t.Fatalf("unexpected rehydrate payload: %+v", payload)
	}
}
