# Yanzi

[![QA Build](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml)
[![Release](https://github.com/chuxorg/yanzi/actions/workflows/release.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/release.yml)

Links: [yanzi (install info)](https://github.com/chuxorg/yanzi) | [yanzi.io](https://yanzi.io) | [chucksailer.me](https://chucksailer.me)

Yanzi is a local workflow state manager for AI-assisted development that enables deterministic resume via projects and checkpoints.

## Why Yanzi
- AI sessions are ephemeral
- Context grows and becomes noisy
- Rehydration is unreliable
- Yanzi provides deterministic resume

## Capabilities
- Global CLI install with `go install github.com/chuxorg/yanzi/cmd/yanzi@latest`.
- Project primitive: `yanzi project create <name>`, `yanzi project use <name>`, `yanzi project list`, `yanzi project current`.
- Active project context stored in `.yanzi/state.json`.
- Capture primitive: `yanzi capture --prompt ... --response ...` (project metadata auto-attached when active; `--author` is required).
- Checkpoint primitive: `yanzi checkpoint create --summary "..."`, `yanzi checkpoint list`.
- Deterministic resume: `yanzi rehydrate`.
- Deterministic project log export: `yanzi export --format markdown`.
- Immutable artifact storage with deterministic hashing and an append-only ledger.
- Unit-tested primitives.

## Installation
Install the latest CLI directly:

```sh
go install github.com/chuxorg/yanzi/cmd/yanzi@latest
```

If you already have the repository checked out:

```sh
go install ./cmd/yanzi
```

Optional release-binary installer scripts are available under `scripts/`:

```sh
./scripts/install.sh
./scripts/uninstall.sh
```

## Code Documentation
- Yanzi uses `gomarkdoc` to generate API docs from Go comments.
- Generated docs are written to `docs/API.md` and include an Index section (TOC).

```sh
make docs
make docs-check
```

## Quick Start
```sh
yanzi project create MyProject
yanzi project use MyProject
yanzi capture --prompt "Build landing page" --response "..."
yanzi checkpoint create --summary "Initial layout complete"
yanzi export --format markdown
yanzi rehydrate
```

- `yanzi project create` creates a project record.
- `yanzi project use` sets the active project in `.yanzi/state.json`.
- `yanzi capture` stores the prompt/response and attaches active project metadata. Add `--author` when running this command.
- `yanzi checkpoint create` saves a checkpoint for the active project.
- `yanzi export --format markdown` generates `YANZI_LOG.md` in project root.
- `yanzi rehydrate` prints the latest checkpoint and artifacts since.

## Typical Workflow
- Build a feature and capture key prompts/responses.
- Create a checkpoint when a milestone is reached.
- Restart your AI session.
- Run `yanzi rehydrate` to load the latest checkpoint and artifacts since.
- Continue work with deterministic context.

## Using Yanzi with AI Coding Agents
```text
You are an AI coding agent with shell access.

Before beginning development:

1. Verify Yanzi is installed:
   yanzi --version

2. If not installed, install it using the documented method.

3. Create and activate a project:
   yanzi project create <ProjectName>
   yanzi project use <ProjectName>

4. Persist all meaningful prompts and responses using:
   yanzi capture

5. When a milestone is reached:
   yanzi checkpoint create --summary "..."

6. If the session restarts:
   yanzi rehydrate

Yanzi must be treated as required workflow infrastructure.
Do not proceed with development if it is unavailable.
-----------------------------------------------------
```
