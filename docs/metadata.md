# Metadata Semantics

## Purpose

Metadata is currently the operational semantic layer in Yanzi.

The datastore remains intentionally simple. Meaningful grouping, filtering, continuity tagging, and lightweight workflow semantics are carried by explicit metadata rather than by hidden inference or lifecycle engines.

## Core Guidance

Prefer small, explicit, exact-match metadata values.

Recommended patterns:

- `area=auth`
- `component=cli`
- `phase=validation`
- `profile=engineer`
- `owner=core`
- `status=in_progress`
- `channel=handoff`
- `to=codex`
- `from=operator`

Avoid:

- large free-form paragraphs in metadata
- values that require fuzzy matching
- storing transient presentation details as metadata

## Continuity-Focused Usage

Good continuity metadata helps answer:

- what subsystem the record belongs to
- which phase of work it represents
- whether it is still open
- which operator or agent owns follow-up

Useful keys for captures:

- `area`
- `component`
- `phase`
- `profile`

Useful keys for explicit intent artifacts:

- `status`
- `owner`
- `priority`
- `area`

## `task` and `change_request` Conventions

Yanzi does not currently impose a workflow engine, but explicit conventions are useful.

Recommended artifact types for open work:

- `task`
- `change_request`

Recommended `status` values:

- `open`
- `in_progress`
- `blocked`
- `in_review`
- `resolved`
- `closed`
- `rejected`

Current status behavior:

- unresolved-work views treat `task` and `change_request` as open unless `status` is clearly terminal
- terminal values currently include `done`, `completed`, `complete`, `closed`, `resolved`, `cancelled`, `canceled`, and `rejected`

If you want deterministic open-work visibility, set `status` explicitly.

## Retrieval and Scoping Implications

Metadata filters are exact-match and AND-combined.

Implications:

- `--meta area=auth --meta phase=validation` returns records matching both values
- `Auth` and `auth` are different values unless you standardize casing yourself
- inconsistent synonyms reduce retrieval quality

For consistency:

- use lowercase keys
- prefer lowercase, underscore-separated values when values are enum-like
- reserve human-readable prose for prompt, response, title, or artifact content

## Reserved and System-Like Metadata

Some metadata is operationally significant:

- `project`
- `scope`
- tombstone fields such as `deleted` and `deleted_at`

Those fields should not be repurposed casually by callers. If you need additional semantics, add new keys instead of overloading system-like ones.
