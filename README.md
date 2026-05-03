# Yanzi

[![QA Build](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml)
[![Release](https://github.com/chuxorg/yanzi/actions/workflows/release.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/release.yml)

Links: [yanzi.io](https://yanzi.io) | [chucksailer.me](https://chucksailer.me)
Agent Setup: [Tell your AI Agent to use Yanzi (Codex, Copilot, etc.)](docs/agent-bootstrap.md)
Tutorial: [Learn Yanzi step-by-step](docs/tutorial.md)

AI-assisted development generates decisions and reasoning that are often lost across chat sessions, commits, and ad hoc notes. Git captures code changes, but not the full decision trail behind those changes. Yanzi provides deterministic logging for AI-assisted development so decisions can be recovered, audited, and shared.

To get started with your favorite AI agent, first provide the seed prompt from [AI_AGENT_SEED.md](/Users/developer/projects/chuxorg/chux-yanzi-cli/prompts/AI_AGENT_SEED.md), then continue with the [Agent Bootstrap](docs/agent-bootstrap.md) and the [Yanzi Tutorial](docs/tutorial.md).

## Using Yanzi With an AI Agent

If you want an AI coding agent to install, initialize, and use Yanzi correctly in this repository, start with the seed prompt:

- [AI_AGENT_SEED.md](/Users/developer/projects/chuxorg/chux-yanzi-cli/prompts/AI_AGENT_SEED.md)

Recommended workflow:

1. Open `prompts/AI_AGENT_SEED.md`.
2. Copy the full prompt contents.
3. Paste the prompt into your AI agent at the beginning of the session.
4. Let the agent verify or install Yanzi.
5. Let the agent read `README.md`, `docs/agent-bootstrap.md`, and `docs/tutorial.md`.
6. Confirm or create an active Yanzi project before capture, checkpoint, or export workflows.

What the seed prompt is for:

- It tells the agent how to verify installation and install Yanzi if needed.
- It tells the agent how to check project state before logging work.
- It introduces the main Yanzi workflows:
  - `project create|use|current`
  - `capture`
  - `checkpoint`
  - `rules add|list|export`
  - `export`
  - `rehydrate`
- It tells the agent not to invent unsupported Yanzi behavior.

If you are using an agent such as Codex or Copilot, the seed prompt should be the first Yanzi-specific instruction you provide in the session.

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
