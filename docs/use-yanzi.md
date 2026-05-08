# Use Yanzi

## For Execs

Problem:
AI-assisted work loses state between sessions. Teams repeat context, miss decisions, and struggle to understand what changed.

Impact:
Yanzi gives teams a local record of captures, checkpoints, and handoff notes so project state is easier to review and recover.

## For Engineers

Install:

```bash
brew install chuxorg/yanzi/yanzi
```

Basic workflow:

```bash
yanzi init demo
yanzi capture --author "Ada" --prompt "What changed?" --response "Updated the API client."
yanzi checkpoint create --summary "API client update complete"
yanzi rehydrate --dry-run
```

## For Builders

Example chat loop:

1. Create or select a project.
2. Capture the prompt and the AI response.
3. Add shared documents with `bootstrap` from `.yanzi/bootstrap.yaml`.
4. Use the message channel for handoff notes.

Example:

```bash
yanzi project create vibe-loop
yanzi project use vibe-loop
yanzi bootstrap
yanzi message send --to claude --from operator --channel handoff --content "Continue from the latest checkpoint."
yanzi message pull --to claude --channel handoff
```

That keeps the project context and handoff notes in one place while the code work continues outside Yanzi.

## Use Cases

- [Use Cases Overview](use-cases/index.md)
- [Agent-Agnostic Communication](use-cases/agent-communication.md)
- [Seeding Context](use-cases/seeding.md)
- [Targeted Retrieval](use-cases/retrieval.md)
