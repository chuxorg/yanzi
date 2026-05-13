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

Capture with stdin:

```bash
echo "Need to validate auth edge cases" \
  | yanzi capture --author "Ada" --response "Clock skew appears likely." --meta area=auth
```

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

Operational notes:

- `yanzi list` and `yanzi checkpoint list` print project headers and tab-separated columns
- `yanzi rehydrate --format json` is the structured continuity output for automation
- stdin is accepted as the prompt source for `yanzi capture`, but it conflicts with `--prompt` and `--prompt-file`
