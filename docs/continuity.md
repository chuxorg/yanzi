# Continuity Philosophy

## What Yanzi Is

Yanzi is a deterministic continuity CLI for AI-assisted development.

It preserves operational cognition as explicit local records:

- captures for prompt and response history
- checkpoints for stable recovery anchors
- artifacts for decisions, tasks, change requests, rules, and references
- exports for review, handoff, and audit

The model is append-only and local-first. Yanzi does not reinterpret stored records with AI. It stores them, orders them deterministically, and renders them in ways that make recovery easier for operators and agents.

## Operational Cognition Preservation

Operational continuity is not only about storage. It is about keeping enough explicit structure to answer:

- what work exists
- what changed most recently
- where the last stable anchor is
- what remains unresolved
- what should be recovered first after an interruption

Yanzi treats captures, checkpoints, and intent artifacts as durable continuity signals. Rehydrate, status, and exports expose those signals directly instead of generating inferred summaries.

## Checkpoint Mental Model

A checkpoint is a recovery anchor, not a controller.

It marks a known boundary in project history so recovery can start from the latest stable point and replay forward. If no checkpoint exists, Yanzi falls back to a deterministic recent-capture window instead of pretending that no continuity exists.

## Local-First Determinism

Yanzi is designed for local inspection and predictable behavior:

- SQLite on the local machine
- short-lived CLI execution
- deterministic ordering
- explicit operator control
- no hidden services

That keeps the operational model understandable during debugging, demos, and handoff scenarios.

## What Yanzi Is Not

Yanzi is not:

- an orchestrator
- an autonomous workflow manager
- an agent runtime
- a background daemon
- an AI reasoning engine
- a vector-memory or embedding system

It does not queue work, execute work on its own, or infer project state beyond explicit stored metadata such as checkpoint records and artifact status fields.

## Recovery Expectations

Recovery in Yanzi means:

- load the active project
- inspect continuity status
- recover from the latest checkpoint when one exists
- otherwise use the deterministic fallback window
- continue from explicit stored evidence

The goal is not to automate the operator away. The goal is to make operational recovery truthful, inspectable, and repeatable.
