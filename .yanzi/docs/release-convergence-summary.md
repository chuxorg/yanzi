# Release Convergence Governance Summary

Date: 2026-05-17

## Current Operational State

- Open PR density is concentrated in `yanzi`, with one active distribution PR in `homebrew-yanzi` and no open PRs in `chux-yanzi-qa`.
- Multiple open PRs are intermediate governance/QA states that now function as provenance evidence rather than promotable heads.

## Canonical Convergence Branch Decision

- Canonical branch: `feature/release-convergence-resolution`
- Canonical PR: `yanzi` PR #118
- New convergence branch creation: **not necessary** at this phase (existing branch already serves the deterministic convergence role).

## Promotable Candidate

- `yanzi` candidate SHA: `688d3f372b2e2f9a644e70bc5bb602dd54758cb6` (PR #118)
- `homebrew-yanzi` companion SHA: `1fe1813044252a4441d24256d64d90f4dc332af5` (PR #1)

## Remaining Blockers

Architectural:
- None material.

Operational:
- Required review approvals for canonical convergence PR(s).
- Confirmation that required feeder content is fully represented in PR #118.

Sequencing:
- Deterministic merge order and release gate execution across repos.

## Recommended PR Closures

Recommended closure as superseded provenance-only lineage after confirmation of absorption:
- `yanzi` PRs: #103-#112, #113, #114, #116, #117

Keep open until canonical merge is complete:
- `yanzi` PR #115 (if needed as explicit certification evidence during approval)
- `yanzi` PR #118 (canonical)
- `homebrew-yanzi` PR #1 (distribution companion)

Out-of-scope for this convergence closure wave:
- `yanzi` development-stream PRs #94-#101
- `yanzi` maintenance PR #96

## Recommended Next Actions

1. Confirm PR #118 includes all required deterministic governance/promotion/distribution semantics.
2. Merge PR #118 into `qa` after approvals/check gate completion.
3. Merge `homebrew-yanzi` PR #1 in deterministic sequence immediately after canonical core merge.
4. Close superseded feeder PRs as provenance-retained, with closure comments pointing to PR #118.
5. Execute final release promotion gate and append certification outcomes (no history rewrites).

## Determinism Validation

- Convergence analysis uses current open PR + branch state as of 2026-05-17.
- Provenance continuity preserved by recommending closure/archival instead of deletion or history mutation.
- Ownership boundaries preserved by keeping repo-specific promotion roles explicit (`yanzi` core, `homebrew-yanzi` distribution, QA repo provenance).
