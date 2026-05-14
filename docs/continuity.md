# Continuity Philosophy

## What Yanzi Is

Yanzi is a deterministic continuity CLI for AI-assisted development.

It preserves:

- prompt and response history
- checkpoints as recovery anchors
- context and rule artifacts
- handoff notes and protocol annotations

Its job is to make recovery, review, and handoff easier without inventing behavior that did not happen.

## What Yanzi Is Not

Yanzi is not:

- an orchestration framework
- a background daemon
- an autonomous workflow engine
- an embeddings or vector retrieval system
- an AI summarization layer

Yanzi records and renders stored state. It does not silently execute workflows behind the operator.

## Deterministic Continuity

Yanzi favors explicit records over inferred meaning.

That means:

- ordering is stable
- checkpoints are explicit boundaries
- metadata is rendered as stored
- recovery output is derived from recorded artifacts
- missing checkpoints fall back to recent project captures instead of guessed summaries

## Checkpoint Mental Model

A checkpoint is a recovery anchor, not a branch rewrite.

When a checkpoint exists, `yanzi rehydrate` starts from the latest checkpoint and replays forward from there.

When no checkpoint exists, `yanzi rehydrate` falls back to the latest capture window for the active project.

## Operational Cognition Preservation

Yanzi preserves operational cognition by keeping the project record readable:

- captures show what was asked and answered
- intent artifacts keep actionable work visible
- protocol annotations record transitions such as pause, resume, role changes, and export intent
- exports preserve chronology and checkpoint boundaries for later review

## Protocol Truthfulness

`@yanzi` lines are protocol annotations, not hidden executable commands.

Examples:

- `@yanzi pause`
- `@yanzi resume`
- `@yanzi role Architect`
- `@yanzi checkpoint "Auth snapshot"`

These annotations are useful continuity markers and may be logged as artifacts or events.

If a workflow wants an actual CLI action, it must run the explicit command separately, for example:

```bash
yanzi checkpoint create --summary "Auth snapshot"
```

## Recovery Expectations

You should expect Yanzi to help with:

- reconstructing recent project state
- understanding where a checkpoint boundary exists
- surfacing recent operational transitions
- reviewing open intent artifacts and continuity metadata

You should not expect Yanzi to:

- decide what matters on your behalf
- infer unfinished work from vague language
- execute protocol annotations automatically
- recover state that was never captured
