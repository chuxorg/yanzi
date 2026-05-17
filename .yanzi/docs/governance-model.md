# 1. Purpose

Yanzi stores and retrieves governance artifacts from the filesystem.

Yanzi is not an orchestration layer for agents. It does not schedule, direct, or supervise agent behavior. It also does not enforce a specific software methodology.

The role of Yanzi governance is deterministic context composition: selecting explicit artifacts and assembling them into operational context in a repeatable way.

Governance artifacts remain plain files, explicit, and inspectable at all times.

# 2. Core Philosophy

- Deterministic workflows: the same selected inputs should produce the same composed governance context.
- Operational simplicity: prefer clear file structures and direct references over complex indirection.
- Explicit composition: context is built from intentionally selected artifacts, not inferred expansion.
- Filesystem-first governance: artifacts live as normal files and directories, versioned with the project.
- Append-only operational history: state and operational records should be additive to preserve traceability.
- Human-readable artifacts: governance content should be understandable without specialized tooling.
- Minimal abstraction: avoid layers that hide what is being loaded or why.
- AI-agent interoperability: artifacts should be readable and usable across different agent systems.
- Governance-first engineering: define operational rules and flows explicitly before relying on automation.

# 3. Governance Artifact Types

## Rules

Purpose: Define stable operational constraints and guardrails.

Scope: Broad, reusable directives that apply across workflows or tasks.

Constraints:
- Must be standalone.
- Must not reference other artifact types.
- Should remain concise and stable.

Examples:
- Coding safety constraints
- Review quality expectations
- Commit hygiene rules

## Workflows

Purpose: Describe repeatable operational sequences for common engineering tasks.

Scope: Step-oriented guidance for execution, handoff, and validation.

Constraints:
- May reference rules and templates.
- Should avoid embedding policy decisions that belong in policies.
- Should not contain runtime orchestration behavior.

Examples:
- Feature delivery flow
- Incident response flow
- Release preparation flow

## Packs

Purpose: Provide declarative manifests for selecting and composing governance context.

Scope: Context packaging for a role, task type, or operational scenario.

Constraints:
- May reference rules, workflows, seeds, templates, and policies.
- Should be declarative and explicit.
- Should not include hidden or recursive expansion behavior.

Examples:
- Engineer pack
- QA pack
- Release coordinator pack

## Seeds

Purpose: Provide baseline initialization content for new governance or project setup states.

Scope: Starting artifacts used as initial context inputs.

Constraints:
- Must be standalone.
- Must not reference other artifact types.
- Should remain generic and reusable.

Examples:
- Initial governance seed
- New repo onboarding seed
- Baseline prompt seed

## Policies

Purpose: Capture project-level governance decisions and operational boundaries.

Scope: Durable decision records that constrain how workflows and rules are applied.

Constraints:
- May reference rules and workflows.
- Should avoid procedural detail that belongs in workflows.
- Should remain explicit and auditable.

Examples:
- Branching policy
- Approval policy
- Artifact retention policy

## Templates

Purpose: Standardize recurring artifact formats and operational documents.

Scope: Reusable structures for consistent authoring.

Constraints:
- Should be format guidance, not execution logic.
- Should remain readable and minimal.
- Should not encode runtime behavior.

Examples:
- PR template
- Incident report template
- Governance change template

## Prompts

Purpose: Store reusable prompt artifacts for agent and human-assisted operations.

Scope: Task-specific or role-specific prompt content.

Constraints:
- Should be explicit and reviewable.
- Should not implicitly pull additional governance context.
- Should avoid hidden decision logic.

Examples:
- Code review prompt
- Debugging prompt
- Documentation update prompt

## State

Purpose: Record runtime, project, or session state relevant to governance operations.

Scope: Temporal operational data and history.

Constraints:
- Runtime/project/session information only.
- Prefer append-only updates for traceability.
- Must not act as a hidden dependency resolver.

Examples:
- Active session notes
- Last composition snapshot
- Operational event log

# 4. Deterministic Context Composition

Packs are declarative manifests. They describe what governance artifacts are included for a given context, rather than how to execute work.

Context must be intentionally loaded. Automatic expansion should be avoided because it reduces inspectability and predictability.

Agents should receive only the operational context required for their current task. Overloading context increases ambiguity and weakens determinism.

Composition must remain inspectable: selected artifacts and dependency paths should be visible and explainable from filesystem artifacts.

# 5. Dependency Rules

Allowed directions:

- Packs may reference rules, workflows, seeds, templates, and policies.
- Rules reference nothing.
- Seeds reference nothing.
- Workflows may reference rules and templates.
- Policies may reference rules and workflows.
- State stores runtime/project/session information only.

Reasoning:

- Prevent recursive context chains that are hard to audit.
- Prevent hidden expansion where one artifact silently pulls many others.
- Keep composition predictable, bounded, and easy to inspect.
- Preserve deterministic behavior by making dependency intent explicit.

# 6. Non-Goals

Yanzi governance is not:

- an orchestration engine
- a workflow runtime
- an autonomous agent framework
- a BPM/workflow engine
- a hidden automation system
- an AI decision engine

# 7. Future Direction

Potential evolution areas, without implementation commitment:

- optional structured metadata to improve artifact discoverability
- JIT context loading patterns that preserve explicit selection
- broader agent interoperability through stable artifact conventions
- deterministic pack resolution refinements for larger repositories
- governance portability across repositories with similar layouts
