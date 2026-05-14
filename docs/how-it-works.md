# How It Works

## Intent vs Context

Yanzi stores two kinds of project state:

- intent: prompt and response pairs captured with `yanzi capture`
- context: documents or notes added with `yanzi context add`, `yanzi rules add`, or `yanzi bootstrap`

Intent tells you what was asked and answered. Context stores the supporting rules, references, and notes that should stay available across sessions.

## Checkpoints

A checkpoint marks a stable point in the active project.

`yanzi checkpoint create --summary "..."` writes a checkpoint record for the current project. Rehydration starts from the latest checkpoint instead of replaying the entire history from the beginning.

## Rehydration

`yanzi rehydrate` loads the active project, finds the latest checkpoint, and prints the captures recorded after that checkpoint.

The output is continuity-oriented:

- checkpoint boundary first
- chronological captures after the boundary
- protocol annotations rendered explicitly
- open intent artifacts surfaced separately

If no checkpoint exists, rehydrate falls back to the latest captures for the active project instead of failing.

`yanzi rehydrate --dry-run` shows what would be loaded without printing the full sequence.

## Message Channel

`yanzi message send`, `yanzi message list`, and `yanzi message pull` store handoff notes as captures with message metadata.

This is useful for passing short instructions between operators and agents without mixing those notes into unrelated captures.

## Protocol Annotations

`@yanzi` lines are continuity annotations, not background automation.

They are recorded to preserve operational transitions such as:

- pause
- resume
- role changes
- export intent
- checkpoint intent

If a workflow requires an actual command, run the corresponding `yanzi` CLI command explicitly.
