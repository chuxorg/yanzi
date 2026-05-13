#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

if ! command -v gomarkdoc >/dev/null 2>&1; then
  echo "gomarkdoc not found in PATH" >&2
  exit 1
fi

mkdir -p docs/api

gomarkdoc -o docs/API.md ./cmd/yanzi ./internal/...
gomarkdoc -o docs/api/cmd.md ./cmd/yanzi
gomarkdoc -o docs/api/internal.md ./internal/...

cat > docs/api/index.md <<'EOF'
# API

## Problem

Yanzi can be called from terminals, scripts, or other tools. The supported entrypoints need to be explicit so callers do not depend on undocumented behavior.

## Solution

The CLI is the primary API surface.

Every stable workflow in Yanzi is exposed as a command plus flags first.

Examples:

```bash
yanzi capture --author "Ada" --prompt "What changed?" --response "Updated docs."
yanzi export --format markdown --meta type=context
yanzi message pull --to codex --channel execution
```

## HTTP Mode

HTTP mode is optional.

When `mode=http` is configured, supported intent-oriented commands use the configured base URL instead of local storage.

Configuration example:

```yaml
mode: http
base_url: http://127.0.0.1:8080
```

CLI example:

```bash
yanzi mode http
yanzi capture --author "Ada" --prompt "Ping" --response "Stored remotely."
```

Some commands remain local-only, including checkpoint, rehydrate, export, delete, and restore.

## Generated References

Detailed generated references remain available here:

- [CLI Package](cmd.md)
- [Internal Packages](internal.md)
- [Combined API Reference](../API.md)
EOF
