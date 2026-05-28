# CAP-001 Phase 2A - Project and Checkpoint Provider Migration

## Scope

CAP-001 Phase 2A migrates the first real operation groups behind the storage provider abstraction:

- project operations
- checkpoint operations

SQLite remains the behavioral reference. Existing CLI output, local configuration, database paths, schema, migrations, checkpoint hashing, checkpoint ordering, and project ordering are preserved.

## Non-Goals

This phase does not include:

- capture or artifact operation migration
- export migration
- verification migration
- rehydration migration except preserving checkpoint compatibility through existing database state
- Postgres or any other datastore provider
- REST
- runtime or daemon work
- MCP
- federation
- UI
- connector behavior
- config changes
- schema changes
- CLI behavior changes

## Deliverables

- provider contract coverage for current project behavior
- provider contract coverage for current checkpoint behavior
- project provider methods implemented by SQLite
- checkpoint provider methods implemented by SQLite
- existing project command execution routed through provider-backed library operations
- existing checkpoint library and command execution routed through provider methods
- storage provider contract documentation updated with Phase 2A status

## Acceptance Criteria

Acceptance requires:

- `yanzi project create` behavior is unchanged
- `yanzi project list` behavior is unchanged
- `yanzi project use` and `yanzi project current` behavior is unchanged
- duplicate project errors remain compatible
- missing project behavior remains compatible
- `yanzi checkpoint create` behavior is unchanged
- `yanzi checkpoint list` behavior is unchanged
- `yanzi checkpoint list --all-projects` ordering is unchanged
- checkpoint project association is preserved
- checkpoint summaries are preserved
- existing SQLite databases remain compatible
- no config changes are introduced
- no schema changes are introduced
- provider tests pass
- SQLite parity is preserved

## Validation Performed

Required validation for this phase:

- `go test ./...`
- `go vet ./...`
- `go build ./cmd/yanzi`
- `make docs-build`
- representative CLI workflow validation for project and checkpoint commands

## Preserved Behavior Notes

Checkpoint creation with nil artifact IDs returns a checkpoint with nil `ArtifactIDs` while storing an empty JSON list in SQLite. This behavior is preserved for compatibility.
