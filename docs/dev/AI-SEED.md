# AI Development Seed Prompt

This repository contains the Yanzi CLI.

Yanzi is a local-first workflow state manager designed to support AI-assisted development.

Before making changes, review the following documents:

Architecture
../architecture.md

Domain Model
../domain-model.md

Contribution Rules
../../CONTRIBUTING.md

Documentation Rules
CODE_DOCUMENTATION.md
RELEASE_PROTOCOL.md
These documents define the invariants of the system.

Do not introduce architectural changes without explicit direction.

---

## Development Workflow

Typical workflow when assisting with development:

1. Read the relevant code and documentation.
2. Propose minimal changes.
3. Implement changes in small, focused commits.
4. Verify CLI behavior locally.
5. Ensure changes respect architecture boundaries.

Avoid speculative features.

Focus on correctness and determinism.

---

## Git Workflow

Development occurs on feature branches.

Example:

feature/<short-description>

Commit messages should be concise and descriptive.

Example:

fix: embed version in CLI binary

Pull requests should contain a single logical change.

---

## Using Yanzi During Development

This project uses Yanzi itself to track development intent.

If Yanzi is installed, follow this workflow:

Create a project:

yanzi project create "cli-dev"

Use the project:

yanzi project use "cli-dev"

Create checkpoints before structural changes:

yanzi checkpoint create --summary "refactor command structure"

When context is lost, use:

yanzi rehydrate

This allows deterministic reconstruction of development history.

---

## CLI Responsibilities

The CLI should remain an adapter layer.

Responsibilities:

- argument parsing
- command routing
- user output

The CLI must not contain domain logic.

All domain behavior lives in the library.

---

## Library Responsibilities

The library owns:

- domain logic
- persistence
- checkpoint management
- deterministic rehydration

The CLI should call library APIs directly.

Do not duplicate domain behavior in the CLI.

---

## Key Design Principles

- Local-first
- Deterministic behavior
- Minimal system surface
- Explicit state management
- No hidden inference

---

## When in Doubt

Prefer the smallest change that maintains system clarity.