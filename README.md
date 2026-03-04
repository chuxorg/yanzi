# Yanzi

[![QA Build](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml)
[![Release](https://github.com/chuxorg/yanzi/actions/workflows/release.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/release.yml)

Links: [yanzi (install info)](https://github.com/chuxorg/yanzi) | [yanzi.io](https://yanzi.io) | [chucksailer.me](https://chucksailer.me)
Agent Setup: [Tell your AI Agent (Codex, Copilot, etc.)](prompts/AI_AGENT_SEED.md)

AI-assisted development generates decisions and reasoning that are often lost across chat sessions, commits, and ad hoc notes. Git captures code changes, but not the full decision trail behind those changes. Yanzi provides deterministic logging for AI-assisted development so decisions can be recovered, audited, and shared.

## What Yanzi Is

Yanzi is a deterministic logging layer for AI-assisted development.

It records:
- Captures (prompt/response records)
- Checkpoints (milestone summaries)
- Role/meta events (agent control intent)
- Optional metadata for capture context

Yanzi exports this event stream in structured formats so AI-assisted work can be reviewed later.

Yanzi is not an orchestration framework or automation engine.

## Installation

Install Yanzi with Go:

```sh
go install github.com/chuxorg/yanzi/cmd/yanzi@latest
```

Binary downloads are also available from GitHub Releases.

## Quick Start

```sh
yanzi project create "alpha"
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

## Using Yanzi Logs in External Systems

Yanzi JSON export can be ingested by external systems such as ELK, Splunk, Sentry, or other logging pipelines.

Yanzi writes JSON to `./YANZI_LOG.json` by default. Example usage patterns:

```sh
yanzi export --format json > log.json
```

```sh
yanzi export --format json | curl -X POST https://example.internal/logs --data-binary @-
```
