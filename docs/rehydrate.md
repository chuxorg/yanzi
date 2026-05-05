# Rehydrate

`yanzi rehydrate` works on the active project.

It does four things:

1. load the active project
2. find the latest checkpoint
3. read captures recorded after that checkpoint
4. print the ordered result

`yanzi rehydrate --dry-run` prints a summary instead of the full list.

If the active project has no checkpoint, the command returns `no checkpoint found for active project`.
