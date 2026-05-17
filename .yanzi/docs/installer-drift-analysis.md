# Installer Drift Analysis

## Drift Summary

Prior certification observed installer drift: candidate under test was `v2.9.1-rc1`, while install script resolved mutable `releases/latest` and installed `v2.9.0`.

## Root Cause

- Installer resolved assets from mutable latest endpoint.
- Certification path depended on implicit release selection.
- No hard mismatch gate between requested candidate and installed runtime version.

## Impact

- Certified candidate and installed runtime diverged.
- Distribution trust boundary was broken.
- Certification evidence could not prove release lineage deterministically.

## Provenance Implications

- Evidence captured installer success but not candidate identity integrity.
- Runtime version check lacked candidate-tag gating.

## Deterministic Risks

- Mutable latest can change between validation runs.
- RC certification can accidentally validate a different stable artifact.

## Hardening Applied

- Installer now supports explicit `--version <tag>` pinning.
- Installer emits provenance: requested tag, resolved tag, asset URL, installed runtime version.
- Installer fails on requested/resolved/installed version mismatch.
