# Specifications

- [Context Library Contract](context-library-contract.md): Canonical operational contract defining artifacts, lineage, provenance, deterministic import/export, and interface/storage independence for Yanzi.
- [Storage Provider Contract](storage-provider-contract.md): Contract for datastore provider behavior that preserves Context Library semantics across SQLite and future provider classes.
- [Operational API](operational-api.md): Canonical API contract for exposing Context Library operations across REST, runtime-hosted integrations, and future transports.
- [MCP Interface](mcp-interface.md): Contract for exposing Context Library operations through MCP as deterministic shared operational state without orchestration semantics.
- [Federation Protocol](federation-protocol.md): Contract for exchanging Context Library artifacts and lineage across independently authoritative Yanzi nodes without central orchestration.
- [Pack Specification](pack-specification.md): Contract for portable operational context bundles imported into the Context Library with deterministic behavior and provenance continuity.
- [Connector Runtime](connector-runtime.md): Contract for optional connector runtime adapters that ingest and expose operational context while preserving provenance continuity and non-orchestration semantics.
- [Runtime / Daemon Architecture](runtime-architecture.md): Contract for optional runtime-hosted services (REST, MCP, connectors, federation, and shared deployments) that preserve core Context Library semantics.
- [UI Operational Model](ui-operational-model.md): Contract for UI operational visibility, governance management, and lineage/provenance exploration over the Context Library without orchestration semantics.
