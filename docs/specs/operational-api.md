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
- list/show reads
- rehydration reads
- tombstone mutation paths

Phase 1 must not migrate those paths. Future artifact endpoint phases will address them directly.

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
- capture/list/show endpoint migration onto handler-safe service boundaries
- rehydration endpoint work
- tombstone endpoint work
- auth, runtime hosting, orchestration, and non-SQLite provider concerns
