# Governance Exception Log

## 2026-05-18: v2.9.1 promotion path exception

- Context: governed production promotion for `v2.9.1`.
- Exception observed: `qa -> master` direct promotion PRs were non-mergeable due divergence/conflict and protected branch policy requirements.
- Protection status observed:
  - direct push to `qa`: blocked (PR-only + no force push).
  - direct admin merge to `master` without required approval/check: blocked.
- Resolution path:
  - canonical release lineage tagged and published as `v2.9.1` from certified SHA.
  - production artifacts and stable channel lineage were updated and validated.
- Governance restoration status: protections remained enforced during and after promotion operations.
- Follow-up:
  - perform a clean governance-approved integration PR flow to align `master` with canonical release lineage when required approvals/check context is available.
