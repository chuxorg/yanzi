# GitHub Actions Governance Philosophy

## Purpose
Define trust assumptions and review expectations for GitHub Actions in Yanzi release engineering.

## SHA Pinning Philosophy
- Prefer commit-SHA pinning for third-party actions.
- Treat tag-based action references as mutable and higher risk.
- Record rationale when pinning exceptions are necessary.

Version-tagged actions are mutable trust boundaries.

## Least-Privilege Expectations
- Workflows should request only required permissions.
- Secrets and tokens should be minimized and scope-limited.
- Privilege escalation in workflows requires explicit review.

## Explicit Workflow Behavior
- Workflow steps should be readable and deterministic.
- Hidden behavior and opaque scripts should be avoided.
- Changes to release workflows require governance-aware review.

## Deterministic Build Expectations
- Build inputs should be explicit and reproducible.
- Dependency and environment changes should be visible in PR review.
- Deterministic QA evidence remains the release trust anchor.

## Mutable Dependency Awareness
- Action tags, external images, and remote scripts can drift.
- Mutable dependencies require periodic trust review.
- Drift signals should trigger recertification consideration.

## Release Workflow Transparency
- Release workflow outcomes should map to human-reviewable artifacts.
- Promotion decisions should not depend on opaque automation-only signals.
- Workflow outputs should support provenance capture.
