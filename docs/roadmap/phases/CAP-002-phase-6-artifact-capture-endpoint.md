# CAP-002 Phase 6 — Artifact Capture API Endpoint

## Scope

Phase 6 exposes the first artifact write API endpoint by routing capture creation through the internal artifact write boundary introduced in CAP-002 Phase 5.

This phase covers:

- `POST /v0/artifacts` for deterministic capture creation
- `GET /v0/artifacts/{id}` for read-after-write verification of created captures
- current author, source, prompt, response, metadata, project, `prev_hash`, and hash behavior
- deterministic JSON success and error responses
- CLI capture regression coverage

## Non-Goals

This phase does not introduce:

- artifact update endpoints
- artifact delete or tombstone endpoints
- rehydration endpoints
- export endpoints
- verification endpoints
- runtime daemonization
- auth or RBAC
- federation
- MCP
- Postgres
- orchestration behavior
- alternate capture, lineage, or hash semantics

## Implementation Boundaries

The endpoint is intentionally narrow:

- `POST /v0/artifacts` accepts a capture-shaped payload and delegates persistence to `ArtifactWriteStore.CreateCapture`
- `GET /v0/artifacts/{id}` delegates reads to `ArtifactReadStore.GetIntentRecord`
- `GET /v0/artifacts` remains deferred
- `PUT`, `PATCH`, and `DELETE` remain unavailable
- handlers open the current local provider but do not reach directly into SQL for persistence
- public tombstone and mutation APIs remain deferred

## Request Shape

`POST /v0/artifacts` accepts:

- `author` required
- `source_type` optional, defaulting to `cli`
- `title` optional
- `prompt` required
- `response` required
- `metadata` optional object of string values
- `project` optional project association stored as capture metadata
- `prev_hash` optional lineage link

## Preserved Semantics

The endpoint preserves current capture behavior:

- capture records use random 16-byte hex IDs and UTC RFC3339Nano timestamps
- capture hashes use the existing `HashIntent` preimage
- metadata is persisted in the current capture metadata column behavior
- `project` association is stored in metadata before hashing
- `prev_hash` is included in the hash preimage when supplied
- capture rows continue to persist artifact-compatible columns
- no CLI behavior changes are introduced

## Acceptance Criteria

Phase 6 is complete when:

- `POST /v0/artifacts` creates a capture record
- the handler delegates through `ArtifactWriteStore`
- malformed payloads return deterministic error responses
- created captures can be read through `GET /v0/artifacts/{id}`
- metadata and project association persist
- deterministic hash behavior is preserved
- CLI capture behavior remains unchanged
- no update, delete, patch, or tombstone endpoint is introduced
- no orchestration semantics are introduced

## Remaining Deferred Debt

The following work remains intentionally deferred after this phase:

- public artifact update endpoints
- public artifact delete/tombstone endpoints
- artifact list endpoint implementation
- rehydration endpoints
- export endpoints
- verification endpoints
- provider contract methods for capture writes
- provider contract methods for tombstone and restore mutations
- runtime daemonization, auth/RBAC, federation, MCP, Postgres, and orchestration behavior

## Implementation Status

Completed in this phase:

- added `POST /v0/artifacts`
- added `GET /v0/artifacts/{id}` for read-after-write
- kept collection list and mutation methods deferred
- added endpoint coverage for success, malformed payloads, validation errors, metadata/project persistence, read-after-write, response shape, and deferred mutation methods
- added CLI capture regression coverage

## Validation

Required validation for this phase:

- `go test ./...`
- `go vet ./...`
- `go build -o bin/yanzi ./cmd/yanzi`
- `make docs-build`
- representative CLI regression flow covering `capture`, `list`, `show`, `verify`, `export markdown/json/html`, and `rehydrate`
- API validation for `POST /v0/artifacts`, `GET /v0/artifacts/{id}`, malformed payloads, and deterministic error responses

## Recommendation For CAP-002 Phase 7

Phase 7 should harden the provider contract for capture writes before expanding mutation behavior. Artifact update, delete, and tombstone endpoints should remain deferred until those operations can delegate through explicit provider-backed write methods.
