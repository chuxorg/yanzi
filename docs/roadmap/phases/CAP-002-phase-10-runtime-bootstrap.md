# CAP-002 Phase 10 - Runtime Bootstrap and Server Execution Mode

## Scope

Phase 10 introduces the first intentional shared-runtime execution mode for Yanzi while preserving local-first CLI semantics.

This phase covers:

- explicit foreground runtime bootstrap
- server execution lifecycle
- optional shared operational access
- API route serving through the existing operational API foundation
- minimal runtime visibility through health responses

## Non-Goals

This phase does not introduce:

- auth or RBAC
- federation
- MCP
- Postgres
- orchestration behavior
- background workers
- distributed coordination
- plugin hosting
- Kubernetes or microservice deployment semantics
- autonomous workflow execution
- mutable sessions

## Runtime Boundaries

The runtime boundary is intentionally narrow:

- `internal/runtime` owns runtime start/shutdown lifecycle
- `internal/api/server` remains the lightweight HTTP server wrapper
- `internal/api/routes` continues to own route registration
- `yanzi serve` starts the shared runtime explicitly in the foreground

The runtime is optional and does not replace the CLI as the canonical interface.

## Current Behavior Audit

The current runtime behavior introduced in this phase:

- `yanzi serve` starts a foreground HTTP server explicitly
- the server binds a configured address and serves the existing Operational API routes
- health responses continue to report provider readiness
- health responses now expose minimal runtime visibility when the shared runtime is active
- shutdown occurs explicitly through process signal or cancellation
- CLI workflows continue to function independently of the runtime

## Acceptance Criteria

Phase 10 is complete when:

- runtime bootstrap exists
- `yanzi serve` starts the server explicitly
- API routes are served through the runtime bootstrap
- runtime health visibility is present
- CLI workflows remain unchanged
- runtime startup/shutdown is deterministic
- no orchestration semantics are introduced

## Deferred Enterprise/Runtime Work

The following remain intentionally deferred after this phase:

- auth and RBAC
- federation
- MCP
- Postgres
- orchestration semantics
- background workers
- distributed coordination
- plugin hosting
- enterprise control-plane behavior

## Implementation Status

Completed in this phase:

- added runtime bootstrap package and server lifecycle helpers
- introduced `yanzi serve`
- routed the existing API foundation through the runtime bootstrap
- added runtime health visibility
- added tests for startup, shutdown, route serving, health visibility, and bind failures

## Validation

Validation to perform for this phase:

- `go test ./...`
- `go vet ./...`
- `go build -o bin/yanzi ./cmd/yanzi`
- `make docs-build`
- representative CLI regression flow covering capture, list, show, checkpoint create/list, rehydrate, verify, and export markdown/json/html
- runtime execution validation covering startup, shutdown, health checks, and route serving

## Recommendation For CAP-002 Stabilization or CAP-003

Keep future work focused on stabilization unless there is a concrete need for more shared-runtime capability. If CAP-003 follows, it should add only narrowly scoped runtime behavior that still preserves CLI primacy and deterministic provider semantics.
