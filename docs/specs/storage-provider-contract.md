# Storage Provider Contract

## Purpose

Yanzi needs a storage abstraction so operational context can be persisted across different deployment environments while preserving one canonical operational model.

This supports Yanzi as Operational Context Infrastructure: artifacts, lineage, and provenance must remain stable whether data is stored in an embedded file, relational service, or object-backed system.

The Context Library contract remains canonical. Storage providers persist and retrieve operational artifacts, but they do not define artifact meaning.

## Relationship to Context Library Contract

This contract depends on the [Context Library Contract](context-library-contract.md).

Storage providers must implement Context Library semantics without changing them. If a provider cannot represent required semantics, the provider is non-conformant until it can represent them explicitly.

## Non-Goals

This contract is not:

- A distributed database design.
- A consensus protocol.
- A replacement for Git.
- A workflow engine.
- An orchestration layer.
- A mandate to abandon SQLite.
- A requirement for cloud infrastructure.

## Core Principles

- Local-first remains valid.
- SQLite remains the default provider.
- Append-only provenance is required.
- Deterministic hashes must remain verifiable where canonical content is unchanged.
- Imports must not silently overwrite lineage.
- Providers must preserve source attribution.
- Storage failures must be explicit.
- Provider differences must not leak into user-facing semantics.

## Provider Responsibilities

A conforming provider is responsible for:

- Artifact persistence: durable creation of artifacts with identity, payload, and provenance fields.
- Artifact retrieval: deterministic read by identity and scoped queries.
- Project persistence: durable project records and project-scoped isolation semantics.
- Checkpoint persistence: durable checkpoint records used by rehydration.
- Lineage relationship persistence: explicit parent/related links and traceable lineage paths.
- Metadata persistence: structured metadata storage without changing contract semantics.
- Import/export support: deterministic ingest and emission of artifact sets with provenance continuity.
- Query/list support: contract-defined filter and ordering behavior.
- Verification support: digest/provenance verification operations and validation surfaces.
- Migration/version reporting: schema/provider version visibility and migration status reporting.
- Health/status reporting where applicable: explicit provider readiness and failure state reporting.

## Conceptual Provider Interface

The provider interface is conceptual and not final code.

```text
CreateArtifact(input) -> ArtifactRecord | Error
GetArtifact(id, project) -> ArtifactRecord | NotFound | Error
ListArtifacts(filter, pagination) -> ArtifactRecord[] | Error
LinkArtifacts(parent_id, child_id, relation_type, project) -> LinkRecord | Error

CreateProject(input) -> ProjectRecord | Error
GetProject(project_id) -> ProjectRecord | NotFound | Error

CreateCheckpoint(input) -> CheckpointRecord | Error
ListCheckpoints(project, filter) -> CheckpointRecord[] | Error

Rehydrate(project, checkpoint_id, options) -> RehydrationResult | Error
VerifyArtifact(id, project, options) -> VerificationResult | Error

ExportArtifacts(project, filter, format) -> ExportBundle | Error
ImportArtifacts(project, bundle, options) -> ImportResult | Error

Health() -> ProviderHealth
Migrate(target_version?) -> MigrationResult | Error
```

Conformance requirements:

- Operations may be implemented differently per datastore.
- Return shapes may evolve in concrete APIs.
- Semantic outcomes must remain consistent with the Context Library contract.

## Provider Classes

Likely provider classes include:

- Embedded local provider (`SQLite`): best for local-first workflows, deterministic portability, low-ops operation.
- Networked relational provider (`Postgres`): best for shared/team workloads requiring centralized query and concurrency management.
- Object-backed provider (S3/blob/object store): best for large payload retention and archive-oriented storage, usually alongside metadata indexing.
- Hybrid provider (metadata database plus object store): best when metadata query needs and large artifact payload needs must both scale.
- Future enterprise providers: may integrate with governed infrastructure and managed platforms while preserving contract semantics.

Each class is valid when it preserves artifact semantics, provenance continuity, and deterministic lineage behavior.

## SQLite Provider

SQLite is the default provider.

It is ideal for individual developers, local workflows, vibe coding, serious solo development, demos, and portable deterministic state. It requires no daemon and remains the baseline compatibility provider.

Provider compatibility expectations should always include SQLite as the baseline behavior reference.

## Postgres Provider

Postgres is the likely first shared/team provider.

It is better suited for multi-user or team context libraries and is useful behind an optional runtime/daemon for shared access. Postgres support must preserve append-only semantics and deterministic lineage behavior.

Using Postgres does not imply orchestration or autonomous control.

## Object Store / Blob Provider

Object or blob-backed storage is useful for large artifacts, exports, packs, archives, attachments, and long-term retention.

It is likely paired with metadata storage for queryability and lineage traversal. It is not necessarily appropriate as the only provider initially when low-latency relational queries are required.

## Append-Only and Mutation Semantics

- Artifacts should not be rewritten casually.
- Corrections should be represented as new artifacts or explicit supersession records.
- Status changes must be traceable.
- Providers must support auditability of lineage and mutation history.
- Destructive operations must be explicit, governed, and observable.

## Transactions and Consistency

Minimum expectations:

- Artifact write and digest persistence should be atomic where provider capabilities permit.
- Parent/child linkage must not create invisible partial lineage.
- Imports should be transactional where provider supports transactions.
- Failures must leave detectable state.
- Eventual consistency providers require explicit caveats in provider capability and operational guidance.

## Query and Index Expectations

Minimum query requirements:

- by `id`
- by `project`
- by `type` and `subtype`
- by `author` and `source`
- by parent/related artifact reference
- by `checkpoint`
- by `created_at`
- by metadata keys where feasible

Future search and index layers may optimize retrieval, but they must not redefine artifact truth.

## Migration Expectations

- Providers must expose schema/provider versions.
- Automatic migrations should be explicit about direction, scope, and compatibility.
- Migration safety requirements must include failure visibility and rollback or recovery guidance.
- Backup or export is expected before destructive changes.
- Providers should report capability and migration status for operational governance.

## Failure and Recovery Semantics

- Errors must be explicit and actionable.
- No silent fallback that risks split-brain provenance.
- Recovery should include post-recovery verification of lineage and digest continuity.
- Export/import should remain a supported recovery path.
- Healthcheck behavior should report degraded and unavailable states where applicable.

## Provider Capability Discovery

Providers should eventually report capabilities such as:

- `supports_transactions`
- `supports_full_text_search`
- `supports_blob_storage`
- `supports_multi_user`
- `supports_remote_access`
- `supports_migrations`
- `supports_locking`
- `supports_event_streaming`

Capability reporting informs deployment choices but does not modify core artifact semantics.

## Security and Access Control Boundaries

- Core local SQLite may provide minimal access control.
- Runtime/daemon layers may enforce authentication and RBAC.
- Providers should not invent role semantics.
- Access policy should live above the provider boundary where possible.

## Runtime Relationship

- Local CLI may use an embedded provider directly.
- An optional runtime/daemon may host shared providers.
- REST, MCP, connectors, and UI should normally access shared providers through runtime/API layers.
- Provider abstraction enables these deployment modes without changing CLI contract semantics.

## Future Compatibility

This contract supports:

- REST API
- MCP interface
- federation protocol
- packs
- connector runtime
- UI
- enterprise deployment modes

## Summary

Storage providers persist the Context Library. They do not define the Context Library. SQLite remains default; additional providers expand deployment options without changing Yanzi's operational model.
