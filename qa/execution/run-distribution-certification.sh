#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REPORTS_DIR="$ROOT_DIR/qa/reports"
TAG="${1:-v2.9.1-rc1}"
SHA="${2:-bceb106b0fa97d6574fa1aa5d419f489f3e935c4}"
RUN_DIR="$REPORTS_DIR/${TAG}-convergence"
REPORT_FILE="$REPORTS_DIR/${TAG}-convergence-certification.md"

mkdir -p "$RUN_DIR"

echo "candidate_tag=$TAG" > "$RUN_DIR/candidate-metadata.env"
echo "candidate_sha=$SHA" >> "$RUN_DIR/candidate-metadata.env"
echo "repo_sha=$(git -C "$ROOT_DIR" rev-parse HEAD)" >> "$RUN_DIR/candidate-metadata.env"
echo "run_at_utc=$(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "$RUN_DIR/candidate-metadata.env"

status="PASS"
findings=()

# Installer deterministic validation
set +e
bash "$ROOT_DIR/install.sh" --version "$TAG" > "$RUN_DIR/installer.log" 2>&1
installer_rc=$?
set -e
if [ "$installer_rc" -ne 0 ]; then
  status="FAIL"
  findings+=("FAIL: installer pinned-tag install failed (see qa/reports/${TAG}-convergence/installer.log)")
fi

installed_line="$(yanzi --version 2>/dev/null || true)"
printf '%s\n' "$installed_line" > "$RUN_DIR/installer-installed-version.log"
installed_tag="$(printf '%s\n' "$installed_line" | awk '/^yanzi / {print $2; exit}')"
if [ -z "$installed_tag" ] || [ "$installed_tag" != "$TAG" ]; then
  status="FAIL"
  findings+=("FAIL: installed version mismatch for installer path (expected $TAG, got ${installed_tag:-<empty>})")
fi

# Homebrew deterministic validation (best-effort in local environment)
set +e
brew --version > "$RUN_DIR/homebrew-version.log" 2>&1
brew_rc=$?
set -e
if [ "$brew_rc" -eq 0 ]; then
  set +e
  brew search chuxorg/yanzi/yanzi > "$RUN_DIR/homebrew-search.log" 2>&1
  brew install chuxorg/yanzi/yanzi > "$RUN_DIR/homebrew-install.log" 2>&1
  hb_install_rc=$?
  hb_version_line="$(yanzi --version 2>/dev/null || true)"
  printf '%s\n' "$hb_version_line" > "$RUN_DIR/homebrew-installed-version.log"
  hb_tag="$(printf '%s\n' "$hb_version_line" | awk '/^yanzi / {print $2; exit}')"
  brew uninstall --force yanzi > "$RUN_DIR/homebrew-uninstall.log" 2>&1
  set -e

  if [ "$hb_install_rc" -ne 0 ]; then
    status="FAIL"
    findings+=("FAIL: Homebrew install failed (see qa/reports/${TAG}-convergence/homebrew-install.log)")
  fi

  if [ -z "$hb_tag" ] || [ "$hb_tag" != "$TAG" ]; then
    status="FAIL"
    findings+=("FAIL: Homebrew lineage mismatch (expected $TAG, got ${hb_tag:-<empty>})")
  fi
else
  if [ "$status" = "PASS" ]; then
    status="WARN"
  fi
  findings+=("WARN: Homebrew not available in environment; channel validation not executed")
fi

# Channel ambiguity gate
if grep -q "requested_tag=latest" "$RUN_DIR/installer.log" 2>/dev/null; then
  status="FAIL"
  findings+=("FAIL: ambiguous installer channel usage detected (latest resolution)")
fi

# Report
{
  echo "# Release Certification Report: $TAG (Distribution Convergence)"
  echo
  echo "- Timestamp (UTC): $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  echo "- Candidate Tag: $TAG"
  echo "- Candidate SHA: $SHA"
  echo "- Environment: $(uname -s) $(uname -m)"
  echo
  echo "## Scenarios Executed"
  echo
  echo "1. Installer pinned-tag distribution validation"
  echo "2. Homebrew channel validation"
  echo "3. Version lineage mismatch gating"
  echo "4. Channel ambiguity gating"
  echo
  echo "## Findings"
  echo
  if [ "${#findings[@]}" -eq 0 ]; then
    echo "- PASS: no blocking findings"
  else
    for f in "${findings[@]}"; do
      echo "- $f"
    done
  fi
  echo
  echo "## Final Result"
  echo
  echo "- Outcome: $status"
  if [ "$status" = "FAIL" ]; then
    echo "- Release recommendation: DO NOT PROMOTE"
  elif [ "$status" = "WARN" ]; then
    echo "- Release recommendation: PROMOTE WITH DOCUMENTED RISKS"
  else
    echo "- Release recommendation: PROMOTE"
  fi
  echo
  echo "## Evidence"
  echo
  echo "- qa/reports/${TAG}-convergence/candidate-metadata.env"
  echo "- qa/reports/${TAG}-convergence/installer.log"
  echo "- qa/reports/${TAG}-convergence/installer-installed-version.log"
  echo "- qa/reports/${TAG}-convergence/homebrew-search.log"
  echo "- qa/reports/${TAG}-convergence/homebrew-install.log"
  echo "- qa/reports/${TAG}-convergence/homebrew-installed-version.log"
  echo "- qa/reports/${TAG}-convergence/homebrew-uninstall.log"
} > "$REPORT_FILE"

echo "certification_report=$REPORT_FILE"
echo "certification_status=$status"
