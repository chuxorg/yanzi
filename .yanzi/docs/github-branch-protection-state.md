# GitHub Branch Protection State

Date: 2026-05-17

## Repository Coverage
- `chuxorg/yanzi`: branch protection applied and verified.
- `chuxorg/yanzi-cli-qa`: repository not resolvable via GitHub API in this environment; no branch rule updates applied.

## Applied Branch Protections (`chuxorg/yanzi`)

### `master`
- PR-only merge path: enforced by required PR reviews.
- Required approvals: `1`.
- Require up-to-date branch before merge: enabled (`strict` status checks).
- Required status checks: `validate`.
- Force pushes: disabled.
- Deletions: disabled.
- Conversation resolution: required.
- Linear history: required.
- Admin enforcement: enabled.

### `qa`
- PR-only merge path: enforced by required PR reviews.
- Required approvals: `1`.
- Require up-to-date branch before merge: enabled (`strict` status checks).
- Required status checks: `validate`.
- Force pushes: disabled.
- Deletions: disabled.
- Conversation resolution: required.
- Linear history: required.
- Admin enforcement: enabled.

## Notes
- A pre-existing bypass allowance for user `csailer` is present on `master` pull-request review enforcement.
- No additional branch-level complexity was introduced.
