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

CAP-002 Phase 6 exposes the first artifact write endpoint, `POST /v0/artifacts`, for capture creation only. The handler delegates to `ArtifactWriteStore.CreateCapture`, and read-after-write uses `GET /v0/artifacts/{id}` through the read boundary. Mutation, tombstone, export, verification, and rehydration endpoints remain deferred.

CAP-002 Phase 7 exposes verification and export read endpoints. Verification and chain handlers delegate through provider-backed verification helpers. Export handlers delegate through provider-backed export reads and shared deterministic renderers. No direct SQL access is introduced in handlers, and mutation, tombstone, and rehydration work remain deferred.

## Implementation Status

Current status: artifact capture, verification, chain, and export read endpoints available; broader operational API remains deferred.

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

- artifact list endpoint implementation
- artifact update/delete/tombstone endpoint implementation
- project and checkpoint endpoint implementation
- rehydration endpoint work
- tombstone endpoint work
- auth, runtime hosting, orchestration, and non-SQLite provider concerns

## Artifact Read Boundary

Current artifact endpoint status:

- `GET /v0/artifacts/{id}` is available for capture read-after-write
- `GET /v0/artifacts` remains deferred
- no artifact list endpoint is exposed yet

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
- `GET /v0/artifacts/{id}`

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
- `meta_<key>=<value>` optional exact-match metadata filters

Deferred or unsupported query behavior:

- checkpoint filters remain unsupported because current CLI export semantics do not provide them

Endpoint behavior:

- export reads delegate through `provider.ListExportItems`
- export serialization delegates through shared deterministic library renderers reused by the CLI
- export ordering, checkpoint interleaving, metadata visibility, meta-event handling, and project scoping preserve the current CLI behavior
- filtered exports continue to omit checkpoints and meta events when current CLI behavior omits them
- handlers do not perform direct SQL access

Remaining read/runtime debt:

- artifact list endpoints remain deferred
- project and checkpoint management endpoints remain deferred
- rehydration endpoints remain deferred
- runtime daemonization, auth/RBAC, federation, MCP, Postgres, and orchestration behavior remain deferred
