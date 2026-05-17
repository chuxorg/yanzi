# Release Integrity Model

## Purpose
Define how Yanzi releases are promoted with deterministic operational validation, explicit governance, and inspectable evidence.

## Operational Trust Philosophy
- Trust is earned through reproducible validation, not assumed from tooling.
- Human review is required at promotion boundaries.
- Release decisions must be auditable from evidence artifacts.

## Deterministic Release Expectations
- Release workflows must be repeatable from documented commands.
- QA certification outcomes must be reproducible from snapshot evidence.
- Output drift must be classified before promotion.

## Provenance Preservation
- Certification and release evidence are append-only.
- Historical decisions are retained, not rewritten.
- Promotion rationale is recorded with PASS/WARN/FAIL context.

## Trusted Promotion Model
1. Engineering work completes in feature branches.
2. Changes merge into `qa`.
3. Deterministic QA scenarios execute.
4. Snapshot comparisons are reviewed.
5. Certification reports are generated.
6. Human governance review approves or rejects promotion.
7. `qa` is promoted to `master`.
8. Tag and release artifacts are validated.
9. Provenance records are retained.

Promotion is governance-driven, not automation-driven.

## Trust Boundaries
### GitHub Actions
Risk: workflow execution context and third-party actions can change behavior.
Governance: keep workflows explicit and review all workflow changes.
Validation: cross-check releases with deterministic QA artifacts.

### External Dependencies
Risk: upstream changes can alter build/runtime behavior.
Governance: dependency updates require review rationale.
Validation: run deterministic scenarios after dependency changes.

### Homebrew
Risk: formula or mirror drift can break install correctness.
Governance: treat formula updates as release-impacting changes.
Validation: certify install, version, and uninstall behavior.

### APT Repositories
Risk: packaging metadata and repository state can drift.
Governance: package publication requires explicit release review.
Validation: verify install path, version output, and upgrade behavior.

### Go Modules
Risk: transitive dependency changes can alter binaries.
Governance: module updates need changelog and review context.
Validation: deterministic QA certification before promotion.

### Release Artifacts
Risk: artifact mismatch or tampering breaks trust.
Governance: artifact inventory and checksum verification required.
Validation: install and runtime version checks against release tag.

### Build Environments
Risk: non-deterministic environment differences impact outputs.
Governance: keep build inputs explicit and review environment drift.
Validation: compare release behavior using deterministic scenario evidence.

### QA Environments
Risk: local machine differences can hide regressions.
Governance: use documented, isolated, reproducible QA execution paths.
Validation: normalized snapshots and certification reports.

## Deterministic Promotion Workflow
1. Engineering phase completion.
2. Merge into `qa`.
3. Deterministic QA certification.
4. Snapshot validation.
5. Certification report generation.
6. Human review.
7. `qa` → `master` promotion.
8. Release tagging.
9. Release artifact validation.
10. Provenance preservation.

## Non-Goals for This Phase
- No hardened CI/CD implementation.
- No signing/SBOM/provenance tooling rollout.
- No automated promotion systems.
