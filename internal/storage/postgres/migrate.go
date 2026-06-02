package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// RunMigrations applies all pending up migrations to the Postgres database
// using golang-migrate with the embedded iofs source.
//
// NOTE: m.Close() is intentionally not called because WithInstance passes in
// an externally-owned *sql.DB; calling m.Close() would close that connection.
func RunMigrations(db *sql.DB) error {
	d, err := pgmigrate.WithInstance(db, &pgmigrate.Config{})
	if err != nil {
		return fmt.Errorf("postgres migrate driver: %w", err)
	}

	src, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("iofs migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", d)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
