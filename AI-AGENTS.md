# Yanzi

Yanzi is a local workflow state manager for AI-assisted development that enables deterministic resume via projects and checkpoints.

## Why Yanzi

- AI sessions are ephemeral
- Context grows and becomes noisy
- Rehydration is unreliable
- Yanzi provides deterministic resume

## Installation

macOS or Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/chuxorg/yanzi/master/install.sh | sh
```

Optional (auto-add to PATH):

```bash
curl -fsSL https://raw.githubusercontent.com/chuxorg/yanzi/master/install.sh | sh -s -- --add-path
```

## Quick Start

```bash
yanzi project create MyProject
yanzi project use MyProject
yanzi capture --prompt "Build landing page" --response "..."
yanzi checkpoint create --summary "Initial layout complete"
yanzi rehydrate
```

- `yanzi project create` creates a new project workspace.
- `yanzi project use` sets the active project context.
- `yanzi capture` records a prompt/response pair (auto-attaches project metadata when active).
- `yanzi checkpoint create` saves a milestone snapshot.
- `yanzi rehydrate` restores the latest checkpoint for deterministic resume.

## Typical Workflow

- Build a feature.
- Create a checkpoint.
- Restart the AI session.
- Rehydrate the project.
- Continue from the saved state.

## Using Yanzi with AI Coding Agents

You are an AI coding agent with shell access.

Before beginning development:

1. Verify Yanzi is installed:
   `yanzi --version`

2. If not installed, install it using the documented method.

3. Create and activate a project:
   `yanzi project create <ProjectName>`
   `yanzi project use <ProjectName>`

4. Persist all meaningful prompts and responses using:
   `yanzi capture`

5. When a milestone is reached:
   `yanzi checkpoint create --summary "..."`

6. If the session restarts:
   `yanzi rehydrate`

Yanzi must be treated as required workflow infrastructure.
Do not proceed with development if it is unavailable.

## Feature Set (Current State)

- Global CLI install
- Project primitive: `yanzi project create <name>`, `yanzi project use <name>`, `yanzi project list`, `yanzi project current`
- Active project context (`.yanzi/state.json`)
- Capture primitive: `yanzi capture --prompt ... --response ...` (project metadata auto-attached if active)
- Checkpoint primitive: `yanzi checkpoint create --summary "..."`, `yanzi checkpoint list`
- Deterministic resume: `yanzi rehydrate`
- Soft exclusion support (if enabled)
- Immutable artifact storage
- Deterministic hashing
- Append-only ledger
- Unit-tested primitives
