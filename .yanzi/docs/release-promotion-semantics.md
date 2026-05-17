# Deterministic Release Promotion Semantics

## Purpose
Define explicit promotion semantics so a certified candidate becomes a trusted operational release lineage object across governed distribution channels.

## Core Rule
A release is not promotable until all governed distribution channels converge on the same certified lineage.

## Release Candidate Lifecycle
1. Engineering Candidate: built from governed source branch for intended release scope.
2. QA Candidate: immutable tag or commit selected for deterministic QA execution.
3. Certified Candidate: deterministic QA and snapshot certification completed with PASS/WARN/FAIL outcome.
4. Promotable Candidate: certified candidate plus distribution convergence PASS and no promotion blockers.
5. Published Release: promoted tag and artifacts published to release channels.
6. Converged Release: all governed channels resolve to the same lineage object and verification evidence is recorded.

## Certification Lifecycle
1. Lock candidate tag + SHA.
2. Execute deterministic scenarios.
3. Classify drift with PASS/WARN/FAIL discipline.
4. Generate append-only provenance artifacts.
5. Validate governed distribution channel convergence.
6. Determine promotable or non-promotable state.
7. Require explicit human approval before release promotion.

## Promotable Candidate Definition
A candidate is promotable only when all are true:
- deterministic scenario certification result is PASS
- snapshot validation is PASS
- installer lineage matches candidate tag + SHA contract
- Homebrew lineage matches candidate tag + SHA contract
- release artifacts and runtime version identity agree
- provenance artifacts are complete and append-only
- no active promotion blocking conditions

## Official Release Lineage Definition
Official release lineage is the tuple:
- release tag
- candidate commit SHA
- checksummed artifacts
- runtime version identity
- governed channel resolution evidence
- certification report and findings

## Distribution Convergence Expectations
- GitHub release artifacts, installer resolution, Homebrew formula, and package channels must agree on release lineage.
- Any mismatch is non-promotable until remediated and recertified.
- Promotion is governance action after convergence validation, not before.

## Governance Semantics
- Promotion decisions are human-governed.
- Provenance records are append-only.
- Corrections are new dated entries, never silent rewrites.
