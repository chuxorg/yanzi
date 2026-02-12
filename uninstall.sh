#!/usr/bin/env sh
set -eu

remove_path() {
  path="$1"
  if [ -z "$path" ]; then
    return 0
  fi
  if [ ! -e "$path" ]; then
    return 0
  fi
  if rm -f "$path" 2>/dev/null; then
    return 0
  fi
  if command -v sudo >/dev/null 2>&1; then
    if sudo -n rm -f "$path" 2>/dev/null; then
      return 0
    fi
  fi
  echo "Failed to remove $path (permission denied)." >&2
  exit 1
}

YANZI_PATH="$(command -v yanzi 2>/dev/null || true)"
EMITTER_PATH="$(command -v yanzi-emitter 2>/dev/null || true)"

if [ -z "$YANZI_PATH" ] && [ -z "$EMITTER_PATH" ]; then
  echo "Yanzi is not installed."
  exit 0
fi

remove_path "$YANZI_PATH"
remove_path "$EMITTER_PATH"

echo "Yanzi has been uninstalled."
