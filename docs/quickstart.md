# Quickstart

## 1. Install

macOS:

```bash
brew install chuxorg/yanzi/yanzi
```

macOS or Linux install script:

```bash
curl -sSL https://raw.githubusercontent.com/chuxorg/yanzi/main/install.sh | bash
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

## Optional: Message Channel

```bash
yanzi message send --to claude --from ada --channel handoff --content "Start from the latest checkpoint."
yanzi message pull --to claude --channel handoff
```
