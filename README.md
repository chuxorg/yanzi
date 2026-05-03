# Yanzi

[![QA Build](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml)
[![Release](https://github.com/chuxorg/yanzi/actions/workflows/release.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/release.yml)

**Docs:** [chuxorg.github.io/yanzi](https://chuxorg.github.io/yanzi) | [Quickstart](docs/quickstart.md) | [CLI Reference](docs/cli.md) | [AI Seed](docs/ai-seed.md)

AI-assisted development generates decisions and reasoning that are often lost across chat sessions, commits, and ad hoc notes. Git captures code changes, but not the full decision trail behind those changes. Yanzi provides deterministic logging for AI-assisted development so decisions can be recovered, audited, and shared.

## Quick Start

```sh
# Install
go install github.com/chuxorg/yanzi/cmd/yanzi@latest

# Create a project and record your first capture
yanzi project create "my-project"
yanzi project use "my-project"
yanzi capture --author "Ada" --prompt "What is the plan?" --response "Start with the data model."
yanzi checkpoint create --summary "Initial scaffolding complete"

# Export the log
yanzi export --format markdown
```

For AI agent setup, see [docs/ai-seed.md](docs/ai-seed.md).
For a step-by-step tutorial, see [docs/tutorial.md](docs/tutorial.md).

## Using Yanzi With an AI Agent

To use Yanzi with an AI coding agent (Codex, Copilot, Claude, etc.), provide the seed prompt at the start of the session:

- [docs/ai-seed.md](docs/ai-seed.md)

Recommended workflow:

1. Open `docs/ai-seed.md` and copy the seed prompt.
2. Paste it into your AI agent at the beginning of the session.
3. Let the agent verify or install Yanzi.
4. Let the agent read `README.md`, `docs/agent-bootstrap.md`, and `docs/tutorial.md`.
5. Confirm or create an active Yanzi project before capture, checkpoint, or export workflows.

The seed prompt tells the agent how to install Yanzi, check project state, and use the core workflows without inventing unsupported behavior.

## What Yanzi Is

Yanzi is a cross-platfor (Windows, Mac OS, and Linux) CLI written in golang and is designed to be used by AI Agents. 
Yanzi is not an MCP. Yanzi is a CLI and can be queried by a human from the command line. 
Yanzi is a deterministic logging layer for AI-assisted development.

It records:
- Captures (prompt/response records)
- Checkpoints (milestone summaries)
- Role/meta events (agent control intent)
- Optional metadata for capture context

Yanzi allows you to pause this capture as well.
Yanzi offers a variety of commands, known as @yanzi commands, that allow both the AI Agent and Human to interact with Yanzi more easily. 
Yanzi allows exports of the event stream, in structured formats, so AI-assisted work can be piped into other systems for later review or analysis.

Yanzi is not an orchestration framework or automation engine.

## Installation

Install Yanzi with Homebrew:

```sh
brew tap chuxorg/yanzi
brew install yanzi
yanzi version
```

Upgrade and uninstall with Homebrew:

```sh
brew upgrade yanzi
brew uninstall yanzi
```

Homebrew manages the installed binary only. It does not delete project data or local state under `~/.yanzi`.

Manual fallback install with the repo script:

```sh
./scripts/install.sh
```

Install with Go:

```sh
go install github.com/chuxorg/yanzi/cmd/yanzi@latest
```

Binary downloads are also available from GitHub Releases.

## Quick Start

```sh
yanzi project create "alpha"
yanzi project use "alpha"
yanzi checkpoint create --summary "Initial state"
```

During development, agents can record captures and checkpoints as work progresses, then export logs when needed.

## Agent Protocol

Yanzi is designed for AI-agent workflows that use explicit control lines, for example:

```text
@yanzi role Engineer
@yanzi checkpoint "Refactor authentication flow"
@yanzi pause
@yanzi resume
@yanzi export
```

Agents translate these protocol lines into Yanzi CLI commands.

## Exporting Logs

Supported export formats:

```sh
yanzi export --format markdown
yanzi export --format json
yanzi export --format html
```

- Markdown: human-readable log
- JSON: structured machine-readable log
- HTML: professional presentation of the log

For a full walkthrough including captures, projects, and the HTML export UI, see [docs/tutorial.md](docs/tutorial.md).

## Using Yanzi Logs in External Systems

Yanzi JSON export can be ingested by external systems such as ELK, Splunk, Sentry, or other logging pipelines.

Yanzi writes JSON to `./YANZI_LOG.json` by default. Example usage patterns:

```sh
yanzi export --format json > log.json
```

```sh
yanzi export --format json | curl -X POST https://example.internal/logs --data-binary @-
```
