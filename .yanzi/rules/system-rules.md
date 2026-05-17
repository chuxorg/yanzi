# Name

SYSTEM_RULES

# Scope

Repository-wide governance constraints for deterministic context composition.

# Rule Statements

- Workflows must be deterministic: equivalent inputs should produce equivalent operational outcomes.
- Context must be explicitly loaded from named artifacts; implicit expansion is not allowed.
- Operational history must be append-only when recording state or run outcomes.
- Hidden automation is not allowed; actions must be visible and reviewable.
- Execution remains human-governed; artifacts guide decisions but do not self-execute.
- Prefer operational simplicity over abstraction-heavy structures.
- Governance remains filesystem-first: artifacts are plain files, inspectable in version control.

# Rationale

These constraints preserve inspectability, reproducibility, and bounded context composition.
