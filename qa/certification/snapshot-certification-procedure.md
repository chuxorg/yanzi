# Snapshot Certification Procedure

## Purpose
Certified snapshots define deterministic operational truth surfaces for Yanzi workflows. They are governance-approved certification artifacts used to evaluate release trustworthiness.

Snapshots are operational certification artifacts, not disposable test outputs.

## Human Review Requirements
- Every baseline candidate must be reviewed by a human operator.
- Review must confirm operational meaning, not only string equality.
- Reviewer must verify normalization did not remove meaningful behavior.
- Reviewer identity and review date must be captured in certification evidence.

## Certification Approval Process
1. Execute deterministic scenario commands.
2. Capture actual outputs.
3. Normalize volatile values.
4. Manually review normalized outputs and associated raw evidence.
5. Approve or reject baseline promotion.
6. Promote approved normalized files into `qa/snapshots/<scenario>/expected/`.
7. Generate certification report with PASS/WARN/FAIL outcome.
8. Preserve append-only provenance records.

## Deterministic Baseline Philosophy
- Baselines represent stable, operator-visible behavior under deterministic inputs.
- Baselines must remain readable and auditable.
- Baseline changes require explicit rationale tied to scenario intent.

## Operational Trust Expectations
- Certified baselines must allow independent reviewers to reproduce outcomes.
- Certification outcomes must be traceable from scenario command to snapshot evidence.
- Trust is earned through explicit review and preserved provenance.

## Recertification Expectations
Recertification is required when:
- operational behavior intentionally changes
- normalization rules change meaningfully
- scenario scope changes
- governance rules require renewed approval

## Drift Handling Philosophy
- Drift is classified before any baseline update.
- Baseline updates are never automatic.
- Governance review is required before accepting drift into certified expected outputs.
