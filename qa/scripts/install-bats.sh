#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
VENDOR_DIR="$ROOT_DIR/qa/vendor"

mkdir -p "$VENDOR_DIR"

ensure_submodule() {
  local repo_url="$1"
  local target_dir="$2"

  if [ -d "$target_dir/.git" ] || [ -f "$target_dir/.git" ]; then
    git -C "$ROOT_DIR" submodule update --init --recursive -- "$target_dir"
  else
    git -C "$ROOT_DIR" submodule add "$repo_url" "$target_dir"
    git -C "$ROOT_DIR" submodule update --init --recursive -- "$target_dir"
  fi
}

ensure_submodule "https://github.com/bats-core/bats-core.git" "qa/vendor/bats-core"
ensure_submodule "https://github.com/bats-core/bats-support.git" "qa/vendor/bats-support"
ensure_submodule "https://github.com/bats-core/bats-assert.git" "qa/vendor/bats-assert"

"$ROOT_DIR/qa/vendor/bats-core/install.sh" "$ROOT_DIR/qa/vendor"

echo "Bats installed locally at: $ROOT_DIR/qa/vendor/bin/bats"
