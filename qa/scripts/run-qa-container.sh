#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE="ubuntu:24.04"

exec docker run --rm -it \
  -v "$ROOT_DIR":/workspace \
  -w /workspace \
  "$IMAGE" \
  bash -lc 'set -euo pipefail; apt-get update >/dev/null && apt-get install -y --no-install-recommends bash ca-certificates curl git >/dev/null && echo "Container ready in /workspace" && exec bash'
