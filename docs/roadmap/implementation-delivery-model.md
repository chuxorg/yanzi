# Implementation Backlog and Delivery Model

## Purpose

Architecture defines boundaries. The backlog converts architecture into execution.

Yanzi preserves operational continuity from planning through delivery by recording explicit artifacts, checkpoints, and release outcomes in a deterministic, inspectable workflow.

## Relationship to Existing Specs

This roadmap model derives from:

- [Platform Architecture](../specs/platform-architecture.md)
- [Context Library Contract](../specs/context-library-contract.md)
- [Runtime / Daemon Architecture](../specs/runtime-architecture.md)
- [Pack Specification](../specs/pack-specification.md)

This roadmap converts contracts into delivery structure and execution flow.

## Non-Goals

This is not:

- a Jira replacement
- workflow automation
- orchestration
- autonomous planning
- project management software

## Delivery Hierarchy

Delivery hierarchy:

`Initiative -> Capability -> Phase -> Task -> Artifact -> Checkpoint -> Release`

Definitions:

- Initiative: long-lived strategic outcome boundary.
- Capability: deliverable system capability aligned to architecture.
- Phase: bounded execution slice for a capability.
- Task: atomic executable unit of work.
- Artifact: explicit output record (spec, code, test, decision, docs, etc.).
- Checkpoint: deterministic continuity anchor after meaningful progress.
- Release: validated published outcome with versioned coverage.

## Initiative Model

Initiative fields:

- ID
- Title
- Status
- Description
- Capabilities

## Capability Model

Capability fields:

- ID
- Title
- Status
- Initiative
- Acceptance
- Dependencies

## Phase Model

Phase fields:

- ID
- Capability
- Scope
- Acceptance
- Deliverables
- Checkpoint

Phases should be executable in small PRs with clear closure criteria.

## Task Model

Task fields:

- ID
- Phase
- Status
- Inputs
- Outputs
- Acceptance
- Artifacts Produced

Tasks are atomic execution units intended for deterministic completion and review.

## Artifact Model

Artifact examples:

- spec
- decision
- implementation
- migration
- test
- docs
- checkpoint
- release

Artifacts remain first-class delivery evidence and continuity anchors.

## Checkpoint Model

Checkpoint model:

- Purpose: preserve deterministic continuity at meaningful execution boundaries.
- Trigger: completion of a phase slice, risk boundary, or release-prep milestone.
- Acceptance: checkpoint references validated artifacts and explicit scope.
- Rehydration expectations: equivalent checkpoint inputs should support equivalent context reconstruction.

## Release Model

Release model fields:

- Version
- Capability coverage
- Validation
- Release notes

## Delivery Workflow

Canonical workflow:

`Architecture -> Capability -> Phase -> Task -> Execute -> Capture -> Checkpoint -> QA -> Release`

## Operational Governance

Governance rule:

- Humans approve.
- Agents execute.
- Yanzi records.

Agents may propose plans or edits. Humans retain final operational authority.

## Ownership Model

Ownership modes:

- human
- agent
- mixed

## Status Model

Status values:

- planned
- active
- blocked
- complete
- deferred

## Backlog Hygiene

Backlog hygiene rules:

- keep phases small
- checkpoint frequently
- close execution loops
- preserve lineage continuity

## Implementation Sequencing

Initial capability sequence:

- CAP-001 Storage Abstraction
- CAP-002 REST API
- CAP-003 Pack Installation
- CAP-004 Runtime Foundation
- CAP-005 MCP Interface
- CAP-006 Connector Runtime
- CAP-007 UI Operational Interface

Sequence is revisable based on validation outcomes and dependency learning.

## Summary

Yanzi preserves operational continuity from planning through release.

## Future Compatibility Direction

Future roadmap compatibility direction includes:

- pack registries
- federation expansion
- runtime-hosted operations
- release governance hardening
- IDE integrations

These are compatibility directions, not implementation commitments.
