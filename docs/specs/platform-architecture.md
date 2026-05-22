# Platform Architecture

## Purpose

This document is the canonical architecture reference for Yanzi.

It consolidates completed Phase 1-9 specifications into one operational model, unifies terminology, defines layer boundaries, reduces overlap, establishes implementation sequencing, explains deployment models, and frames product packaging direction.

Yanzi is operational context infrastructure for deterministic context composition and governance-first engineering workflows.

Yanzi is not orchestration infrastructure, autonomous engineering infrastructure, AI infrastructure, workflow automation infrastructure, or agent execution infrastructure.

## Relationship to Prior Specs

This architecture consolidates and organizes the following contracts:

- [Context Library Contract](context-library-contract.md)
- [Storage Provider Contract](storage-provider-contract.md)
- [Operational API](operational-api.md)
- [MCP Interface](mcp-interface.md)
- [Federation Protocol](federation-protocol.md)
- [Pack Specification](pack-specification.md)
- [Connector Runtime](connector-runtime.md)
- [Runtime / Daemon Architecture](runtime-architecture.md)
- [UI Operational Model](ui-operational-model.md)

Normative precedence remains with the individual contracts for interface-specific details. This document provides a single consolidated architecture view.

## Architecture Overview

Yanzi is a layered architecture around a canonical Context Library boundary.

Operational state is represented as explicit artifacts with append-only lineage and provenance continuity. Interfaces, runtime services, and product experiences are adapters around this boundary and must not redefine artifact semantics.

## Canonical Layer Model

Layer 1 - Context Library (Core)

- canonical artifact semantics
- lineage and provenance continuity
- deterministic import/export and rehydration behavior
- project-scoped operational history

Layer 2 - Interfaces

- CLI
- REST
- MCP
- import/export surfaces

Layer 3 - Runtime Services

- connectors
- federation synchronization services
- API hosting
- auth/RBAC direction (future)

Layer 4 - Product Experiences

- UI
- IDE integrations
- packs
- operational tooling

Layer dependency rule:

- Upper layers depend on lower-layer semantics.
- Lower layers do not depend on upper-layer product choices.

## Context Library as System Boundary

Everything reduces to Context Library operations.

Whether state enters through CLI, REST, MCP, connector ingestion, pack install, federation sync, or UI action, the authoritative effect is a Context Library artifact operation with explicit lineage and provenance.

## Deployment Models

Supported conceptual deployment models:

- Local standalone: CLI + local provider, no runtime required.
- Team runtime: shared runtime hosting REST/MCP and optional shared providers.
- Enterprise runtime: governed runtime deployment with organization controls.
- Federated nodes: independently authoritative nodes exchanging artifacts.
- Offline operation: disconnected local operation with later explicit synchronization.

## Product Packaging Strategy

Conceptual packaging direction (not a committed release matrix):

`yanzi-community`

- CLI
- SQLite baseline provider
- packs
- export/import workflows

`yanzi-runtime`

- REST hosting
- MCP hosting
- federation services
- connector runtime hosting

Future optional packages/extensions:

- UI
- IDE integrations

Packaging is conceptual and future-facing, and does not change canonical contract semantics.

## Interface Map

Conceptual interface matrix:

| Interface | Reads | Writes | Typical User |
| --- | --- | --- | --- |
| CLI | artifacts, checkpoints, projects, lineage views | artifacts, checkpoints, imports/exports | individual engineer, local operator |
| REST | project-scoped Context Library state | contract-defined artifact/project/checkpoint operations | services, internal tools, runtime clients |
| MCP | shared operational context artifacts | prompt/intent/result-style artifacts and related operations | MCP-capable AI tools and assistants |
| UI | runtime/API-exposed state, lineage, health, governance views | explicit governed operational actions | engineering teams, leads, governance operators |
| Connector | external systems and connector state | normalized imported artifacts and sync metadata | integration operators, platform teams |
| Pack Installer | pack manifests/artifacts | deterministic Context Library imports with provenance | team leads, onboarding/governance maintainers |

## Artifact Lifecycle

Canonical lifecycle path:

`seed -> intent -> execution -> result -> checkpoint -> export -> federation -> rehydration`

Lifecycle notes:

- Not every workflow uses every step.
- Lineage links and provenance metadata remain explicit.
- Rehydration is deterministic for equivalent scope and state.

## Governance Model

Governance model principles:

- bounded operational authority
- human-governed decisions
- convergence validation against explicit lineage/rules
- append-only operational lineage evidence

Yanzi records, validates, and exposes operational context; it does not autonomously decide or orchestrate execution.

## Product Positioning

Why Yanzi exists:

AI-assisted engineering workflows often lose continuity across prompts, tools, environments, and handoffs.

Yanzi preserves operational state as explicit, inspectable artifacts with deterministic reconstruction and provenance continuity.

## Implementation Sequencing

Recommended execution sequence for productization:

- Phase A: storage abstraction and provider conformance hardening
- Phase B: REST interface alignment with operational contract
- Phase C: pack install/import semantics implementation
- Phase D: runtime/daemon hosting baseline
- Phase E: MCP interface hosting and shared-state workflows
- Phase F: connector runtime integration surfaces
- Phase G: UI operational model implementation

This sequence is directional and should be refined through implementation feedback and ADR decisions.

## Deferred Topics

Explicitly deferred architecture topics:

- authentication details
- RBAC model details
- search architecture
- signatures and trust verification details
- registry ecosystem details
- IDE implementation details
- monetization model
- cloud offering model

## Architecture Decision Log

ADR placeholders for future decisions:

- ADR-001: canonical artifact identity and digest strategy across providers
- ADR-002: runtime auth and RBAC boundary model
- ADR-003: pack signature and trust verification model
- ADR-004: connector normalization profile strategy
- ADR-005: federation conflict and supersession presentation model
- ADR-006: UI mutation guardrails and governed destructive operations
- ADR-007: multi-provider routing and capability negotiation in runtime

## Consolidation Notes (Inconsistencies and Alignment)

Observed cross-spec inconsistencies/deferred alignments:

- Pack representation appears in two forms: as distributable bundle semantics ([Pack Specification](pack-specification.md)) and as potential pack-related artifact entries in core taxonomy ([Context Library Contract](context-library-contract.md)). Consolidated interpretation: pack distribution is external packaging; installation materializes Context Library artifacts with optional pack-manifest lineage evidence.
- Some interface documents list conceptual operations (for example `install_pack` under MCP) ahead of implementation sequencing. Consolidated interpretation: these are future-facing contract directions, not implementation claims.
- UI and runtime specs define broad visibility/health concepts without finalized endpoint schemas. Consolidated interpretation: visibility surfaces remain contract goals pending dedicated observability/auth ADRs.

No contradiction changes are applied retroactively in prior specs; this section records alignment guidance for implementation.

## Summary

Yanzi preserves deterministic operational context while allowing optional interfaces, runtimes, and integrations around a stable Context Library.
