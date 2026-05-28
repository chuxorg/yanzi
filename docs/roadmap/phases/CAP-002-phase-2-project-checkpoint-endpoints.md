# CAP-002 Phase 2 - Project and Checkpoint API Endpoints

## Scope

CAP-002 Phase 2 implements the first real Operational API endpoints using the provider-backed project and checkpoint operations established during CAP-001 Phase 2A.

This phase introduces:

- provider-backed project endpoints
- provider-backed checkpoint endpoints
- active-project API handling aligned with current local semantics
- deterministic JSON response conventions for successful and failed operations
- API testing patterns for real route execution against local provider behavior

The CLI remains the canonical interface. No existing CLI workflow, lineage behavior, export behavior, verification behavior, or rehydration behavior changes in this phase.

## Non-Goals

This phase does not include:

- artifact CRUD endpoints
- capture endpoints
- artifact list/show endpoints
- rehydration endpoints
- tombstone mutation behavior
- auth or RBAC
- runtime daemonization
- federation
- MCP
- Postgres
- orchestration behavior

## Implementation Boundaries

Implemented routes:

- `GET /v0/projects`
- `POST /v0/projects`
- `GET /v0/projects/current`
- `POST /v0/projects/current`
- `GET /v0/checkpoints`
- `POST /v0/checkpoints`

Artifact routes remain deferred placeholders and were not expanded in this phase.

Project endpoint behavior preserves current semantics:

- project creation uses existing local project creation behavior
- project listing preserves current ordering and serialization
- current-project reads and writes use the same active-project state semantics already used by the CLI
- setting the current project verifies that the target project exists before persisting state

Checkpoint endpoint behavior preserves current semantics:

- checkpoint listing uses the active project by default
- checkpoint listing supports `all_projects=true` for current `--all-projects` parity
- checkpoint creation uses the active project
- checkpoint creation does not introduce artifact or rehydration behavior

## Response Conventions

Success responses remain minimal and operation-shaped:

- collection endpoints return collection objects such as `projects` and `checkpoints`
- create endpoints return the created project or checkpoint object
- current-project reads and writes return a `project` object or `null`

Error responses use the deterministic JSON error envelope introduced in Phase 1:

- `{"error":{"code":"...","message":"..."}}`

Status code conventions in this phase:

- `200` for successful reads and current-project updates
- `201` for successful creates
- `400` for invalid JSON, missing required fields, active-project requirements, and semantic request conflicts
- `404` for missing projects and other not-found conditions
- `409` for duplicate project creation
- `501` remains reserved for deferred artifact route groups

## Deferred Artifact Endpoint Work

Artifact-related operational paths remain intentionally deferred:

- capture writes
- artifact list/show reads
- rehydration reads
- tombstone mutation paths
- artifact CRUD API endpoints

Phase 2 does not migrate or expose those paths.

## Acceptance Criteria

Acceptance requires:

- project endpoints are operational
- checkpoint endpoints are operational
- deterministic JSON responses are implemented
- provider-backed execution is preserved
- CLI behavior remains unchanged
- no artifact endpoints are introduced
- no orchestration semantics are introduced
- deferred SQLDB artifact debt is preserved and documented

## Validation Performed

Required validation for this phase:

- `go test ./...`
- `go vet ./...`
- `go build -o bin/yanzi ./cmd/yanzi`
- `make docs-build`

Additional focus:

- existing CLI behavior unchanged
- provider abstraction remains stable
- project and checkpoint behavior parity preserved
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
- existing CLI behavior remained unchanged under the current command, storage, export, verification, and rehydration test suites
- provider abstraction remained stable and project/checkpoint parity was preserved
- no migration, export, or verification regressions were introduced
