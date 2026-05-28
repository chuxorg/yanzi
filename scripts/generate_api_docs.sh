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

# docs/api/index.md is hand-maintained (REST API reference for /v0 endpoints).
# Do NOT regenerate it here. Run `make docs-check` to verify gomarkdoc output
# is up-to-date without overwriting hand-maintained files.
