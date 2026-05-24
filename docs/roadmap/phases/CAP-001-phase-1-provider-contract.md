# CAP-001 Phase 1 - Provider Contract and Structural Extraction

## Scope

CAP-001 Phase 1 creates the internal storage provider seam while preserving current SQLite behavior.

This phase establishes:

- a small internal provider contract under `internal/storage`
- SQLite as the only implemented provider
- a provider registry that selects SQLite from existing configuration
- migration ownership inside the SQLite provider boundary
- compatibility coverage for existing project, artifact, checkpoint, rehydration, verification, and export storage paths

The phase is structural only. Existing CLI commands, flags, config files, schemas, migrations, exports, checkpoints, rehydration, and verification semantics remain unchanged.

## Non-Goals

This phase does not include:

- Postgres or any additional datastore provider
- REST API work
- runtime or daemon work
- federation
- MCP
- UI
- connector behavior
- export redesign
- persistence redesign
- config format changes
- user-visible workflow changes
- provider-specific optimizations

## Deliverables

- `internal/storage` provider contract and shared provider types
- `internal/storage/sqlite` SQLite provider boundary
- `internal/storage/registry` provider construction layer
- SQLite migration behavior preserved without schema changes
- provider and compatibility tests
- storage provider contract documentation with current implementation status

## Acceptance

Acceptance requires:

- SQLite remains the default and only provider
- existing `~/.yanzi/config.yaml` and `YANZI_DB_PATH` behavior is preserved
- existing migrations remain authoritative and apply automatically
- existing users can upgrade without manual migration
- existing CLI workflows produce the same user-visible behavior
- exports are unchanged
- checkpoints are unchanged
- rehydration is unchanged
- verification is unchanged
- `go test ./...` passes
- `go vet ./...` passes
- `go build ./cmd/yanzi` passes
- `make docs-build` passes
