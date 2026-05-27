# CAP-001 Phase 2C - Export and Verification Provider Migration

## Scope

CAP-001 Phase 2C migrates existing export and verification read paths behind the storage provider abstraction:

- export timeline reads for current markdown, JSON, and HTML log exports
- export metadata filtering and tombstone visibility reads
- checkpoint boundary reads used by current exports
- verification lookup by intent ID
- chain traversal lookup by previous hash

SQLite remains the behavioral reference. Existing CLI output, local configuration, database paths, schema, migrations, export formats, hash preimages, digest verification, chain traversal, ordering, and metadata behavior are preserved.

## Non-Goals

This phase does not include:

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
- export format changes
- hash or digest semantic changes
- CLI behavior changes
- rehydration migration
- tombstone mutation migration
- provider-specific optimizations

## Deliverables

- provider contract coverage for current export timeline behavior
- provider contract coverage for export metadata filtering and deleted-record visibility
- provider contract coverage for verification lookup by ID and by hash
- SQLite export read provider methods for current timeline and checkpoint export inputs
- SQLite verification read provider methods for current verify and chain workflows
- existing export command execution routed through provider-backed reads
- existing verify and chain command execution routed through provider-backed reads
- storage provider contract documentation updated with Phase 2C status

## Acceptance Criteria

Acceptance requires:

- `yanzi export --format markdown` output behavior is unchanged
- `yanzi export --format json` output behavior is unchanged
- `yanzi export --format html` output behavior is unchanged
- export ordering is unchanged
- export metadata filtering is unchanged
- deleted/tombstoned records remain hidden unless explicitly included
- checkpoint boundary rendering remains unchanged
- `yanzi verify` output behavior is unchanged
- `yanzi chain` output behavior is unchanged
- missing intent errors remain compatible
- hash/digest preimage semantics are unchanged
- previous-hash chain traversal semantics are unchanged
- existing SQLite databases remain compatible
- no config changes are introduced
- no schema changes are introduced
- provider tests pass
- SQLite parity is preserved

## Validation Performed

Required validation for this phase:

- `go test ./...`
- `go vet ./...`
- `go build -o bin/yanzi ./cmd/yanzi`
- `make docs-build`
- representative CLI workflow validation for project, capture, list, show, checkpoint, export, verify, chain, and rehydrate commands

## Preserved Behavior Notes

Export timeline reads preserve current handling of metadata from both the `meta` and `metadata` columns. Later metadata values continue to override earlier values when both columns contain the same key.

Metadata-filtered exports continue to omit checkpoint boundaries and meta-command events. Normal exports continue to include meta-command events.

Verification and chain reads preserve the existing intent hash preimage. Chain traversal still follows `prev_hash` by looking up the previous record by stored hash and reports a missing link when a previous hash cannot be found.
