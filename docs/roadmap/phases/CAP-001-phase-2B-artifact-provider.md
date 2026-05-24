# CAP-001 Phase 2B - Artifact Provider Migration

## Scope

CAP-001 Phase 2B migrates existing artifact operation groups behind the storage provider abstraction:

- artifact creation
- artifact read/list/query behavior
- artifact context visibility behavior
- existing deleted/tombstone visibility behavior for artifact lists and context show resolution

SQLite remains the behavioral reference. Existing CLI output, local configuration, database paths, schema, migrations, artifact metadata handling, project association, source/author values, and ordering are preserved.

## Non-Goals

This phase does not include:

- export migration
- verification migration
- rehydration migration
- capture intent migration outside current artifact operations
- Postgres or any other datastore provider
- REST
- runtime or daemon work
- MCP
- federation
- UI
- connector behavior
- provider configuration
- config changes
- schema changes
- CLI behavior changes
- provider-specific optimizations

## Deliverables

- provider contract coverage for current artifact behavior
- provider contract coverage for current context visibility behavior
- provider contract coverage for deleted artifact list/show compatibility
- SQLite artifact provider methods for current create/list/visible-context operations
- existing artifact library execution routed through provider methods
- focused CLI compatibility coverage for intent and context artifact workflows
- storage provider contract documentation updated with Phase 2B status

## Acceptance Criteria

Acceptance requires:

- `yanzi intent add` behavior is unchanged
- `yanzi intent list` behavior is unchanged
- `yanzi context add` behavior is unchanged
- `yanzi context list` behavior is unchanged
- `yanzi context show` behavior is unchanged
- artifact metadata is preserved
- artifact project association is preserved
- context global/project visibility is preserved
- deleted artifacts remain hidden from normal lists
- deleted context artifacts remain resolvable by current `context show` semantics
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
- representative CLI workflow validation for intent and context artifact commands
- checkpoint, export, and verification regression validation without migrating those operation groups

## Preserved Behavior Notes

Deleted artifact state remains encoded in the existing artifact system metadata stored in the `intents.meta` column. Normal artifact and context lists hide deleted artifacts unless explicitly included by internal callers.

Current context show resolution includes deleted context artifacts when resolving by full ID or unique prefix. This behavior is preserved for compatibility and documented as existing behavior rather than new semantics.
