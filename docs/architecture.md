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

## Contract Stability

- machine-readable outputs should expose explicit schema/version identity
- JSON shapes should evolve additively when possible
- checkpoint and continuity semantics should remain aligned across rehydrate, status, and export surfaces
- orchestration concerns should not be added to storage or retrieval layers casually
