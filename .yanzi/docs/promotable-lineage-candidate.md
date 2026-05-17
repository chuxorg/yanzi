# Promotable Lineage Candidate

Date: 2026-05-17

## Canonical Candidate (Current)

Primary repo (`yanzi`) candidate:
- Branch: `feature/release-convergence-resolution`
- PR: [#118](https://github.com/chuxorg/yanzi/pull/118)
- Candidate SHA: `688d3f372b2e2f9a644e70bc5bb602dd54758cb6`
- Target base: `qa`

Distribution companion candidate:
- Repo: `homebrew-yanzi`
- Branch: `feature/release-convergence-resolution`
- PR: [#1](https://github.com/chuxorg/homebrew-yanzi/pull/1)
- Candidate SHA: `1fe1813044252a4441d24256d64d90f4dc332af5`
- Target base: `master`

QA repo (`chux-yanzi-qa`):
- No open PR candidate; repository currently contributes passive provenance state only.

## Why This Is Canonical

- Most operationally complete active lineage tip is PR #118 (it explicitly resolves deterministic convergence for `v2.9.1-rc1`).
- It sits at the end of governance + QA + distribution semantic iterations rather than competing as an earlier partial state.
- Homebrew PR #1 is aligned to the same convergence branch naming and intent, enabling cross-repo release-path determinism.

## Related PR Set

Upstream/feeder lineage with direct relevance:
- `yanzi`: #117, #116, #115, #114, #113
- `homebrew-yanzi`: #1

Historical provenance lineage (superseded but retained):
- `yanzi`: #103-#112, #102

## Certification Status

- Deterministic QA certification evidence exists in active open PRs (#115 plus earlier baseline chain).
- Several feeder PRs show passing `QA Foundation` validate checks, but canonical PR #118 currently has no completed check run listed and remains review-blocked.

## Distribution Convergence Status

- Homebrew lineage has a single clean open PR (#1) aligned to convergence branch intent.
- Publication and propagation are not yet complete until both canonical PRs merge and release publication sequence is executed.

## Remaining Blockers

Architectural blockers:
- None identified for convergence model itself; architecture exists.

Operational blockers:
- Required reviews/approvals on PR #118 and homebrew PR #1.
- Confirm that #118 fully absorbs required artifacts from #113-#117 before closing feeder PRs.

Sequencing blockers:
- Merge order must be deterministic: `yanzi` canonical convergence first, then `homebrew-yanzi` lineage alignment, then promotion/certification gate.

## Release Readiness Assessment

Status: Conditionally promotable pending governance approvals and deterministic merge sequencing.

Readiness score (qualitative): Medium-High
- Strengths: clear canonical branch, explicit convergence intent, lineage traceability retained.
- Gaps: approval gates and final cross-repo merge sequencing not completed.
