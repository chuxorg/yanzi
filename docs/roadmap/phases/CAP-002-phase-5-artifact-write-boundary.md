# CAP-002 Phase 5 — Artifact Write Boundary

## Scope

Phase 5 extracts and hardens the internal write boundary for current capture, artifact creation, and tombstone mutation behavior without exposing public mutation endpoints.

This phase covers:

- current `yanzi capture` local write behavior
- current intent and context artifact creation behavior
- current metadata and author/source persistence behavior
- current hash and digest generation behavior
- current `prev_hash` lineage persistence behavior
- current tombstone and restore mutation behavior
- provider-compatible artifact creation writes where available
- parity with existing CLI capture semantics and output

## Non-Goals

This phase does not introduce:

- public artifact mutation endpoints
- public capture endpoints
- public tombstone APIs
- rehydration endpoints
- runtime daemonization
- auth or RBAC
- federation
- MCP
- Postgres
- orchestration behavior
- alternate capture, lineage, or hash semantics

## Current Behavior Audit

The existing local write behavior before refactoring was:

- `yanzi capture` accepted exactly one prompt source and exactly one response source, rejected stdin, required `--author`, defaulted `--source` to `cli`, and printed only `id:` and `hash:`
- capture metadata came from repeated `--meta key=value` flags, optional `--profile`, and the active project when present
- duplicate capture metadata keys used last-value-wins behavior
- capture rows used random 16-byte hex IDs and UTC RFC3339Nano timestamps
- capture hashes used `HashIntent` over the current intent preimage, including canonicalized metadata and optional `prev_hash`
- capture persisted both the canonical intent fields and artifact-compatible columns: `class = intent`, `type = prompt`, `content = prompt`, and `metadata = meta`
- capture writes updated the local `last_hash` state file after successful persistence
- artifact creation used provider-backed SQLite artifact writes, with `author = yanzi`, `source_type = artifact`, artifact system metadata in `meta`, caller metadata in `metadata`, no `prev_hash`, and the existing artifact hash preimage
- intent artifact creation required a project; project-scoped context creation required a project; global context creation did not
- tombstone behavior wrote `deleted=true` and `deleted_at=<UTC RFC3339Nano>` to `metadata` for normal captures and to `meta` for `source_type = artifact` rows
- restore removed `deleted` and `deleted_at` from the same column selected by tombstone behavior
- cascade tombstone traversal followed `prev_hash` descendants in creation order
- checkpoint-referenced artifacts remained protected unless `--force` was supplied

## Implementation Boundaries

The new seam is intentionally narrow:

- `internal/library/artifact_write_store.go` is the internal write boundary for capture, artifact creation, tombstone, and restore writes
- current CLI capture/delete/restore flows delegate through `ArtifactWriteStore`
- current library artifact creation delegates through `ArtifactWriteStore`
- artifact creation inside the boundary continues to use the provider-compatible SQLite write path
- direct SQL remains inside the boundary only where current capture and tombstone behavior has no provider contract yet
- `/v0/artifacts` remains a deterministic deferred placeholder

The boundary preserves current behavior rather than redesigning capture, lineage, metadata, or tombstone semantics.

## Acceptance Criteria

Phase 5 is complete when:

- artifact write boundary exists
- capture behavior is unchanged
- deterministic hashing is preserved
- lineage semantics are preserved
- CLI behavior is unchanged
- no public mutation APIs are introduced
- provider-compatible artifact writes are routed through the boundary
- remaining SQLDB-backed write debt is isolated and documented
- parity tests cover capture, artifact creation, metadata persistence, hashing, lineage, duplicate captures, malformed metadata, and tombstone behavior

## Remaining Deferred Debt

The following work remains intentionally deferred after this phase:

- public artifact mutation endpoints
- public capture endpoints
- public tombstone APIs
- provider contract methods for capture writes
- provider contract methods for tombstone and restore mutations
- portions of rehydration reads
- public rehydration endpoints
- runtime daemonization, auth/RBAC, federation, MCP, Postgres, and orchestration behavior

## Implementation Status

Completed in this phase:

- added parity coverage for current capture persistence, duplicate capture handling, deterministic hash recomputation, artifact-compatible capture columns, artifact creation persistence, malformed metadata, and tombstone column behavior
- introduced `ArtifactWriteStore` as the internal write boundary
- routed provider-backed artifact creation through the boundary
- routed CLI local capture through the boundary
- routed CLI delete and restore through the boundary
- removed the command-local tombstone mutation implementation
- kept public artifact, capture, and tombstone endpoints deferred

## Validation

Required validation for this phase:

- `go test ./...`
- `go vet ./...`
- `go build -o bin/yanzi ./cmd/yanzi`
- `make docs-build`
- representative CLI regression flow covering `capture`, `list`, `show`, `verify`, `export markdown/json/html`, and `rehydrate`

## Recommendation For CAP-002 Phase 6

Use `ArtifactWriteStore` as the only internal entry point for future write API work. Phase 6 should add provider contract methods for capture and tombstone writes or expose read-only API work already backed by stable boundaries before adding public mutation endpoints.
