# CAP-002 Phase 1 - API Foundation and Routing

## Scope

CAP-002 Phase 1 establishes the foundational Operational API seam while preserving Yanzi's current local-first architecture.

This phase introduces:

- internal API package structure
- lightweight HTTP server foundation
- routing foundations for future endpoint groups
- canonical request and response models
- handler boundaries for provider-backed health/status reads
- deterministic placeholder behavior for deferred route groups

The CLI remains the canonical interface. No existing CLI behavior, provider behavior, lineage semantics, or local storage guarantees are changed.

## Non-Goals

This phase does not include:

- full artifact CRUD endpoints
- runtime daemonization
- auth or RBAC
- federation
- MCP
- Postgres
- connector hosting
- distributed deployment
- orchestration behavior
- migration of remaining SQLDB-backed operational paths

## Implementation Boundaries

The API foundation is intentionally lightweight:

- standard library `net/http`
- no large framework
- no dependency injection system
- internal-only server construction
- local-only friendly behavior

Implemented route groups:

- `/v0/health`
- `/v0/artifacts`
- `/v0/projects`
- `/v0/checkpoints`

Implemented endpoint:

- `GET /v0/health`

Deferred route groups are explicitly registered and return deterministic deferred responses so future work can add endpoints without changing routing topology.

## Provider Integration Approach

The health handler delegates into existing configuration and storage provider seams.

Phase 1 deliberately avoids:

- redesigning `SQLDB` usage
- bypassing provider seams
- adding parallel persistence logic

The server foundation uses existing local provider construction for health reads and does not attempt to migrate deferred CAP-001 carry-forward paths.

## CAP-001 Carry-Forward Debt

The following paths remain intentionally outside this phase:

- capture writes
- list/show reads
- rehydration reads
- tombstone mutation paths

These are deferred to later CAP-002 endpoint phases and must not be force-migrated here.

## Acceptance Criteria

Acceptance requires:

- API foundation exists under `internal/api`
- routing foundation exists
- handler boundaries exist
- health endpoint foundation exists
- provider integration is preserved
- CLI behavior remains unchanged
- no runtime or daemon complexity is introduced
- no orchestration semantics are introduced
- SQLDB carry-forward debt is explicitly documented

## Deferred Work

Deferred beyond Phase 1:

- artifact endpoints
- project endpoints
- checkpoint endpoints
- rehydration endpoints
- full status or admin surfaces
- authentication and authorization
- hosted runtime behavior
- non-SQLite provider concerns

## Validation Performed

Required validation for this phase:

- `go test ./...`
- `go vet ./...`
- `go build -o bin/yanzi ./cmd/yanzi`
- `make docs-build`

Additional focus:

- existing CLI behavior unchanged
- provider abstraction remains stable
- no migration regressions
- no export regressions
- no verification regressions

Validation status:

- passed

Validation findings:

- `go test ./...` passed
- `go vet ./...` passed
- `go build -o bin/yanzi ./cmd/yanzi` passed
- `make docs-build` passed with existing MkDocs informational warnings about nav entries and MkDocs 2.0 compatibility
- existing CLI behavior remained unchanged under the full existing command and storage test suite
- provider abstraction remained stable and no migration, export, or verification regressions were introduced by the API foundation changes
