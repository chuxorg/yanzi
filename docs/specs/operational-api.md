# Operational API

## Purpose

The operational API defines the internal HTTP seam for Yanzi operational workflows while preserving the current local-first architecture and CLI primacy.

The CLI remains the canonical interface. The operational API exists so endpoint work can reuse existing library and provider behavior instead of introducing parallel semantics.

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
- `/v0/verify`
- `/v0/export`
- `/v0/rehydrate`

Implemented endpoints:

- `GET /v0/health`
- `GET /v0/artifacts`
- `GET /v0/artifacts/{id}`
- `POST /v0/artifacts`
- `GET /v0/verify/{id}`
- `GET /v0/chain/{id}`
- `GET /v0/export/markdown`
- `GET /v0/export/json`
- `GET /v0/export/html`
- `GET /v0/rehydrate`

Deferred route groups remain registered so the future endpoint surface has a stable home without implying CRUD completeness.

## Model Boundaries

Canonical models align with current Yanzi semantics for:

- artifacts
- projects
- checkpoints
- health and status
- verification and export payloads
- rehydration payloads

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

CAP-002 began after the storage abstraction seam was sufficiently complete, but the following paths intentionally remained outside provider migration at that point:

- capture writes
- rehydration reads
- tombstone mutation paths

CAP-002 Phase 3 narrowed the `list/show` portion of that debt by introducing a dedicated internal artifact read boundary for current CLI read behavior. The boundary still uses the existing local SQL path where required, but future artifact endpoints must delegate through that seam instead of duplicating local read logic in handlers.

CAP-002 Phase 5 narrowed the write-side portion of that debt by introducing a dedicated internal artifact write boundary for current capture, artifact creation, tombstone, and restore behavior. Provider-compatible artifact creation writes now route through that boundary. Capture writes and tombstone mutations still use SQLite where no provider contract exists, but that SQL is isolated behind the boundary and must not leak into future API handlers or contracts.

CAP-002 Phase 7 added verification and export read endpoints that delegate through existing library and provider behavior. These remain read-only and deterministic.

CAP-002 Phase 8 introduced the rehydration service boundary. Phase 9 exposed that boundary through `GET /v0/rehydrate`.

## Implementation Status

Current status: operational API plus lightweight shared runtime foundation.

Implemented in CAP-002 Phase 1:

- internal `internal/api` package structure
- lightweight `net/http` server foundation
- route registration for `health`, `artifacts`, `projects`, and `checkpoints`
- canonical API model definitions for artifacts, projects, checkpoints, and health/status
- health handler wired through existing config and storage provider seams
- deterministic placeholder responses for deferred route groups
- lightweight routing and response tests

Implemented in CAP-002 Phase 3:

- dedicated artifact read boundary for list/show behavior
- CLI `list` and `show` delegation through that boundary

Implemented in CAP-002 Phase 4:

- read-only artifact list endpoint at `GET /v0/artifacts`
- read-only artifact detail endpoint at `GET /v0/artifacts/{id}`
- artifact handler routing through the internal artifact read boundary
- deterministic JSON list/detail and error response coverage for artifact reads

Implemented in CAP-002 Phase 5:

- dedicated artifact write boundary for capture, artifact creation, tombstone, and restore behavior
- provider-compatible artifact creation writes routed through that boundary
- local `capture`, `delete`, and `restore` delegation through the boundary

Implemented in CAP-002 Phase 6:

- `POST /v0/artifacts` capture endpoint
- read-after-write behavior through `GET /v0/artifacts/{id}`
- deterministic artifact capture response shape

Implemented in CAP-002 Phase 7:

- `GET /v0/verify/{id}`
- `GET /v0/chain/{id}`
- `GET /v0/export/markdown`
- `GET /v0/export/json`
- `GET /v0/export/html`
- provider-compatible verification and export delegation

Implemented in CAP-002 Phase 8:

- internal rehydration service boundary
- CLI `rehydrate` delegation through that boundary

Implemented in CAP-002 Phase 9:

- `GET /v0/rehydrate`
- deterministic reconstruction through the rehydration boundary

Implemented in CAP-002 Phase 10:

- explicit foreground runtime bootstrap
- optional `yanzi serve` execution mode
- runtime health visibility

Implemented in CAP-002 Phase 11:

- stabilized API/runtime response behavior
- repeated startup/shutdown coverage
- release-readiness review and documentation consolidation

Current runtime status:

- optional
- local-first friendly
- foreground execution by default
- not daemonized
- not exposed as a full supported distributed runtime surface

Current health/status behavior:

- exposes version, mode, provider status, and runtime status
- remains deterministic and operational

Deferred endpoint work:

- project and checkpoint endpoint implementation
- artifact update/delete/tombstone endpoint implementation
- public mutation endpoints beyond capture creation
- auth, runtime hosting, orchestration, and non-SQLite provider concerns

## Artifact Read Boundary

Current artifact endpoint status:

- `GET /v0/artifacts` is implemented
- `GET /v0/artifacts/{id}` is implemented
- artifact API behavior is currently read-only for reads and capture-only for writes
- artifact mutation endpoints remain deferred

Current internal read status:

- CAP-002 Phase 3 introduces `internal/library/artifact_read_store.go` as the current local list/show read boundary
- the CLI `list` and `show` commands delegate through that boundary
- the artifact API read handlers also delegate through that boundary
- the boundary preserves existing SQL-backed ordering, filtering, project scoping, and deleted-record handling

Preserved current behavior and quirks:

- `list` remains scoped to the active project unless `--all-projects` is used
- `list` continues to exclude `source_type = 'artifact'` rows
- `list` continues to hide tombstoned records by default and include them only with `--include-deleted`
- `list` ordering remains newest first by `created_at`, with `id` as the deterministic tie-breaker
- `show` continues to look up by record ID without applying a tombstone visibility filter
- `show` preserves the current direct `meta` column behavior and does not synthesize missing metadata from the legacy `metadata` column

Future artifact endpoint work:

- keep artifact read endpoints routed only through the read boundary
- keep capture writes routed only through the write boundary
- keep mutation endpoints deferred until the remaining write and tombstone debt is explicitly addressed

## Artifact Write Boundary

Current artifact mutation endpoint status:

- `POST /v0/artifacts` creates capture records through the write boundary
- no artifact update, patch, delete, or tombstone endpoint is exposed yet

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
- future APIs must delegate through `ArtifactWriteStore` or a provider-backed successor rather than reaching into SQLDB directly

## Verification and Export Read Paths

Current verification and export status:

- verification endpoints are implemented and deterministic
- export endpoints are implemented and deterministic
- CLI and API semantics remain aligned

Preserved current behavior:

- verification keeps the current hash/chain semantics
- export keeps current ordering, metadata visibility, and checkpoint/project scoping behavior
- export serialization remains deterministic for markdown, JSON, and HTML outputs

## Rehydration Read Path

Current rehydration status:

- the rehydration service boundary is implemented
- `GET /v0/rehydrate` exposes deterministic reconstruction through that boundary
- checkpoint anchoring, post-checkpoint composition, and ordering semantics are preserved

Remaining rehydration debt:

- runtime/session/orchestration semantics are intentionally deferred
- no mutable continuation or execution control is exposed through rehydration

## Runtime Foundation

Current runtime status:

- explicit foreground server bootstrap is implemented
- `yanzi serve` is optional and local-first compatible
- API serving lifecycle is deterministic
- startup and shutdown behavior is covered by regression tests

Deferred runtime work:

- auth/RBAC
- federation
- MCP
- Postgres
- orchestration behavior
- background workers
- distributed coordination
- plugin hosting

