# Context Library Contract

## Purpose

The Context Library is the canonical operational boundary of Yanzi.

All interfaces, including CLI, REST, MCP, runtime layers, packs, connectors, and UI surfaces, read from or write to this contract. The contract defines shared semantics for artifacts, lineage, provenance, and deterministic rehydration across interfaces.

Yanzi remains deterministic, local-first, append-only, and human-governed. This document is a contract specification for future-facing architecture and cross-interface consistency. It is not a claim that every capability described here is fully implemented today.

## Non-Goals

The Context Library is not:

- AI.
- Orchestration.
- Workflow automation.
- An autonomous agent framework.
- A distributed coordination system.
- A replacement for Git, Jira, CI/CD, or documentation systems.

## Core Principles

- Artifacts are the atomic unit: all operational state is represented as explicit artifacts.
- Append-only provenance: new operations append lineage evidence rather than rewriting history.
- Deterministic identity: identity and relationships are stable and reproducible from explicit inputs.
- Human-governed operational authority: people and external systems decide actions; Yanzi records and validates.
- Local-first operation: local usage is the default and must remain viable without remote dependencies.
- Federation over distribution: independent libraries can exchange artifacts without requiring one global coordinator.
- Interfaces are adapters, not the product: interfaces expose the same contract and must not redefine semantics.
- Storage is an implementation detail: storage backends implement the contract but do not define it.

## Canonical Concepts

- Context Library: the contract-defined corpus of operational artifacts and lineage within a project scope.
- Artifact: the atomic append-only record unit representing intent, result, context, governance, or packaging state.
- Intent: an artifact representing an explicit requested action, objective, or operational prompt.
- Result: an artifact representing execution output, decision outcome, or observed consequence.
- Context: an artifact representing background information used to shape intent or interpretation.
- Rule: an artifact representing constraints, governance boundaries, or deterministic policy.
- Role: an artifact representing bounded operational authority and responsibility boundaries.
- Workflow: an artifact describing stepwise operational process and validation expectations.
- Seed: an artifact used to initialize baseline context for a project or operation domain.
- Pack: a composed artifact set intended for deterministic import and reusable operational context.
- Project: a scoped operational namespace for artifacts, lineage, and validation boundaries.
- Checkpoint: a time-anchored artifact used as a deterministic rehydration baseline.
- Rehydration: deterministic reconstruction of usable context from checkpointed and subsequent artifacts.
- Operational Lineage: explicit recorded relationships among artifacts across intents, results, context, and checkpoints.
- Provenance Continuity: preservation of origin and relationship evidence across append, import, export, and federation operations.
- Convergence Validation: explicit checks that multiple views or channels resolve to consistent lineage.
- Bounded Operational Authority: explicit limits on what actors may author, modify, install, or promote.

## Artifact Contract

The canonical artifact contract is conceptual and semantic. It is not necessarily the final physical database schema.

Expected artifact fields include:

- `id`: deterministic artifact identity within contract rules.
- `type`: primary semantic category (for example `intent`, `result`, `context`).
- `subtype`: finer-grained classification within a type.
- `title`: concise operator-readable label.
- `body` or `content`: canonical payload.
- `author`: originating actor identity (human, service, or tool identity label).
- `source` or `interface`: interface/system that authored the artifact (`cli`, `mcp`, `rest`, `ui`, connector id).
- `project`: project scope identifier.
- `parent_refs` and `related_refs`: explicit links to parent and related artifacts.
- `metadata`: structured extension fields that do not alter core semantics.
- `created_at`: canonical creation timestamp.
- `hash` or `digest`: deterministic integrity value where possible.
- `status` (where applicable): lifecycle or governance state.

Contract requirements:

- Core semantics are stable even if transport formats evolve.
- Optional fields may vary by artifact type, but provenance-critical fields must remain available.
- Import/export implementations must preserve contract meaning even when backend schemas differ.

## Artifact Type Semantics

- `intent`: declares requested work or decision target and may reference governing context/rules.
- `result`: records outcome of an executed intent or validation operation and links back to source intent.
- `context`: captures supporting knowledge, assumptions, constraints, or external references.
- `rule`: captures enforceable or advisory governance expectations with explicit scope.
- `role`: captures authority boundaries, ownership, and permitted operational actions.
- `workflow`: captures ordered operational procedure and associated validation expectations.
- `seed`: captures initial reusable baseline context for project bootstrapping.
- `checkpoint`: captures a deterministic anchor for subsequent rehydration composition.
- Pack-related artifacts (`pack`, `pack-manifest`, or equivalent): capture reusable composed context units and their lineage-preserving import semantics.

## Lineage Model

Lineage is explicit data, not hidden behavior.

- Parent/child relationships are recorded through explicit references.
- Prompt/intent to result relationships are recorded directly and remain traceable.
- Checkpoints anchor lineage at specific states for deterministic replay/rehydration.
- Lineage is project-scoped by default to avoid ambiguous cross-project inheritance.
- Artifact graph traversal is deterministic over recorded references.
- Lineage is recorded, not inferred magically; absent links are treated as absent lineage.

## Deterministic Import / Export Guarantees

- Imports preserve provenance metadata and relationship references.
- Exports are reproducible from the same artifact set and serialization rules.
- Imported artifacts must not silently overwrite existing lineage.
- Deterministic hashes/digests remain verifiable where canonical data is unchanged.
- Pack installs are Context Library imports and must follow the same provenance and collision rules.

## Rehydration Semantics

Rehydration is deterministic composition of operational context from recorded artifacts.

- A checkpoint provides a baseline context anchor.
- Post-checkpoint artifacts are composed in deterministic order using explicit lineage and scope rules.
- Yanzi guarantees that rehydration behavior is contract-driven, inspectable, and based on recorded artifacts.
- Yanzi does not guarantee semantic correctness of human-authored content, external system availability, or autonomous conflict resolution beyond defined contract rules.

## Multi-Agent Shared State Semantics

The Context Library supports shared operational state across tools and agents.

Example MCP-style flow:

1. ChatGPT writes an `intent` artifact (prompt/request).
2. CodeX or Claude loads that intent artifact from the shared library.
3. The executing agent records a corresponding `result` artifact.
4. ChatGPT reviews lineage and continues from explicit recorded state.

Contract clarification:

- This is shared operational state with explicit provenance.
- This is not orchestration.
- Yanzi does not decide which agent acts next.

## Storage Independence

- SQLite remains the default local provider.
- Future providers may include Postgres or object-store-backed implementations.
- Datastore changes must not change Context Library semantics.
- Storage providers implement the contract; they do not define it.

## Interface Independence

- CLI, REST, MCP, UI, and connectors are interfaces over the same contract.
- No interface receives special semantic authority.
- Interfaces must preserve provenance and source attribution when reading/writing artifacts.

## Governance and Authority

- Yanzi records operational intent and results.
- Yanzi may validate convergence against explicit rules and recorded lineage.
- Yanzi does not execute autonomous decisions.
- Humans and external tools retain operational authority.
- Roles may constrain editing, installation, or promotion behavior in future runtime/UI layers.

## Future Compatibility

This contract is intended to support:

- storage providers
- REST API
- MCP interface
- federation protocol
- pack specification
- connector runtime
- optional runtime/daemon
- UI operational management

## Summary

The Context Library is the stable operational substrate. Everything else is an interface, provider, package, or runtime around it.
