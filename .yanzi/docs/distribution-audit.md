# Distribution Audit

Date: 2026-05-17

## Scope
Assessment of current distribution trust boundaries and validation gaps.

## Homebrew Flow
- Current state: distribution likely externalized to tap/repo process (not implemented in this repository workflow set).
- Risks:
  - Formula drift vs release tag.
  - Install script/version mismatch.
- Validation gaps:
  - No repository-local deterministic Homebrew validation evidence.

## APT / Debian Flow
- Current state: `.deb` package is built via `nfpm` in release workflow.
- Risks:
  - Packaging metadata drift.
  - Repo publication state and package index consistency not validated here.
- Validation gaps:
  - No explicit install/upgrade/remove certification for published apt path.

## Binary Release Flow
- Current state: binaries built for linux/darwin/windows and published via GitHub release.
- Risks:
  - Artifact-version drift risk if release path assumptions change.
  - No signed integrity or manifest bundle yet.
- Validation gaps:
  - No automated post-release download/install verification evidence attached to release.

## Version Drift Risks
- Cross-channel drift between binary assets, package metadata, and install surfaces remains possible.
- Drift should be treated as release-integrity issue and gated by QA certification.

## Install Friction / Operational Trust Concerns
- End-user install instructions must stay aligned with real release channels.
- Uninstall and upgrade behavior needs deterministic certification across channels.

## Conservative Hardening Direction
- Keep distribution workflows simple and explicit.
- Add deterministic validation evidence before broadening automation.
- Preserve human review gates for release promotion decisions.
