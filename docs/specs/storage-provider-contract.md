# Storage Provider Contract

## Purpose

The storage provider contract defines the internal boundary for persisting Yanzi operational context while preserving existing local-first behavior.

CAP-001 Phase 1 introduces the seam only. The contract exists to make storage responsibilities explicit without changing CLI behavior, schema semantics, export formats, checkpoints, rehydration, or verification.

## Current Implementation Status

Current provider: SQLite only.

SQLite remains the default provider and the only implemented provider. No provider configuration key is active. Existing configuration continues to use the current local SQLite path rules:

1. `YANZI_DB_PATH`
2. `db_path` from `~/.yanzi/config.yaml`
3. default `~/.yanzi/yanzi.db`

No Postgres, object storage, runtime-hosted, REST-backed, MCP-backed, federation-backed, connector-backed, or UI-specific provider exists in this phase.

## Contract Boundary

The provider boundary is internal to the Go codebase. It is not a user-facing API.

The current Phase 1 provider contract exposes:

- provider identity
- provider health
- artifact capability
- project capability
- checkpoint capability
- verification capability
- import/export capability
- current SQLite database handle for backward-compatible call sites

The database handle is intentionally retained during Phase 1 to avoid broad persistence rewrites. Future phases may move individual operations behind narrower provider methods after compatibility coverage is in place.

## Required Current Capabilities

A provider conforming to the current Yanzi behavior must preserve the following capabilities.

### Artifacts

- create intent artifacts
- create context artifacts
- list project-scoped artifacts
- list all-project artifacts where current commands require it
- preserve artifact metadata, class, type, title, content, scope, project, and created timestamp semantics

### Projects

- create projects
- list projects
- detect missing projects consistently
- preserve project hash and creation ordering behavior

### Checkpoints

- create checkpoints
- list checkpoints
- preserve checkpoint hash generation inputs
- preserve previous-checkpoint linkage
- preserve checkpoint ordering used by CLI and rehydration

### Verification

- retrieve stored intent records for digest verification
- preserve hash preimage compatibility
- preserve verification result semantics

### Import and Export

- preserve existing deterministic export inputs and ordering
- preserve artifact directory export behavior
- preserve metadata filtering behavior
- avoid changing generated export formats

### Health

- expose internal readiness for the provider
- report unavailable state when a provider handle cannot service requests
- do not expose provider health through a CLI surface in this phase

## SQLite Provider Requirements

The SQLite provider must preserve:

- current database file creation behavior
- current directory permissions
- current database file permissions
- current SQLite pragmas: WAL, foreign keys, busy timeout
- current `schema_version` behavior
- current embedded SQL migrations
- current schema and indexes
- automatic migration on open
- existing `InitDB` and `InitDBAtPath` call behavior

## Registry Requirements

The provider registry must:

- return SQLite for current local configuration
- reject unsupported provider names internally
- avoid changing config file schema
- avoid changing CLI selection behavior
- avoid silently falling back from a configured future provider to SQLite

Because no future provider config is active in Phase 1, existing users continue to receive SQLite.

## Non-Goals

This contract does not define:

- a distributed datastore design
- a network protocol
- a runtime API
- a REST API
- MCP behavior
- federation behavior
- UI behavior
- connector behavior
- Postgres behavior
- object storage behavior
- new query semantics
- new user workflows

## Compatibility Guarantees

CAP-001 Phase 1 must preserve:

- zero CLI behavior changes
- backward-compatible local config behavior
- existing SQLite schema
- existing migrations
- existing export output contracts
- existing checkpoint behavior
- existing rehydration behavior
- existing verification behavior
- existing data upgrade without manual migration

## CAP-001 Phase 2A Implementation Status

Project and checkpoint operations are now routed through the SQLite provider.

Migrated operation groups:

- project creation
- project listing
- project existence checks used by current storage paths
- checkpoint creation
- project-scoped checkpoint listing
- all-project checkpoint listing

Unmigrated operation groups remain intentionally outside provider methods in this phase:

- capture and artifact operations
- export operations
- verification operations
- rehydration query operations

SQLite remains the only provider. No config keys, schemas, migrations, command outputs, or user workflows changed.

## Implementation Notes

The current implementation lives in:

- `internal/storage/provider.go`
- `internal/storage/types.go`
- `internal/storage/errors.go`
- `internal/storage/registry/registry.go`
- `internal/storage/sqlite/provider.go`

The existing library database entry points remain available and backward compatible:

- `internal/library.InitDB`
- `internal/library.InitDBAtPath`
- `internal/library.Initialize`

These entry points now resolve SQLite through the provider boundary while preserving their existing return types and caller behavior.
