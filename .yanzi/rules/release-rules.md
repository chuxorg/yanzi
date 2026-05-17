# Name

RELEASE_RULES

# Scope

Release governance constraints for promotion from validated code to published release artifacts.

# Rule Statements

- Releases must be reproducible from tagged source and documented commands.
- Version and tag discipline is required: one release version maps to one immutable tag.
- Promotion flow is `qa` to `master` only after QA approval.
- Release artifacts must be verified for completeness and integrity.
- Release validation must be deterministic and repeatable.

# Rationale

These constraints preserve release integrity, traceability, and predictable operations.
