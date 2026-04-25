# Domain Model

Yanzi revolves around three primary entities.

## Intent

An immutable record of a prompt/response interaction.

Intent is the smallest artifact tracked by the system.

Properties:

- immutable
- timestamped
- ordered

## Project

Projects define context boundaries.

All intents belong to a project.

Projects allow developers to isolate workstreams and avoid context explosion.

## Checkpoint

A checkpoint marks a stable point in a project.

Checkpoints allow deterministic reconstruction of the project state.