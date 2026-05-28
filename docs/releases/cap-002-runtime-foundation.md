# CAP-002 Runtime Foundation

CAP-002 closes the Operational API and runtime foundation for Yanzi.

The result is a local-first operational surface with optional shared runtime access, deterministic behavior, and no change to the CLI-first operating model.

## Capabilities Completed

- deterministic artifact capture through an internal write boundary
- artifact read access through the internal read boundary
- verification and chain read endpoints
- deterministic export read endpoints
- deterministic rehydration read endpoint
- explicit foreground runtime bootstrap via `yanzi serve`
- runtime health visibility for provider readiness and runtime state
- route, response, and serialization stabilization

## API Surface Summary

Implemented operational routes:

- `GET /v0/health`
- `POST /v0/artifacts`
- `GET /v0/artifacts/{id}`
- `GET /v0/verify/{id}`
- `GET /v0/chain/{id}`
- `GET /v0/export/markdown`
- `GET /v0/export/json`
- `GET /v0/export/html`
- `GET /v0/rehydrate`

Deferred route groups:

- `/v0/projects`
- `/v0/checkpoints`

## Runtime Capabilities

- `yanzi serve` starts the shared runtime explicitly
- runtime startup and shutdown are foreground and deterministic
- the runtime is optional and does not alter CLI usage
- health output includes provider readiness and runtime visibility when active

## Operational Guarantees

- CLI primacy remains intact
- API handlers delegate through existing provider and library boundaries
- deterministic serialization and error envelopes are preserved
- runtime behavior does not imply orchestration or workflow continuation

## Local-First Guarantees

- the shared runtime is optional
- the local SQLite-backed provider remains the source of truth for current semantics
- no endpoint requires daemonization or distributed coordination
- existing CLI workflows continue to operate independently of runtime mode

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

## Upgrade Compatibility Expectations

- no breaking CLI behavior was introduced in CAP-002
- the API surface is additive and route-scoped
- deterministic operational semantics remain stable across minor and patch releases
- the next release should be a minor bump from `2.9.0` to `2.10.0`

