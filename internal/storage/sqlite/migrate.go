package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	sqlitemigrate "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// migrationsTable is the golang-migrate tracking table name. A distinct name
// avoids collision with the legacy schema_migrations table used by the old
// custom runner so that both fresh installs and existing databases work.
const migrationsTable = "_yanzi_migrations"

// RunMigrations applies all pending up migrations using golang-migrate with
// the embedded iofs source and the modernc SQLite database driver.
//
// For databases previously migrated by the old custom runner, it detects the
// effective schema version via table/column introspection and bootstraps
// golang-migrate to that version before applying any new ones.
//
// NOTE: m.Close() is intentionally not called here. WithInstance passes in an
// externally-owned *sql.DB; calling m.Close() would close that connection.
func RunMigrations(db *sql.DB) error {
	d, err := sqlitemigrate.WithInstance(db, &sqlitemigrate.Config{
		MigrationsTable: migrationsTable,
	})
	if err != nil {
		return fmt.Errorf("sqlite migrate driver: %w", err)
	}

	src, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("iofs migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "sqlite", d)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	// Detect the schema version via introspection. This handles:
	//   - Fresh databases: version 0, run all migrations.
	//   - Databases from the old custom runner: version > 0, skip already-applied
	//     migrations. Schema introspection is used instead of the old
	//     schema_migrations table because it is reliable even when the table is
	//     absent or when golang-migrate left a dirty flag after a failed attempt.
	version := introspectSchemaVersion(db)
	if version > 0 {
		if err := m.Force(version); err != nil {
			return fmt.Errorf("bootstrap schema version %d: %w", version, err)
		}
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// introspectSchemaVersion returns the highest migration version whose effects
// are present in the database, based on schema introspection. Returns 0 when
// no migrations have been applied. This is used to bootstrap golang-migrate on
// databases previously migrated by the old custom runner, and to recover from
// a dirty migration state.
func introspectSchemaVersion(db *sql.DB) int {
	// Migration 4: class/type/content/metadata columns on intents.
	if columnExists(db, "intents", "class") {
		return 4
	}
	// Migration 3: checkpoints table.
	if tableExists(db, "checkpoints") {
		return 3
	}
	// Migration 2: projects table.
	if tableExists(db, "projects") {
		return 2
	}
	// Migration 1: intents table.
	if tableExists(db, "intents") {
		return 1
	}
	return 0
}

func tableExists(db *sql.DB, name string) bool {
	var count int
	_ = db.QueryRow(
		`SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name=?`, name,
	).Scan(&count)
	return count > 0
}

func columnExists(db *sql.DB, table, column string) bool {
	var count int
	_ = db.QueryRow(
		`SELECT COUNT(1) FROM pragma_table_info(?) WHERE name=?`, table, column,
	).Scan(&count)
	return count > 0
}
