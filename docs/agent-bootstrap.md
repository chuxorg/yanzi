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

## Protocol Annotation Grammar

- protocol annotations start at the beginning of the line
- protocol annotations use the prefix `@yanzi`
- protocol annotations are single-line records
- protocol annotations are not hidden executable commands

Supported protocol annotations:

- `@yanzi pause`
- `@yanzi resume`
- `@yanzi checkpoint "Summary"`
- `@yanzi export`
- `@yanzi role <RoleName>`

## State Rules

- `@yanzi pause` and `@yanzi resume` are continuity markers only
- protocol annotations are allowed while paused
- if the workflow needs a real CLI action, run the explicit `yanzi ...` command separately
- major structural decisions should still use `yanzi checkpoint create --summary "..."`

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
