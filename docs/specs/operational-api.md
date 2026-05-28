# Operational API

## Purpose

The operational API defines the internal HTTP seam for Yanzi operational workflows while preserving the current local-first architecture and CLI primacy.

The CLI remains the canonical interface. The operational API exists so endpoint work can reuse existing library and provider behavior instead of introducing parallel semantics.

## Scope

The operational API covers:

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

The `/v0` API currently exposes the following deterministic operational routes:

- `GET /v0/health`
- `POST /v0/artifacts`
- `GET /v0/artifacts/{id}`
- `GET /v0/verify/{id}`
- `GET /v0/chain/{id}`
- `GET /v0/export/markdown`
- `GET /v0/export/json`
- `GET /v0/export/html`
- `GET /v0/rehydrate`

The following route groups remain deferred placeholders:

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

Future phases must keep those paths behind internal boundaries rather than exposing direct SQLDB access in handlers.

## CAP-002 Boundary Status

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

Implemented in CAP-002 Phase 3:

- internal artifact read service boundary for CLI `list` and `show`

Implemented in CAP-002 Phase 5:

- internal artifact write boundary for capture, artifact creation, tombstone, and restore behavior

Implemented in CAP-002 Phase 6:

- `POST /v0/artifacts`
- `GET /v0/artifacts/{id}` for read-after-write

Implemented in CAP-002 Phase 7:

- `GET /v0/verify/{id}`
- `GET /v0/chain/{id}`
- `GET /v0/export/markdown`
- `GET /v0/export/json`
- `GET /v0/export/html`

- provider-backed verification and export delegation

Implemented in CAP-002 Phase 8:

- rehydration service boundary for CLI reconstruction

Implemented in CAP-002 Phase 9:

- deterministic `GET /v0/rehydrate`

Implemented in CAP-002 Phase 10:

- explicit `yanzi serve` foreground runtime bootstrap
- lightweight server execution mode for the existing Operational API
- runtime health visibility in the shared runtime path
- runtime lifecycle tests for startup, shutdown, route serving, and bind failure handling

Implemented in CAP-002 Phase 11:

- API and runtime stabilization
- route consistency audit
- response consistency audit
- runtime lifecycle regression coverage
- CAP-002 completion assessment and release-readiness documentation

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
- auth, runtime hosting, orchestration, and non-SQLite provider concerns

## Artifact Read Boundary

Current artifact endpoint status:

- `GET /v0/artifacts` is implemented
- `GET /v0/artifacts/{id}` is implemented
- `POST /v0/artifacts` is implemented
- artifact API behavior is currently read-only for reads and capture-only for writes
- artifact mutation endpoints remain deferred

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

- implement artifact list endpoints only through the read boundary
- implement future write endpoints only through the write boundary
- keep public mutation endpoints, rehydration reads, and public tombstone APIs deferred to later CAP-002 phases

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
- `POST /v0/artifacts` delegates to `ArtifactWriteStore.CreateCapture`

Preserved current behavior and quirks:

- `capture` keeps the existing required flags, prompt/response source validation, metadata shaping, `id:`/`hash:` output, and `last_hash` update behavior
- capture hashes keep the existing `HashIntent` preimage, canonical metadata behavior, and optional `prev_hash` lineage semantics
- capture rows still persist artifact-compatible columns with `class = intent`, `type = prompt`, `content = prompt`, and `metadata = meta`
- artifact creation still persists `author = yanzi`, `source_type = artifact`, system metadata in `meta`, caller metadata in `metadata`, and no `prev_hash`
- tombstone still writes deletion metadata to `metadata` for normal captures and to `meta` for `source_type = artifact` rows
- restore still removes deletion metadata from the same column selected by tombstone behavior

## Artifact Capture Endpoint

Implemented in CAP-002 Phase 6:

- `POST /v0/artifacts`

`POST /v0/artifacts` request fields:

- `author` required
- `source_type` optional, default `cli`
- `title` optional
- `prompt` required
- `response` required
- `metadata` optional string map
- `project` optional project association stored in capture metadata before hashing
- `prev_hash` optional lineage link

`POST /v0/artifacts` response fields:

- `id`
- `created_at`
- `author`
- `source_type`
- `title` when supplied
- `prompt`
- `response`
- `metadata` when present
- `prev_hash` when supplied
- `hash`

Endpoint behavior:

- writes are routed through `ArtifactWriteStore.CreateCapture`
- read-after-write is routed through `ArtifactReadStore.GetIntentRecord`
- malformed payloads use deterministic API error envelopes
- collection `GET /v0/artifacts` remains deferred
- `PUT`, `PATCH`, `DELETE`, and tombstone APIs remain unavailable
- no daemonization, auth/RBAC, orchestration, federation, MCP, Postgres, export, verification, or rehydration behavior is introduced

Remaining write-side debt:

- capture writes do not yet have a provider contract method
- tombstone and restore mutations do not yet have provider contract methods
- direct SQLite writes remain inside `ArtifactWriteStore` for those deferred operations
- artifact update/delete/tombstone APIs remain deferred
- future APIs must delegate through `ArtifactWriteStore` or a provider-backed successor rather than reaching into SQLDB directly

## Verification Endpoints

Implemented in CAP-002 Phase 7:

- `GET /v0/verify/{id}`
- `GET /v0/chain/{id}`

Compatibility aliases:

- `GET /v0/intents/{id}/verify`
- `GET /v0/intents/{id}/chain`

Endpoint behavior:

- verification delegates through shared library helpers backed by `provider.GetVerificationIntent`
- chain traversal delegates through shared library helpers backed by `provider.GetVerificationIntentByHash`
- verify preserves the current stored-hash vs computed-hash behavior and invalid-hash error semantics
- chain preserves the current oldest-to-newest ordering and missing-link reporting behavior
- handlers do not reach into SQL directly

## Export Endpoints

Implemented in CAP-002 Phase 7:

- `GET /v0/export/markdown`
- `GET /v0/export/json`
- `GET /v0/export/html`

Supported query behavior:

- `project` required
- `include_deleted` optional boolean
- `profile` optional
- `meta_<key>` exact-match metadata filters

Endpoint behavior:

- handlers delegate through provider-backed export reads
- deterministic renderer helpers preserve current export ordering and serialization
- no direct SQL access is introduced in handlers
- checkpoint filtering is not added beyond existing CLI semantics

## Rehydration Status

Implemented in CAP-002 Phase 9:

- `GET /v0/rehydrate`

Endpoint behavior:

- handler delegates through the rehydration service boundary
- checkpoint anchoring and fallback behavior remain deterministic
- missing active project and missing project behavior remain stable
- no runtime, auth, federation, MCP, or orchestration semantics are implied

## Future Work

- keep artifact list and mutation work deferred to later CAP-002 phases
- keep project/checkpoint mutation work deferred to later CAP-002 phases
- keep tombstone mutation work deferred to later CAP-002 phases
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
- the local SQLite-backed provider remains the source of truth for current semantics
- no endpoint requires daemonization or distributed coordination
