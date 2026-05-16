#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
EXPECTED_DIR="$ROOT_DIR/qa/snapshots/project-lifecycle/expected"
NORMALIZED_DIR="$ROOT_DIR/qa/snapshots/project-lifecycle/normalized"
REPORTS_DIR="$ROOT_DIR/qa/reports"
DIFF_FILE="$REPORTS_DIR/project-lifecycle-diff.txt"
CLASS_FILE="$REPORTS_DIR/project-lifecycle-drift-classification.txt"

mkdir -p "$REPORTS_DIR"
: > "$DIFF_FILE"

status="PASS"

if ! find "$EXPECTED_DIR" -type f ! -name 'README.md' | grep -q .; then
  status="WARN"
  {
    echo "No expected snapshots found."
    echo "Seed expected snapshots after review before PASS certification."
  } > "$DIFF_FILE"
else
  while IFS= read -r expected_path; do
    name="$(basename "$expected_path")"
    actual_path="$NORMALIZED_DIR/$name"

    if [ ! -f "$actual_path" ]; then
      status="FAIL"
      echo "missing normalized output: $name" >> "$DIFF_FILE"
      continue
    fi

    if ! diff -u "$expected_path" "$actual_path" >> "$DIFF_FILE"; then
      if [ "$status" != "FAIL" ]; then
        status="WARN"
      fi
    fi
  done < <(find "$EXPECTED_DIR" -type f ! -name 'README.md' | sort)

  while IFS= read -r actual_path; do
    name="$(basename "$actual_path")"
    expected_path="$EXPECTED_DIR/$name"
    if [ ! -f "$expected_path" ]; then
      if [ "$status" = "PASS" ]; then
        status="WARN"
      fi
      echo "unexpected normalized output not in expected/: $name" >> "$DIFF_FILE"
    fi
  done < <(find "$NORMALIZED_DIR" -type f ! -name 'README.md' | sort)
fi

case "$status" in
  PASS) drift="Expected Drift" ;;
  WARN) drift="Formatting Drift" ;;
  FAIL) drift="Regression Drift" ;;
  *) drift="Operational Drift" ; status="FAIL" ;;
esac

{
  echo "status=$status"
  echo "drift='$drift'"
} > "$CLASS_FILE"

if [ ! -s "$DIFF_FILE" ]; then
  echo "No drift detected." > "$DIFF_FILE"
fi

echo "snapshot comparison complete"
echo "status: $status"
echo "drift classification: $drift"
echo "diff report: $DIFF_FILE"
