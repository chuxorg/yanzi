# Runtime / Daemon Architecture

## Purpose

The runtime/daemon layer defines optional operational infrastructure services around the Context Library.

The runtime enables integrations and shared deployments by hosting interfaces and services such as REST, MCP, connectors, and federation coordination.

Yanzi Core remains independently useful without the runtime. The runtime extends interoperability and deployment options without redefining core operational semantics.

## Relationship to Existing Contracts

This specification depends on:

- [Context Library Contract](context-library-contract.md)
- [Storage Provider Contract](storage-provider-contract.md)
- [Operational API](operational-api.md)
- [MCP Interface](mcp-interface.md)
- [Federation Protocol](federation-protocol.md)
- [Pack Specification](pack-specification.md)
- [Connector Runtime](connector-runtime.md)

Contract statements:

- The runtime exposes Context Library interfaces.
- The Context Library remains canonical.
- Runtime behavior must preserve provenance continuity and lineage semantics.

## Non-Goals

The runtime is not:

- orchestration infrastructure
- Kubernetes replacement
- workflow automation
- distributed scheduling
- autonomous execution infrastructure
- a centralized AI control plane
- a mandatory deployment requirement

## Core Principles

- local-first compatibility
- optional runtime architecture
- append-only operational provenance
- deterministic operational semantics
- explicit operational visibility
- runtime isolation from core
- federation over centralized control
- operational transparency

## Yanzi Core vs Runtime

### Yanzi Core

- CLI
- local storage
- Context Library
- checkpoints
- exports
- deterministic operational workflows
- standalone operation

### Runtime/Daemon

- REST hosting
- MCP hosting
- connector hosting
- federation hosting
- auth/RBAC (future)
- UI backend support
- shared/team deployments

Clarifications:

- Runtime extends deployment models.
- Runtime does not replace Core semantics.

## Runtime Hosting Model

The runtime may execute as a daemon/service.

The runtime may host one or more storage providers and may expose multiple interfaces simultaneously. Runtime behavior should remain operationally understandable and avoid unnecessary distributed complexity.

## REST API Hosting

The runtime may expose Operational API endpoints.

REST is the likely initial runtime-hosted interface. API semantics derive from the Context Library contract, and transport neutrality remains a future-compatible requirement.

## MCP Hosting

The runtime may expose MCP interfaces.

MCP remains an adapter layer over Context Library semantics. Runtime-hosted MCP enables shared operational context between tools and does not imply orchestration.

## Connector Hosting

The runtime may host connector processes/services.

Connector hosting remains optional. Connector failures should remain isolated from core lineage integrity, and connector operations should remain visible and auditable.

## Federation Hosting

The runtime may coordinate synchronization between nodes.

Federation behavior remains explicit and inspectable. Synchronization must preserve provenance continuity and does not imply centralized operational authority.

## Multi-Provider Deployment Direction

Conceptual deployment directions include:

- embedded SQLite
- shared Postgres provider
- hybrid metadata/blob storage
- organizational runtime deployments
- offline/local deployments

Provider evolution should preserve deterministic semantics and contract compatibility.

## Runtime Deployment Modes

Conceptual runtime modes include:

- standalone local runtime
- team/shared runtime
- enterprise runtime
- disconnected/offline runtime
- federated runtime nodes

These modes are deployment options, not semantic forks of Yanzi behavior.

## Authentication and RBAC Direction

Future-facing direction:

- authentication likely belongs in runtime layers
- token/session concepts
- role-based editing/install permissions
- organizational governance
- auditability
- runtime-scoped operational authority

This section is directional and does not claim full current implementation.

## UI Relationship

Future UI layers will likely connect through runtime/API services.

Runtime may host UI backend services. UI remains an operational management interface and must preserve operational transparency.

## Operational Visibility

Future-facing runtime visibility should include:

- health/status
- synchronization visibility
- connector visibility
- provider visibility
- operational auditability
- import/export visibility
- lineage verification visibility

## Runtime Isolation and Failure Semantics

Runtime failures should not destroy Context Library semantics.

Requirements:

- local exports/imports remain recovery paths
- failures remain explicit
- partial failures remain inspectable
- disconnected operation remains valid

Isolation protects operational continuity across local and shared deployment modes.

## Security and Trust Direction

Future direction includes:

- trust boundaries
- runtime-scoped credentials
- secure federation
- pack verification
- connector credential management
- organizational governance layers

Security layers should reinforce transparent, deterministic operational governance.

## Performance and Scalability Philosophy

Runtime architecture should avoid premature distributed systems complexity.

Priorities:

- operational clarity over infrastructure theater
- simple deployment models where possible
- scaling that preserves deterministic semantics
- runtime operation that remains comprehensible to engineering teams

## Future Compatibility

This contract is intended to remain compatible with:

- REST API
- MCP
- federation
- packs
- connectors
- UI
- IDE integrations
- enterprise/shared deployments

## Comparison Philosophy

The runtime extends Yanzi into shared operational infrastructure.

Yanzi prioritizes deterministic operational continuity over orchestration and centralized automation.

Runtime architecture is closer to operational middleware than autonomous platform infrastructure.

## Summary

The runtime/daemon layer hosts optional operational infrastructure services around the Context Library.

Yanzi Core remains lightweight, local-first, and independently useful.

The runtime enables integrations, federation, and shared operational deployments without introducing orchestration semantics.
