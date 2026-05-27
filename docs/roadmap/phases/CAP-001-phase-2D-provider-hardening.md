# CAP-001 Phase 2D - Provider Hardening and Compatibility Closure

## Scope

CAP-001 Phase 2D closes the SQLite storage abstraction hardening work before any future provider or runtime capability is introduced.

This phase focuses on:

- provider surface audit and cleanup
- behavioral parity checks for projects, checkpoints, artifacts, exports, verification, rehydration, and chain traversal
- migration and upgrade compatibility coverage
- provider lifecycle and health hardening
- regression protection for repeated export, verification, and rehydration flows

SQLite remains the only provider and the behavioral reference. Existing CLI output, local configuration, database paths, schema, migrations, export formats, checkpoint hashes, verification hashes, rehydration semantics, ordering, and metadata behavior are preserved.

## Non-Goals

This phase does not include:

- Postgres or any other datastore provider
- runtime or daemon work
- REST
- federation
- MCP
- provider configuration switching
- config schema changes
- schema redesign
- export format changes
- checkpoint hash changes
- verification hash changes
- CLI behavior changes
- user-visible provider health exposure

## Provider Surface Audit

The provider interface remains intentionally narrow for current CAP-001 needs:

- artifact operations
- project operations
- checkpoint operations
- export read operations
- verification read operations
- internal provider identity, health, SQL handle, and lifecycle

The `SQLDB` escape hatch remains by design for backward-compatible call sites that still operate directly on SQLite. Removing it is deferred until capture writes, generic list/show reads, rehydration reads, and tombstone mutation paths have provider methods and compatibility coverage.

No provider configuration surface was added. No new user-facing selection behavior exists.

## Behavioral Parity Audit

Parity was preserved for:

- projects: create, list, duplicate handling, and missing-project checks
- checkpoints: creation, newest-first project ordering, all-project ordering, previous-checkpoint linkage, and nil artifact ID return behavior
- artifacts: intent/context creation, project and all-project listing, context visibility, deleted artifact visibility quirks, and unique-prefix context lookup
- exports: markdown, JSON, and HTML timeline behavior, checkpoint boundaries, meta events, metadata filtering, repeated execution, and deleted-record visibility
- verification: lookup by ID, digest recomputation inputs, repeated execution, and missing-record errors
- chain traversal: previous-hash lookup behavior and missing-link reporting
- rehydration: latest-checkpoint selection, post-checkpoint composition, fallback window behavior, and repeated deterministic output

Intentionally preserved quirks:

- metadata-filtered exports omit checkpoint boundary events and meta-command events
- malformed export metadata rows are skipped instead of failing the full export
- malformed checkpoint `artifact_ids` still fail checkpoint reads with a decode error
- normal artifact and context lists hide deleted artifacts, while current context show resolution can still resolve deleted context artifacts by visible ID or unique prefix
- checkpoint creation with nil artifact IDs returns nil `ArtifactIDs` while SQLite stores an empty JSON list

## Provider Health Hardening

Provider health remains internal-only.

SQLite health now reports:

- provider identity
- resolved path when available
- connectivity through `PingContext`
- migration/schema readiness for current local tables and columns
- writable state through SQLite `query_only`
- unavailable state for closed or nil provider handles

This is not exposed through CLI or runtime APIs in this phase.

## Regression Coverage Added

New or strengthened coverage includes:

- existing DB reuse without reinitialization
- repeated export reads
- repeated verification reads
- malformed export metadata compatibility
- malformed checkpoint artifact ID behavior
- closed provider lifecycle health
- missing migration/schema health
- legacy database upgrade before rehydration
- repeated checkpoint-based rehydration composition

## Operational Continuity

Yanzi-backed execution flow was run during the phase using project `yanzi-dev`.

Prompt captured:

```text
CAP-001 / Phase 2D - Provider Hardening and Compatibility Closure
```

Implementation decisions captured:

- keep SQLite as the only provider
- keep `SQLDB` for compatibility until remaining direct SQLite call sites are migrated
- harden internal health without adding user-visible provider behavior
- prefer regression tests that pin existing quirks instead of abstraction-only tests
- preserve export, verification, chain, and rehydration semantics exactly

Validation findings captured:

- `go test ./...` passed
- `go vet ./...` passed
- `go build -o bin/yanzi ./cmd/yanzi` passed
- `make docs-build` passed with existing MkDocs informational warnings about nav entries and MkDocs 2.0 compatibility
- fresh DB representative CLI workflow passed
- repeated verify, rehydrate, and JSON export workflow passed
- upgraded DB compatibility was validated by regression coverage

Completion checkpoint:

- completion checkpoint created after validation using project `yanzi-dev`

## CAP-001 Completion Assessment

Completed abstraction work:

- internal provider contract exists
- SQLite registry path exists and remains the only provider path
- projects and checkpoints route through provider operations
- artifact creation and listing route through provider operations
- export read paths route through provider operations
- verification and chain lookup reads route through provider operations
- provider health and lifecycle behavior are hardened internally
- migration, upgrade, repeated execution, and malformed-data regression coverage are in place

Intentionally deferred:

- capture intent write migration
- generic list/show intent read migration
- rehydration query migration
- tombstone mutation migration
- removal of `SQLDB`
- provider configuration switching
- additional datastore providers
- runtime, REST, federation, MCP, or operational API work

Readiness for CAP-002 Operational API:

CAP-001 is ready to close from the storage abstraction perspective once final validation passes and human phase completion is approved. CAP-002 can build on a stable SQLite-backed provider boundary while deferring future-provider expansion and remaining direct SQLite call-site migrations to explicitly scoped work.

## Validation Performed

Required validation:

- `go test ./...`
- `go vet ./...`
- `go build -o bin/yanzi ./cmd/yanzi`
- `make docs-build`

Representative CLI workflow validation:

- fresh DB: project create/list/use/current, capture, list, show, checkpoint create/list, rehydrate, verify, chain, export markdown/json/html
- upgraded DB: legacy SQLite database upgrade and rehydration compatibility
- repeated execution: export, verify, and rehydrate repeated without output-shape changes

Validation status:

- passed

Validation findings:

- Fresh DB workflow passed for project create/list/use/current, capture, list, show, checkpoint create/list, rehydrate, verify, chain, and markdown/json/html export.
- Repeated execution passed for verify, rehydrate, and JSON export output stability.
- Upgraded DB compatibility passed through legacy SQLite upgrade and rehydration regression coverage.
- Documentation build passed. Existing MkDocs warnings are informational and unrelated to CAP-001 Phase 2D.
