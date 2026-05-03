# AI Agent Seed

This document is a seed prompt for AI agents working on projects that use Yanzi.

Paste the content below at the start of an agent session (Codex, Copilot, Claude, etc.) to orient the agent to the Yanzi workflow.

---

## Seed Prompt

```
This project uses Yanzi to track development intent.

Yanzi is installed at: yanzi (available in PATH)

Before starting work:

1. Read SYSTEM_RULES.md in the project root (if present).
2. Check the active project: yanzi project current
3. Review recent captures: yanzi list --limit 20
4. Rehydrate if context is missing: yanzi rehydrate

During work:

- Record significant decisions as captures:
  yanzi capture --author "<your-name>" --prompt "<question or task>" --response "<decision or result>"

- Create checkpoints at phase boundaries:
  yanzi checkpoint create --summary "<milestone description>"

- Do not make commits directly to development or master.
- Work on feature branches created from development.
- Each logical change gets a single PR.

When done with a phase:

- Export the log: yanzi export --format markdown
- Review YANZI_LOG.md for completeness.
- Create a PR to development.
```

---

## How Yanzi Fits Into Workflows

Yanzi runs as a local CLI alongside your development environment. It does not require network access and stores all data in `~/.yanzi/`.

**For AI agents:** Pass the seed prompt above at session start. The agent reads SYSTEM_RULES.md, checks project state, and begins recording captures automatically.

**For humans:** Run `yanzi list` at any time to review what was recorded. Use `yanzi rehydrate` to reconstruct context after a break or context reset.

**Export pipeline:** JSON export can be piped into external systems (ELK, Splunk, custom dashboards):

```sh
yanzi export --format json | curl -X POST https://logs.internal/ingest --data-binary @-
```

See [Workflow](./workflow.md) for the full phased execution model.
