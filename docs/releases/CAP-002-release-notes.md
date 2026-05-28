# CAP-002 Release Notes

Yanzi 2.10.0 closes the CAP-002 Operational API and Runtime Foundation arc.

This release keeps Yanzi local-first by default while adding an optional shared runtime for centralized operational access.

## What Is Included

- project and checkpoint API endpoints
- artifact capture and artifact read API endpoints
- verification and chain API endpoints
- export API endpoints for markdown, JSON, and HTML
- deterministic rehydration API endpoint
- explicit foreground runtime bootstrap via `yanzi serve`
- runtime health visibility for provider readiness and runtime state

## Operational Guarantees

- CLI behavior remains canonical
- the API surface is additive and route-scoped
- deterministic serialization and error behavior are preserved
- runtime startup and shutdown remain explicit and foreground-only
- local-first operation remains the default mode

## Shared Runtime Capability

- `yanzi serve` provides an optional shared runtime
- the runtime serves the same deterministic operational API used by the CLI
- the runtime does not imply orchestration, background workers, or distributed coordination

## Deferred Enterprise Features

- auth/RBAC
- federation
- MCP
- enterprise control-plane behavior
- connector runtime
- plugin hosting

## Deferred Distributed Features

- Postgres
- orchestration behavior
- distributed coordination
- background workers
- queue systems

## Known Limitations

- CAP-002 does not add orchestration semantics
- CAP-002 does not add mutable runtime sessions
- CAP-002 does not add public tombstone mutation APIs
- CAP-002 does not change the CLI-first operating model

## Upgrade Compatibility

- no breaking CLI behavior was introduced
- existing local workflows remain stable
- this release is a backward-compatible minor release
