# Rules

## SYSTEM_RULES.md

`SYSTEM_RULES.md` is a governance file placed in the project root. It defines the rules AI agents must follow when working on the project.

Example content:

```markdown
# AI Execution Rules

Before executing any development task:
- Confirm work is on a feature branch created from development branch.
- Confirm no direct commits to development or master.
- After each task: stage, commit, push.
- At phase completion: create PR to development.
```

Agents read this file at the start of each session (via the [seed prompt](./ai-seed.md)). The rules act as persistent constraints that survive context resets.

## What Rules Control

Rules in `SYSTEM_RULES.md` can specify:

- **Branch discipline** — which branches are allowed, when PRs are required
- **Commit behavior** — when to stage, commit, and push
- **Phase gates** — what must be true before a phase is considered complete
- **Documentation requirements** — when to update comments or docs
- **Test requirements** — what must pass before a push
- **Agent restrictions** — files or areas the agent must not touch

## How Rules Are Stored and Retrieved

`SYSTEM_RULES.md` is a plain markdown file committed to the repository. It is versioned with the code.

Agents retrieve it by reading the file directly:

```sh
cat SYSTEM_RULES.md
```

The [Yanzi seed prompt](./ai-seed.md) instructs agents to read `SYSTEM_RULES.md` before starting any task.

## Profiles *(planned for v2.5.0)*

Profiles are named sets of rules that allow different workflows to apply different constraints. For example, a `release` profile might enforce stricter review requirements than a `feature` profile.

Profile support will be available in a future release. Until then, manage rules directly in `SYSTEM_RULES.md`.
