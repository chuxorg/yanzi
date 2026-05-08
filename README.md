# Yanzi

[![QA Build](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml)
[![Release](https://github.com/chuxorg/yanzi/actions/workflows/release.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/release.yml)

Yanzi is a CLI for recording intent, storing context, creating checkpoints, and rehydrating AI-assisted work.

## What It Does

Yanzi stores project history locally in SQLite. The main workflow is:

- capture prompt and response pairs
- add context documents or notes
- create checkpoints at stable boundaries
- rehydrate the active project from the latest checkpoint

It does not orchestrate agents or modify your code for you.

## Install

macOS:

```bash
brew install chuxorg/yanzi/yanzi
```

Linux:

```bash
sudo dpkg -i yanzi_*.deb
```

Windows:

1. Download `yanzi-windows-amd64.zip` from the latest release.
2. Extract `yanzi.exe`.
3. Add the extract directory to `PATH`.

Technical docs: https://chuxorg.github.io/yanzi/

## Quickstart

Create and select a project:

```bash
yanzi project create demo
yanzi project use demo
```

Capture a prompt and response:

```bash
yanzi capture \
  --author "Ada" \
  --prompt "Summarize the current task" \
  --response "Add distribution docs and validate examples."
```

Create a checkpoint and preview rehydration:

```bash
yanzi checkpoint create --summary "Initial project state"
yanzi rehydrate --dry-run
```

Optional message example:

```bash
yanzi message send --to claude --from ada --channel handoff --content "Review the latest checkpoint."
yanzi message pull --to claude --channel handoff
```

## Docs

- Website and overview: https://chuxorg.github.io/yanzi/
- Quickstart: https://chuxorg.github.io/yanzi/quickstart/
- CLI reference: https://chuxorg.github.io/yanzi/cli/
- Install: https://chuxorg.github.io/yanzi/install/

## License

MIT
