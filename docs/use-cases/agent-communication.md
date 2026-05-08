# Agent-Agnostic Communication

## Problem

Multiple AI agents operate independently and cannot share state.

## Approach

Agents write messages to Yanzi:

```bash
yanzi message send \
  --to codex \
  --from operator \
  --channel execution \
  --content "Continue from the stored handoff."
```

Other agents retrieve messages:

```bash
yanzi message pull --to codex --channel execution
```

## Result

Agents communicate through shared persistent state.

## Note

This is a technique, not a required pattern.
