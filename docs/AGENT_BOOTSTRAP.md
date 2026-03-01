# Yanzi Agent Bootstrap

Purpose: provide deterministic operating rules for AI agents using the Yanzi CLI.

## 1. Role Declaration
- Agents must declare role at session start:
  - `Role: Engineer`
- If no role is declared, default role is `Engineer`.

## 2. Meta-Command Grammar
- Meta-commands must:
  - start at the beginning of the line
  - use prefix `@yanzi`
  - be single-line commands

Supported meta-commands:
- `@yanzi pause`
- `@yanzi resume`
- `@yanzi checkpoint "Summary"`
- `@yanzi export`
- `@yanzi role <RoleName>`

## 3. State Rules
- Pause affects capture only.
- Meta-commands are allowed while paused.
- State-changing commands must acknowledge execution.
- Meta-commands must be captured as intent events.

## 4. Capture Expectations
- Major structural decisions must be checkpointed.
- Role switches must be explicit.
- Avoid silent structural changes.

## 5. CLI Command Surface (Current)
Primary usage:
- `yanzi <command> [args]`

Commands:
- `capture` create a new intent record
- `verify` verify an intent by id
- `chain` print intent chain by id
- `list` list intent records
- `show` show intent details by id
- `mode` show/set runtime mode (`local|http`)
- `project` manage project context
- `checkpoint` manage checkpoints
- `rehydrate` rehydrate active project context
- `export` export active project history
- `version` print CLI version

## 6. CLI Arguments (Current)
Capture:
- `--author <name>` required
- `--prompt <text>` exclusive with `--prompt-file`
- `--prompt-file <path>` exclusive with `--prompt`
- `--response <text>` exclusive with `--response-file`
- `--response-file <path>` exclusive with `--response`
- `--title <title>` optional
- `--source <source>` optional; default `cli`
- `--prev-hash <hash>` optional
- `--meta key=value` optional, repeatable
  - duplicate keys: last value wins
  - malformed value without `=`: returns error

List:
- `--author <name>` optional
- `--source <source>` optional
- `--meta k=v` optional, repeatable, exact match, AND semantics
- `--limit <n>` optional, default `20`

Verify:
- `yanzi verify <intent-id>`

Chain:
- `yanzi chain <intent-id>`

Show:
- `yanzi show <intent-id>`

Mode:
- `yanzi mode`
- `yanzi mode local`
- `yanzi mode http`

Project:
- `yanzi project create <name>`
- `yanzi project use <name>`
- `yanzi project current`
- `yanzi project list`

Checkpoint:
- `yanzi checkpoint create --summary "..."` for active project
- `yanzi checkpoint list` for active project

Rehydrate:
- `yanzi rehydrate`

Export:
- `yanzi export --format markdown`

Version:
- `yanzi version`
- `yanzi --version`

## 7. Canonical Agent Examples
Capture with files and metadata:
```bash
yanzi capture \
  --author "Ada" \
  --prompt-file prompt.txt \
  --response-file response.txt \
  --meta area=auth \
  --meta decision_type=refactor \
  --meta tags=migration,security
```

Capture with inline content:
```bash
yanzi capture --author "Ada" --prompt "Hello" --response "World"
```

List captures with metadata filters:
```bash
yanzi list --meta area=auth --meta decision_type=refactor --limit 10
```

Export project log:
```bash
yanzi export --format markdown
```

## 8. Output and Behavior Notes
- Export ordering is chronological.
- Metadata is optional and should not change capture ordering.
- Runtime `mode http` does not start `libraryd`.
- Keep outputs deterministic and machine-auditable.
