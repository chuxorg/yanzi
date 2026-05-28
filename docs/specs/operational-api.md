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

Implemented endpoints:

- `GET /v0/health`
- `GET /v0/artifacts`
- `GET /v0/artifacts/{id}`

Deferred route groups remain registered so the future endpoint surface has a stable home without implying CRUD completeness.

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

## Implementation Status

Current status: health plus read-only artifact endpoints.

Implemented in CAP-002 Phase 1:

- internal `internal/api` package structure
- lightweight `net/http` server foundation
- route registration for `health`, `artifacts`, `projects`, and `checkpoints`
- canonical API model definitions for artifacts, projects, checkpoints, and health/status
- health handler wired through existing config and storage provider seams
- deterministic placeholder responses for deferred route groups
- lightweight routing and response tests

Implemented in CAP-002 Phase 4:

- read-only artifact list endpoint at `GET /v0/artifacts`
- read-only artifact detail endpoint at `GET /v0/artifacts/{id}`
- artifact handler routing through the internal artifact read boundary introduced in CAP-002 Phase 3
- deterministic JSON list/detail and error response coverage for artifact reads

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

- artifact mutation endpoints
- capture endpoint implementation
- project and checkpoint endpoint implementation
- rehydration endpoint work
- tombstone endpoint work
- auth, runtime hosting, orchestration, and non-SQLite provider concerns

## Artifact Read Boundary

Current artifact endpoint status:

- `GET /v0/artifacts` is implemented
- `GET /v0/artifacts/{id}` is implemented
- artifact API behavior is currently read-only
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
- add mutation endpoints only after the remaining write and tombstone debt is explicitly addressed
- keep capture writes, rehydration reads, and tombstone mutation work deferred to later CAP-002 phases
