# CAP-002 Phase 4 — Artifact Read API Endpoints

## Scope

Phase 4 exposes the first read-only artifact API endpoints on top of the artifact read boundary introduced in CAP-002 Phase 3.

This phase covers:

- `GET /v0/artifacts`
- `GET /v0/artifacts/{id}`
- deterministic JSON response and error behavior
- handler delegation through the existing artifact read boundary
- regression coverage for current list/show semantics exposed through the API

## Non-Goals

This phase does not introduce:

- artifact create or capture endpoints
- artifact mutation or tombstone mutation endpoints
- rehydration endpoints
- export endpoints
- verification endpoints
- auth, runtime hosting, orchestration, or provider switching

## Implementation Boundaries

The Phase 4 implementation stays within the merged foundation:

- handlers delegate through the CAP-002 Phase 3 artifact read boundary
- handlers do not issue direct SQL queries
- `/v0/projects` and `/v0/checkpoints` remain deferred on the active base branch
- artifact API behavior remains read-only

## Acceptance Criteria

Phase 4 is complete when:

- artifact read endpoints are operational
- handlers delegate through the read boundary
- CLI behavior remains unchanged
- no artifact mutation endpoints are introduced
- deterministic response and error behavior is covered by tests
- remaining deferred artifact debt is documented

## Remaining Deferred Artifact Debt

The following work remains intentionally deferred after this phase:

- capture writes
- artifact mutation endpoints
- tombstone mutation paths
- rehydration read migration
- export and verification endpoint work

## Implementation Status

Completed in this phase:

- implemented `GET /v0/artifacts`
- implemented `GET /v0/artifacts/{id}`
- preserved active-project scoping, filter behavior, and missing-ID handling through the API
- routed artifact API reads through explicit handler dependencies backed by the Phase 3 read boundary
- kept unsupported artifact methods deferred by leaving the route group read-only

## Validation

Validation performed for this phase:

- `go test ./...`
- `go vet ./...`
- `go build -o bin/yanzi ./cmd/yanzi`
- `make docs-build`
- representative CLI regression flow covering capture, list, show, verify, export markdown/json/html, and rehydrate
- API behavior validation for `GET /v0/artifacts`, `GET /v0/artifacts/{id}`, missing IDs, and deterministic JSON responses

## Recommendation For CAP-002 Phase 5

Use the read-only artifact API as the stable entry point for future artifact endpoint work.

Phase 5 should either:

- introduce the first artifact mutation surface only after the remaining write and tombstone debt is explicitly isolated, or
- continue hardening the remaining artifact operational paths before exposing any mutation API surface
