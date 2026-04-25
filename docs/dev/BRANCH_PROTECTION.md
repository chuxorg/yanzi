# Branch Protection

This document describes recommended GitHub branch protection settings for the default branch.

## Recommended Settings
- Require a pull request before merging
- Require at least one approval
- Require status checks to pass before merging
- Dismiss stale approvals when new commits are pushed
- Prevent force pushes
- Prevent branch deletion
- Optionally require review from code owners

## How To Enable (GitHub UI)
1. Open the repository on GitHub.
2. Go to Settings.
3. Select Branches in the left sidebar.
4. Under Branch protection rules, click Add rule.
5. Enter the default branch name, for example `master`.
6. Enable the settings listed in the Recommended Settings section.
7. Click Create or Save changes.
