# Architectural Guardrails

## Purpose

These guardrails explain what belongs in Yanzi today and what does not.

They are intended for contributors making architectural changes, not for marketing.

## Current Boundary

Yanzi today is:

- a local-first CLI
- a deterministic library over explicit stored records
- a continuity and retrieval tool
- an export and recovery surface

Yanzi today is not:

- an orchestrator
- an autonomous agent runtime
- a background service
- a queue
- a workflow engine
- a vector-memory system

## Library vs Future Agentd

Library and CLI responsibilities today:

- record persistence
- deterministic retrieval
- checkpoint handling
- continuity rendering
- machine-readable contract output
- local operational diagnostics

Potential future `agentd` responsibilities, if ever added later:

- orchestration
- background coordination
- long-lived runtime state
- scheduling or lifecycle control

Those responsibilities should not be smuggled into the current library or CLI as “small helpers.” If a feature requires background coordination or hidden control flow, it is outside the current architecture.

## Why Orchestration Is Excluded

The current shape is intentional:

- it keeps failure modes inspectable
- it preserves deterministic local behavior
- it keeps the storage model understandable
- it avoids aspirational semantics that runtime behavior cannot uphold

This constraint matters more than adding convenience features that blur the contract.

## Schema Evolution Expectations

Storage and machine-readable output contracts should evolve conservatively.

Guidelines:

- prefer additive changes
- preserve existing field meanings
- version contracts explicitly when semantics change
- do not redesign storage casually to support hypothetical future systems

## Governance References

Repository execution and release discipline are tracked in:

- `/system-rules.md`
- [docs/dev/RELEASE_PROTOCOL.md](dev/RELEASE_PROTOCOL.md)
