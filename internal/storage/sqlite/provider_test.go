package sqlite_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
	"github.com/chuxorg/yanzi/internal/storage"
	"github.com/chuxorg/yanzi/internal/storage/sqlite"
	_ "modernc.org/sqlite"
)

func TestProviderSatisfiesContractAndReportsHealth(t *testing.T) {
	path := filepath.Join(t.TempDir(), "yanzi.db")
	provider, initialized, err := sqlite.Open(context.Background(), path, yanzilibrary.MigrationsFS())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer provider.Close()

	var _ storage.Provider = provider
	if !initialized {
		t.Fatalf("expected first provider open to initialize schema")
	}
	if provider.Name() != storage.ProviderSQLite {
		t.Fatalf("expected sqlite provider name, got %q", provider.Name())
	}
	if provider.SQLDB() == nil {
		t.Fatalf("expected SQLDB handle")
	}
	if !provider.Artifacts() || !provider.Projects() || !provider.Checkpoints() || !provider.Verification() || !provider.ImportExport() {
		t.Fatalf("expected provider to advertise current local capabilities")
	}

	health := provider.Health(context.Background())
	if health.Status != storage.HealthReady {
		t.Fatalf("expected ready health, got %+v", health)
	}
	if health.Path != path {
		t.Fatalf("expected health path %q, got %q", path, health.Path)
	}
	if health.MigrationState != storage.HealthMigrationApplied {
		t.Fatalf("expected applied migration health, got %+v", health)
	}
	if !health.Writable {
		t.Fatalf("expected writable health, got %+v", health)
	}
	if err := provider.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	closedHealth := provider.Health(context.Background())
	if closedHealth.Status != storage.HealthUnavailable {
		t.Fatalf("expected unavailable health after close, got %+v", closedHealth)
	}
}

func TestProviderAppliesArtifactColumnMigrationToExistingDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	createLegacyDatabase(t, path)

	provider, initialized, err := sqlite.Open(context.Background(), path, yanzilibrary.MigrationsFS())
	if err != nil {
		t.Fatalf("Open legacy db: %v", err)
	}
	defer provider.Close()
	if initialized {
		t.Fatalf("expected existing schema_version to avoid first-time initialization")
	}

	db := provider.SQLDB()
	for _, column := range []string{"class", "type", "content", "metadata"} {
		var count int
		if err := db.QueryRow(`SELECT COUNT(1) FROM pragma_table_info('intents') WHERE name = ?`, column).Scan(&count); err != nil {
			t.Fatalf("check column %s: %v", column, err)
		}
		if count != 1 {
			t.Fatalf("expected migrated column %s", column)
		}
	}

	var class, artifactType, content, metadata string
	if err := db.QueryRow(`SELECT class, type, content, metadata FROM intents WHERE id = 'legacy-1'`).Scan(&class, &artifactType, &content, &metadata); err != nil {
		t.Fatalf("read migrated legacy row: %v", err)
	}
	if class != "intent" || artifactType != "prompt" || content != "legacy prompt" || metadata != `{"project":"alpha"}` {
		t.Fatalf("unexpected migrated row: class=%q type=%q content=%q metadata=%q", class, artifactType, content, metadata)
	}
}

func TestProviderHealthReportsMissingMigrationState(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "empty.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	provider := sqlite.FromDB(db)
	health := provider.Health(context.Background())
	if health.Status != storage.HealthUnavailable {
		t.Fatalf("expected unavailable health for unmigrated db, got %+v", health)
	}
	if health.MigrationState != storage.HealthMigrationMissing {
		t.Fatalf("expected missing migration state, got %+v", health)
	}
}

func createLegacyDatabase(t *testing.T, path string) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	defer db.Close()

	statements := []string{
		`CREATE TABLE schema_version (version INTEGER NOT NULL, applied_at TIMESTAMP NOT NULL);`,
		`INSERT INTO schema_version (version, applied_at) VALUES (1, '2026-01-01T00:00:00Z');`,
		`CREATE TABLE schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT NOT NULL);`,
		`INSERT INTO schema_migrations (version, applied_at) VALUES ('0001_create_intent_table.sql', '2026-01-01T00:00:00Z');`,
		`INSERT INTO schema_migrations (version, applied_at) VALUES ('0002_create_projects_table.sql', '2026-01-01T00:00:00Z');`,
		`INSERT INTO schema_migrations (version, applied_at) VALUES ('0003_create_checkpoints_table.sql', '2026-01-01T00:00:00Z');`,
		`CREATE TABLE intents (id TEXT PRIMARY KEY, created_at TEXT NOT NULL, author TEXT NOT NULL, source_type TEXT NOT NULL, title TEXT, prompt TEXT NOT NULL, response TEXT NOT NULL, meta TEXT, prev_hash TEXT, hash TEXT NOT NULL);`,
		`CREATE TABLE projects (name TEXT PRIMARY KEY, description TEXT, created_at TEXT NOT NULL, prev_hash TEXT, hash TEXT NOT NULL);`,
		`CREATE TABLE checkpoints (hash TEXT PRIMARY KEY, project TEXT NOT NULL, summary TEXT NOT NULL, created_at TEXT NOT NULL, artifact_ids TEXT NOT NULL, previous_checkpoint_id TEXT);`,
		`INSERT INTO intents (id, created_at, author, source_type, title, prompt, response, meta, prev_hash, hash) VALUES ('legacy-1', '2026-01-01T00:00:00Z', 'yanzi', 'capture', 'Legacy', 'legacy prompt', 'legacy response', '{"project":"alpha"}', NULL, 'hash-1');`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("exec legacy statement %q: %v", statement, err)
		}
	}
}

func openTestProvider(t *testing.T) *sqlite.Provider {
	t.Helper()
	provider, _, err := sqlite.Open(context.Background(), filepath.Join(t.TempDir(), "yanzi.db"), yanzilibrary.MigrationsFS())
	if err != nil {
		t.Fatalf("Open provider: %v", err)
	}
	t.Cleanup(func() { _ = provider.Close() })
	return provider
}
