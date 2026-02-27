

# Yanzi

[Yanzi](https://www.yanzi.io/) | [Chuck Sailer](https://chucksailer.me)

[![Release](https://img.shields.io/github/v/release/chuxorg/yanzi)](https://github.com/chuxorg/yanzi/releases/latest)
[![Build Status](https://img.shields.io/github/actions/workflow/status/chuxorg/yanzi/ci.yml?branch=main&label=build%20status)](https://github.com/chuxorg/yanzi/actions)

<small>Yanzi is a local-first workflow state manager for AI-assisted development.</small>


It captures prompt/response pairs as immutable artifacts, groups them into projects, and allows deterministic reconstruction of work using checkpoints.

- No daemon required.
- No orchestration layer.
- No hidden state.

---

## The Problem

If you use Copilot, CodeX, ChatGPT, or any LLM for development, you’ve likely experienced:

* Context drift
* Re-explaining the same architecture repeatedly
* Losing the thread after a session reset
* Not knowing which decisions came first
* “Almost working” states that you can’t reconstruct cleanly

LLMs do not remember your project.
They approximate it from context windows.

Yanzi solves a narrow problem:

> Preserve intent and enable deterministic recovery.

It does not summarize.
It does not infer.
It does not reinterpret history.

It records what happened.

---

## Core Concepts

### Intent

An immutable prompt/response pair.

### Project

A scoped boundary grouping intents.

### Checkpoint

A stability marker within a project.

### Rehydrate

Mechanical reconstruction of:

* Project
* Latest checkpoint
* All intents since checkpoint

No ML.
No heuristics.
Deterministic ordering only.

---

## Architecture

Yanzi is delivered as a single CLI with local SQLite-backed state.

HTTP mode is optional.
Local mode is default.

SQLite database location:

```
~/.yanzi/yanzi.db
```

---

## Installation

```bash
go install github.com/chuxorg/yanzi/cmd/yanzi@latest
```

Verify:

```bash
yanzi version
```

## Uninstall

```bash
rm $(which yanzi)
```

---

## Quick Start

Create and use a project:

```bash
yanzi project create "alpha"
yanzi project use "alpha"
```

Create a checkpoint:

```bash
yanzi checkpoint create --summary "Stabilized pipeline"
```

Rehydrate:

```bash
yanzi rehydrate
```

---

## Intended Workflow

Yanzi works best when paired with AI-assisted development.

A typical loop:

1. Prompt AI
2. Implement changes
3. Capture intent
4. Checkpoint before structural shifts
5. Rehydrate when context is lost

It is safe for humans to use directly.

It becomes valuable when used consistently during AI-driven work.

---

## Prompt for AI Systems

You may copy the following when onboarding an AI agent:

---

You are assisting with development in a repository that uses Yanzi.

Yanzi is a local-first workflow state manager.

Rules:

* Each prompt/response cycle may be captured as intent.
* Projects define context boundaries.
* Create checkpoints before major structural changes.
* Use `yanzi rehydrate` instead of summarizing history.
* Do not assume implicit memory.
* Treat state as mechanical and reconstructable.

Installation:
`go install github.com/chuxorg/yanzi/cmd/yanzi@latest`

---

## Design Principles

* Local-first
* Deterministic behavior
* No daemon required
* Agent-agnostic
* Shell-friendly
* Minimal surface area
* No speculative orchestration

---

## Non-Goals

Yanzi is not:

* An agent framework
* A memory embedding system
* A vector database
* A workflow engine
* A summarization layer
* A project management tool

It solves one problem:

Preserve intent. Enable recovery.

---

## Status

Stabilization phase.

Focus areas:

* Release pipeline hardening
* REST surface alignment
* Store abstraction (future)

No feature sprawl.

---

---
