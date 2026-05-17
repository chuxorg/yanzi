# Homebrew Drift Analysis

## Drift Summary

Prior certification observed Homebrew drift: candidate under test was `v2.9.1-rc1`, while tap formula installed `v2.7.0`.

## Root Cause

- Tap formula version and checksums were stale relative to certified candidate.
- Release promotion lacked an explicit release-to-tap synchronization gate.

## Impact

- Homebrew channel produced materially different runtime version.
- QA distribution validation and operator installs diverged.
- Release promotion risk increased due to channel inconsistency.

## Provenance Implications

- Certification evidence showed channel mismatch but governance did not block promotion early enough.
- Tap lineage was not governed as part of release readiness.

## Deterministic Risks

- Channel drift can persist across release cycles.
- Operators receive non-certified artifacts from commonly documented install path.

## Hardening Applied

- Release governance now requires explicit tap synchronization validation.
- Certification gating fails on channel ambiguity or lineage mismatch.
- RC policy requires explicit declaration of Homebrew behavior (supported or intentionally blocked for RC).
