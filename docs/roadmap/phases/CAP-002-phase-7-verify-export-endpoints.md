# CAP-002 Phase 7 — Verification and Export Read Endpoints

## Scope

Phase 7 exposes verification and export read functionality through the Operational API while preserving the current deterministic local operational behavior.

This phase covers:

- `GET /v0/verify/{id}`
- `GET /v0/chain/{id}`
- `GET /v0/export/markdown`
- `GET /v0/export/json`
- `GET /v0/export/html`
- provider-backed verification reads
- provider-backed export timeline reads and deterministic rendering
- CLI regression coverage for verify, chain, and export behavior

## Non-Goals

This phase does not introduce:

- export mutation behavior
- rehydration endpoints
- artifact mutation APIs
- tombstone APIs
- runtime daemonization
- auth or RBAC
- federation
- MCP
- Postgres
- orchestration behavior
- alternate hash, lineage, or export semantics

## Implementation Boundaries

The new API surface is intentionally narrow:

- verification handlers delegate to provider-backed verification through shared library helpers
- export handlers delegate to provider-backed export reads through shared library rendering helpers
- handlers do not perform direct SQL reads or writes
- export endpoints require explicit `project` query scoping
- existing `profile`, `meta_*`, and `include_deleted` export filters are preserved
- checkpoint filtering remains unsupported because current CLI export semantics do not provide it
- artifact mutation, tombstone, and rehydration APIs remain deferred

## Acceptance Criteria

Phase 7 is complete when:

- verification endpoints are operational
- export endpoints are operational
- deterministic verification behavior is preserved
- deterministic export behavior is preserved
- CLI behavior remains unchanged
- no mutation APIs are introduced
- no orchestration semantics are introduced
- provider-compatible execution remains preserved

## Remaining Deferred Operational Debt

The following work remains intentionally deferred after this phase:

- artifact list endpoint implementation
- artifact update, delete, and tombstone endpoints
- rehydration endpoints
- project and checkpoint management endpoints
- provider contract methods for capture writes
- provider contract methods for tombstone and restore mutations
- runtime daemonization, auth/RBAC, federation, MCP, Postgres, and orchestration behavior

## Implementation Status

Completed in this phase:

- added `GET /v0/verify/{id}`
- added `GET /v0/chain/{id}`
- preserved compatibility with current HTTP CLI verification paths through `/v0/intents/{id}/verify` and `/v0/intents/{id}/chain`
- extracted shared verification helpers so CLI and API use the same verify and chain semantics
- extracted shared deterministic log export rendering so CLI and API use the same export behavior
- added `GET /v0/export/markdown`
- added `GET /v0/export/json`
- added `GET /v0/export/html`
- kept checkpoint filtering deferred because it is not part of current CLI export behavior

## Validation

Required validation for this phase:

- `go test ./...`
- `go vet ./...`
- `go build -o bin/yanzi ./cmd/yanzi`
- `make docs-build`
- representative CLI regression flow covering `capture`, `list`, `show`, `verify`, `chain`, `export markdown/json/html`, and `rehydrate`
- API validation for verify, chain, markdown export, JSON export, HTML export, project scoping, supported filters, and deterministic responses

## Recommendation For CAP-002 Phase 8

Phase 8 should focus on remaining read-side asymmetry only if it can stay inside existing library and provider seams. If the next step moves back to mutation work, add explicit provider-backed mutation contracts before exposing tombstone or broader artifact update APIs.
