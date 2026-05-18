# Name

release-promotion-flow

# Purpose

Define the deterministic operational release promotion sequence and authority boundaries for Release Engineer governance.

# Preconditions

- Candidate tag and SHA are explicitly declared.
- QA certification inputs are complete and reviewable.
- Required release pack artifacts are loaded.

# Steps

1. Engineering candidate: Engineering produces candidate artifacts and change scope.
2. QA candidate: QA validates behavior and produces certification evidence.
3. Certification: PASS | WARN | FAIL status is recorded with blocking issues.
4. Convergence validation: installer, binary, and Homebrew lineage are verified.
5. Promotable lineage determination: Release Engineer verifies tag/SHA/artifact/runtime convergence.
6. Governed promotion: Release Engineer approves or blocks `qa` to `master` transition.
7. Artifact publication: Publish release assets for certified lineage only.
8. Post-release validation: Re-check channel installs and runtime lineage.
9. Governance restoration: Record results, risks, and follow-up actions in append-only reports.

# Operational Ownership Boundaries

- Engineering owns candidate implementation and engineering validation inputs.
- QA owns behavioral certification evidence and PASS/WARN/FAIL disposition.
- Release Engineer owns promotion authority, convergence authority, and lineage authority.
- No role may bypass explicit governance evidence requirements.

# Handoff Semantics

- Engineering -> QA handoff includes candidate tag, SHA, and change intent.
- QA -> Release Engineer handoff includes certification report and blocking issue status.
- Release Engineer -> Maintainers handoff includes promotable decision and provenance links.

# Provenance Expectations

- Every handoff references immutable candidate identity.
- Certification and convergence outputs are append-only.
- Promotion decision references exact evidence files used.

# Blocking Conditions

- Certification outcome is FAIL.
- Convergence outcome is not PASS.
- Tag/SHA/runtime/version lineage mismatch exists.
- Distribution channel ambiguity exists.
- Provenance continuity is missing or corrupted.
