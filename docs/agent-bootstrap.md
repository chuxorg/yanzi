# Yanzi Agent Bootstrap

Purpose: define the current Yanzi command surface and operating rules for AI agents working in a repository that uses Yanzi.

Before use, provide the agent seed prompt from the repository:

- https://github.com/chuxorg/yanzi/blob/master/prompts/AI_AGENT_SEED.md

For the user-facing walkthrough, see:

- [Quickstart](./quickstart.md)
- [CLI Reference](./cli.md)

## Role Declaration

- agents should declare a role at session start
- if no role is declared, default to `Engineer`

## Meta-Command Grammar

- meta-commands start at the beginning of the line
- meta-commands use the prefix `@yanzi`
- meta-commands are single-line commands

Supported meta-commands:

- `@yanzi pause`
- `@yanzi resume`
- `@yanzi checkpoint "Summary"`
- `@yanzi export`
- `@yanzi role <RoleName>`

## State Rules

- pause affects capture only
- meta-commands are allowed while paused
- state-changing commands should acknowledge execution
- major structural decisions should be checkpointed

## Current Command Surface

Primary usage:

- `yanzi <command> [args]`

Commands:

- `capture`
- `verify`
- `chain`
- `list`
- `show`
- `delete`
- `restore`
- `mode`
- `project`
- `intent`
- `context`
- `bootstrap`
- `rules`
- `types`
- `message`
- `checkpoint`
- `rehydrate`
- `export`
- `version`

## Current Examples

Capture with files:

```bash
yanzi capture \
  --author "Ada" \
  --prompt-file prompt.txt \
  --response-file response.txt \
  --meta area=auth
```

Create a checkpoint:

```bash
yanzi checkpoint create --summary "refactor complete"
```

Pull handoff notes:

```bash
yanzi message pull --to codex --channel handoff
```

Export project state:

```bash
yanzi export --format markdown
```

## Install Check

Verify the CLI first:

```bash
yanzi --version
```

If it is missing, install from:

- Homebrew: `brew install chuxorg/yanzi/yanzi`
- Canonical repository: `chuxorg/yanzi`
- Source builds require Go >= 1.25
- Releases: https://github.com/chuxorg/yanzi/releases

## Project Setup

```bash
yanzi project create cli-development
yanzi project use cli-development
yanzi checkpoint create --summary "starting development session"
```

If context is lost:

```bash
yanzi rehydrate
```
