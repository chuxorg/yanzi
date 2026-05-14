# How It Works

## Intent vs Context

Yanzi stores two kinds of project state:

- intent: prompt and response pairs captured with `yanzi capture`
- context: documents or notes added with `yanzi context add`, `yanzi rules add`, or `yanzi bootstrap`

Intent tells you what was asked and answered. Context stores the supporting rules, references, and notes that should stay available across sessions.

## Local-First Runtime

Yanzi runs against a local SQLite database by default.

This keeps the operational model small and explicit:

- local file
- deterministic ordering
- short-lived CLI processes
- no background worker or orchestration layer

See [Local-First Operation](local-first.md) for concurrency expectations and lock behavior.

## Checkpoints

A checkpoint marks a stable point in the active project.

`yanzi checkpoint create --summary "..."` writes a checkpoint record for the current project. Rehydration starts from the latest checkpoint instead of replaying the entire history from the beginning.

## Rehydration

`yanzi rehydrate` loads the active project, finds the latest checkpoint, and prints the captures recorded after that checkpoint.

`yanzi rehydrate --dry-run` shows what would be loaded without printing the full sequence.

## Message Channel

`yanzi message send`, `yanzi message list`, and `yanzi message pull` store handoff notes as captures with message metadata.

This is useful for passing short instructions between operators and agents without mixing those notes into unrelated captures.
