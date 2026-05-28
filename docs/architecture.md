# Architecture

## Problem

AI-assisted work needs durable state without depending on orchestration services or hidden runtime behavior.

## Solution

Yanzi is a local-first append-only datastore with a CLI over that state.

## Local-First

- default storage is local SQLite under the Yanzi state directory
- local mode does not require background services
- HTTP mode is optional and does not change the local-first model

## Append-Only

- intent records are appended with hashes
- checkpoints add project boundaries without rewriting earlier records
- context and intent artifacts are added as new records

## Datastore Only

- Yanzi stores and retrieves project state
- filtering is deterministic and explicit
- it does not rank, summarize, or orchestrate agents

## No Interpretation

- stored content is not interpreted by Yanzi
- metadata is matched exactly when filters are used
- meaning is left to the caller or agent using the stored data

## Delivery Authority

Repository delivery governance preserves separation of authority:

```text
Architect       -> Capability
Release Steward -> Phase Approval
Execution Agent -> Delivery
QA              -> Validation
Release Steward -> Release Decision
```

The Release Steward governs PR review, merge approval, checkpoint approval, release readiness, backlog state transitions, and convergence validation. The role does not redefine architecture, orchestrate agents, override branch protections, autonomously release, or invent requirements.

See [Release Steward Role](roles/release-steward.md).
