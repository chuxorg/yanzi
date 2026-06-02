# Storage Providers

## Overview

Yanzi supports pluggable storage providers. The active provider is configured
via `config.yaml` or environment variables. Environment variables take
precedence over config file values.

## SQLite (default)

No configuration required for basic use. Data is stored at
`~/.yanzi/yanzi.db` by default.

**Override the database path:**

```bash
# Environment variable (highest precedence)
export YANZI_DB_PATH=/path/to/yanzi.db

# Or via config.yaml
storage:
  sqlite:
    path: /path/to/yanzi.db
```

Migrations run automatically on startup. The first open of a new database
applies all four schema migrations and returns the provider ready for use.

## Postgres

Requires a running Postgres instance (version 13 or later recommended).

**Configure via environment variables (recommended for production):**

```bash
export YANZI_STORAGE_PROVIDER=postgres
export YANZI_POSTGRES_DSN=postgres://user:pass@host:5432/dbname?sslmode=disable
```

**Or via config.yaml:**

```yaml
storage:
  provider: postgres
  postgres:
    dsn: postgres://user:pass@host:5432/dbname?sslmode=disable
    max_open_conns: 25
    max_idle_conns: 5
    conn_max_lifetime: 300   # seconds
```

**Connection pool defaults:**

| Setting           | Default | Environment override        |
|-------------------|---------|-----------------------------|
| max_open_conns    | 25      | `YANZI_POSTGRES_MAX_CONNS`  |
| max_idle_conns    | 5       | —                           |
| conn_max_lifetime | 300s    | —                           |

Migrations run automatically on first startup against a new database.

**Validation:** Yanzi fails at startup with a clear error if the postgres
provider is selected but `DSN` is empty:

```
postgres provider requires YANZI_POSTGRES_DSN or storage.postgres.dsn in config
```

## Provider parity

Both providers produce identical output for the same logical data.

- Timestamps are stored as RFC3339Nano TEXT strings in both providers.
- JSON fields (`meta`, `metadata`, `artifact_ids`) are stored as TEXT with
  marshaling and unmarshaling in the Go layer.
- Artifact hashes use the same SHA-256 preimage format.
- Checkpoint hashes use the same canonical JSON preimage.

Hash chain integrity is maintained across providers. A corpus created with
SQLite can be migrated to Postgres without hash invalidation, provided the
raw field values are copied verbatim.

## Running Postgres tests

Postgres tests skip automatically when no DSN is set. To opt in:

```bash
export YANZI_TEST_POSTGRES_DSN=postgres://user:pass@localhost:5432/yanzi_test?sslmode=disable
go test ./internal/storage/postgres/...
```

The test suite creates and cleans up its own rows within the target database.
Use a dedicated test database to avoid data loss.
