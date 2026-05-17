# Branch Governance

## Purpose
Define branch expectations for deterministic promotion, review discipline, and release integrity.

## Branch Roles
### `master`
- Role: release source of truth.
- Expectation: accepts only governance-approved promotions from `qa`.
- Gate: release integrity checks and human approval.

### `qa`
- Role: certification integration branch.
- Expectation: all release-candidate changes are validated here first.
- Gate: deterministic scenario execution and certification report review.

### `feature/*`
- Role: engineering development branches.
- Expectation: focused scope, explicit PRs into `qa`.
- Gate: review and deterministic validation readiness.

### `hotfix/*`
- Role: urgent remediation.
- Expectation: minimal, targeted fixes with expedited but explicit review.
- Gate: post-fix deterministic certification before merge/promotion.

## PR Discipline
- Use explicit PR descriptions with validation evidence.
- Include operational impact and deterministic verification notes.
- Avoid bundling unrelated changes in one PR.

## Review Requirements
- At least one qualified reviewer for non-trivial changes.
- Governance-sensitive changes (release, workflow, security docs) require elevated review attention.
- Certification-affecting snapshot updates require explicit reviewer acknowledgment.

## Protected Branch Expectations
- `master` and `qa` should be protected.
- Direct pushes should be disallowed.
- Merge requires passing agreed review gates.

## Force-Push Restrictions
- No force-push to `master` or `qa`.
- Force-push on feature branches should be rare and communicated when used.

## Release Gating
- Release promotion depends on deterministic QA certification.
- Unresolved regression or governance drift blocks release promotion.

## Merge Expectations
- Prefer clear, reviewable history.
- Promotion from `qa` to `master` is explicit and auditable.
- Hotfix merges must preserve provenance and follow-up documentation.

This document defines governance expectations only. Repository settings are configured separately.
