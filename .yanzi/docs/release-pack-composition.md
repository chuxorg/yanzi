# Release Pack Composition

## Role

Release Engineer

## Deterministic Composition Example

Pack inputs:
- `.yanzi/packs/release.pack.md`
- `.yanzi/seeds/release-governance.seed.md`
- `.yanzi/seeds/distribution-governance.seed.md`
- `.yanzi/seeds/provenance-governance.seed.md`
- `.yanzi/workflows/release-promotion-flow.md`
- `.yanzi/RELEASE_RULES.md`

## Deterministic Context Composition

- Context is loaded intentionally from explicit artifact paths.
- No recursive expansion is permitted.
- Loaded context is bounded to release-governance responsibilities.

## Operational Scoping

- Scope is limited to release lineage, convergence, and promotion integrity.
- Engineering implementation authority remains outside this role context.
- QA certification authority remains distinct and upstream.

## Bounded Authority

- Release Engineer can block or approve promotion based on explicit evidence.
- Release Engineer cannot override missing evidence or ambiguous lineage.
- Mutation scope is limited to release-governance artifacts and reports.

## Release Governance Isolation

- Certification and promotion are separate governance stages.
- Promotion requires upstream certification evidence.
- Provenance continuity is preserved through append-only operational records.
