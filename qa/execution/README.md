# Deterministic Scenario Execution

## Purpose
This directory contains explicit shell workflows for deterministic operational certification. The focus is operational trust, not generic test automation.

## Execution Philosophy
- explicit execution steps
- filesystem-first evidence capture
- deterministic command paths
- human-readable artifacts
- append-only friendly reporting

## Execution Model
1. Run scenario workflow to capture operational outputs.
2. Normalize volatile values while preserving operational meaning.
3. Compare normalized outputs against expected snapshots.
4. Generate a human-readable certification report with PASS/WARN/FAIL outcome.

## Scripts
- `run-project-lifecycle.sh`: executes project lifecycle scenario and captures actual outputs.
- `normalize-output.sh`: normalizes volatile fields from actual outputs.
- `compare-snapshots.sh`: compares normalized outputs against expected baselines and classifies drift.
- `generate-report.sh`: appends certification results to report history.

## Snapshot and Normalization Flow
- Actual outputs: `qa/snapshots/project-lifecycle/actual/`
- Normalized outputs: `qa/snapshots/project-lifecycle/normalized/`
- Expected baselines: `qa/snapshots/project-lifecycle/expected/`

Normalization is scoped to volatile values (timestamps, IDs, temp paths, machine-specific values) and must not remove operationally meaningful data.

## Certification Flow
Run in order:

```bash
qa/execution/run-project-lifecycle.sh
qa/execution/normalize-output.sh
qa/execution/compare-snapshots.sh
qa/execution/generate-report.sh
```

The resulting report is `qa/reports/project-lifecycle-certification.md`.
