# CLI Reference

## Global

```sh
yanzi --version      # Print CLI version
yanzi --help         # Show help
yanzi version        # Print CLI version
```

---

## capture

Record a prompt/response pair as an intent.

```sh
yanzi capture --author <name> --prompt <text> --response <text>
yanzi capture --author <name> --prompt-file <path> --response-file <path>
```

Options:

| Flag | Description |
|------|-------------|
| `--author <name>` | Required. Author name. |
| `--prompt <text>` | Prompt text (exclusive with `--prompt-file`). |
| `--prompt-file <path>` | Prompt file path (exclusive with `--prompt`). |
| `--response <text>` | Response text (exclusive with `--response-file`). |
| `--response-file <path>` | Response file path (exclusive with `--response`). |
| `--title <title>` | Optional title. |
| `--source <source>` | Optional source type (default: `cli`). |
| `--meta key=value` | Optional metadata (repeatable). |

---

## list

List captured intent records.

```sh
yanzi list
yanzi list --limit 20
yanzi list --author <name>
yanzi list --source <source>
yanzi list --meta key=value
```

Options:

| Flag | Description |
|------|-------------|
| `--author <name>` | Filter by author. |
| `--source <source>` | Filter by source type. |
| `--meta k=v` | Filter by metadata key/value (exact match, AND logic, repeatable). |
| `--limit <n>` | Max records to return (default: 20). |

---

## show

Show full details for a single capture.

```sh
yanzi show <intent-id>
```

---

## export

Export the active project history.

```sh
yanzi export --format markdown
yanzi export --format json
yanzi export --format html
yanzi export --format html --open
```

Options:

| Flag | Description |
|------|-------------|
| `--format <format>` | Required. `markdown`, `json`, or `html`. |
| `--open` | Open the HTML export in the default browser (only valid with `--format html`). |

Output files:

- `markdown` → `YANZI_LOG.md`
- `json` → `YANZI_LOG.json`
- `html` → `YANZI_LOG.html`

---

## checkpoint

Manage project checkpoints (milestone markers).

```sh
yanzi checkpoint create --summary "Description"
yanzi checkpoint list
```

---

## project

Manage project context.

```sh
yanzi project create <name>
yanzi project use <name>
yanzi project current
yanzi project list
```

---

## rehydrate

Reconstruct the active project context from recorded history.

```sh
yanzi rehydrate
```

---

## intent

Manage intent artifacts attached to the active project.

```sh
yanzi intent add --title "Description" --content "..."
yanzi intent list
```

---

## context

Manage context artifacts attached to the active project.

```sh
yanzi context add --type policy --title "Description" --file ./policy.md
yanzi context list
yanzi context list --type policy
```

---

## verify / chain

Verify or trace the hash chain for a capture.

```sh
yanzi verify <intent-id>
yanzi chain <intent-id>
```

---

## mode

Show or set the runtime mode.

```sh
yanzi mode           # Show current mode
yanzi mode local     # Set to local (default)
yanzi mode http      # Set to http
```
