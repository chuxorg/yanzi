# Yanzi Architecture

Yanzi is a local-first workflow state manager designed to support AI-assisted development.

The system captures prompt/response cycles as immutable artifacts and provides deterministic recovery through checkpoints.

The architecture intentionally avoids orchestration layers and distributed complexity.

## Component Structure

core → library → cli

### core

Defines domain primitives.

Examples:
- Intent
- Project
- Checkpoint identifiers
- Domain errors

This package contains no storage or transport logic.

### library

Implements domain behavior and persistence.

Responsibilities:

- SQLite storage
- migrations
- project management
- checkpoint management
- deterministic rehydration

The library exposes a Go API used by the CLI and optional HTTP server.

### cli

User-facing interface.

Responsibilities:

- argument parsing
- command routing
- terminal output

The CLI contains no domain logic.

All domain operations call into the library.

## Storage

Default storage location:

~/.yanzi/yanzi.db

SQLite is used to keep installation simple and deterministic.

No daemon is required.

## Rehydration

Rehydration reconstructs project state by loading:

1. the active project
2. the most recent checkpoint
3. all artifacts created after the checkpoint

This operation is purely mechanical and contains no inference or summarization.