# Release Workflow Audit

Date: 2026-05-17

## Workflow Reviewed
- `.github/workflows/release.yml`

## Deterministic Behavior Review
- Version derivation is explicit:
  - tag pushes use `GITHUB_REF_NAME`
  - merged-PR path uses `VERSION` file prefixed with `v`
- Version format validation exists (`vX.Y.Z` or `vX.Y.Z-qa(.N)`).
- Artifact naming is explicit and consistent (`yanzi-<os>-<arch>`, plus Windows zip).

## Identified Risks
- Release path depends on merge assumptions (`base=master`, `head=development`) that may drift from branch governance practices.
- `sed -i` behavior can differ across environments; currently runs on Ubuntu so behavior is stable, but portability is constrained.
- External tool fetch (`go install nfpm@v2.43.1`) is version-pinned but still remote-dependent.
- Release workflow has `contents: write` (necessary) and must remain tightly scoped.

## Hardening Applied
- All key actions pinned by SHA (`checkout`, `setup-go`, `action-gh-release`).
- No trigger redesign introduced in this conservative phase.

## Recommended Next Hardening (Future)
- Align release trigger branch assumptions with current branch governance model.
- Add explicit artifact manifest generation/check step before publish.
- Add post-publish validation checklist artifact in workflow summary.
