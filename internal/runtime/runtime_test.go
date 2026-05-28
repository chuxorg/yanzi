package runtime

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chuxorg/yanzi/internal/api/handlers"
	"github.com/chuxorg/yanzi/internal/api/models"
	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/core/hash"
	"github.com/chuxorg/yanzi/internal/core/model"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
	"github.com/chuxorg/yanzi/internal/storage"
	_ "modernc.org/sqlite"
)

func TestRuntimeStartServeAndShutdown(t *testing.T) {
	workdir := t.TempDir()
	withRuntimeWorkdir(t, workdir)
	t.Setenv("HOME", workdir)
	t.Setenv(config.LocalDBPathEnvVar, filepath.Join(workdir, "yanzi.db"))

	db := openRuntimeDB(t, workdir)
	defer db.Close()

	seedRuntimeProject(t, db, "alpha")
	seedRuntimeCheckpoint(t, db, "alpha", "2026-01-01T00:00:10Z", "runtime checkpoint")
	seedRuntimeIntent(t, db, runtimeSeedIntent{
		ID:        "intent-1",
		CreatedAt: "2026-01-01T00:00:11Z",
		Project:   "alpha",
		Author:    "Ada",
		Prompt:    "Prompt body",
		Response:  "Response body",
		Meta:      map[string]string{"project": "alpha"},
	})
	writeRuntimeProjectBinding(t, workdir, "alpha")

	rt := New(Options{
		Addr:    "127.0.0.1:0",
		Version: "v0.0.0-test",
		Dependencies: handlers.Dependencies{
			LoadConfig: func() (config.Config, error) {
				return config.Config{Mode: config.ModeLocal}, nil
			},
			OpenProvider: func(context.Context, config.Config) (storage.Provider, error) {
				return &runtimeStubProvider{
					health: storage.Health{Provider: storage.ProviderSQLite, Status: storage.HealthReady},
					db:     db,
				}, nil
			},
		},
	})

	inst, err := rt.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		if err := inst.Shutdown(context.Background()); err != nil {
			t.Fatalf("Shutdown: %v", err)
		}
		if err := inst.Wait(); err != nil {
			t.Fatalf("Wait: %v", err)
		}
	}()

	health := waitForRuntimeJSON(t, "http://"+inst.Addr()+"/v0/health")
	var healthResp models.HealthResponse
	if err := json.Unmarshal(health, &healthResp); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if healthResp.Runtime == nil || healthResp.Runtime.Mode != "shared" || healthResp.Runtime.StartedAt == "" {
		t.Fatalf("unexpected runtime health: %+v", healthResp)
	}

	rehydrate := waitForRuntimeJSON(t, "http://"+inst.Addr()+"/v0/rehydrate")
	var rehydrateResp models.RehydrateResponse
	if err := json.Unmarshal(rehydrate, &rehydrateResp); err != nil {
		t.Fatalf("decode rehydrate: %v", err)
	}
	if rehydrateResp.Project != "alpha" || !rehydrateResp.HasCheckpoint || len(rehydrateResp.Intents) != 1 {
		t.Fatalf("unexpected rehydrate payload: %+v", rehydrateResp)
	}
}

func TestRuntimeStartFailsWhenPortUnavailable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	rt := New(Options{
		Addr:    ln.Addr().String(),
		Version: "v0.0.0-test",
		Dependencies: handlers.Dependencies{
			LoadConfig: func() (config.Config, error) {
				return config.Config{Mode: config.ModeLocal}, nil
			},
		},
	})

	if _, err := rt.Start(); err == nil {
		t.Fatal("expected start failure")
	}
}

func waitForRuntimeJSON(t *testing.T, url string) []byte {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err != nil {
			lastErr = err
			time.Sleep(50 * time.Millisecond)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status: %d", resp.StatusCode)
			_ = resp.Body.Close()
			time.Sleep(50 * time.Millisecond)
			continue
		}
		body := mustReadBody(t, resp)
		_ = resp.Body.Close()
		return body
	}
	t.Fatalf("endpoint never became ready: %v", lastErr)
	return nil
}

func mustReadBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return body
}

func withRuntimeWorkdir(t *testing.T, dir string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})
}

func openRuntimeDB(t *testing.T, dir string) *sql.DB {
	t.Helper()
	path := filepath.Join(dir, "yanzi.db")
	t.Setenv(config.LocalDBPathEnvVar, path)
	db, err := yanzilibrary.InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	return db
}

type runtimeSeedIntent struct {
	ID        string
	CreatedAt string
	Project   string
	Author    string
	Prompt    string
	Response  string
	Meta      map[string]string
}

func seedRuntimeProject(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	if _, err := yanzilibrary.CreateProject(name, ""); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
}

func seedRuntimeCheckpoint(t *testing.T, db *sql.DB, project, createdAt, summary string) {
	t.Helper()
	checkpoint := yanzilibrary.Checkpoint{
		Project:     project,
		Summary:     summary,
		CreatedAt:   createdAt,
		ArtifactIDs: []string{},
	}
	hashValue, err := yanzilibrary.HashCheckpoint(checkpoint)
	if err != nil {
		t.Fatalf("HashCheckpoint: %v", err)
	}
	artifactIDsJSON, err := json.Marshal([]string{})
	if err != nil {
		t.Fatalf("marshal artifact ids: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO checkpoints (hash, project, summary, created_at, artifact_ids, previous_checkpoint_id)
		VALUES (?, ?, ?, ?, ?, ?)`,
		hashValue,
		project,
		summary,
		createdAt,
		string(artifactIDsJSON),
		nil,
	); err != nil {
		t.Fatalf("seed checkpoint: %v", err)
	}
}

func seedRuntimeIntent(t *testing.T, db *sql.DB, input runtimeSeedIntent) {
	t.Helper()
	meta := input.Meta
	if meta == nil {
		meta = map[string]string{"project": input.Project}
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal meta: %v", err)
	}
	record := model.IntentRecord{
		ID:         input.ID,
		CreatedAt:  input.CreatedAt,
		Author:     input.Author,
		SourceType: "cli",
		Prompt:     input.Prompt,
		Response:   input.Response,
		Meta:       metaJSON,
	}
	record.Hash, err = hash.HashIntent(record)
	if err != nil {
		t.Fatalf("HashIntent: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO intents (id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ID,
		record.CreatedAt,
		record.Author,
		record.SourceType,
		nil,
		record.Prompt,
		record.Response,
		string(record.Meta),
		nil,
		record.Hash,
	); err != nil {
		t.Fatalf("seed intent: %v", err)
	}
}

func writeRuntimeProjectBinding(t *testing.T, dir, project string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".yanzi"), 0o700); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".yanzi", "project"), []byte(project+"\n"), 0o644); err != nil {
		t.Fatalf("write project binding: %v", err)
	}
}

type runtimeStubProvider struct {
	health storage.Health
	db     *sql.DB
}

func (p *runtimeStubProvider) Name() storage.ProviderName { return p.health.Provider }
func (p *runtimeStubProvider) Health(context.Context) storage.Health {
	return p.health
}
func (p *runtimeStubProvider) SQLDB() *sql.DB     { return p.db }
func (p *runtimeStubProvider) Close() error       { return nil }
func (p *runtimeStubProvider) Artifacts() bool    { return false }
func (p *runtimeStubProvider) Projects() bool     { return false }
func (p *runtimeStubProvider) Checkpoints() bool  { return false }
func (p *runtimeStubProvider) Verification() bool { return false }
func (p *runtimeStubProvider) ImportExport() bool { return false }
func (p *runtimeStubProvider) CreateArtifact(context.Context, storage.CreateArtifactInput) (storage.Artifact, error) {
	return storage.Artifact{}, nil
}
func (p *runtimeStubProvider) ListArtifacts(context.Context, storage.ArtifactQuery) ([]storage.Artifact, error) {
	return nil, nil
}
func (p *runtimeStubProvider) ListVisibleContextArtifacts(context.Context, storage.ContextArtifactQuery) ([]storage.Artifact, error) {
	return nil, nil
}
func (p *runtimeStubProvider) GetVisibleContextArtifact(context.Context, string, string) (storage.Artifact, error) {
	return storage.Artifact{}, nil
}
func (p *runtimeStubProvider) CreateProject(context.Context, storage.CreateProjectInput) (storage.Project, error) {
	return storage.Project{}, nil
}
func (p *runtimeStubProvider) ListProjects(context.Context) ([]storage.Project, error) {
	return nil, nil
}
func (p *runtimeStubProvider) ProjectExists(context.Context, string) (bool, error) {
	return false, nil
}
func (p *runtimeStubProvider) CreateCheckpoint(context.Context, storage.CreateCheckpointInput) (storage.Checkpoint, error) {
	return storage.Checkpoint{}, nil
}
func (p *runtimeStubProvider) ListCheckpoints(context.Context, string) ([]storage.Checkpoint, error) {
	return nil, nil
}
func (p *runtimeStubProvider) ListAllCheckpoints(context.Context) ([]storage.Checkpoint, error) {
	return nil, nil
}
func (p *runtimeStubProvider) GetVerificationIntent(context.Context, string) (storage.IntentRecord, error) {
	return storage.IntentRecord{}, nil
}
func (p *runtimeStubProvider) GetVerificationIntentByHash(context.Context, string) (storage.IntentRecord, error) {
	return storage.IntentRecord{}, nil
}
func (p *runtimeStubProvider) ListExportItems(context.Context, storage.ExportQuery) ([]storage.ExportItem, int, error) {
	return nil, 0, nil
}
