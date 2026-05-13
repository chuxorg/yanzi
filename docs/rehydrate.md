# Rehydrate

`yanzi rehydrate` works on the active project.

It does four things:

1. load the active project
2. find the latest checkpoint
3. read captures recorded after that checkpoint
4. print the ordered result

`yanzi rehydrate --dry-run` prints a summary instead of the full list.

Default output includes the active project, checkpoint summary, continuity count, chronology note, and readable continuity blocks for each post-checkpoint capture with author, title, prompt snippet, response snippet, and metadata summary.

If the active project has no checkpoint, the command warns and falls back to the latest capture window for that project instead of failing.

`yanzi rehydrate --format json` prints the same rehydrate payload as machine-readable JSON, including checkpoint data when present, `intent_count`, and fallback indicators when a checkpoint is missing.

The continuity list is rendered oldest to newest so an interrupted operator or agent can read forward from the stable checkpoint boundary.
