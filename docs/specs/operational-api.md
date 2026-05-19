# Operational API

## Purpose

The Operational API exposes Context Library operations through a stable operational interface.

The CLI remains first-class. The API is an additional interface over the same contract, not a replacement for CLI workflows.

The API enables integrations across REST clients, future UI, MCP adapters, connector runtimes, shared/team deployments, and optional runtime/daemon layers.

## Relationship to Existing Contracts

This specification derives from and depends on:

- [Context Library Contract](context-library-contract.md)
- [Storage Provider Contract](storage-provider-contract.md)

API semantics derive from the Context Library contract. Datastore behavior and backend-specific details must not leak into API semantics.

## Non-Goals

The Operational API is not:

- Orchestration.
- Workflow automation.
- An AI runtime.
- An agent execution engine.
- A distributed scheduler.
- A Kubernetes control plane.
- A replacement for GitHub, Jira, or CI systems.

## Core Principles

- Interface neutrality: API semantics match CLI/MCP/UI semantics for the same operations.
- Deterministic behavior: equivalent requests with equivalent state produce equivalent operational outcomes.
- Provenance preservation: write operations preserve artifact provenance and attribution.
- Append-only operational lineage: lineage expands through appended records and explicit links.
- Explicit operations: actions are explicit API calls, not hidden side effects.
- Source attribution: each write is attributable to originating interface and actor metadata.
- Human-governed authority: the API records and serves state; it does not assume autonomous control.
- Transport independence in the future: protocol evolution must preserve contract semantics.

## API Architecture Direction

- REST is the initial API model.
- JSON payloads are expected initially.
- Future transports may include gRPC, event streams, or other protocols.
- Transport changes must not change operational semantics.

This is a contract direction statement, not a promise that all transports exist today.

## Canonical Resource Types

Conceptual API resources include:

- `artifacts`
- `projects`
- `checkpoints`
- `packs`
- `lineage`
- `exports`
- `providers`
- `runtime` or `node` status
- future federation endpoints

Concrete resource shapes may evolve, but contract semantics remain stable.

## Artifact Operations

Conceptual operations include:

- create artifact
- retrieve artifact
- list artifacts
- filter artifacts
- verify artifact
- retrieve lineage
- relate artifacts
- import artifacts
- export artifacts

Conceptual endpoint naming conventions (not final API):

- `POST /artifacts`
- `GET /artifacts/{id}`
- `GET /artifacts`
- `POST /artifacts/{id}/verify`
- `GET /artifacts/{id}/lineage`
- `POST /artifacts/relations`
- `POST /artifacts/import`
- `POST /artifacts/export`

Behavior expectations:

- Writes append operational history and preserve attribution.
- Reads reflect recorded lineage and project scope.
- Import/export operations preserve provenance continuity and report conflicts explicitly.

## Project Operations

Conceptual operations include:

- create project
- list projects
- resolve current/default project semantics
- retrieve project lineage scope
- read/write project metadata

Behavior expectations:

- Artifact operations are project-scoped unless explicitly cross-project.
- Default project resolution is explicit and inspectable.
- Project metadata must not override artifact lineage semantics.

## Checkpoint Operations

Conceptual operations include:

- create checkpoint
- list checkpoints
- retrieve checkpoint summaries
- rehydrate from checkpoint
- compose post-checkpoint state deterministically

Behavior expectations:

- Checkpoints are deterministic anchors.
- Post-checkpoint composition uses explicit lineage and ordering rules.
- Checkpoint summaries expose enough information to support operational review.

## Rehydration Operations

Conceptual rehydration operations should support:

- project-scoped rehydration
- checkpoint-scoped rehydration
- scope options for artifact types, ranges, and lineage depth
- exportable rehydration context bundles

Behavior expectations:

- Rehydration reconstructs context from explicit artifacts and lineage.
- Rehydration behavior is deterministic for equivalent scope and state.
- Rehydration does not infer missing lineage links.

## Import / Export Operations

Conceptual import/export operations should enforce:

- deterministic export behavior for equivalent artifact sets and format rules
- future format extensibility without semantic drift
- provenance preservation across export and re-import
- compatibility with pack install/import workflows
- lineage-safe import behavior with explicit conflict handling

Import/export is a contract surface and may be implemented by multiple interfaces.

## Lineage Traversal Operations

Lineage operations should support:

- parent/child traversal
- related artifact traversal
- project-scoped traversal
- checkpoint traversal
- graph exploration concepts for operational inspection

Lineage is explicit, not inferred magically. Traversal results are derived from recorded links.

## Query and Filtering Semantics

Conceptual filtering support includes:

- `type` and `subtype`
- `project`
- `author`
- `source` or `interface`
- metadata fields
- `checkpoint`
- `created_at` ranges
- `status`
- lineage relationships

Filtering behavior should remain deterministic and consistently documented across interfaces.

## Source Attribution

- Every write operation should preserve source/interface attribution.
- CLI, API, MCP, UI, and connectors must remain distinguishable in provenance records.
- Provenance continuity must survive imports, exports, and synchronization flows.

## Runtime and Deployment Semantics

- The API may run embedded or via an optional runtime/daemon.
- Local CLI usage may bypass the API entirely.
- Enterprise/shared deployments are likely to use a runtime-hosted API.
- API hosting must not become mandatory for local-first operation.

## Authentication and Authorization Direction

Future direction:

- Authentication likely belongs in runtime/daemon layers.
- RBAC likely belongs above provider layers.
- API should support future auth headers/tokens.
- Local embedded usage may remain lightweight.

This section is directional and does not claim current full auth implementation.

## Multi-Agent Shared State Semantics

Intended shared-state flow:

1. ChatGPT writes a prompt artifact.
2. CodeX or Claude retrieves that prompt artifact.
3. The executing agent records a result artifact.
4. Lineage remains traceable for continuation.

Clarifications:

- This is shared operational state.
- This is not orchestration.
- Yanzi does not schedule or direct agents.

## Error and Failure Semantics

- Failures must be explicit.
- Error reporting should be deterministic and inspectable.
- No silent lineage corruption.
- Verification operations should expose integrity/provenance failure details.
- Import/export failures must be visible with actionable outcomes.

## Versioning Strategy

- API versioning should provide explicit compatibility boundaries.
- Backward compatibility is a goal for stable contract surfaces.
- Contract stability takes precedence over transport-specific convenience.
- Semantic evolution must avoid lineage breakage.

## Future Compatibility

This contract is intended to remain compatible with:

- MCP
- federation
- connector runtime
- packs
- UI
- IDE extensions
- enterprise runtime hosting

## Summary

The Operational API exposes the Context Library through a stable interface. The API preserves operational lineage and provenance continuity. The API is an interface layer, not an orchestration system.
