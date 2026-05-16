# Name

release.pack

# Role

Release management context pack.

# Purpose

Compose the minimum deterministic governance context for QA-approved release promotion and verification.

# Loads

- .yanzi/rules/SYSTEM_RULES.md
- .yanzi/rules/RELEASE_RULES.md
- .yanzi/workflows/release-flow.md
- .yanzi/seeds/semantic-versioning.seed.md
- .yanzi/seeds/github-actions.seed.md

# Responsibilities

- Apply release constraints for promotion and tagging.
- Verify release artifacts and deterministic validation outcomes.
- Preserve traceable release decisions.

# Constraints

- Declarative only; no execution logic.
- No recursive pack references.
- No hidden dependencies or dynamic expansion.
- No runtime state mutation.
