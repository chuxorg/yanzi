# CAP-002 Readiness

CAP-002 is the Operational API arc for Yanzi. It now has a stabilized local-first foundation with optional shared runtime access.

## Completed Capabilities

- deterministic artifact capture through an internal write boundary
- artifact read access through the internal read boundary
- verification and chain read endpoints
- deterministic export read endpoints
- deterministic rehydration read endpoint
- explicit foreground runtime bootstrap via `yanzi serve`
- health visibility for provider readiness and runtime state

## Runtime Capabilities

- `yanzi serve` starts the shared runtime explicitly
- startup and shutdown are foreground and deterministic
- runtime behavior is optional and does not alter CLI usage
- runtime health visibility is informational only

## API Capabilities

- `GET /v0/health`
- `POST /v0/artifacts`
- `GET /v0/artifacts/{id}`
- `GET /v0/verify/{id}`
- `GET /v0/chain/{id}`
- `GET /v0/export/markdown`
- `GET /v0/export/json`
- `GET /v0/export/html`
- `GET /v0/rehydrate`

## Operational Guarantees

- local-first operation remains canonical
- deterministic serialization is preserved
- response envelopes stay machine-readable
- API handlers delegate through existing boundaries instead of duplicating storage logic

## Deferred Enterprise Features

- auth/RBAC
- federation
- MCP
- Postgres
- orchestration behavior
- distributed coordination
- connector runtime
- plugin hosting
- enterprise control-plane features

## Upgrade Compatibility Expectations

- no breaking CLI behavior was introduced in CAP-002
- the API surface is additive and route-scoped
- existing local workflow semantics remain stable
- future work must continue to preserve deterministic provider-backed behavior
