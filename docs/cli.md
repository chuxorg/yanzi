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

Record a prompt and response pair.

Flags:

- `--author <name>` required
- `--prompt <text>` or `--prompt-file <path>` required
- `--response <text>` or `--response-file <path>` required
- `--title <title>` optional
- `--source <source>` optional, default `cli`
- `--profile <name>` optional
- `--prev-hash <hash>` optional
- `--meta key=value` optional and repeatable

Examples:

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

## checkpoint

Create or list checkpoints for the active project.

```bash
yanzi checkpoint create --summary "Initial project state"
yanzi checkpoint list
```

## rehydrate

Load the latest checkpoint and the captures recorded after it.

Flags:

- `--dry-run` preview what would load

Examples:

```bash
yanzi rehydrate --dry-run
yanzi rehydrate
```

## export

Write project history or filtered context to a file in the current working directory.

Flags:

- `--format <markdown|json|html|claude-context>` optional
- `--type <type[,type...]>` optional context type filter
- `--profile <name>` optional
- `--meta key=value` optional and repeatable
- `--fields <field[,field...]>` optional context field selection
- `--order <created_at|updated_at>` optional deterministic order
- `--limit <n>` optional result limit after filtering
- `--include-deleted` optional
- `--open` only valid with `--format html`

Outputs:

- `YANZI_LOG.md`
- `YANZI_LOG.json`
- `YANZI_LOG.html`
- `CLAUDE_CONTEXT.md`

Examples:

```bash
yanzi export --format markdown
yanzi export --format json
yanzi export --format html --open
yanzi export --format claude-context
```

## Retrieving Context

Yanzi supports filtering and retrieving stored context using:

- `--type`
- `--meta`
- `--fields`
- `--order`
- `--limit`

Get engineering rules:

```bash
yanzi export --type process_rule --meta role=engineer
```

Limit output:

```bash
yanzi export --limit 5
```

Select fields:

```bash
yanzi export --fields title,content
```

Combined:

```bash
yanzi export \
  --type process_rule \
  --meta role=engineer \
  --fields title,content \
  --limit 5
```

Yanzi does not interpret or rank results. It only filters and returns stored data.

## `yanzi export --help`

```text
export args:
  --format <markdown|json|html|claude-context>
                        Export format. Defaults to claude-context when omitted.
  --type <type[,type...]>
                        Optional context type filter list.
  --profile <name>      Optional profile filter.
  --meta key=value      Optional metadata filter (repeatable; exact match; AND).
  --fields <field[,field...]>
                        Optional context field list.
  --order <field>       Deterministic order field: created_at|updated_at.
  --limit <n>           Optional result limit after filtering.
  --include-deleted     Include tombstoned records.
  --open                Open generated html export in the default browser.
```

## message

Store and retrieve handoff notes as message captures.

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

Examples:

```bash
yanzi message send \
  --to claude \
  --from operator \
  --channel handoff \
  --content "Continue from the latest checkpoint."

yanzi message list --to claude --channel handoff
yanzi message pull --to claude --channel handoff
```

## context

Add, list, and show context artifacts.

Examples:

```bash
yanzi context add --type process_rule --title "Release rule" --file ./SYSTEM_RULES.md
yanzi context list --scope project
yanzi context show abc123def456
```

## intent

Add or list intent artifacts for the active project.

Examples:

```bash
yanzi intent add --title "Clarify export scope" --content "Export only deterministic artifacts."
yanzi intent list --type decision
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
yanzi rules add ./SYSTEM_RULES.md --scope global --priority critical
yanzi rules list --scope global
yanzi rules export --format markdown
yanzi rules export --format html --compose --profile engineer
```

## list

List captured intent records.

Flags:

- `--author <name>` optional
- `--source <source>` optional
- `--profile <name>` optional
- `--meta key=value` optional and repeatable
- `--include-deleted` optional
- `--limit <n>` optional, default `20`

Example:

```bash
yanzi list --limit 10
```

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
