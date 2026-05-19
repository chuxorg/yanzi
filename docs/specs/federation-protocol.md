# Federation Protocol

## Purpose

Yanzi may operate across multiple nodes and environments.

Federation allows nodes to exchange operational artifacts and lineage so teams, tools, environments, and runtime layers can maintain operational continuity without abandoning local-first workflows.

Federation preserves local-first operation: each node remains useful independently, including during disconnected operation.

## Relationship to Existing Contracts

This specification depends on:

- [Context Library Contract](context-library-contract.md)
- [Storage Provider Contract](storage-provider-contract.md)
- [Operational API](operational-api.md)
- [MCP Interface](mcp-interface.md)

Federation exchanges Context Library artifacts. Artifact semantics remain canonical under the Context Library contract. Storage/provider details must not redefine lineage semantics.

## Non-Goals

Federation is not:

- Orchestration.
- Cluster management.
- Distributed scheduling.
- Consensus infrastructure.
- Kubernetes-style control planes.
- Autonomous synchronization intelligence.
- Workflow automation.
- Shared mutable global state.

## Core Principles

- Federation over distribution.
- Independently authoritative nodes.
- Append-only provenance.
- Deterministic lineage preservation.
- Explicit synchronization.
- Source attribution preservation.
- Disconnected operation remains valid.
- Operational continuity over operational centralization.

## Federation Model

- Yanzi nodes exchange operational artifacts.
- Synchronization is additive.
- Lineage is propagated, not rewritten.
- Nodes remain operational when disconnected.
- No node becomes globally authoritative automatically.

Federation is an exchange mechanism, not a centralized command model.

## Node Identity

Conceptual node identity requirements include:

- Stable node identifiers for provenance tracking.
- Node-level source attribution in synchronized artifacts.
- Distinction between provider identity and runtime identity where applicable.
- Future trust/domain concepts for scoped federation boundaries.
- Import provenance tagging that records origin node context.

## Artifact Exchange Model

- Artifacts are the unit of exchange.
- Imports preserve lineage.
- Imported artifacts preserve original attribution where possible.
- Synchronization should avoid silent mutation.
- Imported artifacts become part of local lineage history.

Exchange semantics prioritize traceable provenance continuity across nodes.

## Synchronization Modes

Conceptual synchronization modes include:

- Manual import/export.
- Push synchronization.
- Pull synchronization.
- Bidirectional synchronization.
- Snapshot exchange.
- Pack distribution.
- Selective or project-scoped synchronization.

Clarifications:

- Synchronization behavior should remain explicit.
- Any automation must remain observable and governed.

## Conflict Semantics

- Append-only lineage reduces destructive conflicts.
- Conflicting operational states should remain visible.
- Synchronization must not silently rewrite provenance.
- Supersession should be explicit where supported.
- Yanzi prefers traceability over hidden reconciliation.

## Deterministic Import Semantics

- Imported artifacts preserve IDs and hashes where possible.
- Imports should remain verifiable.
- Duplicate detection should prevent silent lineage divergence.
- Import provenance tagging should record source node context.
- Failed imports must remain visible and detectable.

## Federation Scope Concepts

Synchronization scopes may include:

- project-scoped
- checkpoint-scoped
- pack-scoped
- artifact-type-scoped
- organization-scoped
- runtime-scoped

Scope must be explicit for deterministic synchronization behavior.

## Runtime Relationship

- Federation will likely be hosted by optional runtime/daemon layers.
- Local CLI usage may remain entirely standalone.
- Federation should not become mandatory.
- Runtime/API layers may coordinate synchronization operations.

## Transport Neutrality

- Federation semantics are independent of transport.
- Initial implementations may use REST.
- Future transports may include gRPC, streams, sockets, queues, or file exchange.
- Transport changes must not alter federation semantics.

## Security and Trust Direction

Future direction includes:

- Trust boundaries for node-to-node exchange.
- Organization/domain concepts.
- Authentication and token concepts.
- Signed artifacts or packs in later phases.
- Synchronization permissions.
- RBAC above the provider layer.

This section is directional and does not claim complete implementation.

## Operational Lineage Across Nodes

- Lineage continuity may span multiple nodes.
- Source attribution remains critical for multi-node history.
- Operational history should remain inspectable across synchronization boundaries.
- Federation preserves provenance continuity rather than flattening it.

## Multi-Agent and Multi-Tool Continuity

- Multiple tools and agents may interact through federated Context Libraries.
- Federation supports operational continuity between environments.
- Yanzi still does not orchestrate tools or workflows.

## Federation Failure Semantics

- Synchronization failures must be explicit.
- Partial synchronization must be visible.
- Retry behavior should be explicit and auditable.
- Verification should support post-sync integrity checks.
- No silent lineage loss.
- Disconnected operation remains valid.

## Future Compatibility

This contract is intended to remain compatible with:

- REST API
- MCP
- packs
- connector runtime
- UI
- runtime/daemon
- enterprise/shared deployments
- offline/export workflows

## Comparison Philosophy

Yanzi federation is conceptually closer to Git-style artifact exchange than distributed orchestration systems.

Yanzi prioritizes operational traceability over centralized coordination.

## Summary

Federation allows independently authoritative Yanzi nodes to exchange operational lineage and Context Library artifacts. Federation preserves provenance continuity and deterministic operational history. Federation does not centralize operational authority or orchestrate workflows.
