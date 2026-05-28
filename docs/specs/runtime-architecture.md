# Runtime Architecture

## Purpose

Yanzi keeps CLI execution as the default operating model while exposing an optional foreground shared runtime for API access.

The runtime exists to serve the already-defined Operational API with deterministic local behavior. It is intentionally small and does not attempt to become a scheduler, daemon orchestrator, or distributed control plane.

## Scope

The runtime architecture covers:

- explicit foreground server bootstrap
- operational API serving lifecycle
- graceful shutdown on process signal or explicit cancellation
- runtime visibility on the health endpoint
- optional shared access to the existing API routes

## Non-Goals

This architecture does not define:

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
- mutable session state

## Execution Modes

Yanzi supports two local-first modes:

- CLI mode, which remains the canonical workflow for capture, list, show, checkpoint, verify, chain, export, and rehydrate operations
- runtime mode, which exposes the Operational API in a foreground process via `yanzi serve`

The runtime is optional. CLI workflows continue to function independently of the runtime process.

## Bootstrap Boundary

The runtime bootstrap is intentionally narrow:

- `internal/runtime` owns the start/shutdown lifecycle
- `internal/api/server` provides the HTTP server wrapper used by the runtime
- `internal/api/routes` remains the source of truth for route registration
- handlers continue to delegate into existing library and provider seams

The bootstrap does not introduce separate scheduling or coordination semantics. It simply binds a listener, serves requests, and stops cleanly when asked.

## Health Visibility

The health endpoint continues to report provider readiness and now exposes minimal runtime visibility when the shared runtime is active.

Runtime visibility remains informational:

- runtime mode
- startup timestamp
- provider readiness
- CLI version

No session state, orchestration state, or workflow execution state is implied by the health response.

## CAP-002 Completion Assessment

Yanzi now has a stabilized and release-ready runtime foundation for the CAP-002 arc.

Operationally available:

- CLI mode remains the canonical workflow
- `yanzi serve` provides an optional foreground shared runtime
- operational API routes are served through the runtime bootstrap
- runtime startup and shutdown behavior is deterministic and regression-tested

Intentionally deferred:

- auth and RBAC
- federation
- MCP
- Postgres
- orchestration semantics
- worker or queue systems
- distributed coordination
- plugin hosting
- enterprise control-plane behavior

Local-first guarantees:

- the runtime is optional
- the CLI does not depend on the runtime process
- the shared runtime does not introduce autonomous workflow state
- current provider-backed semantics remain the source of truth

## Deferred Work

The following remain intentionally deferred:

- auth and RBAC
- federation
- MCP
- Postgres
- orchestration semantics
- worker or queue systems
- distributed coordination
- plugin hosting
- enterprise control-plane behavior

## Operational Note

The runtime is a shared access point, not a new operational authority. The local database, provider seams, and CLI semantics remain the source of truth for deterministic behavior.
