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
mode        Show or set runtime mode.
project     Manage project context.
status      Show continuity and observability status.
intent      Manage intent artifacts.
context     Manage context artifacts.
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

## project

Create, select, and inspect the active project.

```bash
yanzi project create demo
yanzi project use demo
yanzi project current
yanzi project list
```

## capture

### Problem

Prompt and response history is easy to lose when work happens across terminals, editors, and agent sessions.

### Solution

`yanzi capture` stores one prompt/response pair as a durable record.

### Example

```bash
yanzi capture \
  --author "Ada" \
  --prompt "Summarize the current task" \
  --response "Rewrite the documentation."

yanzi capture \
  --author "Ada" \
  --prompt-file prompt.txt \
  --response-file response.txt \
  --meta area=docs
```

### Flags

- `--author <name>` required
- `--prompt <text>` or `--prompt-file <path>` required
- `--response <text>` or `--response-file <path>` required
- `--title <title>` optional
- `--source <source>` optional, default `cli`
- `--profile <name>` optional
- `--prev-hash <hash>` optional
- `--meta key=value` optional and repeatable

## checkpoint

### Problem

Replaying all project history from the beginning is unnecessary when a stable boundary already exists.

### Solution

`yanzi checkpoint` records named project boundaries and lists them later.

### Example

```bash
yanzi checkpoint create --summary "Initial project state"
yanzi checkpoint list
yanzi checkpoint list --all-projects
```

### Flags

- `create --summary "..."` required summary for a new checkpoint
- `list --all-projects` optional cross-project retrieval

## rehydrate

### Problem

Agents need the current project state without manually reconstructing it from every earlier record.

### Solution

`yanzi rehydrate` loads the latest checkpoint and the intent records that follow it.

### Example

```bash
yanzi status
yanzi rehydrate --dry-run
yanzi rehydrate
```

### Flags

- `--dry-run` preview what would load

`yanzi rehydrate` now prints a continuity summary before the checkpoint and capture blocks. The dry-run mode also shows continuity mode, depth, latest activity, and open work count.

`yanzi rehydrate --format json` emits a machine contract with `schema_version`, `kind`, `project`, and aligned checkpoint fields for deterministic consumers.

## status

### Problem

Operators and agents need a quick deterministic view of project continuity without exporting or replaying the entire timeline.

### Solution

`yanzi status` reports continuity mode, latest checkpoint anchor, last activity, continuity depth, recent activity, and unresolved task or change-request artifacts.

### Example

```bash
yanzi status
yanzi status --recent 10
yanzi status --format json
```

### Flags

- `--format <text|json>` optional, default `text`
- `--recent <n>` optional, default `5`

`yanzi status --format json` emits a stable machine-readable contract with explicit schema/version identity and deterministic activity ordering.

## export

### Problem

Project history and explicit context exports need deterministic output for scripts and agents.

### Solution

`yanzi export` writes project history or explicit claude-context exports to files in the current working directory.

### Example

```bash
yanzi export --format markdown
yanzi export --format json
yanzi export --format html --open
yanzi export --format claude-context
yanzi export --format markdown --meta type=context
```

### Flags

- `--format <markdown|json|html|claude-context>` required
- `--profile <name>` optional
- `--meta key=value` optional and repeatable
- `--include-deleted` optional
- `--open` only valid with `--format html`

Outputs:

- `YANZI_LOG.md`
- `YANZI_LOG.json`
- `YANZI_LOG.html`
- `CLAUDE_CONTEXT.md`

Yanzi does not interpret or rank results. It only filters and returns stored data.

`yanzi export --format json` includes `schema_version`, `kind`, and an explicit continuity summary so machine consumers can distinguish export shape from other JSON surfaces.

## `yanzi export --help`

```text
export args:
  --format <markdown|json|html|claude-context>
                        Export format. Required.
  --profile <name>      Optional profile filter.
  --meta key=value      Optional metadata filter (repeatable; exact match; AND).
  --include-deleted     Include tombstoned records.
  --open                Open generated html export in the default browser.
```

## message

### Problem

Independent agents or operators need a shared note channel without introducing a separate messaging service.

### Solution

`yanzi message` stores handoff notes as captures with message metadata.

### Example

```bash
yanzi message send \
  --to claude \
  --from operator \
  --channel handoff \
  --content "Continue from the latest checkpoint."

yanzi message list --to claude --channel handoff
yanzi message pull --to claude --channel handoff
```

### Flags

Subcommands:

- `send`
- `list`
- `pull`

`send` flags:

- `--to <name>` required
- `--from <name>` required
- `--channel <name>` optional
- `--title <title>` optional
- `--file <path>` or `--content <text>` required

`list` flags:

- `--to <name>` optional
- `--from <name>` optional
- `--channel <name>` optional
- `--include-deleted` optional
- `--limit <n>` optional

`pull` flags:

- `--to <name>` optional
- `--from <name>` optional
- `--channel <name>` optional
- `--include-deleted` optional

## context

Add, list, and show context artifacts.

Examples:

```bash
yanzi context add --type process_rule --title "Release rule" --file ./system-rules.md
yanzi context list --scope project
yanzi context list --all-projects
yanzi context show abc123def456
```

## intent

Add or list intent artifacts for the active project.

Examples:

```bash
yanzi intent add --title "Clarify export scope" --content "Export only deterministic artifacts."
yanzi intent list --type decision
yanzi intent list --all-projects
```

## bootstrap

Load context documents from `.yanzi/bootstrap.yaml`.

Flags:

- `--dry-run` validate bootstrap documents without loading them

Example:

```bash
yanzi bootstrap --dry-run
yanzi bootstrap
```

## rules

Capture rules files and export only rule records.

Examples:

```bash
yanzi rules add ./system-rules.md --scope global --priority critical
yanzi rules list --scope global
yanzi rules export --format markdown
yanzi rules export --format html --compose --profile engineer
```

## list

List captured intent records for the active project by default.

Flags:

- `--author <name>` optional
- `--source <source>` optional
- `--profile <name>` optional
- `--meta key=value` optional and repeatable
- `--all-projects` optional cross-project retrieval
- `--include-deleted` optional
- `--limit <n>` optional, default `20`

Example:

```bash
yanzi list --limit 10
yanzi list --all-projects
```

## local db resolution

Local commands resolve the SQLite database in deterministic order:

1. `YANZI_DB_PATH`
2. `db_path` from `~/.yanzi/config.yaml`
3. default `~/.yanzi/yanzi.db`

## show

Show a full intent record by id.

```bash
yanzi show 01HZX9Q4X8N9JZ1K2G9N8M4V3P
```

## verify

Verify the stored hash for an intent.

```bash
yanzi verify 01HZX9Q4X8N9JZ1K2G9N8M4V3P
```

## chain

Print a chain of intent records from oldest to newest.

```bash
yanzi chain 01HZX9Q4X8N9JZ1K2G9N8M4V3P
```

## delete and restore

```bash
yanzi delete 01HZX9Q4X8N9JZ1K2G9N8M4V3P --cascade
yanzi restore 01HZX9Q4X8N9JZ1K2G9N8M4V3P
```

## mode

```bash
yanzi mode
yanzi mode local
yanzi mode http
```

`http` mode writes the config only. It does not start `libraryd`.

## types

```bash
yanzi types list --json
```
