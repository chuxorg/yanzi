# Expected Behavior

## Expected Reconstruction Behavior
- All captures included before checkpoint remain discoverable after rehydration.
- Rehydration restores expected active project context deterministically.
- Checkpoint metadata remains queryable and consistent.

## Deterministic Recovery Requirements
- Rehydrate command exits `0` under valid checkpoint state.
- Recovery output provides explicit success indicators.
- Repeated rehydrate on unchanged state produces consistent results.

## Validation Expectations
- Artifact counts before and after rehydration are equivalent for in-scope data.
- No unexpected mutation or record loss is observed.
- Any drift is classified as FAIL unless explicitly documented as accepted WARN.
