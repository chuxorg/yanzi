#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REPORTS_DIR="$ROOT_DIR/qa/reports"
mkdir -p "$REPORTS_DIR"

CANDIDATE_TAG="${1:-${CANDIDATE_TAG:-}}"
CANDIDATE_SHA="${2:-${CANDIDATE_SHA:-unknown}}"

if [ -z "$CANDIDATE_TAG" ]; then
  echo "usage: qa/execution/validate-release-convergence.sh <candidate-tag> [candidate-sha]" >&2
  exit 1
fi

STATUS_FILE="$REPORTS_DIR/release-convergence-status.env"
FINDINGS_FILE="$REPORTS_DIR/release-convergence-findings.txt"
: > "$FINDINGS_FILE"

status="PASS"
promotable="yes"
installer_version_line=""
homebrew_lineage="unknown"

INSTALL_BIN_DIR="$ROOT_DIR/qa/tmp/convergence-bin"
mkdir -p "$INSTALL_BIN_DIR"
ORIG_PATH="$PATH"
export PATH="$INSTALL_BIN_DIR:$PATH"

log() { echo "$*" | tee -a "$FINDINGS_FILE"; }

log "candidate_tag=$CANDIDATE_TAG"
log "candidate_sha=$CANDIDATE_SHA"

if ! "$ROOT_DIR/install.sh" --version "$CANDIDATE_TAG" > "$REPORTS_DIR/release-convergence-installer.log" 2>&1; then
  status="FAIL"
  promotable="no"
  log "FAIL installer: pinned install failed for $CANDIDATE_TAG"
else
  installer_version_line="$(yanzi --version 2>/dev/null | head -n 1 || true)"
  log "installer_version=$installer_version_line"
  if [[ "$installer_version_line" != *"$CANDIDATE_TAG"* ]]; then
    status="FAIL"
    promotable="no"
    log "FAIL installer lineage mismatch: expected $CANDIDATE_TAG got $installer_version_line"
  fi
fi

if "$ROOT_DIR/uninstall.sh" > "$REPORTS_DIR/release-convergence-uninstall.log" 2>&1; then
  if command -v yanzi >/dev/null 2>&1; then
    log "WARN uninstall left yanzi in PATH (may be channel overlap)"
    if [ "$status" = "PASS" ]; then
      status="WARN"
      promotable="no"
    fi
  else
    log "uninstall_check=pass"
  fi
else
  status="FAIL"
  promotable="no"
  log "FAIL uninstall failed"
fi

if "$ROOT_DIR/install.sh" --version "$CANDIDATE_TAG" > "$REPORTS_DIR/release-convergence-reinstall.log" 2>&1; then
  reinstall_version_line="$(yanzi --version 2>/dev/null | head -n 1 || true)"
  log "reinstall_version=$reinstall_version_line"
  if [[ "$reinstall_version_line" != *"$CANDIDATE_TAG"* ]]; then
    status="FAIL"
    promotable="no"
    log "FAIL reinstall lineage mismatch: expected $CANDIDATE_TAG got $reinstall_version_line"
  fi
else
  status="FAIL"
  promotable="no"
  log "FAIL reinstall failed"
fi

if command -v brew >/dev/null 2>&1; then
  brew search yanzi > "$REPORTS_DIR/release-convergence-homebrew-search.log" 2>&1 || true
  if brew search yanzi 2>/dev/null | grep -q '^chuxorg/yanzi/yanzi$'; then
    homebrew_lineage="tap:chuxorg/yanzi/yanzi"
    if brew info --json=v2 chuxorg/yanzi/yanzi > "$REPORTS_DIR/release-convergence-homebrew-info.json" 2>/dev/null; then
      formula_version="$(tr -d '\n' < "$REPORTS_DIR/release-convergence-homebrew-info.json" | sed -n 's/.*"versions"[[:space:]]*:[[:space:]]*{[^}]*"stable"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1 || true)"
      if [ -n "$formula_version" ]; then
        log "homebrew_formula_version=$formula_version"
        expected_no_v="${CANDIDATE_TAG#v}"
        if [ "$formula_version" != "$expected_no_v" ]; then
          log "FAIL homebrew lineage mismatch: expected $expected_no_v got $formula_version"
          status="FAIL"
          promotable="no"
        fi
      fi
    fi
  else
    homebrew_lineage="missing"
    log "FAIL homebrew formula missing: chuxorg/yanzi/yanzi"
    status="FAIL"
    promotable="no"
  fi
else
  homebrew_lineage="not-installed"
  log "WARN brew not installed; homebrew channel not validated"
  if [ "$status" = "PASS" ]; then
    status="WARN"
    promotable="no"
  fi
fi

export PATH="$ORIG_PATH"

cat > "$STATUS_FILE" <<EOF_STATUS
status=$status
promotable=$promotable
candidate_tag=$CANDIDATE_TAG
candidate_sha=$CANDIDATE_SHA
installer_version_line='${installer_version_line//\'/}'
homebrew_lineage=$homebrew_lineage
checked_at_utc=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
EOF_STATUS

log "status=$status"
log "promotable=$promotable"
log "status_file=$STATUS_FILE"

case "$status" in
  PASS) exit 0 ;;
  WARN) exit 0 ;;
  FAIL) exit 1 ;;
  *) exit 1 ;;
esac
