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

## Ensure Yanzi Is Installed

This repository uses **Yanzi** to track development intent and checkpoints during AI-assisted work.
Before proceeding with development tasks, ensure that the `yanzi` CLI is installed and available in the system PATH.

### 1. Check if Yanzi is installed

Run:

```bash
yanzi version
```

If the command returns a version number, Yanzi is already installed and ready to use and disregard the rest of this prompt. 

If there was no version returned do the following:

### 2. Install Yanzi (if not installed)

If the command is not found, install Yanzi from the official releases page:

https://github.com/chuxorg/chux-yanzi-cli/releases

Download the archive that matches your operating system.

Example installation (macOS / Linux):

```bash
curl -L https://github.com/chuxorg/chux-yanzi-cli/releases/latest/download/yanzi_darwin_arm64.tar.gz -o yanzi.tar.gz
tar -xzf yanzi.tar.gz
sudo mv yanzi /usr/local/bin/
```

You may need to select a different archive depending on your platform.

---

### 3. Verify installation

Confirm the CLI is available:

```bash
yanzi version
```

The command should return the installed version.

---

### 4. Initialize a Yanzi development project

Once installed, initialize a project for tracking development work in this repository:

```bash
yanzi project create "cli-development"
yanzi project use "cli-development"
```

---

### 5. Record development checkpoints

Before significant changes, record a checkpoint describing the work being done:

```bash
yanzi checkpoint create --summary "starting development session"
```

Checkpoints make it possible to reconstruct development state later.

If context is lost during development, use:

```bash
yanzi rehydrate
```

This reconstructs the current project state from the most recent checkpoint and recorded artifacts.




