# API Reference

This is the authoritative reference for the Yanzi HTTP API (v0).

For CLI usage, see the [CLI Reference](../cli.md).

## Overview

The Yanzi HTTP API is a local-only REST interface exposed by `yanzi serve`. It is not a remote or cloud API.

Start the server before making requests:

```bash
yanzi serve
# Runtime listening on http://127.0.0.1:8080
```

All requests and responses use `application/json`. All timestamps are RFC3339Nano UTC strings.

**Base URL:** `http://127.0.0.1:8080` (default)

**API version prefix:** `/v0`

## Error Format

All error responses share a common envelope:

```json
{
  "error": {
    "code": "machine_readable_code",
    "message": "human-readable description"
  }
}
```

Common error codes: `validation_failed`, `not_found`, `method_not_allowed`, `config_load_failed`, `malformed_json`.

---

## Endpoints

### `GET /v0/health`

Returns runtime and storage provider health status.

**Response `200`:**

```json
{
  "version": "2.10.0",
  "mode": "local",
  "runtime": {
    "mode": "shared",
    "started_at": "2026-05-28T12:00:00.000000000Z"
  },
  "provider": {
    "name": "sqlite",
    "status": "ready",
    "error": ""
  }
}
```

- `runtime` — present only when `yanzi serve` is running
- `provider.status` — `ready` or `unavailable`

---

### `GET /v0/rehydrate`

Returns the active project's latest checkpoint and all intent records captured after it, oldest-first.

Falls back to the last 10 captures if no checkpoint exists. Requires an active project.

**Response `200`:**

```json
{
  "project": "my-project",
  "has_checkpoint": true,
  "fallback": false,
  "fallback_reason": "",
  "fallback_limit": 0,
  "checkpoint": {
    "hash": "abc123...",
    "project": "my-project",
    "summary": "Initial project state",
    "created_at": "2026-05-28T10:00:00.000000000Z",
    "artifact_ids": [],
    "previous_checkpoint_id": ""
  },
  "intents": [
    {
      "id": "01HZX9...",
      "timestamp": "2026-05-28T11:00:00.000000000Z",
      "author": "Ada",
      "source_type": "cli",
      "title": "Capture note",
      "prompt": "Full prompt text",
      "response": "Full response text",
      "prompt_snippet": "Full prompt t...",
      "response_snippet": "Full response...",
      "metadata": {"phase": "1"},
      "hash": "def456...",
      "prev_hash": ""
    }
  ]
}
```

- `checkpoint` — `null` in fallback mode
- `prompt_snippet` / `response_snippet` — truncated to 160 characters
- `fallback` — `true` when no checkpoint exists

**Errors:**
- `400 active_project_not_set` — no active project
- `404 project_not_found` — active project does not match a stored project

---

### `GET /v0/artifacts`

List artifact records for the active project.

**Query parameters:**

| Parameter | Description | Default |
|---|---|---|
| `author` | Exact-match author filter | — |
| `source` | Exact-match source type filter | — |
| `limit` | Maximum records to return | `20` |
| `profile` | Profile metadata filter | — |
| `meta` | `key=value` metadata filter; repeatable; AND-matched | — |
| `all-projects` | `true` to list across all projects | `false` |
| `include-deleted` | `true` to include tombstoned records | `false` |

**Response `200`:**

```json
{
  "artifacts": [
    {
      "id": "01HZX9...",
      "created_at": "2026-05-28T11:00:00.000000000Z",
      "project": "my-project",
      "author": "Ada",
      "source": "cli",
      "title": "Capture note",
      "metadata": {"phase": "1"}
    }
  ]
}
```

**Errors:**
- `400 invalid_request` — invalid query parameter
- `400 unsupported_mode` — server is not in local mode

---

### `POST /v0/artifacts`

Capture a new artifact record.

**Request body:**

```json
{
  "author": "Ada",
  "prompt": "Summarize the current task",
  "response": "Write the documentation.",
  "title": "Optional title",
  "source_type": "api",
  "metadata": {
    "phase": "1",
    "capability": "CAP-100"
  },
  "prev_hash": "",
  "project": ""
}
```

Required: `author`, `prompt`, `response`. Optional: `title`, `source_type` (defaults to `cli`), `metadata`, `prev_hash`, `project`.

**Response `201`:**

```json
{
  "id": "01HZX9...",
  "created_at": "2026-05-28T11:00:00.000000000Z",
  "author": "Ada",
  "source_type": "api",
  "title": "Optional title",
  "prompt": "Summarize the current task",
  "response": "Write the documentation.",
  "metadata": {"phase": "1"},
  "prev_hash": "",
  "hash": "abc123..."
}
```

**Errors:**
- `400 validation_failed` — missing required field
- `400 malformed_json` — invalid JSON body

---

### `GET /v0/artifacts/:id`

Get a single artifact record by ID.

**Response `200`:**

```json
{
  "artifact": {
    "id": "01HZX9...",
    "created_at": "2026-05-28T11:00:00.000000000Z",
    "project": "my-project",
    "author": "Ada",
    "source": "cli",
    "title": "Capture note",
    "prompt": "Full prompt text",
    "response": "Full response text",
    "metadata": {"phase": "1"},
    "prev_hash": "",
    "hash": "abc123..."
  }
}
```

**Errors:**
- `404 artifact_not_found`

---

### `GET /v0/verify/:id`

Recompute and compare the SHA-256 hash for a stored artifact.

Also accessible at `GET /v0/intents/:id/verify`.

**Response `200`:**

```json
{
  "id": "01HZX9...",
  "valid": true,
  "stored_hash": "abc123...",
  "computed_hash": "abc123...",
  "prev_hash": "",
  "error": null
}
```

- `valid` — `true` if hashes match
- `error` — non-null string if hash computation failed

**Errors:**
- `404 intent_not_found`

---

### `GET /v0/chain/:id`

Walk the `prev_hash` chain from the given artifact, returning records oldest-first.

Also accessible at `GET /v0/intents/:id/chain`.

**Response `200`:**

```json
{
  "head_id": "01HZX9...",
  "length": 3,
  "intents": [
    {
      "id": "01HABC...",
      "created_at": "...",
      "author": "Ada",
      "source_type": "cli",
      "prompt": "...",
      "response": "...",
      "hash": "aaa...",
      "prev_hash": ""
    }
  ],
  "missing_links": []
}
```

- `missing_links` — hashes that could not be resolved in storage

**Errors:**
- `404 intent_not_found`

---

### `GET /v0/export/markdown`
### `GET /v0/export/json`
### `GET /v0/export/html`

Export project history. Returns the file body directly (not a JSON envelope).

**Query parameters:**

| Parameter | Description |
|---|---|
| `project` | **Required.** Project name to export. |
| `include_deleted` | `true` to include tombstoned records |
| `profile` | Filter by profile metadata value |
| `meta_<key>` | Metadata filter; e.g. `meta_type=context` |

**Response `200`** content types:
- `markdown` → `text/markdown; charset=utf-8`
- `json` → `application/json; charset=utf-8`
- `html` → `text/html; charset=utf-8`

**Errors:**
- `400 validation_failed` — `project` parameter missing
- `404 export_not_found` — unsupported format

---

### `GET /v0/projects`

List all projects.

**Response `200`:**

```json
{
  "projects": [
    {
      "name": "my-project",
      "description": "Optional description",
      "created_at": "2026-05-28T10:00:00.000000000Z"
    }
  ]
}
```

---

### `POST /v0/projects`

Create a new project.

**Request body:**

```json
{
  "name": "my-project",
  "description": "Optional description"
}
```

**Response `201`:** same shape as a single project object.

---

### `GET /v0/projects/current`

Get the active project.

**Response `200`:**

```json
{
  "project": {
    "name": "my-project",
    "description": "",
    "created_at": "2026-05-28T10:00:00.000000000Z"
  }
}
```

`project` is `null` when no active project is set.

---

### `POST /v0/projects/current`

Set the active project.

**Request body:**

```json
{ "name": "my-project" }
```

**Response `200`:** same shape as `GET /v0/projects/current`.

**Errors:**
- `404 not_found` — project does not exist

---

### `GET /v0/checkpoints`

List checkpoints for the active project.

**Query parameters:**

| Parameter | Description |
|---|---|
| `all_projects` | `true` to list across all projects |

**Response `200`:**

```json
{
  "checkpoints": [
    {
      "hash": "abc123...",
      "project": "my-project",
      "summary": "Initial project state",
      "created_at": "2026-05-28T10:00:00.000000000Z",
      "artifact_ids": [],
      "previous_checkpoint_id": ""
    }
  ]
}
```

---

### `POST /v0/checkpoints`

Create a new checkpoint for the active project.

**Request body:**

```json
{
  "summary": "End of sprint 1",
  "artifact_ids": []
}
```

Required: `summary`. `artifact_ids` defaults to `[]`. The `project` field is ignored — the active project is always used.

**Response `201`:** same shape as a single checkpoint object.

**Errors:**
- `400` — no active project set, or summary is empty

---

## Constraints

- The API is **local-only** by default (`127.0.0.1`). No authentication or TLS.
- Only `local` storage mode is supported by the server.
- CLI commands `checkpoint`, `rehydrate`, `export`, `delete`, and `restore` remain local-only and do not route through the HTTP API.
- The `claude-context` export format is CLI-only; the API supports `markdown`, `json`, and `html`.

## Generated Package References

Low-level Go package documentation (generated by gomarkdoc):

- [CLI Package](cmd.md)
- [Internal Packages](internal.md)
- [Combined Reference](../API.md)
