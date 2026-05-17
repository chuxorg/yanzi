# JSON Contracts

## Contract Rules

Yanzi JSON outputs are intended for deterministic machine consumption.

Current contract rules:

- every machine-oriented JSON surface includes `schema_version`
- every machine-oriented JSON surface includes a stable `kind`
- top-level field order is stable because outputs are encoded from structs, not ad hoc maps
- timestamps are emitted in UTC using RFC3339Nano-compatible strings
- arrays are deterministically ordered before encoding
- metadata maps are string-to-string maps and are rendered with deterministic key ordering by Go's JSON encoder

Backward compatibility policy:

- additive fields are acceptable in minor releases
- existing field names should not be renamed casually
- existing meanings should not drift without a schema version change
- removals or semantic breaks should be treated as major-version work

Future extension philosophy:

- prefer adding new optional fields over changing existing ones
- prefer new `kind` values over overloading one payload shape for multiple purposes
- keep contract shape explicit even when the data could be inferred indirectly

## `yanzi export --format json`

Top-level fields:

- `schema_version`
- `kind`
- `project`
- `exported_at`
- `version`
- `summary` optional
- `events`

Current `kind`:

- `history_export`

`summary` fields:

- `continuity_mode`
- `continuity_depth`
- `total_captures`
- `total_protocol_annotations`
- `total_checkpoints`
- `total_intent_artifacts`
- `visible_context_artifacts`
- `last_activity_at` optional
- `last_capture_at` optional
- `latest_checkpoint` optional

Event shapes:

- `capture`
- `checkpoint`
- `meta`

## `yanzi rehydrate --format json`

Top-level fields:

- `schema_version`
- `kind`
- `project`
- `has_checkpoint`
- `fallback`
- `fallback_reason` optional
- `fallback_limit` optional
- `checkpoint` optional
- `intents`

Current `kind`:

- `rehydrate`

Intent fields:

- `id`
- `timestamp`
- `author`
- `source_type`
- `title` optional
- `prompt`
- `response`
- `prompt_snippet`
- `response_snippet`
- `metadata` optional
- `hash`
- `prev_hash` optional

## `yanzi status --format json`

Top-level fields:

- `schema_version`
- `kind`
- `project`
- `project_created_at`
- `continuity_mode`
- `continuity_depth`
- `total_captures`
- `total_protocol_annotations`
- `total_checkpoints`
- `total_intent_artifacts`
- `visible_context_artifacts`
- `last_activity_at` optional
- `last_capture_at` optional
- `latest_checkpoint` optional
- `recent_activity`
- `unresolved_work`

Current `kind`:

- `status`

## `yanzi types list --json`

Top-level fields:

- `schema_version`
- `kind`
- `intent`
- `context`
- `aliases`

Current `kind`:

- `artifact_types`

## Shared Checkpoint Shape

Where a checkpoint object appears in JSON, the shape is intentionally aligned:

- `hash`
- `project`
- `summary`
- `created_at`
- `artifact_ids`
- `previous_checkpoint_id` optional

## Timestamp Contract

Yanzi emits timestamps as UTC strings.

Contract expectation:

- machine consumers should parse timestamps as RFC3339Nano-compatible values
- second-only timestamps remain valid because they are a subset of the same format family
- ordering logic should not rely on display formatting; it should rely on the encoded values directly
