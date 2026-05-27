package sqlite_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/core/model"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
	"github.com/chuxorg/yanzi/internal/storage"
	"github.com/chuxorg/yanzi/internal/storage/sqlite"
)

func TestProviderRegressionReusesExistingDatabase(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "yanzi.db")

	provider, initialized, err := sqlite.Open(ctx, path, yanzilibrary.MigrationsFS())
	if err != nil {
		t.Fatalf("Open first provider: %v", err)
	}
	if !initialized {
		t.Fatalf("expected first open to initialize schema")
	}
	if _, err := provider.CreateProject(ctx, storage.CreateProjectInput{Name: "alpha"}); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := provider.Close(); err != nil {
		t.Fatalf("Close first provider: %v", err)
	}

	reopened, initialized, err := sqlite.Open(ctx, path, yanzilibrary.MigrationsFS())
	if err != nil {
		t.Fatalf("Open existing provider: %v", err)
	}
	defer reopened.Close()
	if initialized {
		t.Fatalf("expected existing database open to avoid first-time initialization")
	}
	exists, err := reopened.ProjectExists(ctx, "alpha")
	if err != nil {
		t.Fatalf("ProjectExists after reopen: %v", err)
	}
	if !exists {
		t.Fatalf("expected project from existing database")
	}
}

func TestProviderRegressionRepeatedExportAndVerifyReadsAreStable(t *testing.T) {
	provider := openTestProvider(t)
	ctx := context.Background()
	if _, err := provider.CreateProject(ctx, storage.CreateProjectInput{Name: "alpha"}); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	intent := seedProviderIntent(t, provider.SQLDB(), testIntent("repeat-1", "2026-01-01T00:00:01Z", map[string]string{"project": "alpha"}))

	firstItems, firstCount, err := provider.ListExportItems(ctx, storage.ExportQuery{Project: "alpha"})
	if err != nil {
		t.Fatalf("first ListExportItems: %v", err)
	}
	secondItems, secondCount, err := provider.ListExportItems(ctx, storage.ExportQuery{Project: "alpha"})
	if err != nil {
		t.Fatalf("second ListExportItems: %v", err)
	}
	if firstCount != secondCount || !reflect.DeepEqual(firstItems, secondItems) {
		t.Fatalf("repeated exports differed: first count=%d items=%+v second count=%d items=%+v", firstCount, firstItems, secondCount, secondItems)
	}

	firstRecord, err := provider.GetVerificationIntent(ctx, intent.ID)
	if err != nil {
		t.Fatalf("first GetVerificationIntent: %v", err)
	}
	secondRecord, err := provider.GetVerificationIntent(ctx, intent.ID)
	if err != nil {
		t.Fatalf("second GetVerificationIntent: %v", err)
	}
	if !reflect.DeepEqual(firstRecord, secondRecord) {
		t.Fatalf("repeated verification reads differed: first=%+v second=%+v", firstRecord, secondRecord)
	}
}

func TestProviderRegressionMalformedExportMetadataIsSkipped(t *testing.T) {
	provider := openTestProvider(t)
	ctx := context.Background()
	if _, err := provider.CreateProject(ctx, storage.CreateProjectInput{Name: "alpha"}); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	valid := seedProviderIntent(t, provider.SQLDB(), testIntent("valid-1", "2026-01-01T00:00:02Z", map[string]string{"project": "alpha"}))
	malformed := seedProviderIntent(t, provider.SQLDB(), testIntent("malformed-1", "2026-01-01T00:00:01Z", map[string]string{"project": "alpha"}))
	if _, err := provider.SQLDB().ExecContext(ctx, `UPDATE intents SET meta = ? WHERE id = ?`, `{not-json`, malformed.ID); err != nil {
		t.Fatalf("seed malformed metadata: %v", err)
	}

	items, count, err := provider.ListExportItems(ctx, storage.ExportQuery{Project: "alpha"})
	if err != nil {
		t.Fatalf("ListExportItems: %v", err)
	}
	if count != 1 || len(items) != 1 || items[0].Capture.ID != valid.ID {
		t.Fatalf("expected malformed metadata row to be skipped, got count=%d items=%+v", count, items)
	}
}

func TestProviderRegressionMalformedCheckpointArtifactIDsReturnDecodeError(t *testing.T) {
	provider := openTestProvider(t)
	ctx := context.Background()
	if _, err := provider.CreateProject(ctx, storage.CreateProjectInput{Name: "alpha"}); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if _, err := provider.SQLDB().ExecContext(ctx, `INSERT INTO checkpoints (hash, project, summary, created_at, artifact_ids, previous_checkpoint_id)
		VALUES ('bad-checkpoint', 'alpha', 'bad artifact ids', '2026-01-01T00:00:01Z', '{not-json', NULL)`); err != nil {
		t.Fatalf("seed malformed checkpoint: %v", err)
	}

	_, err := provider.ListCheckpoints(ctx, "alpha")
	if err == nil || !strings.Contains(err.Error(), "invalid character") {
		t.Fatalf("expected checkpoint artifact decode error, got %v", err)
	}
}

func testIntent(id, createdAt string, meta map[string]string) model.IntentRecord {
	return model.IntentRecord{
		ID:         id,
		CreatedAt:  createdAt,
		Author:     "Ada",
		SourceType: "cli",
		Title:      id,
		Prompt:     id + " prompt",
		Response:   id + " response",
		Meta:       mustJSON(meta),
	}
}

func mustJSON(values map[string]string) []byte {
	data, err := json.Marshal(values)
	if err != nil {
		panic(err)
	}
	return data
}
