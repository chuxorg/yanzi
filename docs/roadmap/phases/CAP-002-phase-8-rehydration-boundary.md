# CAP-002 Phase 8 - Rehydration Service Boundary

## Scope

Phase 8 introduces a minimal internal rehydration boundary for the current deterministic `yanzi rehydrate` flow without exposing any public rehydration endpoint.

This phase covers:

- current project-scoped rehydration behavior
- latest-checkpoint selection
- post-checkpoint intent composition
- fallback capture-window behavior when no checkpoint exists
- deterministic ordering and output parity for text and JSON rendering
- a lightweight internal boundary future API handlers can call

## Non-Goals

This phase does not introduce:

- `GET /v0/rehydrate`
- `POST /v0/rehydrate`
- checkpoint redesign
- output redesign
- runtime or daemon behavior
- auth or RBAC
- federation
- MCP
- Postgres
- orchestration semantics
- tombstone mutation behavior

## Implementation Boundaries

The new seam is intentionally narrow:

- `internal/library/rehydration_service.go` isolates the SQL-backed rehydration flow behind a small service object
- the CLI continues to own user-facing formatting, but it now calls the service directly
- the library compatibility wrappers remain in place for existing call sites and future reuse

The boundary preserves current behavior rather than redesigning it.

## Current Behavior Audit

The existing local rehydration path behaves as follows and is preserved in this phase:

- `yanzi rehydrate` uses the active project from the local state file
- missing active project state remains an error
- HTTP mode continues to reject local rehydration execution
- the latest checkpoint for the active project is selected by `created_at DESC`
- if a checkpoint exists, only intents strictly after that checkpoint timestamp are rendered
- if no checkpoint exists, the command falls back to the latest capture window using the deterministic fallback limit
- records remain ordered by `created_at ASC, id ASC` before rendering
- cross-project records remain excluded by the project metadata filter
- text output preserves the current checkpoint and fallback headings
- JSON output preserves the current payload shape, including `fallback_reason` and checkpoint details

## Acceptance Criteria

Phase 8 is complete when:

- the rehydration boundary exists
- CLI `rehydrate` behavior remains unchanged
- parity tests cover checkpoint and fallback rehydration behavior
- deterministic ordering remains unchanged
- no public rehydration endpoint is exposed
- remaining direct SQL debt is isolated and documented

## Deferred Rehydration Debt

The following work remains intentionally deferred after this phase:

- public rehydration endpoints
- runtime hosting and daemon behavior
- auth, RBAC, federation, MCP, and Postgres support
- any checkpoint semantic redesign
- any output format redesign
- any orchestration-specific rehydration behavior

## Implementation Status

Completed in this phase:

- added parity tests for missing-project handling and deterministic ordering
- introduced `RehydrationService` as the internal rehydration boundary
- routed CLI `rehydrate` through the service boundary
- kept the existing compatibility wrappers available for library callers

## Validation

Validation to perform for this phase:

- `go test ./...`
- `go vet ./...`
- `go build -o bin/yanzi ./cmd/yanzi`
- `make docs-build`
- representative CLI regression flow covering capture, list, show, checkpoint create/list, rehydrate, verify, and export markdown/json/html

## Recommendation For CAP-002 Phase 9

Keep rehydration endpoint work deferred until the API layer can delegate directly through `RehydrationService` without duplicating project selection, checkpoint scoping, or output semantics.

Phase 9 should focus on any remaining rehydration adjacency only if it can reuse this boundary unchanged.
