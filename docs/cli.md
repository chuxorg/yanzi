# CLI Reference

Yanzi commands operate on local state by default. Several commands require an active project.

## Command List

```text
capture     Create a new intent record.
verify      Verify an intent by id.
chain       Print an intent chain by id.
list        List intent records.
show        Show intent details by id.
delete      Tombstone an intent or artifact by id.
restore     Remove tombstone metadata by id.
mode        Show or set runtime mode (local | http).
serve       Start the shared operational runtime.
project     Manage project context.
init        Create or bind a project to the current directory.
intent      Manage intent artifacts.
context     Manage context artifacts.
pack        Apply or export portable context packs.
bootstrap   Load ordered context documents from .yanzi/bootstrap.yaml.
rules       Manage rule metadata wrappers.
types       List canonical artifact types and aliases.
message     Manage thin message wrappers.
checkpoint  Manage checkpoints.
rehydrate   Rehydrate active project context.
export      Export active project history.
version     Print the CLI version.
```

## Global

```bash
yanzi --help
yanzi --version
yanzi version
```

---

## project

Create, select, and inspect the active project.

```bash
yanzi project create <name>
yanzi project use <name>
yanzi project current
yanzi project list
```

## init

Create or reuse a project and bind the current directory to it. Writes `.yanzi/project` in the working directory.

```bash
yanzi init
yanzi init <name>
```

If `<name>` is omitted, the command uses the existing bound project or prompts for one.

---

## capture

### Problem

Prompt and response history is easy to lose when work happens across terminals, editors, and agent sessions.

### Solution

`yanzi capture` stores one prompt/response pair as a durable append-only record.

### Flags

- `--author <name>` required
- `--prompt <text>` or `--prompt-file <path>` required (mutually exclusive)
- `--response <text>` or `--response-file <path>` required (mutually exclusive)
- `--title <title>` optional
- `--source <source>` optional, default `cli`
- `--profile <name>` optional
- `--prev-hash <hash>` optional, explicitly sets the chain predecessor
- `--meta key=value` optional, repeatable

### Examples

```bash
yanzi capture \
  --author "Ada" \
  --prompt "Summarize the current task" \
  --response "Rewrite the documentation."

yanzi capture \
  --author "Ada" \
  --prompt-file prompt.txt \
  --response-file response.txt \
  --meta area=docs \
  --meta phase=2
```

---

## checkpoint

### Problem

Replaying all project history from the beginning is unnecessary when a stable boundary already exists.

### Solution

`yanzi checkpoint` records named project boundaries and lists them.

### Subcommands

**create**

```bash
yanzi checkpoint create --summary "Initial project state"
```

Flags:
- `--summary "..."` required

**list**

```bash
yanzi checkpoint list
yanzi checkpoint list --all-projects
```

Flags:
- `--all-projects` optional, list across every project

---

## rehydrate

### Problem

Agents need the current project state without manually reconstructing it from every earlier record.

### Solution

`yanzi rehydrate` loads the latest checkpoint and the intent records that follow it, in chronological order.

If no checkpoint exists, it falls back to the most recent captures for the project.

### Flags

- `--dry-run` preview what would load without rendering it
- `--format text|json` output format, default `text`

### Examples

```bash
yanzi rehydrate --dry-run
yanzi rehydrate
yanzi rehydrate --format json
```

---

## export

### Problem

Project history and explicit context exports need deterministic output for scripts and agents.

### Solution

`yanzi export` writes project history or targeted context to files in the current working directory.

### Flags

- `--format <markdown|json|html|claude-context>` required
- `--profile <name>` optional filter
- `--meta key=value` optional, repeatable, AND-matched
- `--include-deleted` include tombstoned records

### Examples

```bash
yanzi export --format markdown
yanzi export --format json
yanzi export --format html
yanzi export --format claude-context
yanzi export --format markdown --meta type=context
yanzi export --meta type=context --meta subtype=rules --format markdown
```

### Outputs

| Flag | Output file |
|---|---|
| `--format markdown` | `YANZI_LOG.md` |
| `--format json` | `YANZI_LOG.json` |
| `--format html` | `YANZI_LOG.html` |
| `--format claude-context` | `CLAUDE_CONTEXT.md` |

Yanzi does not interpret or rank results. It filters and returns stored data only.

---

## serve

Start the shared operational HTTP runtime on localhost.

```bash
yanzi serve
yanzi serve --addr 127.0.0.1:9090
yanzi serve --shutdown-timeout 10s
```

### Flags

- `--addr <host:port>` optional, default `127.0.0.1:8080`
- `--shutdown-timeout <duration>` optional, default `5s`

### Behavior

- Binds to localhost only by default.
- Exposes the `/v0` REST API (see [API Reference](api/index.md)).
- Runs in the foreground; send `SIGINT` or `SIGTERM` to stop.
- Does not affect local mode operation — local commands remain available while the server is running.

---

## mode

Show or change the runtime mode.

```bash
yanzi mode           # show current mode
yanzi mode local     # set to local
yanzi mode http      # set to http
```

In `http` mode, supported intent commands route to the configured `base_url` instead of local storage. `http` mode does not start the server.

HTTP mode configuration (`~/.yanzi/config.yaml`):

```yaml
mode: http
base_url: http://127.0.0.1:8080
```

Commands that remain local-only regardless of mode: `checkpoint`, `rehydrate`, `export`, `delete`, `restore`.

---

## intent

Add or list intent artifacts for the active project.

### Subcommands

**add**

```bash
yanzi intent add --title "Clarify export scope"
yanzi intent add --title "Design decision" --type decision --content "Use append-only writes."
yanzi intent add --title "Spec note" --file ./spec.md
```

Flags:
- `--title <title>` required
- `--type <type>` optional, default `note`; see `yanzi types list --json` for valid values
- `--content <text>` or `--file <path>` optional
- `--metadata <json>` optional

**list**

```bash
yanzi intent list
yanzi intent list --type decision
yanzi intent list --all-projects
```

Flags:
- `--type <type>` optional filter
- `--all-projects` optional

---

## context

Add, list, and show context artifacts.

### Subcommands

**add**

```bash
yanzi context add --type process_rule --title "Release rule" --file ./SYSTEM_RULES.md
yanzi context add --type governance --title "Release rule" --file ./policy.md
yanzi context add --type coding_standard --title "Go style" --scope global --content "Use table-driven tests."
```

Flags:
- `--type <type>` required; see `yanzi types list --json` for valid values; `governance` is an alias for `process_rule`
- `--title <title>` required
- `--scope global|project` optional, default `project`
- `--file <path>` or `--content <text>` optional
- `--metadata <json>` optional

**list**

```bash
yanzi context list
yanzi context list --scope global
yanzi context list --all-projects
```

Flags:
- `--all-projects` optional

**show**

```bash
yanzi context show <id>
```

---

## pack

Apply or export portable context packs. A pack is a YAML file that describes a set of context artifacts and their sidecar content files.

### Subcommands

**apply**

Load all entries from a pack into the active project idempotently.

```bash
yanzi pack apply ./my-context-pack.yaml
```

Skips entries that already exist (matched by type and title).

**export**

Export the active project's visible context artifacts into a pack YAML and sidecar files.

```bash
yanzi pack export --output ./my-context-pack.yaml
```

Flags:
- `--output <file>` required

### Pack YAML format

```yaml
name: my-project
seed: base-context
version: "1.0"
context:
  - type: process_rule
    title: Release Rules
    file: 01-process_rule-release-rules.md
    scope: global
  - type: coding_standard
    title: Go Style
    file: 02-coding_standard-go-style.md
```

---

## bootstrap

Load ordered context documents from `.yanzi/bootstrap.yaml`.

```bash
yanzi bootstrap
yanzi bootstrap --dry-run
```

Flags:
- `--dry-run` validate documents without loading them

### bootstrap.yaml format

```yaml
documents:
  - type: governance
    title: System Rules
    path: SYSTEM_RULES.md
    scope: global
  - type: coding_standard
    title: Go Style
    path: docs/go-style.md
    scope: project
```

---

## rules

Capture rules files and operate on only rule records.

### Subcommands

**add**

```bash
yanzi rules add ./SYSTEM_RULES.md
yanzi rules add ./SYSTEM_RULES.md --scope global --priority critical
yanzi rules add ./SYSTEM_RULES.md --profile engineer
```

Flags:
- `<file>` required positional argument
- `--scope global|project` optional, default `global`
- `--priority <value>` optional
- `--profile <name>` optional

**list**

```bash
yanzi rules list
yanzi rules list --scope global
yanzi rules list --profile engineer
yanzi rules list --include-deleted
```

Flags:
- `--scope global|project` optional
- `--profile <name>` optional
- `--include-deleted` optional

**export**

```bash
yanzi rules export --format markdown
yanzi rules export --format html --compose --profile engineer
yanzi rules export --format markdown --scope global
```

Flags:
- `--format markdown|json|html` required
- `--compose` optional; groups output into system and profile sections (markdown/html only)
- `--scope global|project` optional
- `--profile <name>` optional
- `--include-deleted` optional

---

## message

Store and retrieve handoff notes between agents or operators.

### Subcommands

**send**

```bash
yanzi message send \
  --to claude \
  --from operator \
  --channel handoff \
  --content "Continue from the latest checkpoint."

yanzi message send --to codex --from ada --channel execution --file ./ready.md
```

Flags:
- `--to <name>` required
- `--from <name>` required
- `--channel <name>` optional
- `--title <title>` optional
- `--content <text>` or `--file <path>` required

**list**

```bash
yanzi message list
yanzi message list --to claude --channel handoff
```

Flags:
- `--to <name>` optional
- `--from <name>` optional
- `--channel <name>` optional
- `--include-deleted` optional
- `--limit <n>` optional

**pull**

Retrieve message notes as markdown.

```bash
yanzi message pull --to claude --channel handoff
```

Flags:
- `--to <name>` optional
- `--from <name>` optional
- `--channel <name>` optional
- `--include-deleted` optional

---

## list

List captured intent records.

```bash
yanzi list
yanzi list --limit 10
yanzi list --all-projects
yanzi list --meta type=context --meta subtype=rules
```

Flags:
- `--author <name>` optional
- `--source <source>` optional
- `--profile <name>` optional
- `--meta key=value` optional, repeatable, AND-matched
- `--all-projects` optional
- `--include-deleted` optional
- `--limit <n>` optional, default `20`

---

## show

Show a full record by id.

```bash
yanzi show <id>
```

---

## verify

Recompute and compare the SHA-256 hash for a stored intent.

```bash
yanzi verify <id>
```

Output: `✔ VALID` or `✖ INVALID`, plus stored and computed hashes.

---

## chain

Print the full chain of linked intent records from oldest to newest, following `prev_hash` links.

```bash
yanzi chain <id>
```

---

## delete

Tombstone a record without removing it from storage.

```bash
yanzi delete <id>
yanzi delete <id> --cascade
yanzi delete <id> --force
```

Flags:
- `--cascade` also tombstone dependent chain records
- `--force` allow tombstoning checkpoint-referenced artifacts

---

## restore

Remove a tombstone marker from a record.

```bash
yanzi restore <id>
```

---

## types

List canonical artifact types and alias mappings.

```bash
yanzi types list --json
```

Intent types: `change_request`, `checkpoint`, `decision`, `note`, `prompt`, `task`

Context types: `coding_standard`, `note`, `process_rule`, `reference`, `requirement`

Aliases: `governance` → `process_rule`

---

## Local Database Resolution

Local commands resolve the SQLite database in this order:

1. `YANZI_DB_PATH` environment variable
2. `db_path` in `~/.yanzi/config.yaml`
3. Default: `~/.yanzi/yanzi.db`
