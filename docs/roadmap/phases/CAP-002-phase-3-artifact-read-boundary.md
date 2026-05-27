# CAP-002 Phase 3 — Artifact Read Service Boundary

## Scope

Phase 3 introduces a minimal internal read boundary for current artifact-adjacent `list` and `show` behavior without exposing public artifact API endpoints.

This phase covers:

- current `yanzi list` and `yanzi show` read behavior
- current artifact lookup-by-ID behavior for those commands
- current project-scoped list behavior
- ordering, metadata shaping, and deleted-record visibility parity
- a lightweight internal boundary future API handlers can call

## Non-Goals

This phase does not introduce:

- artifact API endpoints
- capture or other artifact write endpoints
- rehydration endpoint work
- export endpoint work
- verification endpoint work
- tombstone mutation refactors
- auth, runtime hosting, orchestration, or provider switching

## Implementation Boundaries

The new seam is intentionally narrow:

- `internal/library/artifact_read_store.go` isolates the local SQL-backed read path used by current `list` and `show`
- the CLI remains the canonical caller in this phase
- `/v0/artifacts` remains a deterministic deferred placeholder

The boundary preserves current behavior rather than redesigning it.

## Current Behavior Audit

The existing local read path behaves as follows and is preserved in this phase:

- `yanzi list` is scoped to the active project unless `--all-projects` is set
- `yanzi list` injects `project=<active-project>` into metadata filters for scoped reads
- `yanzi list` rejects conflicting explicit `--meta project=...` values unless `--all-projects` is used
- `yanzi list` excludes rows where `source_type = 'artifact'`
- `yanzi list` hides tombstoned rows by default and includes them only with `--include-deleted`
- `yanzi list` orders results by `created_at DESC, id DESC`
- `yanzi list` displays only exportable metadata fields in the summary column
- `yanzi show` reads by record ID and does not apply tombstone filtering
- `yanzi show` preserves the current `meta` column behavior and does not merge fallback values from the legacy `metadata` column

## Acceptance Criteria

Phase 3 is complete when:

- artifact read/list/show boundary exists
- CLI `list` and `show` behavior remains unchanged
- parity tests cover current list/show behavior
- no artifact API endpoints are exposed
- `/v0/artifacts` remains deferred
- remaining non-read artifact debt is documented

## Deferred Artifact Debt

The following work remains intentionally deferred after this phase:

- capture writes
- public artifact API endpoints
- artifact mutation and tombstone paths
- rehydration reads
- export and verification endpoint work

## Implementation Status

Completed in this phase:

- added parity tests for `show` formatting, missing-ID behavior, and deleted-record visibility by ID
- introduced `ArtifactReadStore` as the internal read boundary
- routed CLI local `list` and `show` reads through the new boundary
- kept `/v0/artifacts` unchanged as a deferred placeholder

## Validation

Validation performed for this phase:

- `go test ./...`
- `go vet ./...`
- `go build -o bin/yanzi ./cmd/yanzi`
- `make docs-build`
- representative CLI regression flow covering project create/list/use/current, capture, list, show, checkpoint create/list, export markdown/json/html, verify, and rehydrate

## Recommendation For CAP-002 Phase 4

Use the new read boundary as the only entry point for future artifact list/show API handlers.

Phase 4 should either:

- expose read-only artifact endpoints through this seam, or
- continue by extracting the remaining artifact write and tombstone debt before any public artifact API surface is added
