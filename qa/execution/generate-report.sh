#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REPORT_FILE="$ROOT_DIR/qa/reports/project-lifecycle-certification.md"
DIFF_FILE="$ROOT_DIR/qa/reports/project-lifecycle-diff.txt"
CLASS_FILE="$ROOT_DIR/qa/reports/project-lifecycle-drift-classification.txt"
NORMALIZED_DIR="$ROOT_DIR/qa/snapshots/project-lifecycle/normalized"
EXPECTED_DIR="$ROOT_DIR/qa/snapshots/project-lifecycle/expected"

status="WARN"
drift="Operational Drift"
if [ -f "$CLASS_FILE" ]; then
  # shellcheck disable=SC1090
  source "$CLASS_FILE"
fi

timestamp="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

cat >> "$REPORT_FILE" <<EOF_REPORT
## Run: $timestamp

- Scenario: project-lifecycle deterministic operational certification
- Environment: $(uname -s) $(uname -m)
- Repository: $ROOT_DIR
- Commands Executed:
  - qa/execution/run-project-lifecycle.sh
  - qa/execution/normalize-output.sh
  - qa/execution/compare-snapshots.sh
  - qa/execution/generate-report.sh
- Snapshots Certified:
  - Expected: $EXPECTED_DIR
  - Normalized: $NORMALIZED_DIR
- Normalization Applied:
  - timestamp tokens (<TIMESTAMP>)
  - generated identifier tokens (<ID32>, <ID64>, <ULID>)
  - machine/path tokens (<REPO_PATH>, <HOME_PATH>, <WORKSPACE_PATH>, <TMP_PATH>)
- Drift Findings Classification: $drift
- Result: $status
- Certification Notes: First human-reviewed deterministic baseline established for project-lifecycle scenario.

### Drift Findings

\`\`\`text
$(cat "$DIFF_FILE" 2>/dev/null || echo "No diff file found.")
\`\`\`

---

EOF_REPORT

echo "certification report updated: $REPORT_FILE"
echo "result: $status"
