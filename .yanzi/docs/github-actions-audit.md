# GitHub Actions Audit

Date: 2026-05-17

## Repository Coverage
- `chuxorg/yanzi`: audited and hardened.
- `chuxorg/yanzi-cli-qa`: repository was not resolvable via GitHub API in this environment; hardening could not be directly applied.

## Workflow Findings

### `.github/workflows/ci.yml` (`QA Build`)
- Purpose: post-merge QA build/test gate on merged PRs to `development`.
- Identified risks:
  - Mutable action tags (`@v4`, `@v5`) before hardening.
  - No explicit workflow permissions before hardening.
  - Trigger is `pull_request.closed`; behavior depends on merge event semantics.
- Hardening actions applied:
  - Pinned `actions/checkout` and `actions/setup-go` by commit SHA.
  - Added `permissions: { contents: read }`.
- Remaining risk notes:
  - Merge-only trigger still relies on PR event shape; monitor for missed runs.

### `.github/workflows/qa-foundation.yml` (`QA Foundation`)
- Purpose: main QA validation (vet/test/build/e2e/doc validation).
- Identified risks:
  - Mutable action tags before hardening.
  - No explicit minimal permissions before hardening.
  - Broad pull_request trigger can increase compute usage.
- Hardening actions applied:
  - Pinned `actions/checkout` and `actions/setup-go` by commit SHA.
  - Added `permissions: { contents: read }`.
- Remaining risk notes:
  - Trigger breadth is operationally acceptable but should be reviewed as repo activity scales.

### `.github/workflows/docs.yml` (`Deploy Docs`)
- Purpose: build/deploy docs to GitHub Pages on `master` push.
- Identified risks:
  - Mutable action tags before hardening.
  - Pages deployment requires elevated permissions (`pages:write`, `id-token:write`).
- Hardening actions applied:
  - Pinned `checkout`, `setup-python`, `configure-pages`, `upload-pages-artifact`, `deploy-pages` by SHA.
  - Kept explicit permissions already present because they are operationally required.
- Remaining risk notes:
  - `pip install` pulls mutable dependencies from requirements lock level; dependency integrity remains a trust boundary.

### `.github/workflows/release.yml` (`Release`)
- Purpose: build and publish release artifacts on tag push or merged PR path.
- Identified risks:
  - Mutable action tags before hardening.
  - Release publication requires write permission to contents.
  - Trigger logic contains implicit branch assumptions (`master` and `development`).
- Hardening actions applied:
  - Pinned `checkout`, `setup-go`, and `softprops/action-gh-release` by SHA.
  - Kept `permissions: { contents: write }` (required for release publication).
- Remaining risk notes:
  - `nfpm` is installed from module version pin, but external fetch availability remains a dependency risk.

## Action Pinning Summary
Pinned actions:
- `actions/checkout@34e114876b0b11c390a56381ad16ebd13914f8d5` (v4)
- `actions/setup-go@40f1582b2485089dde7abd97c1529aa768e1baff` (v5)
- `actions/setup-python@a26af69be951a213d495a4c3e4e4022e16d87065` (v5)
- `actions/configure-pages@983d7736d9b0ae728b81ab479565c72886d7745b` (v5)
- `actions/upload-pages-artifact@56afc609e74202658d3ffba0e8f6dda462b719fa` (v3)
- `actions/deploy-pages@d6db90164ac5ed86f2b6aed7e0febac5b3c0c03e` (v4)
- `softprops/action-gh-release@3bb12739c298aeb8a4eeaf626c5b8d85266b0e65` (v2)

Remaining mutable dependency boundaries:
- `pip install -r docs/requirements.txt` content trust.
- `go install ...@v2.43.1` fetches from remote module infrastructure.
