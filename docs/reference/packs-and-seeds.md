# Packs and Seeds

## Overview

Packs and Seeds are yanzi's prompt composition layer. They are optional — yanzi functions as an operational log without them. When used, they add deterministic, composable, cost-efficient prompt assembly on top of the corpus.

## Concepts

### Seed

A discrete reusable context unit. Has a type, structured content sections, role access permissions, and a token estimate. Seeds are the unit of authorship and reuse.

### Pack

A named composition of Seeds for a specific agent role. Supports inheritance from a parent Pack. The unit of operational identity — packs represent the full context for one kind of agent or workflow.

### Prompt

The assembled result of Pack + Seeds + Task. Delivered to an agent. Captured to corpus as an artifact.

---

## Seed Types

| Type | Purpose |
|------|---------|
| `yanzi` | How to use yanzi (recommended in all packs) |
| `process` | How work is done |
| `guardrail` | What the agent must not do |
| `skill` | Domain knowledge |
| `personality` | Tone and communication style |

---

## System Roles (permission bits)

Roles are a permission bitmask. A higher-bit role includes all lower-bit permissions.

| Name | Bits | Description |
|------|------|-------------|
| `observer` | 1 | Read only, no destructive operations |
| `agent` | 3 | Read and write artifacts |
| `engineer` | 7 | Agent permissions plus code operations |
| `qa` | 15 | Engineer permissions plus validation |
| `pm` | 31 | QA permissions plus decisions |
| `release-steward` | 63 | PM permissions plus release operations |
| `admin` | 255 | All permissions |

Role access checks on seeds are advisory. A seed with a higher role requirement will produce a `role_access_violation` warning at compose time but will not be excluded.

---

## Pack Inheritance

A Pack can extend a parent Pack via `extends_id`. The child inherits all Seeds from the parent and can:
- Add new Seeds (appended after parent seeds)
- Override parent Seeds by name (matched by seed `name` field)
- Replace pack context (child context takes precedence)

Maximum inheritance depth: 10 levels. Circular inheritance returns an error.

---

## API Reference

### Seeds

| Method | Path | Scope | Description |
|--------|------|-------|-------------|
| `POST` | `/v0/seeds` | write | Create a seed |
| `GET` | `/v0/seeds` | read | List seeds |
| `GET` | `/v0/seeds/:id` | read | Get seed by artifact ID |
| `DELETE` | `/v0/seeds/:id` | admin | Delete seed |

Query params for `GET /v0/seeds`: `type`, `name`, `role_bits`

### Creating a Seed (JSON)

```json
POST /v0/seeds
Content-Type: application/json

{
  "name": "git-workflow",
  "seed_type": "process",
  "role_access_bits": 1,
  "description": "Git branching and commit workflow",
  "content": {
    "sections": [
      {
        "section": "overview",
        "type": "instruction",
        "text": "Always work on a feature branch from development."
      },
      {
        "section": "constraints",
        "type": "guardrail",
        "text": "Never commit directly to development or master."
      }
    ]
  }
}
```

### Creating a Seed (YAML)

```yaml
POST /v0/seeds
Content-Type: application/yaml

name: git-workflow
seed_type: process
role_access_bits: 1
description: Git branching and commit workflow
content:
  sections:
    - section: overview
      type: instruction
      text: |
        Always work on a feature branch from development.
    - section: constraints
      type: guardrail
      text: |
        Never commit directly to development or master.
```

---

### Packs

| Method | Path | Scope | Description |
|--------|------|-------|-------------|
| `POST` | `/v0/packs` | write | Create a pack |
| `GET` | `/v0/packs` | read | List packs |
| `GET` | `/v0/packs/:id` | read | Get pack by artifact ID |
| `DELETE` | `/v0/packs/:id` | admin | Delete pack |
| `POST` | `/v0/packs/compose` | write | Compose a pack into a prompt |

### Composing a Pack

```json
POST /v0/packs/compose
Content-Type: application/json

{
  "pack_artifact_id": "art_abc123",
  "task_content": "implement the login feature",
  "model_hint": "claude-opus-4-8",
  "options": {
    "include_assembled_prompt": true,
    "include_clipboard_string": true,
    "include_sections": true
  }
}
```

The response includes:
- `pack` — the resolved pack record
- `sections` — ordered list of content sections (when `include_sections: true`)
- `assembled_prompt` — full prompt with trust boundary markers (when `include_assembled_prompt: true`)
- `clipboard_string` — formatted for paste into any AI chat interface (when `include_clipboard_string: true`)
- `token_estimate` — approximate token counts by section
- `warnings` — advisory issues (missing seeds, role violations, injection patterns)

Token usage is automatically recorded when a project is active.

---

### Token Usage

| Method | Path | Scope | Description |
|--------|------|-------|-------------|
| `GET` | `/v0/tokens` | read | Get token usage summary |

Query params: `project`, `phase`, `task`, `since` (RFC3339)

```json
GET /v0/tokens?project=yanzi-cli-dev

{
  "project": "yanzi-cli-dev",
  "total_tokens": 42150,
  "by_phase": {
    "phase-1": 12000,
    "phase-2": 30150
  },
  "by_task": {
    "implementation": 20000,
    "review": 22150
  },
  "approximate": true
}
```

---

## Prompt Injection Protection

yanzi scans Seed content for known injection patterns at compose time:

- `ignore all previous instructions`
- `ignore previous instructions`
- `disregard the above`
- `forget everything`
- `you are now`
- `new persona`

When a pattern is found, a `injection_pattern` warning is added to the compose result. The seed is included anyway — warnings are advisory, not blocks.

Trust boundary markers are inserted in assembled prompts to reduce blast radius:

```
=== SYSTEM CONTEXT (trusted) ===
<pack context and seeds>

=== TASK ===
<task content>
=== END TASK ===
```

---

## Compose Warnings

| Code | Severity | Meaning |
|------|----------|---------|
| `missing_yanzi_seed` | advisory | Pack has no `yanzi`-type seed |
| `role_access_violation` | advisory | Seed requires higher role bits than pack |
| `injection_pattern` | warning | Suspicious content detected in seed |
| `missing_seed` | warning | A seed reference could not be resolved |
| `circular_inheritance` | error | Pack inheritance loop detected |

---

## Auth Scopes

Seed and Pack endpoints follow standard yanzi auth scope rules:

- `GET` endpoints → `read` scope
- `POST` endpoints → `write` scope
- `DELETE` endpoints → `admin` scope
- `POST /v0/packs/compose` → `write` scope
