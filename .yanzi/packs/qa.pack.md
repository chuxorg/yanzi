# Name

qa.pack

# Role

Quality assurance validation context pack.

# Purpose

Compose the minimum deterministic governance context for QA validation and reporting.

# Loads

- .yanzi/rules/SYSTEM_RULES.md
- .yanzi/rules/QA_RULES.md
- .yanzi/workflows/qa-flow.md
- .yanzi/seeds/semantic-versioning.seed.md

# Responsibilities

- Enforce QA reporting discipline.
- Validate CLI, regression, and documentation outcomes.
- Produce explicit PASS/WARN/FAIL dispositions.

# Constraints

- Declarative only; no execution logic.
- No recursive pack references.
- No hidden dependencies or dynamic expansion.
- No runtime state mutation.
