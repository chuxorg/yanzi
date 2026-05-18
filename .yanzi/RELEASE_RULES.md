# Name

RELEASE_RULES

# Purpose

Define deterministic release governance expectations for promotion authority, lineage integrity, and provenance continuity.

# Release Governance Expectations

- Release promotion is governance-driven and human-approved.
- Promotion decisions must be based on explicit certification and convergence evidence.
- Release operations must remain inspectable and reproducible.

# Promotable Lineage Requirements

- Candidate tag, candidate SHA, release assets, and runtime version must converge.
- A promotable candidate must resolve from canonical release metadata.
- Ambiguous channel resolution disqualifies promotion.

# Certification Requirements

- Deterministic certification must produce explicit PASS | WARN | FAIL outcomes.
- Blocking findings require FAIL with promotion blocked.
- Certification evidence must remain append-only and reviewable.

# Distribution Convergence Requirements

- Installer, direct binary, and Homebrew channels must resolve to the same candidate lineage.
- Distribution mismatch is a promotion blocker.
- Channel convergence checks must be recorded in release evidence.

# Provenance Preservation Requirements

- Prior certification evidence must not be rewritten or deleted.
- New runs append additional evidence with timestamp and candidate identity.
- Promotion records must reference the exact evidence set used for decision.

# Branch Protection Expectations

- Promotion path is `qa` to `master` only after governance gates pass.
- Protected branches must prevent direct bypass of certification gates.
- Release decisions must be traceable to reviewed PR evidence.

# Release Publication Rules

- Publish release artifacts for the certified candidate lineage only.
- Asset metadata must be complete and verifier-friendly.
- Published artifacts must be accessible to governed installer and distribution checks.

# Post-Release Validation Expectations

- Validate installed runtime version across governed channels post-publication.
- Record any propagation delays and operational impact.
- Open a governance follow-up for any post-release drift.

# Promotion Eligibility Rule

A release may only be promoted from:

- certified canonical lineage
- deterministic convergence PASS
- preserved provenance continuity
