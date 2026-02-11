#!/usr/bin/env sh
set -eu

BINARIES="yanzi yanzi-emitter"
LOCATIONS="/usr/local/bin ${HOME}/.local/bin"

removed_any=false

remove_file() {
  path="$1"
  if [ -e "$path" ]; then
    if rm -f "$path" 2>/dev/null; then
      removed_any=true
      return 0
    fi
    if command -v sudo >/dev/null 2>&1; then
      if sudo -n rm -f "$path" 2>/dev/null; then
        removed_any=true
        return 0
      fi
    fi
    echo "Failed to remove $path (permission denied)." >&2
    exit 1
  fi
}

for dir in $LOCATIONS; do
  for bin in $BINARIES; do
    remove_file "$dir/$bin"
  done
done

if [ "$removed_any" = true ]; then
  echo "Yanzi has been uninstalled."
else
  echo "Yanzi was not found on this system."
fi
