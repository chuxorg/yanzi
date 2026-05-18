# Name

release-governance.seed

# Purpose

Provide bounded deterministic guidance for governed release promotion authority.

# Guidance

- Treat promotion as a governance decision, not a mechanical step.
- Promote only when certification and convergence evidence are explicit.
- Require immutable candidate identity (tag + SHA) in every release decision.
- Record promotion outcomes in append-only operational reports.

# Constraints

- No orchestration semantics.
- No autonomous promotion behavior.
- No references to other artifacts.
