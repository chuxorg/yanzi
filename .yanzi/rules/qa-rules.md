# Name

QA_RULES

# Scope

Quality validation constraints for changes promoted through the `qa` branch.

# Rule Statements

- QA must validate real repository workflows, not synthetic-only checks.
- CLI validation is required for user-facing command behavior.
- Regression verification is required for modified or adjacent behavior.
- Documentation validation is required for changed commands, flags, or workflows.
- Release validation requirements must be checked before promotion to release flow.
- QA reporting must use explicit `PASS`, `WARN`, and `FAIL` outcomes with rationale.

# Rationale

Clear QA discipline reduces regressions and makes promotion decisions auditable.
