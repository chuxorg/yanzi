# Operational API

## Purpose

The operational API defines the internal HTTP seam for Yanzi operational workflows while preserving the current local-first architecture and CLI primacy.

The CLI remains the canonical interface. The operational API exists so future endpoint work can reuse existing library and provider behavior instead of introducing parallel semantics.

## Scope

The operational API foundation covers:

- lightweight internal HTTP server structure
- deterministic routing foundations
- canonical request and response models
- handler boundaries that delegate to existing library and provider seams
- minimal operational health and status exposure

## Non-Goals

This specification does not define:

- daemonization
- auth or RBAC
- orchestration behavior
- federation
- MCP
- Postgres
- connector hosting
- distributed deployment
- speculative provider-specific fields
- alternate lineage semantics

## API Shape

The API foundation uses the existing `/v0` prefix and introduces the following route groups:

- `/v0/health`
- `/v0/artifacts`
- `/v0/projects`
- `/v0/checkpoints`

Phase 1 only implements `GET /v0/health`.

The other route groups are explicitly registered as deferred placeholders so the future endpoint surface has a stable home without implying CRUD completeness.

## Model Boundaries

Canonical models align with current Yanzi semantics for:

- artifacts
- projects
- checkpoints
- health and status

Models intentionally avoid future-provider fields and deployment-oriented metadata.

## Handler Boundaries

Handlers must:

- delegate to existing configuration, library, and provider seams
- avoid duplicating storage or domain logic
- preserve current ordering, validation, and compatibility behavior

Handlers must not:

- redefine operational semantics
- bypass provider behavior
- introduce parallel persistence paths

## CAP-001 Carry-Forward Debt

CAP-002 begins after the storage abstraction seam is sufficiently complete, but the following paths intentionally remain outside provider migration at this point:

- capture writes
- rehydration reads
- tombstone mutation paths

Phase 1 must not migrate those paths. Future artifact endpoint phases will address them directly.

CAP-002 Phase 3 narrows the `list/show` portion of that debt by introducing a dedicated internal artifact read boundary for current CLI read behavior. The boundary still uses the existing local SQL path where required, but future artifact endpoints must delegate through that seam instead of duplicating local read logic in handlers.

CAP-002 Phase 5 narrows the write-side portion of that debt by introducing a dedicated internal artifact write boundary for current capture, artifact creation, tombstone, and restore behavior. Provider-compatible artifact creation writes now route through that boundary. Capture writes and tombstone mutations still use SQLite where no provider contract exists, but that SQL is isolated behind the boundary and must not leak into future API handlers or contracts.

## Implementation Status

Current status: API foundation only.

Implemented in CAP-002 Phase 1:

- internal `internal/api` package structure
- lightweight `net/http` server foundation
- route registration for `health`, `artifacts`, `projects`, and `checkpoints`
- canonical API model definitions for artifacts, projects, checkpoints, and health/status
- health handler wired through existing config and storage provider seams
- deterministic placeholder responses for deferred route groups
- lightweight routing and response tests

Current runtime status:

- internal-only
- local-only friendly
- experimental
- not daemonized
- not exposed as a full supported runtime surface

Current health/status limitation:

- the Phase 1 health response intentionally exposes provider name, provider status, and provider error only
- richer provider health fields remain deferred until the CAP-001 provider hardening work is present on the active base branch

Deferred endpoint work:

- full artifact endpoint implementation
- project and checkpoint endpoint implementation
- capture endpoint migration
- rehydration endpoint work
- tombstone endpoint work
- auth, runtime hosting, orchestration, and non-SQLite provider concerns

## Artifact Read Boundary

Current artifact endpoint status:

- `/v0/artifacts` remains a deferred placeholder
- no public artifact list, show, or create endpoint is exposed yet

Current internal read status:

- CAP-002 Phase 3 introduces `internal/library/artifact_read_store.go` as the current local list/show read boundary
- the CLI `list` and `show` commands delegate through that boundary
- the boundary preserves existing SQL-backed ordering, filtering, project scoping, and deleted-record handling

Preserved current behavior and quirks:

- `list` remains scoped to the active project unless `--all-projects` is used
- `list` continues to exclude `source_type = 'artifact'` rows
- `list` continues to hide tombstoned records by default and include them only with `--include-deleted`
- `list` ordering remains newest first by `created_at`, with `id` as the deterministic tie-breaker
- `show` continues to look up by record ID without applying a tombstone visibility filter
- `show` preserves the current direct `meta` column behavior and does not synthesize missing metadata from the legacy `metadata` column

Future artifact endpoint work:

- implement artifact read endpoints only through the read boundary
- implement future write endpoints only through the write boundary
- keep public capture endpoints, public mutation endpoints, rehydration reads, and public tombstone APIs deferred to later CAP-002 phases

## Artifact Write Boundary

Current artifact mutation endpoint status:

- `/v0/artifacts` remains a deferred placeholder
- no public artifact create, update, patch, delete, capture, or tombstone endpoint is exposed yet

Current internal write status:

- CAP-002 Phase 5 introduces `internal/library/artifact_write_store.go` as the current local write boundary
- local `yanzi capture` delegates through that boundary
- library artifact creation delegates through that boundary
- local `yanzi delete` and `yanzi restore` delegate through that boundary
- artifact creation inside the boundary uses provider-compatible SQLite artifact writes

Preserved current behavior and quirks:

- `capture` keeps the existing required flags, prompt/response source validation, metadata shaping, `id:`/`hash:` output, and `last_hash` update behavior
- capture hashes keep the existing `HashIntent` preimage, canonical metadata behavior, and optional `prev_hash` lineage semantics
- capture rows still persist artifact-compatible columns with `class = intent`, `type = prompt`, `content = prompt`, and `metadata = meta`
- artifact creation still persists `author = yanzi`, `source_type = artifact`, system metadata in `meta`, caller metadata in `metadata`, and no `prev_hash`
- tombstone still writes deletion metadata to `metadata` for normal captures and to `meta` for `source_type = artifact` rows
- restore still removes deletion metadata from the same column selected by tombstone behavior

Remaining write-side debt:

- capture writes do not yet have a provider contract method
- tombstone and restore mutations do not yet have provider contract methods
- direct SQLite writes remain inside `ArtifactWriteStore` for those deferred operations
- public capture, artifact mutation, and tombstone APIs remain deferred
- future APIs must delegate through `ArtifactWriteStore` or a provider-backed successor rather than reaching into SQLDB directly
