# MCP Interface

## Purpose

MCP is treated as an input/output interface to the Context Library.

Yanzi exposes operational context through MCP so MCP-capable tools can read and write shared deterministic state. MCP support enables prompt/result exchange and lineage continuity across tools without redefining Yanzi semantics.

## Relationship to Existing Contracts

This specification depends on:

- [Context Library Contract](context-library-contract.md)
- [Storage Provider Contract](storage-provider-contract.md)
- [Operational API](operational-api.md)

MCP semantics derive from the Context Library contract. MCP is another interface over the same operational substrate.

## Non-Goals

MCP support is not:

- Orchestration.
- Workflow automation.
- Autonomous execution.
- Multi-agent scheduling.
- Distributed coordination.
- A reasoning engine.
- An AI platform.

## Core Principles

- Deterministic operational state.
- Append-only provenance.
- Explicit lineage.
- Source attribution.
- Human-governed authority.
- Interface neutrality.
- Operational continuity across tools.
- Local-first compatibility.

## MCP Positioning

Yanzi is not an MCP platform.

MCP is a transport/interface mechanism. Yanzi provides operational memory and provenance continuity behind MCP-capable systems.

Yanzi focuses on operational determinism and lineage integrity, not agent behavior.

## Canonical MCP Use Case

Intended operational flow:

- ChatGPT writes a prompt artifact.
- CodeX loads the pending prompt.
- CodeX executes work.
- CodeX records result artifacts.
- ChatGPT reviews lineage and continues.

Clarifications:

- This is shared state, not orchestration.
- Yanzi records operational lineage.
- Humans and tools retain operational authority.

## MCP Resource Concepts

Conceptual MCP-accessible resources include:

- artifacts
- prompts/intents
- results
- projects
- checkpoints
- packs
- lineage
- rehydration context

These resources map to Context Library semantics and do not introduce separate artifact meaning.

## Conceptual MCP Operations

Conceptual MCP operations may include:

- `list_projects`
- `current_project`
- `create_artifact`
- `get_artifact`
- `list_artifacts`
- `list_pending_prompts`
- `claim_prompt`
- `complete_prompt`
- `list_checkpoints`
- `rehydrate`
- `export_context`
- `retrieve_lineage`
- `install_pack`

These are conceptual only and are not final API signatures.

## Prompt/Intent Semantics

- Prompts are operational artifacts.
- Prompts should preserve author and source attribution.
- Prompts may conceptually target a tool or role.
- Prompts remain part of operational lineage.
- Prompt retrieval does not imply orchestration ownership or scheduling authority.

## Result Semantics

- Execution results become lineage-linked artifacts.
- Outputs preserve tool attribution.
- Result artifacts should reference originating intent where possible.
- Verification and convergence checks may operate over these artifacts later.

## Shared Operational State Model

- Multiple tools may interact with the same project lineage.
- Lineage continuity matters more than session continuity.
- Artifacts are authoritative, not ephemeral chat sessions.
- Tools may disconnect and reconnect without losing operational state.

## Source Attribution Requirements

- Every MCP write should preserve source metadata.
- Tools must remain distinguishable in artifact provenance.
- Imports and exports should preserve provenance where possible.
- Operational lineage should remain inspectable.

## Rehydration Through MCP

- MCP tools may request deterministic operational context.
- Rehydration should compose checkpoints and post-checkpoint lineage.
- Yanzi provides reconstructable operational state.
- Yanzi does not provide hidden or magical memory.

## Runtime Relationship

- MCP interfaces will likely be hosted by optional runtime/daemon layers.
- Local CLI usage may bypass MCP entirely.
- MCP should not become mandatory for Yanzi usage.
- Runtime-hosted MCP can enable shared/team workflows.

## Security and Authority Boundaries

Future direction:

- Authentication and RBAC likely belong in runtime layers.
- MCP tools should not receive implicit operational authority.
- Write permissions may eventually be role-scoped.
- Operational governance remains human-controlled.

## Failure and Consistency Semantics

- Failures must be explicit.
- No silent lineage mutation.
- Imports should be deterministic where possible.
- Verification support should be available for provenance and integrity checks.
- Append-only expectations remain in force.
- Conflicts must be visible.

## Federation Compatibility

- MCP may operate against federated Yanzi nodes in the future.
- MCP should remain transport-neutral.
- Federation semantics belong to federation protocols, not MCP itself.

## Future Compatibility

This contract is intended to remain compatible with:

- REST API
- federation
- packs
- connector runtime
- UI
- IDE integrations
- enterprise/shared deployments

## Summary

MCP support allows AI tools to share deterministic operational state through Yanzi. Yanzi preserves provenance continuity and operational lineage. Yanzi does not orchestrate agents or automate operational authority.
