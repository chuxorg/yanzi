# Local-First Operation

## Philosophy

Yanzi is local-first by design.

It uses a single SQLite database on the local machine so captures, checkpoints, context, and exports remain deterministic and easy to inspect.

This favors:

- simple installation
- explicit operator control
- deterministic recovery
- filesystem-local debugging

It does not target distributed coordination or background orchestration.

## SQLite Runtime Model

Yanzi configures SQLite for practical CLI usage:

- WAL mode for better read/write coexistence
- foreign keys enabled
- bounded busy timeout for short lock contention
- single-connection handles per process for deterministic local behavior

Startup and write operations use short bounded retries when SQLite reports transient lock contention.

## Supported Concurrency Expectations

Primary supported mode:

- sequential CLI workflows

Also supported:

- light concurrent local usage from multiple terminal processes
- short-lived overlapping writes where one process finishes quickly

Not yet a target:

- heavy concurrent orchestration
- many long-running writer processes
- distributed or daemon-managed coordination

## Contention Behavior

When two local writers overlap:

- the active writer keeps the lock until its write finishes
- the competing writer waits for SQLite busy handling and bounded Yanzi retries
- if the lock clears quickly, the competing write succeeds
- if the lock does not clear in time, Yanzi fails deterministically with an actionable lock message

Example operator message:

`sqlite database is locked by another writer; retry shortly`

## Operational Diagnostics

Yanzi tries to keep SQLite failures actionable.

Examples:

- lock contention: retry shortly after the active writer finishes
- invalid path: verify the configured path exists and is writable
- permission failure: check directory and file permissions
- unreadable database: inspect for corruption or a non-SQLite file at the configured path

## Recommended Multi-Agent Usage

Recommended:

- one active writer at a time when practical
- short CLI commands instead of long-held transactions
- checkpoints at meaningful boundaries
- exports and rehydrate for handoff/recovery rather than concurrent write-heavy coordination

Acceptable:

- two or a few local agents writing occasionally
- short overlapping capture operations

Avoid for now:

- high-frequency concurrent writes from many agents
- using Yanzi as a queue or workflow engine
