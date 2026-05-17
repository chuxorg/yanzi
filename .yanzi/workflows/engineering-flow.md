# Name

engineering-flow

# Purpose

Define the standard feature development path from implementation to PR into `qa`.

# Preconditions

- Work begins from an up-to-date base branch.
- A dedicated feature branch exists.
- Required packs/rules are explicitly loaded.

# Steps

1. Create or switch to a feature branch for the scoped change.
2. Implement the change with minimal, focused edits.
3. Run local validation relevant to touched behavior.
4. Update docs when user-facing behavior changes.
5. Commit with clear scope and message discipline.
6. Push branch and open a PR targeting `qa`.

# Validation

- Local checks pass for touched areas.
- Diff scope matches stated intent.
- PR includes concise change and validation summary.

# Failure Handling

- If local validation fails, stop promotion and fix before commit.
- If scope drifts, split into smaller commits or follow-up PRs.
