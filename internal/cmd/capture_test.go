package cmd

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/core/hash"
	"github.com/chuxorg/yanzi/internal/core/model"
	_ "modernc.org/sqlite"
)

func TestKVPairsToJSONLastValueWins(t *testing.T) {
	pairs := kvPairs{"area=auth", "area=billing", "tags=migration,security"}

	raw, err := pairs.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}

	var meta map[string]string
	if err := json.Unmarshal(raw, &meta); err != nil {
		t.Fatalf("unmarshal meta: %v", err)
	}
	if meta["area"] != "billing" {
		t.Fatalf("expected last value to win for duplicate key, got %q", meta["area"])
	}
	if meta["tags"] != "migration,security" {
		t.Fatalf("unexpected tags value: %q", meta["tags"])
	}
}

func TestKVPairsToJSONMalformedArgument(t *testing.T) {
	pairs := kvPairs{"missing-separator"}

	_, err := pairs.ToJSON()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "expected key=value") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCaptureStoresProfileMetadata(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	if err := RunCapture([]string{
		"--author", "Ada",
		"--prompt", "Hello",
		"--response", "World",
		"--profile", "engineer",
		"--meta", "area=auth",
	}); err != nil {
		t.Fatalf("RunCapture: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var metaText string
	if err := db.QueryRow(`SELECT meta FROM intents ORDER BY rowid DESC LIMIT 1`).Scan(&metaText); err != nil {
		t.Fatalf("query meta: %v", err)
	}

	var meta map[string]string
	if err := json.Unmarshal([]byte(metaText), &meta); err != nil {
		t.Fatalf("decode meta: %v", err)
	}
	if meta["profile"] != "engineer" {
		t.Fatalf("expected profile metadata, got %#v", meta)
	}
	if meta["area"] != "auth" {
		t.Fatalf("expected existing metadata to remain, got %#v", meta)
	}
	if meta["type"] == "context" {
		t.Fatalf("did not expect generic capture to force context metadata: %#v", meta)
	}
}

func TestRunCapturePersistsCurrentWriteSemantics(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	output, err := captureStdout(func() error {
		return RunCapture([]string{
			"--author", "Ada",
			"--source", "agent",
			"--title", "Boundary audit",
			"--prompt", "What changed?",
			"--response", "Captured current write behavior.",
			"--prev-hash", "previous-hash",
			"--meta", "area=capture",
		})
	})
	if err != nil {
		t.Fatalf("RunCapture: %v", err)
	}
	if !strings.Contains(output, "id: ") || !strings.Contains(output, "hash: ") {
		t.Fatalf("unexpected capture output: %q", output)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var record model.IntentRecord
	var metaText string
	var metadataText string
	var class string
	var artifactType string
	var content string
	if err := db.QueryRow(
		`SELECT id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash, class, type, content, metadata
		FROM intents ORDER BY rowid DESC LIMIT 1`,
	).Scan(
		&record.ID,
		&record.CreatedAt,
		&record.Author,
		&record.SourceType,
		&record.Title,
		&record.Prompt,
		&record.Response,
		&metaText,
		&record.PrevHash,
		&record.Hash,
		&class,
		&artifactType,
		&content,
		&metadataText,
	); err != nil {
		t.Fatalf("query capture: %v", err)
	}
	record.Meta = json.RawMessage(metaText)

	if record.Author != "Ada" || record.SourceType != "agent" || record.Title != "Boundary audit" {
		t.Fatalf("unexpected record identity fields: %+v", record)
	}
	if record.Prompt != "What changed?" || record.Response != "Captured current write behavior." {
		t.Fatalf("unexpected record payload: %+v", record)
	}
	if record.PrevHash != "previous-hash" {
		t.Fatalf("expected prev hash to persist, got %q", record.PrevHash)
	}
	if class != "intent" || artifactType != "prompt" || content != record.Prompt {
		t.Fatalf("unexpected artifact compatibility columns: class=%q type=%q content=%q", class, artifactType, content)
	}
	if metadataText != metaText {
		t.Fatalf("expected metadata column to mirror meta column, metadata=%q meta=%q", metadataText, metaText)
	}

	var meta map[string]string
	if err := json.Unmarshal([]byte(metaText), &meta); err != nil {
		t.Fatalf("decode meta: %v", err)
	}
	if meta["area"] != "capture" || meta["project"] != "alpha" {
		t.Fatalf("unexpected metadata: %#v", meta)
	}

	computedHash, err := hash.HashIntent(record)
	if err != nil {
		t.Fatalf("HashIntent: %v", err)
	}
	if computedHash != record.Hash {
		t.Fatalf("expected deterministic hash %q, got %q", record.Hash, computedHash)
	}

	stateDir, err := config.StateDir()
	if err != nil {
		t.Fatalf("StateDir: %v", err)
	}
	lastHash, err := os.ReadFile(filepath.Join(stateDir, "last_hash"))
	if err != nil {
		t.Fatalf("read last_hash: %v", err)
	}
	if strings.TrimSpace(string(lastHash)) != record.Hash {
		t.Fatalf("expected last_hash %q, got %q", record.Hash, strings.TrimSpace(string(lastHash)))
	}
}

func TestRunCaptureAllowsDuplicatePayloadsAsDistinctRecords(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	args := []string{"--author", "Ada", "--prompt", "same prompt", "--response", "same response"}
	if _, err := captureStdout(func() error { return RunCapture(args) }); err != nil {
		t.Fatalf("first RunCapture: %v", err)
	}
	if _, err := captureStdout(func() error { return RunCapture(args) }); err != nil {
		t.Fatalf("second RunCapture: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, hash FROM intents WHERE prompt = ? AND response = ? ORDER BY rowid ASC`, "same prompt", "same response")
	if err != nil {
		t.Fatalf("query captures: %v", err)
	}
	defer rows.Close()

	var ids []string
	var hashes []string
	for rows.Next() {
		var id string
		var hashValue string
		if err := rows.Scan(&id, &hashValue); err != nil {
			t.Fatalf("scan capture: %v", err)
		}
		ids = append(ids, id)
		hashes = append(hashes, hashValue)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected duplicate payloads to create two records, got ids=%v", ids)
	}
	if ids[0] == ids[1] || hashes[0] == hashes[1] {
		t.Fatalf("expected distinct ids and hashes, ids=%v hashes=%v", ids, hashes)
	}
}
