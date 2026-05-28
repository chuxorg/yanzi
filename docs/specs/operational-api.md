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
- `/v0/rehydrate`
- `/v0/artifacts`
- `/v0/verify`
- `/v0/export`
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

CAP-002 Phase 7 exposes verification and export read endpoints. Verification and chain handlers delegate through provider-backed verification helpers. Export handlers delegate through provider-backed export reads and shared deterministic renderers. No direct SQL access is introduced in handlers, and mutation, tombstone, and rehydration work remain deferred.

CAP-002 Phase 8 narrows the rehydration portion of that debt by introducing `internal/library/rehydration_service.go` as the current local rehydration boundary. The CLI continues to use existing rehydration semantics, but future rehydration endpoints must delegate through that boundary instead of duplicating project selection, checkpoint scoping, or output behavior in handlers.

## Implementation Status

Current status: artifact capture, verification, export, and rehydration read endpoints available; broader operational API remains deferred.

Implemented in CAP-002 Phase 1:

- internal `internal/api` package structure
- lightweight `net/http` server foundation
- route registration for `health`, `artifacts`, `projects`, and `checkpoints`
- canonical API model definitions for artifacts, projects, checkpoints, and health/status
- health handler wired through existing config and storage provider seams
- deterministic placeholder responses for deferred route groups
- lightweight routing and response tests

Implemented in CAP-002 Phase 7:

- `GET /v0/verify/{id}`
- `GET /v0/chain/{id}`
- `GET /v0/export/markdown`
- `GET /v0/export/json`
- `GET /v0/export/html`
- provider-backed verification and export delegation

Implemented in CAP-002 Phase 9:

- deterministic `GET /v0/rehydrate` endpoint
- rehydration handler wiring through the rehydration service boundary
- deterministic API response shaping for checkpoint and fallback reconstruction
- API tests for deterministic reconstruction, missing active-project handling, missing-project handling, and fallback behavior

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

- project and checkpoint endpoint implementation
- tombstone endpoint work
- auth, runtime hosting, orchestration, and non-SQLite provider concerns

## Artifact Read Boundary

Current artifact endpoint status:

- `GET /v0/artifacts` is implemented
- `GET /v0/artifacts/{id}` is implemented
- `POST /v0/artifacts` is implemented
- artifact API behavior is currently read-only for reads and capture-only for writes
- artifact mutation endpoints remain deferred

Current rehydration status:

- `GET /v0/rehydrate` is implemented
- CLI `rehydrate` and the API endpoint both route through the internal rehydration service boundary
- the endpoint preserves deterministic checkpoint anchoring, fallback behavior, and serialization

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

- keep artifact read endpoints routed only through the read boundary
- keep capture writes routed only through the write boundary
- keep mutation endpoints deferred until the remaining write and tombstone debt is explicitly addressed

Future rehydration endpoint work:

- preserve deterministic checkpoint and fallback behavior for any future expansion
- keep runtime, auth, federation, MCP, and Postgres work deferred
