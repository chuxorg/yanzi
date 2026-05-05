# Architecture

Yanzi stores data locally and exposes a CLI over that local state.

## Storage

- SQLite database under the Yanzi state directory
- append-only intent records with hashes
- checkpoint records per project
- context and intent artifacts for structured documents and notes

## Runtime Modes

- `local`: CLI reads and writes local SQLite state
- `http`: CLI sends supported intent commands to a configured server

Some commands are local-only, including `checkpoint`, `rehydrate`, `export`, `delete`, and `restore`.
