#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ACTUAL_DIR="$ROOT_DIR/qa/snapshots/project-lifecycle/actual"
NORMALIZED_DIR="$ROOT_DIR/qa/snapshots/project-lifecycle/normalized"
WORKSPACE="$ROOT_DIR/qa/tmp/project-lifecycle-workspace"

mkdir -p "$NORMALIZED_DIR"
find "$NORMALIZED_DIR" -type f ! -name 'README.md' -delete

normalize_file() {
  local input="$1"
  local output="$2"

  perl -pe '
    s/[0-9]{4}-[0-9]{2}-[0-9]{2}[T ][0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]+)?(?:Z|[+-][0-9]{2}:[0-9]{2})?/<TIMESTAMP>/g;
    s/\b[0-9A-HJKMNP-TV-Z]{26}\b/<ULID>/g;
    s/\b[a-f0-9]{64}\b/<ID64>/g;
    s/\b[a-f0-9]{32}\b/<ID32>/g;
    s#/tmp/[A-Za-z0-9_.-]+#<TMP_PATH>#g;
  ' "$input" \
    | sed -E \
      -e "s#${ROOT_DIR//\//\\/}#<REPO_PATH>#g" \
      -e "s#${HOME//\//\\/}#<HOME_PATH>#g" \
      -e "s#${WORKSPACE//\//\\/}#<WORKSPACE_PATH>#g" \
      -e 's#[[:alnum:]_./-]*/qa/tmp/project-lifecycle-workspace[^ ]*#<WORKSPACE_PATH>#g' \
    > "$output"
}

shopt -s nullglob
files=("$ACTUAL_DIR"/*)
if [ ${#files[@]} -eq 0 ]; then
  echo "error: no files found in $ACTUAL_DIR" >&2
  exit 1
fi

for path in "${files[@]}"; do
  [ -f "$path" ] || continue
  name="$(basename "$path")"
  [ "$name" = "README.md" ] && continue
  normalize_file "$path" "$NORMALIZED_DIR/$name"
done

echo "normalization complete"
echo "normalized outputs: $NORMALIZED_DIR"
