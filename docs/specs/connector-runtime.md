# Connector Runtime

## Purpose

Connector runtimes define how Yanzi integrates with external engineering systems while preserving Context Library semantics.

Connectors ingest, expose, synchronize, or transform operational context so Yanzi can interoperate with real engineering ecosystems without changing core artifact, lineage, and provenance behavior.

Connectors extend interoperability as adapters/interfaces and are expected to live in optional runtime/daemon layers rather than Yanzi Core.

## Relationship to Existing Contracts

This specification depends on:

- [Context Library Contract](context-library-contract.md)
- [Storage Provider Contract](storage-provider-contract.md)
- [Operational API](operational-api.md)
- [MCP Interface](mcp-interface.md)
- [Federation Protocol](federation-protocol.md)
- [Pack Specification](pack-specification.md)

Contract statements:

- Connectors operate against Context Library semantics.
- Connectors preserve provenance continuity.
- Connector behavior must not redefine artifact semantics.

## Non-Goals

Connectors are not:

- orchestration systems
- autonomous agents
- deployment automation systems
- workflow engines
- hidden synchronization layers
- AI reasoning systems
- centralized control planes

## Core Principles

- interface neutrality
- deterministic ingestion behavior
- provenance preservation
- source attribution
- append-only operational lineage
- explicit synchronization behavior
- local-first compatibility
- runtime optionality

## Connector Runtime Model

Connectors likely execute within optional runtime/daemon layers.

Runtime layers host integrations without bloating Yanzi Core. Local standalone CLI operation remains valid without connector runtime dependencies.

Runtime layers may expose REST and MCP interfaces simultaneously while preserving one shared Context Library contract.

## Connector Categories

### Context Source Connectors

Examples:

- GitHub
- Jira
- ChatGPT exports
- Claude exports
- Slack
- Notion
- Confluence
- filesystem imports

### Event Connectors

Examples:

- CI/CD systems
- deployment systems
- incident systems
- PR merge events
- release systems

### Storage Connectors

Examples:

- object stores
- blob stores
- archives
- enterprise storage systems

### Runtime/Interface Connectors

Examples:

- MCP bridges
- IDE integrations
- editor extensions
- API bridges

## Context Ingestion Semantics

Ingested context becomes operational artifacts.

Requirements:

- imports preserve attribution where possible
- ingestion remains inspectable
- ingestion must not silently rewrite lineage
- connector normalization behavior remains explicit

## Event Ingestion Semantics

External operational events may be represented as artifacts.

Requirements:

- event lineage remains traceable
- timestamps and source attribution are preserved
- event ingestion does not imply orchestration ownership

## Source Attribution Requirements

Attribution requirements:

- every connector preserves source identity
- imported artifacts retain external origin metadata where available
- operational lineage remains inspectable
- connector identity remains visible

## Connector Identity and Metadata

Conceptual metadata includes:

- `connector_id`
- `connector_type`
- `source_system`
- runtime/node source
- ingestion timestamps
- synchronization metadata
- transformation metadata where applicable

Metadata should support deterministic traceability and operational auditability.

## Synchronization and Polling Concepts

Future-facing connector synchronization concepts include:

- polling connectors
- push/webhook connectors
- manual synchronization
- scheduled synchronization
- snapshot imports
- replay/import recovery

Synchronization must remain observable and governed. Hidden automation should be avoided.

## Transformation and Normalization Boundaries

Connectors may normalize external formats into Context Library artifact shapes.

Requirements:

- transformations remain traceable
- original source context remains recoverable where possible
- connectors avoid inventing hidden semantics

Normalization adapts representation, not operational meaning.

## Runtime Isolation

Connectors belong outside Yanzi Core.

Runtime isolation requirements:

- connector failures should not corrupt core operational lineage
- optional runtime layers enable extended enterprise capabilities
- local CLI usage remains lightweight and viable without connector runtime

## Security and Trust Direction

Future direction only:

- connector credentials
- scoped permissions
- RBAC integration in later phases
- runtime-hosted secrets/configuration
- auditability
- organizational trust boundaries

This section is directional and does not claim full current implementation.

## Federation Compatibility

Connectors may operate across federated nodes.

Requirements:

- federation preserves imported provenance
- synchronization between runtimes remains explicit
- connector-mediated exchanges follow federation lineage continuity constraints

## UI and Operational Visibility

Future-facing operational visibility expectations include:

- UI visibility into connectors
- connector status
- synchronization history
- ingestion visibility
- operational auditability
- health and failure visibility

## IDE and AI Tool Integration Direction

Conceptual future integrations include:

- VSCode
- JetBrains
- Cursor
- ChatGPT
- Claude
- MCP-capable systems

Clarifications:

- integrations expose operational continuity and shared context access
- Yanzi remains operational infrastructure rather than AI orchestration

## Failure and Recovery Semantics

Failure behavior requirements:

- connector failures remain explicit
- retries remain observable
- ingestion conflicts remain visible
- no silent lineage corruption
- replay/reimport paths remain available for recovery

Recovery behavior prioritizes traceability over hidden repair.

## Future Compatibility

This contract is intended to remain compatible with:

- REST API
- MCP
- federation
- packs
- runtime/daemon
- UI
- enterprise/shared deployments

## Comparison Philosophy

Connectors integrate operational context systems.

Yanzi prioritizes provenance continuity and operational traceability over orchestration and automation.

## Summary

Connector runtimes integrate Yanzi into external engineering ecosystems.

Connectors preserve provenance continuity and operational lineage.

Connectors expose and ingest operational context without orchestrating workflows or operational authority.
