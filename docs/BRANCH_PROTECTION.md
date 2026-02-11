# Branch Protection Recommendations

This document describes recommended GitHub branch protection settings for the default branch.

## Suggested Settings

- Require a pull request before merging
- Require at least one approval
- Require status checks to pass before merging
- Dismiss stale approvals when new commits are pushed
- Prevent force pushes
- Prevent branch deletion
- Optionally require review from code owners

## How to Enable in GitHub

1. Go to the repository on GitHub.
2. Open Settings.
3. Select Branches in the sidebar.
4. Under Branch protection rules, select Add rule.
5. Set the branch name pattern (for example, `master` or `main`).
6. Enable the settings listed above.
7. Click Create (or Save changes).
