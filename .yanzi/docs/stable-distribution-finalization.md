# Stable Distribution Finalization - v2.9.1

Date: 2026-05-17
Branch: feature/stable-distribution-finalization

## Scope
Post-release stable-channel normalization for production lineage convergence to `v2.9.1`.

## Audit Findings

1. Homebrew active tap lineage was still resolving `2.9.1-rc1` before retap/reinstall validation.
2. `homebrew-yanzi/Formula/yanzi.rb` already targeted `v2.9.1` URL lineage, but embedded SHA256 values were stale and failed checksum validation.
3. Installer default behavior in `install.sh` previously selected the first matching asset from paged releases; this could drift and was not deterministic stable promotion semantics.
4. `scripts/install.sh` path did not exist in `yanzi`; stable installer entrypoint existed only as top-level `install.sh`.
5. No `v2.9.1-rc1` references were found in tracked files in `yanzi` or `homebrew-yanzi` during audit.

## Normalization Actions

1. Installer default promoted to deterministic stable tag via `DEFAULT_STABLE_VERSION="v2.9.1"`.
2. Installer now supports explicit version override with `--version=<tag>` for RC opt-in while preserving stable default.
3. Added `scripts/install.sh` wrapper to preserve stable installer path expectations.
4. Homebrew formula SHA256 updated to production artifact checksums for `v2.9.1`.

## Result
Production stable distribution lineage now converges on `v2.9.1` for:

- Homebrew formula lineage
- installer default lineage
- direct release binary lineage

RC lineage remains available only via explicit version request (for example `--version=v2.9.1-rc1`).
