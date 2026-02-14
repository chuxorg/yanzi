

# Yanzi

**Managed Intent for AI-Assisted Development**

AI accelerates generation.
Yanzi preserves the *why*.

---

## What Yanzi Is

Yanzi is a small, composable system that captures, verifies, and chains the reasoning behind AI-generated decisions.

It does not generate code.
It does not orchestrate agents.
It does not summarize or reinterpret reasoning.

It simply refuses to forget.

## 90-Second Quickstart

These commands assume embedded local mode (no server).

1. Install
```bash
curl -fsSL https://raw.githubusercontent.com/chuxorg/yanzi/master/install.sh | sh -s -- --add-path
```
2. Capture
```bash
yanzi capture --author "Ada" --prompt-file prompt.txt --response-file response.txt
```
3. Verify
```bash
yanzi verify <intent-id>
```
4. Chain
```bash
yanzi chain <intent-id>
```
5. Optional: enable HTTP mode
```bash
git clone https://github.com/chuxorg/chux-yanzi-library
cd chux-yanzi-library
go run ./cmd/libraryd -addr :8080 -db yanzi.db
```

---

## Use Cases

Yanzi is useful for:

* Individual developers using AI tools locally.
* Teams sharing AI-generated decisions.
* Resuming context after closing an AI session.
* Preserving architectural reasoning.

---

## End-to-End Example: Local Mode (Default)

1. Install:
```bash
curl -sSL https://raw.githubusercontent.com/chuxorg/yanzi/master/install.sh | sh
```
2. Verify version:
```bash
yanzi --version
```
3. Capture an interaction:
```bash
echo "Create a minimal HTML page that says Hello World" > prompt.txt
echo "<!DOCTYPE html><html><body>Hello World</body></html>" > response.txt

yanzi capture \
  --author "developer" \
  --source "manual" \
  --prompt-file prompt.txt \
  --response-file response.txt \
  --meta project=hello-world \
  --meta type=feature
```
4. Browse:
```bash
yanzi list
```
5. Inspect:
```bash
yanzi show <id>
```
6. Verify:
```bash
yanzi verify <id>
```

Local mode is the default and requires no server.

---

## End-to-End Example: Server Mode (Optional)

1. Install with server enabled:
```bash
curl -sSL https://raw.githubusercontent.com/chuxorg/yanzi/master/install.sh | sh -s -- --server
```
2. Start server:
```bash
libraryd
```
3. Confirm config:
```bash
cat ~/.yanzi/config.yaml
```
4. Capture via server:
```bash
yanzi capture ...
```

If the server is stopped, capture will fail (no silent fallback).

---

## AI Agent Integration Example

```bash
PROMPT_FILE=$(mktemp)
RESPONSE_FILE=$(mktemp)

echo "$HUMAN_PROMPT" > "$PROMPT_FILE"
echo "$AI_RESPONSE" > "$RESPONSE_FILE"

yanzi capture \
  --author "codex" \
  --source "codex" \
  --prompt-file "$PROMPT_FILE" \
  --response-file "$RESPONSE_FILE"

rm "$PROMPT_FILE" "$RESPONSE_FILE"
```

Files are ephemeral. Yanzi stores structured, immutable records. Each capture represents one atomic interaction.

---

## Why This Is Necessary

If you are using AI to generate production artifacts but not preserving the reasoning behind them, you are creating undocumented architectural drift.

## The Problem

AI tools move fast.
But the reasoning behind decisions often disappears.

Weeks later, someone asks:

> Why does this exist?

And no one can answer.

Yanzi captures intent at the moment decisions are made — and preserves it immutably.

---

## Install Yanzi

### For Humans

macOS or Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/chuxorg/yanzi/master/install.sh | sh
```

Optional (auto-add to PATH):

```bash
curl -fsSL https://raw.githubusercontent.com/chuxorg/yanzi/master/install.sh | sh -s -- --add-path
```

To uninstall:

```bash
curl -fsSL https://raw.githubusercontent.com/chuxorg/yanzi/master/uninstall.sh | sh
```

---

## Server Mode (Optional)

Default install uses embedded local mode.

Enable server mode:

```bash
./install.sh --server
```

Then start the server:

```bash
libraryd
```

Server mode is activated via:
`~/.yanzi/config.yaml`

---

### Tell Your AI to Install Yanzi

If your AI assistant has shell access (Cursor, Codex CLI, etc.), you can give it this prompt:

```
Install Yanzi on this machine.

If macOS or Linux:
Run:
curl -fsSL https://raw.githubusercontent.com/chuxorg/yanzi/master/install.sh | sh -s -- --add-path

If Windows:
Download the latest release from:
https://github.com/chuxorg/yanzi/releases

After installation:
Run:
yanzi --help

Then reply:
"Yanzi is installed and ready. I will save important decisions using it."
```

---

## Quick Example

Capture intent:

```bash
yanzi capture --author "Ada" --prompt-file prompt.txt --response-file response.txt
```

Verify integrity:

```bash
yanzi verify <intent-id>
```

Inspect evolution:

```bash
yanzi chain <intent-id>
```

---

## Architecture

Yanzi consists of:

* **Yanzi Library** — append-only intent ledger
* **Yanzi CLI** — human interface
* **Yanzi Emitter** — machine ingestion boundary

Each component is intentionally small and independent.

---

## Philosophy

Small. Composable. Verifiable.

No magic.
No orchestration theater.
No agent hype.

Just memory.

---

## Repositories

* [https://github.com/chuxorg/chux-yanzi-library](https://github.com/chuxorg/chux-yanzi-library)
* [https://github.com/chuxorg/chux-yanzi-cli](https://github.com/chuxorg/chux-yanzi-cli)
* [https://github.com/chuxorg/chux-yanzi-emitter](https://github.com/chuxorg/chux-yanzi-emitter)

---
