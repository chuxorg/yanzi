# Rehydrate

`yanzi rehydrate` works on the active project.

It does four things:

1. load the active project
2. find the latest checkpoint
3. read captures recorded after that checkpoint
4. print the ordered result

`yanzi rehydrate --dry-run` prints a summary instead of the full list.

Default output includes the active project, checkpoint summary, and readable continuity blocks for each post-checkpoint capture with author, title, prompt snippet, response snippet, and metadata summary.

It also surfaces protocol annotations explicitly and lists open intent artifacts so recovery output keeps unresolved work visible.

If the active project has no checkpoint, the command warns and falls back to the latest capture window for that project instead of failing.

`yanzi rehydrate --format json` prints the same rehydrate payload as machine-readable JSON, including checkpoint data when present and fallback indicators when a checkpoint is missing.
