# Drift Review Guidelines

## Purpose
Define how snapshot drift is evaluated for certification impact and release risk.

## Drift Categories
### Acceptable Drift
Meaning: intentional and documented behavior change aligned with approved scope.
Action: recertification required.
Review level: engineering review + governance review.
Release impact: may proceed after approval.

### Operational Drift
Meaning: workflow-visible behavior changed and may affect operators.
Action: investigate and classify impact.
Review level: engineering review required; governance review if baseline changes.
Release impact: block until disposition is clear.

### Regression Drift
Meaning: previously certified behavior degraded or failed.
Action: treat as defect.
Review level: engineering review mandatory.
Release impact: blocking.

### Formatting Drift
Meaning: output formatting changed with no semantic contract break.
Action: verify no operational impact.
Review level: engineering review; governance review if expected baseline update is proposed.
Release impact: non-blocking only when explicitly accepted.

### Non-Deterministic Drift
Meaning: repeated runs under identical inputs produce inconsistent results.
Action: investigate determinism failure.
Review level: engineering review mandatory.
Release impact: blocking until determinism restored.

### Governance Drift
Meaning: baseline changed without required approval trail.
Action: reject baseline change and restore governance path.
Review level: governance review mandatory.
Release impact: blocking.

## When Recertification Is Required
- Any approved change to expected snapshots.
- Any update to normalization logic that can alter comparison meaning.
- Any scenario command or scope change.

## Decision Discipline
- PASS: no unresolved harmful drift.
- WARN: accepted low-risk drift with explicit rationale.
- FAIL: unresolved operational, regression, non-deterministic, or governance drift.
