#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TAG="${1:-v2.9.1-rc1}"
SHA="${2:-bceb106b0fa97d6574fa1aa5d419f489f3e935c4}"
REPORT_DIR="$ROOT_DIR/qa/reports/${TAG}"
mkdir -p "$REPORT_DIR"
REPORT_FILE="$ROOT_DIR/qa/reports/${TAG}-convergence-validation.md"

status="PASS"
installer_status="PASS"
homebrew_status="PASS"
binary_status="PASS"
sha_status="PASS"

fail() {
  status="FAIL"
  printf -- "%s\n" "$1" >> "$REPORT_DIR/convergence-findings.log"
}

: > "$REPORT_DIR/convergence-findings.log"

release_json="$(gh release view "$TAG" -R chuxorg/yanzi --json tagName,targetCommitish,url,assets)"
printf '%s\n' "$release_json" > "$REPORT_DIR/release-metadata.json"

resolved_tag="$(printf '%s\n' "$release_json" | jq -r '.tagName')"
resolved_sha="$(printf '%s\n' "$release_json" | jq -r '.targetCommitish')"

if [ "$resolved_tag" != "$TAG" ]; then
  sha_status="FAIL"
  fail "release tag mismatch: expected=$TAG got=$resolved_tag"
fi
if [ "$resolved_sha" != "$SHA" ]; then
  sha_status="FAIL"
  fail "candidate SHA mismatch: expected=$SHA got=$resolved_sha"
fi

required_assets=("yanzi-darwin-amd64" "yanzi-darwin-arm64" "yanzi-linux-amd64" "yanzi-windows-amd64.zip" "sha256sums.txt")
for a in "${required_assets[@]}"; do
  if ! printf '%s\n' "$release_json" | jq -r '.assets[].name' | grep -Fx "$a" >/dev/null; then
    sha_status="FAIL"
    fail "missing release asset: $a"
  fi
done

os_raw="$(uname -s)"
arch_raw="$(uname -m)"
case "$os_raw" in
  Darwin) os="darwin" ;;
  Linux) os="linux" ;;
  *) os="" ;;
esac
case "$arch_raw" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) arch="" ;;
esac
if [ -n "$os" ] && [ -n "$arch" ]; then
  asset="yanzi-${os}-${arch}"
  url="https://github.com/chuxorg/yanzi/releases/download/${TAG}/${asset}"
  tmpbin="$(mktemp)"
  curl -fsSL "$url" -o "$tmpbin"
  chmod +x "$tmpbin"
  direct_ver="$($tmpbin --version 2>/dev/null || true)"
  printf '%s\n' "$direct_ver" > "$REPORT_DIR/direct-binary-version.log"
  direct_tag="$(printf '%s\n' "$direct_ver" | awk '/^yanzi / {print $2; exit}')"
  rm -f "$tmpbin"
  if [ "$direct_tag" != "$TAG" ]; then
    binary_status="FAIL"
    fail "direct binary lineage mismatch: expected=$TAG got=${direct_tag:-<empty>}"
  fi
fi

install_bin_dir="$REPORT_DIR/installer-bin"
rm -rf "$install_bin_dir"
mkdir -p "$install_bin_dir"
set +e
YANZI_INSTALL_DIR="$install_bin_dir" bash "$ROOT_DIR/install.sh" --version "$TAG" > "$REPORT_DIR/installer-convergence.log" 2>&1
ins_rc=$?
set -e
if [ "$ins_rc" -ne 0 ]; then
  installer_status="FAIL"
  fail "installer execution failed for pinned tag"
else
  iv="$($install_bin_dir/yanzi --version 2>/dev/null || true)"
  printf '%s\n' "$iv" > "$REPORT_DIR/installer-convergence-version.log"
  itag="$(printf '%s\n' "$iv" | awk '/^yanzi / {print $2; exit}')"
  if [ "$itag" != "$TAG" ]; then
    installer_status="FAIL"
    fail "installer lineage mismatch: expected=$TAG got=${itag:-<empty>}"
  fi
fi

# Homebrew via local tap repo branch to validate channel propagation deterministically pre-merge.
tap_repo="${HOMEBREW_YANZI_TAP_REPO:-/Users/developer/projects/chuxorg/homebrew-yanzi}"
set +e
HOMEBREW_NO_AUTO_UPDATE=1 brew update > "$REPORT_DIR/homebrew-update.log" 2>&1
HOMEBREW_NO_AUTO_UPDATE=1 brew tap --custom-remote chuxorg/yanzi "$tap_repo" > "$REPORT_DIR/homebrew-tap.log" 2>&1
brew_tap_repo="$(brew --repository chuxorg/yanzi)"
cp "$tap_repo/Formula/yanzi.rb" "$brew_tap_repo/Formula/yanzi.rb"
printf '%s
' "tap_repo=$tap_repo" > "$REPORT_DIR/homebrew-propagation.log"
printf '%s
' "tap_checkout=$brew_tap_repo" >> "$REPORT_DIR/homebrew-propagation.log"
printf '%s
' "formula_source_commit=$(git -C "$tap_repo" rev-parse HEAD)" >> "$REPORT_DIR/homebrew-propagation.log"
printf '%s
' "formula_applied_locally=1" >> "$REPORT_DIR/homebrew-propagation.log"
HOMEBREW_NO_AUTO_UPDATE=1 brew uninstall --force yanzi > "$REPORT_DIR/homebrew-uninstall-pre.log" 2>&1
HOMEBREW_NO_AUTO_UPDATE=1 HOMEBREW_NO_INSTALL_FROM_API=1 brew install chuxorg/yanzi/yanzi > "$REPORT_DIR/homebrew-install-convergence.log" 2>&1
hb_rc=$?
hb_ver="$(yanzi --version 2>/dev/null || true)"
printf '%s\n' "$hb_ver" > "$REPORT_DIR/homebrew-convergence-version.log"
# reinstall path
HOMEBREW_NO_AUTO_UPDATE=1 brew uninstall --force yanzi > "$REPORT_DIR/homebrew-uninstall-mid.log" 2>&1
HOMEBREW_NO_AUTO_UPDATE=1 HOMEBREW_NO_INSTALL_FROM_API=1 brew install chuxorg/yanzi/yanzi > "$REPORT_DIR/homebrew-reinstall-convergence.log" 2>&1
hb_re_rc=$?
hb_re_ver="$(yanzi --version 2>/dev/null || true)"
printf '%s\n' "$hb_re_ver" > "$REPORT_DIR/homebrew-reinstall-convergence-version.log"
HOMEBREW_NO_AUTO_UPDATE=1 brew uninstall --force yanzi > "$REPORT_DIR/homebrew-uninstall-post.log" 2>&1
set -e
if [ "$hb_rc" -ne 0 ] || [ "$hb_re_rc" -ne 0 ]; then
  homebrew_status="FAIL"
  fail "homebrew install/reinstall failed using tap repo $tap_repo"
else
  htag="$(printf '%s\n' "$hb_ver" | awk '/^yanzi / {print $2; exit}')"
  htag_re="$(printf '%s\n' "$hb_re_ver" | awk '/^yanzi / {print $2; exit}')"
  if [ "$htag" != "$TAG" ] || [ "$htag_re" != "$TAG" ]; then
    homebrew_status="FAIL"
    fail "homebrew lineage mismatch: expected=$TAG got install=${htag:-<empty>} reinstall=${htag_re:-<empty>}"
  fi
fi

if [ "$installer_status" = "FAIL" ] || [ "$homebrew_status" = "FAIL" ] || [ "$binary_status" = "FAIL" ] || [ "$sha_status" = "FAIL" ]; then
  status="FAIL"
fi

{
  echo "# Release Convergence Validation: $TAG"
  echo
  echo "- Timestamp (UTC): $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  echo "- Candidate Tag: $TAG"
  echo "- Candidate SHA: $SHA"
  echo
  echo "## Channel Results"
  echo
  echo "- Installer lineage: $installer_status"
  echo "- Homebrew lineage: $homebrew_status"
  echo "- Direct binary lineage: $binary_status"
  echo "- Candidate SHA alignment: $sha_status"
  echo
  echo "## Convergence Status"
  echo
  echo "- Convergence status: $status"
  if [ "$status" = "PASS" ]; then
    echo "- Promotable: yes"
  else
    echo "- Promotable: no"
  fi
  echo
  echo "## Findings"
  echo
  if [ -s "$REPORT_DIR/convergence-findings.log" ]; then
    sed 's/^/- /' "$REPORT_DIR/convergence-findings.log"
  else
    echo "- No blocking findings."
  fi
} > "$REPORT_FILE"

echo "convergence_report=$REPORT_FILE"
echo "convergence_status=$status"
