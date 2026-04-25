package yanzilibrary

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateArtifactAndListArtifacts(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envDBPath, "")

	if _, err := Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if _, err := CreateProject("alpha", ""); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	artifact, err := CreateArtifact("alpha", ArtifactClassIntent, "decision", "Keep export stable", "Preserve legacy export output.", `{"owner":"core"}`)
	if err != nil {
		t.Fatalf("CreateArtifact: %v", err)
	}
	if artifact.Class != ArtifactClassIntent {
		t.Fatalf("expected intent artifact, got %q", artifact.Class)
	}

	_, err = CreateArtifact("alpha", ArtifactClassContext, "policy", "Release policy", "Never rewrite history.", "")
	if err != nil {
		t.Fatalf("CreateArtifact context: %v", err)
	}

	intents, err := ListArtifacts("alpha", ArtifactClassIntent, "")
	if err != nil {
		t.Fatalf("ListArtifacts intent: %v", err)
	}
	if len(intents) != 1 {
		t.Fatalf("expected 1 intent artifact, got %d", len(intents))
	}
	if intents[0].Type != "decision" || intents[0].Title != "Keep export stable" {
		t.Fatalf("unexpected intent artifact: %+v", intents[0])
	}

	contexts, err := ListArtifacts("alpha", ArtifactClassContext, "policy")
	if err != nil {
		t.Fatalf("ListArtifacts context: %v", err)
	}
	if len(contexts) != 1 {
		t.Fatalf("expected 1 context artifact, got %d", len(contexts))
	}
}

func TestCreateArtifactRejectsInvalidType(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envDBPath, "")

	if _, err := Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if _, err := CreateProject("alpha", ""); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	_, err := CreateArtifact("alpha", ArtifactClassContext, "note", "Bad context type", "content", "")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "invalid context type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMigrationAddsArtifactColumns(t *testing.T) {
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

	var count int
	for _, column := range []string{"class", "type", "content", "metadata"} {
		if err := db.QueryRow(`SELECT COUNT(1) FROM pragma_table_info('intents') WHERE name = ?`, column).Scan(&count); err != nil {
			t.Fatalf("check column %s: %v", column, err)
		}
		if count != 1 {
			t.Fatalf("expected column %s to exist", column)
		}
	}

	dbPath := filepath.Join(home, defaultDBDirName, defaultDBFile)
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("stat db: %v", err)
	}
}
