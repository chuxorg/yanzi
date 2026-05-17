# Promotion Blocking Conditions

## Blocking Policy
Promotion is blocked when any deterministic trust boundary is violated.

## Blocking Conditions
1. Lineage mismatch across channels.
2. Snapshot regression or deterministic scenario FAIL.
3. Distribution drift (installer/Homebrew/package mismatch).
4. Uncertified artifacts in release path.
5. Unresolved governance drift or missing review evidence.
6. Version inconsistency between tag, artifact, and runtime output.

## Why Promotion Is Blocked
- Prevents uncertified artifacts from being treated as trusted releases.
- Preserves provenance continuity and operator trust.
- Maintains deterministic release semantics.

## Required Remediation Path
1. Identify mismatch root cause and impacted channels.
2. Patch source of drift (artifact publish, installer resolver, formula metadata, or docs).
3. Re-run deterministic convergence validation.
4. Regenerate append-only certification evidence.
5. Obtain human governance approval before retrying promotion.
