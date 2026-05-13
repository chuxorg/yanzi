# Yanzi

[![QA Build](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml)
[![Release](https://github.com/chuxorg/yanzi/actions/workflows/release.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/release.yml)

Yanzi is a CLI for recording intent, storing context, creating checkpoints, and rehydrating AI-assisted work.

## Quick Start

```bash
yanzi project create my-project
yanzi project use my-project
yanzi capture --author "Ada" --prompt "hello" --response "world"
yanzi export --format html
open YANZI_LOG.html
```

This opens an interactive HTML UI.
No separate UI install is required after Yanzi is installed.
The exported file works offline.

## What It Does

Yanzi stores project history locally in SQLite. The main workflow is:

- capture prompt and response pairs
- add context documents or notes
- create checkpoints at stable boundaries
- rehydrate the active project from the latest checkpoint

It does not orchestrate agents or modify your code for you.

## UI Overview

`yanzi export --format html` writes an interactive HTML file.
Artifacts are collapsible, search is built in, and checkpoints are separated in the timeline.

## Install

Requires Go >= 1.25 for source builds.

Canonical GitHub repository: `chuxorg/yanzi`

macOS:

```bash
brew install chuxorg/yanzi/yanzi
```

macOS or Linux installer:

```bash
curl -fL -o /tmp/yanzi-install.sh https://raw.githubusercontent.com/chuxorg/yanzi/main/install.sh
test -s /tmp/yanzi-install.sh
sh /tmp/yanzi-install.sh
```

If you already cloned `chuxorg/yanzi`, you can also run:

```bash
./scripts/install.sh
```

That wrapper installs the current checkout and requires Go >= 1.25.

Windows:

1. Download `yanzi-windows-amd64.zip` from the latest release.
2. Extract `yanzi.exe`.
3. Add the extract directory to `PATH`.

Technical docs: [https://chuxorg.github.io/yanzi/](https://chuxorg.github.io/yanzi/)

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

echo "Need to validate auth edge cases" \
  | yanzi capture --author "Ada" --response "Clock skew appears likely." --meta area=auth
```

Create a checkpoint and preview rehydration:

```bash
yanzi checkpoint create --summary "Initial project state"
yanzi rehydrate --dry-run
yanzi rehydrate --format json
```

Optional message example:

```bash
yanzi message send --to claude --from ada --channel handoff --content "Review the latest checkpoint."
yanzi message pull --to claude --channel handoff
```

## Docs

- Website and overview: [https://chuxorg.github.io/yanzi/](https://chuxorg.github.io/yanzi/)
- Demo flow: [https://chuxorg.github.io/yanzi/demo-flow/](https://chuxorg.github.io/yanzi/demo-flow/)
- Quickstart: [https://chuxorg.github.io/yanzi/quickstart/](https://chuxorg.github.io/yanzi/quickstart/)
- CLI reference: [https://chuxorg.github.io/yanzi/cli/](https://chuxorg.github.io/yanzi/cli/)
- Install: [https://chuxorg.github.io/yanzi/install/](https://chuxorg.github.io/yanzi/install/)

Homebrew upgrades depend on the tap formula being refreshed. If `brew upgrade yanzi` does not move you to the latest release immediately, use the install script above.

## License

MIT
