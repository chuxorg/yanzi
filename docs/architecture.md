# Architecture

## Problem

AI-assisted work needs durable state without depending on orchestration services or hidden runtime behavior.

## Solution

Yanzi is a local-first append-only datastore with a CLI over that state. It exposes that same state through an optional local HTTP API for agent integration.

## Local-First

- Default storage is local SQLite under the Yanzi state directory (`~/.yanzi/yanzi.db`)
- Local mode requires no background services
- HTTP mode is optional and does not change the local-first model — it routes CLI operations to a running `yanzi serve` instance rather than opening SQLite directly

## Append-Only

- Intent records are appended with SHA-256 hashes over a canonical JSON preimage
- Each record optionally links to a predecessor via `prev_hash`, forming a verifiable chain
- Checkpoints add project boundaries without rewriting earlier records
- Context and intent artifacts are added as new records; deletion leaves a tombstone

## Storage Provider Abstraction

Introduced in CAP-001. All storage access flows through the `internal/storage.Provider` interface rather than hitting SQLite directly:

```text
Provider interface
  ├── ArtifactOperations   — create, list, get artifacts
  ├── ProjectOperations    — create, list, check projects
  ├── CheckpointOperations — create, list checkpoints
  ├── VerificationOperations — hash verification
  └── ImportExportOperations — export rendering

Implementations
  └── internal/storage/sqlite/  — SQLite-backed provider (default)

Registry
  └── internal/storage/registry/ — provider selection at startup
```

The `SQLDB()` method on the interface preserves existing call sites that need raw SQL access during the CAP-001 transition. Future phases can narrow those callers behind the interface without changing CLI contracts.

## HTTP Runtime

Introduced in CAP-002. `yanzi serve` starts a shared operational API server at `127.0.0.1:8080` (default):

```text
yanzi serve
  └── internal/runtime.Runtime
        └── internal/api/routes  — registers all /v0 handlers
              ├── GET  /v0/health
              ├── GET  /v0/rehydrate
              ├── GET  /v0/artifacts
              ├── POST /v0/artifacts
              ├── GET  /v0/artifacts/:id
              ├── GET  /v0/verify/:id
              ├── GET  /v0/chain/:id
              ├── GET  /v0/export/:format
              ├── GET  /v0/projects
              ├── POST /v0/projects
              ├── GET  /v0/projects/current
              ├── POST /v0/projects/current
              ├── GET  /v0/checkpoints
              └── POST /v0/checkpoints
```

The server binds to localhost only. It carries no authentication or TLS; callers that need these must add a local proxy. All responses use `application/json`; all timestamps are RFC3339Nano UTC.

CLI commands `checkpoint`, `rehydrate`, `export`, `delete`, and `restore` remain local-only regardless of mode. See [API Reference](api/index.md).

## Datastore Only

- Yanzi stores and retrieves project state
- Filtering is deterministic and explicit
- It does not rank, summarize, or orchestrate agents

## No Interpretation

- Stored content is not interpreted by Yanzi
- Metadata is matched exactly when filters are used
- Meaning is left to the caller or agent using the stored data

## Delivery Authority

Repository delivery governance preserves separation of authority:

```text
Architect       -> Capability
Release Steward -> Phase Approval
Execution Agent -> Delivery
QA              -> Validation
Release Steward -> Release Decision
```

The Release Steward governs PR review, merge approval, checkpoint approval, release readiness, backlog state transitions, and convergence validation. The role does not redefine architecture, orchestrate agents, override branch protections, autonomously release, or invent requirements.

See [Release Steward Role](roles/release-steward.md).
