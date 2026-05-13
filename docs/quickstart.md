# Quickstart

## 1. Install

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

## 2. Create a Project

```bash
yanzi project create demo
yanzi project use demo
```

## 3. Capture Work

```bash
yanzi capture \
  --author "Ada" \
  --prompt "Summarize the current task" \
  --response "Set up the project and record the first state."

echo "Need to validate auth edge cases" \
  | yanzi capture --author "Ada" --response "Clock skew appears likely." --meta area=auth
```

## 4. Create a Checkpoint

```bash
yanzi checkpoint create --summary "Initial project state"
```

## 5. Rehydrate

```bash
yanzi rehydrate --dry-run
yanzi rehydrate
```

Notes:

- `yanzi list` scopes to the active project by default and prints tab-separated columns
- `yanzi rehydrate --format json` is the machine-readable continuity form

## Optional: Message Channel

```bash
yanzi message send --to claude --from ada --channel handoff --content "Start from the latest checkpoint."
yanzi message pull --to claude --channel handoff
```
