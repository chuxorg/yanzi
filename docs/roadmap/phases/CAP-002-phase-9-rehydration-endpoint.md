# CAP-002 Phase 9 - Rehydration API Endpoint

## Scope

Phase 9 exposes deterministic rehydration reads through the Operational API while preserving current CLI semantics and the existing rehydration service boundary.

This phase covers:

- current project-scoped rehydration reconstruction behavior
- checkpoint anchoring and post-checkpoint composition
- fallback reconstruction when no checkpoint exists
- deterministic ordering and serialization
- provider-compatible read execution through the existing service boundary

## Non-Goals

This phase does not introduce:

- runtime daemonization
- auth or RBAC
- federation
- MCP
- orchestration behavior
- workflow execution semantics
- mutable session semantics
- long-running execution state
- rehydration mutation behavior

## Implementation Boundaries

The new seam is intentionally narrow:

- `internal/api/handlers/rehydrate.go` exposes `GET /v0/rehydrate`
- the handler resolves the current active project using the shared library state helper
- the handler delegates to `internal/library/rehydration_service.go`
- the handler only shapes deterministic JSON response data and API errors

The boundary preserves current behavior rather than redesigning it.

## Current Behavior Audit

The existing local rehydration path behaves as follows and is preserved in this phase:

- `rehydrate` is project-scoped via the current active project binding
- missing active project state remains a validation error
- missing project state remains a not-found error
- the latest checkpoint for the active project is selected by `created_at DESC`
- if a checkpoint exists, only intents strictly after that checkpoint timestamp are returned
- if no checkpoint exists, the deterministic fallback capture window is returned
- results remain ordered deterministically by creation time and ID
- JSON serialization preserves the current payload shape

## Acceptance Criteria

Phase 9 is complete when:

- `GET /v0/rehydrate` is operational
- the endpoint delegates through `RehydrationService`
- CLI `rehydrate` behavior remains unchanged
- checkpoint and fallback reconstruction behavior remains deterministic
- missing active-project and missing-project behaviors remain compatible
- no orchestration or runtime-session semantics are introduced

## Deferred Runtime Debt

The following work remains intentionally deferred after this phase:

- runtime daemonization
- auth, RBAC, federation, MCP, and Postgres support
- orchestration semantics
- workflow execution semantics
- mutable session semantics
- any public mutation-oriented rehydration work

## Implementation Status

Completed in this phase:

- added parity tests for deterministic rehydration API behavior
- introduced `GET /v0/rehydrate`
- routed the endpoint through the rehydration service boundary
- preserved deterministic JSON reconstruction semantics
- kept CLI behavior unchanged

## Validation

Validation to perform for this phase:

- `go test ./...`
- `go vet ./...`
- `go build -o bin/yanzi ./cmd/yanzi`
- `make docs-build`
- representative CLI regression flow covering capture, list, show, checkpoint create/list, rehydrate, verify, and export markdown/json/html
- API regression flow covering `GET /v0/rehydrate`, missing active project, missing project, and deterministic reconstruction output

## Recommendation For CAP-002 Phase 10

Keep the next phase away from runtime/session semantics unless there is a concrete provider-backed contract to attach them to. If more operational API work follows, preserve the read-only reconstruction boundary and avoid turning rehydration into a workflow engine.
