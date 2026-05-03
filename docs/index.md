# Yanzi

Yanzi is a cross-platform CLI (Windows, macOS, Linux) written in Go. It is a deterministic logging layer for AI-assisted development.

AI-assisted development generates decisions and reasoning that are often lost across chat sessions, commits, and ad hoc notes. Git captures code changes, but not the full decision trail behind those changes. Yanzi records that trail so it can be recovered, audited, and shared.

## Core Concepts

**Intent**
An intent (also called a capture) is a prompt/response record. Every interaction that matters to the project can be captured as an intent, stored with a hash, author, source, and optional metadata.

**Context**
Context artifacts attach structured policy, background, or reference material to a project. They provide stable anchors for rehydration.

**Rules**
SYSTEM_RULES.md is a governance file in the project root that defines constraints for AI agents working on the project. It is read by agents at session start and specifies what they must and must not do.

**Profiles** *(planned for v2.5.0)*
Profiles are named sets of rules and context that can be applied to a project. They allow different agents or workflows to operate under different constraint sets.

## Navigation

- [Quickstart](./quickstart.md) — install and first capture
- [CLI Reference](./cli.md) — full command reference
- [Rules](./rules.md) — SYSTEM_RULES and project governance
- [AI Seed](./ai-seed.md) — seed prompt for AI agents
- [Workflow](./workflow.md) — phased execution model
