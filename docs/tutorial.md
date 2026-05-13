# Yanzi Tutorial

This tutorial walks through the current local workflow: project setup, capture, checkpoint, export, message handoff, and rehydration.

## Install

Requires Go >= 1.25 for source builds.

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

Windows:

- download `yanzi-windows-amd64.zip`
- extract `yanzi.exe`
- add the extract directory to `PATH`

Verify:

```bash
yanzi --version
```

## Create a Project

```bash
yanzi project create alpha
yanzi project use alpha
yanzi project current
```

## Capture Work

Inline capture:

```bash
yanzi capture \
  --author "Ada" \
  --prompt "Summarize the auth refactor plan" \
  --response "Split login, session, and token validation into separate handlers."
```

Capture from files:

```bash
yanzi capture \
  --author "Ada" \
  --prompt-file prompt.txt \
  --response-file response.txt \
  --meta area=docs
```

Capture from stdin:

```bash
echo "Need to validate the reconnect edge cases." \
  | yanzi capture --author "Ada" --response "Focus on retry timing and clock skew." --meta area=auth
```

stdin currently supports the prompt side only. Responses remain explicit with `--response` or `--response-file`.

## Create a Checkpoint

```bash
yanzi checkpoint create --summary "Initial stable state"
yanzi checkpoint list
```

## Export

```bash
yanzi export --format markdown
yanzi export --format json
yanzi export --format html
```

## Message Handoff

```bash
yanzi message send --to claude --from operator --channel handoff --content "Continue from the latest checkpoint."
yanzi message list --to claude --channel handoff
yanzi message pull --to claude --channel handoff
```

## Rehydrate

```bash
yanzi rehydrate --dry-run
yanzi rehydrate
yanzi rehydrate --format json
```

## Next

- [Quickstart](quickstart.md)
- [CLI Reference](cli.md)
- [How It Works](how-it-works.md)
