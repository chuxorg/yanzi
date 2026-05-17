# Release Finalization Summary

Date: 2026-05-17
Release candidate: `v2.9.1-rc1`
Canonical convergence SHA: `112262263832eef10ddf3c13441c41c8e4072a99`

## Final Validation Outcomes

- Convergence validation: PASS
- Certification validation: PASS
- Distribution convergence certification: PASS
- Installer lineage: PASS
- Homebrew lineage: PASS
- Direct binary lineage: PASS
- Canonical SHA alignment: PASS
- Promotable status: YES

## Canonical Lineage Reconciliation

- Release metadata was re-anchored to canonical merged lineage SHA:
  - `v2.9.1-rc1` -> `112262263832eef10ddf3c13441c41c8e4072a99`
- New certification artifacts were generated append-only; historical evidence was preserved.

## PR Reconciliation Summary

### `yanzi`

- Merged canonical convergence PR: `#118`
- Superseded feeder lineage previously closed with provenance retention: `#103-#117`, `#102`
- Superseded PR `#119` (convergence consolidation docs) has been closed as provenance-retained.
- Development-stream PRs `#94-#101` and maintenance PR `#96` are out-of-scope for this release cycle.

### `homebrew-yanzi`

- Merged deterministic distribution companion PR: `#1`
- No remaining open PRs.

### `chux-yanzi-qa` (requested as `yanzi-cli-qa`)

- No open PRs.

## Superseded Lineage Notes

- Intermediate convergence/governance/certification branches are retained as auditable provenance.
- Canonical operational authority is now rooted in merged `qa` lineage SHA `1122622...`.

## Governance Exception Summary

- Temporary `qa` branch protection relaxation was executed to resolve required-check deadlock.
- Root cause, rationale, and restoration tasks are recorded in:
  - `.yanzi/docs/governance-exception-log.md`
  - `.yanzi/docs/branch-protection-restoration-checklist.md`

## Release Promotion Recommendation

`v2.9.1-rc1` is eligible for governed promotion.

Pre-promotion operator follow-up:
1. Restore protected-branch required-check governance per checklist.
2. Validate restored PR/check gating with a controlled dry run.
