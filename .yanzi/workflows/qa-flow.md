# Name

qa-flow

# Purpose

Define QA validation steps for candidate changes before release promotion.

# Preconditions

- Validation branch is checked out and synchronized.
- QA and system rules are explicitly loaded.
- Test environment is available.

# Steps

1. Check out the validation branch and confirm target commit.
2. Execute the validation suite for affected and critical paths.
3. Verify CLI behavior for changed commands and flags.
4. Verify export/output behavior for affected interfaces.
5. Verify documentation accuracy for changed behavior.
6. Produce a QA report using `PASS | WARN | FAIL` outcomes.

# Validation

- Required checks completed and recorded.
- CLI behavior matches documented behavior.
- Report includes clear evidence and disposition.

# Failure Handling

- Any `FAIL` blocks promotion until resolved.
- `WARN` requires explicit acceptance or follow-up issue.
