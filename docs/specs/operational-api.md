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
- explicit foreground runtime bootstrap for optional shared access

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

The API foundation uses the existing `/v0` prefix and currently exposes the following deterministic read and write surfaces:

- `GET /v0/health`
- `POST /v0/artifacts`
- `GET /v0/artifacts/{id}`
- `GET /v0/verify/{id}`
- `GET /v0/chain/{id}`
- `GET /v0/export/markdown`
- `GET /v0/export/json`
- `GET /v0/export/html`
- `GET /v0/rehydrate`

The following route groups remain as deferred placeholders so the future endpoint surface has a stable home without implying CRUD completeness:

- `/v0/projects`
- `/v0/checkpoints`

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

CAP-002 Phase 8 narrows the rehydration portion of that debt by introducing `internal/library/rehydration_service.go` as the current local rehydration boundary. The CLI continues to use existing rehydration semantics, but future rehydration endpoints must delegate through that boundary instead of duplicating project selection, checkpoint scoping, or output behavior in handlers.

## Implementation Status

Current status: API and runtime stabilization complete for CAP-002.

Implemented in CAP-002 Phase 1:

- internal `internal/api` package structure
- lightweight `net/http` server foundation
- route registration for `health`, `artifacts`, `projects`, and `checkpoints`
- canonical API model definitions for artifacts, projects, checkpoints, and health/status
- health handler wired through existing config and storage provider seams
- deterministic placeholder responses for deferred route groups
- lightweight routing and response tests

Implemented in CAP-002 Phase 9:

- deterministic `GET /v0/rehydrate` endpoint
- rehydration handler wiring through the rehydration service boundary
- deterministic API response shaping for checkpoint and fallback reconstruction
- API tests for deterministic reconstruction, missing active-project handling, missing-project handling, and fallback behavior

Implemented in CAP-002 Phase 10:

- explicit `yanzi serve` foreground runtime bootstrap
- lightweight server execution mode for the existing Operational API
- runtime health visibility in the shared runtime path
- runtime lifecycle tests for startup, shutdown, route serving, and bind failure handling

Current runtime status:

- internal-only
- local-first friendly
- stabilized foreground shared runtime available via `yanzi serve`
- not daemonized
- not exposed as a full supported runtime surface

Current health/status limitation:

- the health response reports the active configuration mode and provider readiness
- when the shared runtime is active, a nested runtime block reports runtime mode and startup timestamp
- the top-level mode remains the configuration mode, preserving the distinction between local CLI operation and shared runtime serving

Deferred endpoint work:

- project and checkpoint endpoint implementation
- tombstone mutation APIs
- auth, orchestration, and non-SQLite provider concerns

## Artifact Read Boundary

Current artifact endpoint status:

- `POST /v0/artifacts` is implemented
- `GET /v0/artifacts/{id}` is implemented
- public artifact list, update, and delete endpoints remain deferred

Current rehydration status:

- `GET /v0/rehydrate` is implemented
- CLI `rehydrate` and the API endpoint both route through the internal rehydration service boundary
- the endpoint preserves deterministic checkpoint anchoring, fallback behavior, and serialization

Current runtime/health status:

- `GET /v0/health` remains the primary readiness surface
- health responses now include minimal runtime visibility when the shared runtime is active
- runtime visibility is informational only and does not imply workflow orchestration

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

- keep artifact list and mutation work deferred to later CAP-002 phases
- keep tombstone mutation work deferred to later CAP-002 phases

Future rehydration endpoint work:

- preserve deterministic checkpoint and fallback behavior for any future expansion
- keep runtime, auth, federation, MCP, and Postgres work deferred

Future runtime work:

- keep runtime bootstrap lightweight and foreground only
- avoid background workers, orchestration, or distributed coordination
- preserve CLI primacy and local-first behavior

## CAP-002 Completion Assessment

What is operational:

- artifact capture through the write boundary
- artifact read access through the current read boundary
- verification and chain reads
- deterministic export reads
- deterministic rehydration reads
- optional foreground runtime serving through `yanzi serve`
- health visibility for provider and runtime state

What remains intentionally deferred:

- public project and checkpoint mutation endpoints
- public tombstone and delete endpoints
- auth/RBAC
- federation
- MCP
- Postgres
- orchestration semantics
- distributed coordination
- connector runtime or plugin hosting

Enterprise readiness notes:

- the API surface is deterministic and local-first
- shared runtime access is optional and foreground only
- the runtime does not introduce orchestration state or workflow continuation semantics
- response envelopes remain stable and machine-readable

Shared runtime readiness notes:

- `yanzi serve` is the explicit bootstrap entry point
- startup and shutdown are deterministic and covered by regression tests
- health output exposes runtime mode and startup timestamp when the runtime is active

Local-first guarantees:

- CLI workflows remain canonical
- the shared runtime is optional
- local SQLite-backed behavior remains the source of truth for current semantics
- no endpoint requires daemonization or distributed coordination
