#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
YANZI_BIN="$ROOT_DIR/yanzi"
WORKSPACE="$ROOT_DIR/qa/tmp/project-lifecycle-workspace"
ACTUAL_DIR="$ROOT_DIR/qa/snapshots/project-lifecycle/actual"

if [ ! -x "$YANZI_BIN" ]; then
  echo "error: yanzi binary not found or not executable at $YANZI_BIN" >&2
  exit 1
fi

rm -rf "$WORKSPACE"
mkdir -p "$WORKSPACE/home" "$ACTUAL_DIR"

# Clean only generated run outputs while keeping README placeholders.
find "$ACTUAL_DIR" -type f ! -name 'README.md' -delete

export HOME="$WORKSPACE/home"
export PATH="$ROOT_DIR:$PATH"

run_and_capture() {
  local name="$1"
  shift

  {
    echo "== COMMAND =="
    printf '%q ' "$@"
    echo
    echo "== OUTPUT =="
    "$@"
  } >"$ACTUAL_DIR/${name}.txt" 2>&1
}

cd "$WORKSPACE"

run_and_capture "01_project_create" "$YANZI_BIN" project create deterministic-project
run_and_capture "02_project_use" "$YANZI_BIN" project use deterministic-project
run_and_capture "03_capture" "$YANZI_BIN" capture \
  --author "qa-operator" \
  --title "Deterministic capture" \
  --prompt "Create deterministic QA baseline." \
  --response "Baseline created for operational certification." \
  --profile "deterministic" \
  --meta scenario=project-lifecycle \
  --meta stage=foundation
run_and_capture "04_checkpoint_create" "$YANZI_BIN" checkpoint create --summary "Deterministic checkpoint"
run_and_capture "05_export_markdown" "$YANZI_BIN" export --format markdown
run_and_capture "06_export_html" "$YANZI_BIN" export --format html
run_and_capture "07_export_json" "$YANZI_BIN" export --format json
run_and_capture "08_project_list" "$YANZI_BIN" project list
run_and_capture "09_checkpoint_list" "$YANZI_BIN" checkpoint list
run_and_capture "10_rehydrate_dry_run" "$YANZI_BIN" rehydrate --dry-run

for file in YANZI_LOG.md YANZI_LOG.html YANZI_LOG.json; do
  if [ -f "$WORKSPACE/$file" ]; then
    cp "$WORKSPACE/$file" "$ACTUAL_DIR/$file"
  fi
done

echo "project lifecycle scenario execution complete"
echo "workspace: $WORKSPACE"
echo "actual outputs: $ACTUAL_DIR"
