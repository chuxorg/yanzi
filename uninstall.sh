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

if [ -z "$YANZI_PATH" ]; then
  echo "Yanzi is not installed."
  exit 0
fi

remove_file() {
  name="$1"
  path="$2"
  if [ -n "$path" ] && [ -e "$path" ]; then
    echo "Removing $name..."
    remove_path "$path"
  fi
}

remove_file "yanzi" "$YANZI_PATH"

echo "Uninstall complete."
