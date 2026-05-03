# Quickstart

## Install

Install with Go:

```sh
go install github.com/chuxorg/yanzi/cmd/yanzi@latest
```

Binary downloads are available on the [GitHub Releases](https://github.com/chuxorg/yanzi/releases) page.

## First Capture

Create a project and set it as active:

```sh
yanzi project create "my-project"
yanzi project use "my-project"
```

Record a capture (prompt/response pair):

```sh
yanzi capture --author "Ada" --prompt "What is the plan?" --response "Start with the data model."
```

Create a checkpoint to mark a milestone:

```sh
yanzi checkpoint create --summary "Initial scaffolding complete"
```

List recent captures:

```sh
yanzi list --limit 10
```

## Export

Export the project log as markdown, JSON, or HTML:

```sh
yanzi export --format markdown
yanzi export --format json
yanzi export --format html
```

The HTML export opens as a standalone log viewer with search and filtering.

## SYSTEM_RULES

Add a `SYSTEM_RULES.md` file to your project root to define governance rules for AI agents:

```sh
# Create or edit SYSTEM_RULES.md
# Yanzi will include it automatically when agents read project context.
```

To include rules in a composed export context, export to JSON and pipe it to your toolchain:

```sh
yanzi export --format json > project-log.json
```

## Rehydrate

When context is lost between sessions, rehydrate:

```sh
yanzi rehydrate
```

This reconstructs the active project state from recorded captures and checkpoints.
