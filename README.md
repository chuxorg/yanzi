# Yanzi

[![QA Build](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml)
[![Release](https://github.com/chuxorg/yanzi/actions/workflows/release.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/release.yml)

> Intent Integrity for AI-Assisted Development

---

## The Problem

AI-assisted development has changed how software is built.

We no longer write every line of code — we collaborate with AI agents.

But this introduces a critical failure:

**Intent is ephemeral.**

* Prompts are not stored
* Decisions are not preserved
* Context is lost between sessions
* Agents drift from original goals

---

## The Insight

This is not an AI problem.

It is a **state management problem**.

There is no system of record for:

* What we asked
* Why we asked it
* What the AI decided

---

## The Solution: Yanzi

Yanzi introduces:

**Intent as a first-class artifact**

Every AI interaction becomes:

* Structured
* Immutable
* Chronological
* Verifiable

---

## What Yanzi Is

* Local-first CLI
* Append-only ledger of intent
* Deterministic recovery system
* Agent-agnostic

---

## What Yanzi Is Not

* Not an orchestrator
* Not a framework
* Not tied to any AI vendor
* Not automating decisions

---

## Core Concepts

### Intent

Prompt + response pair

### Context

Rules, standards, references

### Checkpoints

Stable recovery points

### Rehydration

Reconstructs working state

---

## Why It Matters

Without Yanzi:

* You guess where you left off
* You re-explain context
* Agents behave inconsistently

With Yanzi:

* State is deterministic
* Context is preserved
* Workflows are reproducible

---

## Example

Without Yanzi:
AI deletes a production database with no trace.

With Yanzi:

* Rule exists: "Do not drop production database"
* Context persists
* Behavior is predictable

---

## Quick Start

```bash
brew install chuxorg/yanzi/yanzi
```

```bash
yanzi project create my-project
yanzi capture --prompt-file prompt.txt --response-file response.txt
yanzi checkpoint create --summary "Initial state"
yanzi rehydrate
```

---

## Documentation

See full documentation:

* [/docs](docs/index.md)
* [/docs/whitepaper.md](docs/whitepaper.md)

---

## Philosophy

Yanzi focuses on:

**Preserving intent, not orchestrating it**

---

## Status

Active development
Local-first
Production-focused

---

## License

MIT
