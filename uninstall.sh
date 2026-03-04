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

remove_file() {
  name="$1"
  path="$2"
  if [ -n "$path" ] && [ -e "$path" ]; then
    echo "Removing $name..."
    remove_path "$path"
    return 0
  fi
  return 1
}

removed_any=false
for bin in yanzi yanzi-emitter libraryd; do
  path="$(command -v "$bin" 2>/dev/null || true)"
  if remove_file "$bin" "$path"; then
    removed_any=true
    continue
  fi

  for dir in /usr/local/bin "$HOME/.local/bin"; do
    candidate="$dir/$bin"
    if remove_file "$bin" "$candidate"; then
      removed_any=true
      break
    fi
  done
done

if [ "$removed_any" = false ]; then
  echo "Yanzi is not installed."
  exit 0
fi

echo "Uninstall complete."
