package sqlite_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/core/hash"
	"github.com/chuxorg/yanzi/internal/core/model"
	"github.com/chuxorg/yanzi/internal/storage"
)

type exportReadProvider interface {
	ListExportItems(context.Context, storage.ExportQuery) ([]storage.ExportItem, int, error)
}

type verificationReadProvider interface {
	GetVerificationIntent(context.Context, string) (storage.IntentRecord, error)
	GetVerificationIntentByHash(context.Context, string) (storage.IntentRecord, error)
}

func TestExportProviderContractTimelineParity(t *testing.T) {
	provider := openTestProvider(t)
	ctx := context.Background()
	exportProvider, ok := any(provider).(exportReadProvider)
	if !ok {
		t.Fatalf("provider does not implement export read contract")
	}
	if _, err := provider.CreateProject(ctx, storage.CreateProjectInput{Name: "alpha"}); err != nil {
		t.Fatalf("CreateProject alpha: %v", err)
	}
	if _, err := provider.CreateProject(ctx, storage.CreateProjectInput{Name: "beta"}); err != nil {
		t.Fatalf("CreateProject beta: %v", err)
	}

	first := seedProviderIntent(t, provider.SQLDB(), model.IntentRecord{
		ID:         "alpha-first",
		CreatedAt:  "2026-01-01T00:00:01Z",
		Author:     "Ada",
		SourceType: "cli",
		Title:      "First capture",
		Prompt:     "prompt one",
		Response:   "response one",
		Meta:       rawMeta(t, map[string]string{"project": "alpha", "area": "auth"}),
	})
	checkpoint, err := provider.CreateCheckpoint(ctx, storage.CreateCheckpointInput{Project: "alpha", Summary: "checkpoint one", ArtifactIDs: []string{first.ID}})
	if err != nil {
		t.Fatalf("CreateCheckpoint: %v", err)
	}
	if _, err := provider.SQLDB().ExecContext(ctx, `UPDATE checkpoints SET created_at = ? WHERE hash = ?`, "2026-01-01T00:00:02Z", checkpoint.Hash); err != nil {
		t.Fatalf("fix checkpoint time: %v", err)
	}
	checkpoint.CreatedAt = "2026-01-01T00:00:02Z"
	metaEvent := seedProviderIntent(t, provider.SQLDB(), model.IntentRecord{
		ID:         "alpha-meta",
		CreatedAt:  "2026-01-01T00:00:03Z",
		Author:     "Yanzi",
		SourceType: "meta-command",
		Title:      "Meta event",
		Prompt:     "@yanzi pause",
		Response:   "true",
		Meta:       rawMeta(t, map[string]string{"project": "alpha"}),
	})
	seedProviderIntent(t, provider.SQLDB(), model.IntentRecord{
		ID:         "beta-capture",
		CreatedAt:  "2026-01-01T00:00:04Z",
		Author:     "Ben",
		SourceType: "cli",
		Title:      "Beta capture",
		Prompt:     "beta prompt",
		Response:   "beta response",
		Meta:       rawMeta(t, map[string]string{"project": "beta"}),
	})

	items, captureCount, err := exportProvider.ListExportItems(ctx, storage.ExportQuery{Project: "alpha"})
	if err != nil {
		t.Fatalf("ListExportItems: %v", err)
	}
	if captureCount != 1 {
		t.Fatalf("expected one alpha capture, got %d", captureCount)
	}
	if len(items) != 3 {
		t.Fatalf("expected capture, checkpoint, and meta event, got %+v", items)
	}
	if items[0].Kind != storage.ExportItemCapture || items[0].Capture.ID != first.ID || items[0].Capture.Hash != first.Hash {
		t.Fatalf("unexpected first export item: %+v", items[0])
	}
	if items[0].Capture.Metadata["area"] != "auth" || items[0].Capture.Metadata["project"] != "alpha" {
		t.Fatalf("metadata was not preserved: %+v", items[0].Capture.Metadata)
	}
	if items[1].Kind != storage.ExportItemCheckpoint || items[1].Checkpoint.Hash != checkpoint.Hash || items[1].Checkpoint.Summary != "checkpoint one" {
		t.Fatalf("unexpected checkpoint item: %+v", items[1])
	}
	if items[2].Kind != storage.ExportItemMeta || items[2].Meta.Command != metaEvent.Prompt || items[2].Meta.Value != metaEvent.Response {
		t.Fatalf("unexpected meta item: %+v", items[2])
	}

	filtered, filteredCount, err := exportProvider.ListExportItems(ctx, storage.ExportQuery{Project: "alpha", MetaFilters: map[string]string{"area": "auth"}})
	if err != nil {
		t.Fatalf("ListExportItems filtered: %v", err)
	}
	if filteredCount != 1 || len(filtered) != 1 || filtered[0].Capture.ID != first.ID {
		t.Fatalf("expected only matching capture without checkpoints/meta events, got count=%d items=%+v", filteredCount, filtered)
	}
}

func TestExportProviderContractDeletedVisibility(t *testing.T) {
	provider := openTestProvider(t)
	ctx := context.Background()
	exportProvider, ok := any(provider).(exportReadProvider)
	if !ok {
		t.Fatalf("provider does not implement export read contract")
	}
	if _, err := provider.CreateProject(ctx, storage.CreateProjectInput{Name: "alpha"}); err != nil {
		t.Fatalf("CreateProject alpha: %v", err)
	}
	visible := seedProviderIntent(t, provider.SQLDB(), model.IntentRecord{
		ID:         "visible-capture",
		CreatedAt:  "2026-01-01T00:00:01Z",
		Author:     "Ada",
		SourceType: "cli",
		Title:      "Visible",
		Prompt:     "visible prompt",
		Response:   "visible response",
		Meta:       rawMeta(t, map[string]string{"project": "alpha"}),
	})
	deleted := seedProviderIntent(t, provider.SQLDB(), model.IntentRecord{
		ID:         "deleted-capture",
		CreatedAt:  "2026-01-01T00:00:02Z",
		Author:     "Ada",
		SourceType: "cli",
		Title:      "Deleted",
		Prompt:     "deleted prompt",
		Response:   "deleted response",
		Meta:       rawMeta(t, map[string]string{"project": "alpha"}),
	})
	if _, err := provider.SQLDB().ExecContext(ctx, `UPDATE intents SET metadata = ? WHERE id = ?`, `{"deleted":"true","project":"alpha"}`, deleted.ID); err != nil {
		t.Fatalf("mark deleted: %v", err)
	}

	items, captureCount, err := exportProvider.ListExportItems(ctx, storage.ExportQuery{Project: "alpha"})
	if err != nil {
		t.Fatalf("ListExportItems: %v", err)
	}
	if captureCount != 1 || len(items) != 1 || items[0].Capture.ID != visible.ID {
		t.Fatalf("expected deleted capture hidden by default, got count=%d items=%+v", captureCount, items)
	}

	withDeleted, withDeletedCount, err := exportProvider.ListExportItems(ctx, storage.ExportQuery{Project: "alpha", IncludeDeleted: true})
	if err != nil {
		t.Fatalf("ListExportItems include deleted: %v", err)
	}
	if withDeletedCount != 2 || len(withDeleted) != 2 || withDeleted[1].Capture.ID != deleted.ID {
		t.Fatalf("expected deleted capture when requested, got count=%d items=%+v", withDeletedCount, withDeleted)
	}
	if withDeleted[1].Capture.Metadata["deleted"] != "true" {
		t.Fatalf("deleted metadata was not preserved: %+v", withDeleted[1].Capture.Metadata)
	}
}

func TestVerificationProviderContractIntentLookupAndChain(t *testing.T) {
	provider := openTestProvider(t)
	ctx := context.Background()
	verifyProvider, ok := any(provider).(verificationReadProvider)
	if !ok {
		t.Fatalf("provider does not implement verification read contract")
	}
	if _, err := provider.CreateProject(ctx, storage.CreateProjectInput{Name: "alpha"}); err != nil {
		t.Fatalf("CreateProject alpha: %v", err)
	}

	first := seedProviderIntent(t, provider.SQLDB(), model.IntentRecord{
		ID:         "chain-first",
		CreatedAt:  "2026-01-01T00:00:01Z",
		Author:     "Ada",
		SourceType: "cli",
		Title:      "First",
		Prompt:     "first prompt",
		Response:   "first response",
		Meta:       rawMeta(t, map[string]string{"project": "alpha"}),
	})
	second := seedProviderIntent(t, provider.SQLDB(), model.IntentRecord{
		ID:         "chain-second",
		CreatedAt:  "2026-01-01T00:00:02Z",
		Author:     "Ada",
		SourceType: "cli",
		Title:      "Second",
		Prompt:     "second prompt",
		Response:   "second response",
		Meta:       rawMeta(t, map[string]string{"project": "alpha"}),
		PrevHash:   first.Hash,
	})

	got, err := verifyProvider.GetVerificationIntent(ctx, second.ID)
	if err != nil {
		t.Fatalf("GetVerificationIntent: %v", err)
	}
	if got.ID != second.ID || got.PrevHash != first.Hash || string(got.Meta) != string(second.Meta) {
		t.Fatalf("unexpected verification intent: %+v", got)
	}
	computed, err := hash.HashIntent(model.IntentRecord{
		ID:         got.ID,
		CreatedAt:  got.CreatedAt,
		Author:     got.Author,
		SourceType: got.SourceType,
		Title:      got.Title,
		Prompt:     got.Prompt,
		Response:   got.Response,
		Meta:       got.Meta,
		PrevHash:   got.PrevHash,
	})
	if err != nil {
		t.Fatalf("HashIntent: %v", err)
	}
	if computed != got.Hash {
		t.Fatalf("expected hash parity, got computed=%s stored=%s", computed, got.Hash)
	}

	previous, err := verifyProvider.GetVerificationIntentByHash(ctx, got.PrevHash)
	if err != nil {
		t.Fatalf("GetVerificationIntentByHash: %v", err)
	}
	if previous.ID != first.ID || previous.Hash != first.Hash {
		t.Fatalf("unexpected previous intent: %+v", previous)
	}

	_, err = verifyProvider.GetVerificationIntent(ctx, "missing")
	if !errors.Is(err, storage.ErrNotFound) || !strings.Contains(err.Error(), "intent not found: missing") {
		t.Fatalf("expected missing intent not-found error, got %v", err)
	}
	_, err = verifyProvider.GetVerificationIntentByHash(ctx, "missing-hash")
	if !errors.Is(err, storage.ErrNotFound) || !strings.Contains(err.Error(), "intent hash not found: missing-hash") {
		t.Fatalf("expected missing hash not-found error, got %v", err)
	}
}

func rawMeta(t *testing.T, values map[string]string) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(values)
	if err != nil {
		t.Fatalf("encode meta: %v", err)
	}
	return data
}

func seedProviderIntent(t *testing.T, db *sql.DB, record model.IntentRecord) model.IntentRecord {
	t.Helper()
	if record.Hash == "" {
		sum, err := hash.HashIntent(record)
		if err != nil {
			t.Fatalf("hash intent %s: %v", record.ID, err)
		}
		record.Hash = sum
	}
	var title any
	if record.Title != "" {
		title = record.Title
	}
	var meta any
	if len(record.Meta) > 0 {
		meta = string(record.Meta)
	}
	var prevHash any
	if record.PrevHash != "" {
		prevHash = record.PrevHash
	}
	if _, err := db.ExecContext(
		context.Background(),
		`INSERT INTO intents (id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash, class, type, content, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ID,
		record.CreatedAt,
		record.Author,
		record.SourceType,
		title,
		record.Prompt,
		record.Response,
		meta,
		prevHash,
		record.Hash,
		"intent",
		"prompt",
		record.Prompt,
		meta,
	); err != nil {
		t.Fatalf("seed intent %s: %v", record.ID, err)
	}
	return record
}
