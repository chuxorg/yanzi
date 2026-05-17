# Canonical Governance Artifact Specifications

This document defines canonical artifact shape for deterministic context composition in Yanzi.

Scope of this document:
- artifact structure and operational contracts only
- no runtime behavior
- no orchestration logic
- no loader or parser definitions
- no YAML, JSON schema, or DSL concepts

## Pack

Purpose:
- Declarative governance manifest.
- Role-oriented context composition unit.
- Deterministic context loader by explicit declaration.

Allowed references:
- rules
- workflows
- seeds
- templates
- policies

Forbidden references:
- recursive references to other packs
- runtime state artifacts
- dynamic or implicit context expansion sources

Required sections:
- Name
- Role
- Purpose
- Loads
- Responsibilities
- Constraints

Optional sections:
- Notes
- Recommended Workflows
- Stack Guidance

Operational constraints:
- Must remain declarative and inspectable.
- Must not execute logic.
- Must not contain runtime state.
- Must not dynamically expand context.
- Must not mutate operational state.

## Seed

Purpose:
- Atomic operational knowledge artifact.
- Narrowly scoped reusable context fragment.

Allowed references:
- none

Forbidden references:
- other seeds
- packs
- runtime execution definitions

Required sections:
- Name
- Purpose
- Guidance
- Constraints

Optional sections:
- Notes

Operational constraints:
- Must remain focused and bounded.
- Must avoid composite behavior.
- Must avoid orchestration semantics.
- Must remain directly readable without expansion.

## Workflow

Purpose:
- Ordered operational procedure.
- Human-readable and audit-readable execution flow.

Allowed references:
- rules
- templates

Forbidden references:
- packs
- seeds
- policies as executable dependencies
- hidden branching constructs

Required sections:
- Name
- Purpose
- Preconditions
- Steps
- Validation
- Failure Handling

Optional sections:
- Notes
- Escalation

Operational constraints:
- Must not execute automatically.
- Must not define autonomous behavior.
- Must not contain hidden branching logic.
- Must keep sequence explicit and reviewable.

## Rule

Purpose:
- Non-overridable governance constraint.

Allowed references:
- none

Forbidden references:
- rules
- workflows
- packs
- seeds
- policies
- templates
- prompts
- state

Required sections:
- Name
- Scope
- Rule Statements
- Rationale

Optional sections:
- Examples
- Notes

Operational constraints:
- Must remain globally understandable.
- Must avoid ambiguity.
- Must avoid implementation-specific coupling.
- Must not contain workflow logic.

## Policy

Purpose:
- Project-specific operational decision.
- Mutable governance preference.

Allowed references:
- rules
- workflows

Forbidden references:
- packs
- seeds
- runtime state dependencies
- hidden composition mechanisms

Required sections:
- Name
- Purpose
- Policy Statements
- Applicability

Optional sections:
- Rationale
- Review Cadence
- Notes

Operational constraints:
- Must remain explicit.
- Must remain human-governed.
- Must remain inspectable.
- Must not define runtime execution behavior.

## Deterministic Dependency Direction

Canonical direction:
- Packs -> rules, workflows, seeds, templates, policies
- Workflows -> rules, templates
- Policies -> rules, workflows
- Rules -> none
- Seeds -> none

Deterministic constraints:
- No recursive pack composition.
- No hidden or automatic transitive expansion.
- No artifact-defined execution behavior.
- Composition intent must be visible from direct references.
