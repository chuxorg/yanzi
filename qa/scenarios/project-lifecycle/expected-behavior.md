# Expected Behavior

## Deterministic Expectations
- Project creation/binding returns exit code `0`.
- Capture command returns exit code `0` and stable artifact identity fields.
- Artifact listing includes captured record deterministically.
- Checkpoint creation returns exit code `0` and checkpoint identifier.
- Checkpoint list includes newly created checkpoint.
- Rehydrate returns exit code `0` and confirms active state restoration.

## Operational Validation Criteria
- CLI output includes expected command-specific markers (`project`, `checkpoint`, `rehydrate`).
- No panic, crash, or inconsistent ordering in repeated runs under the same inputs.
- All artifacts remain inspectable through CLI retrieval commands.

## Expected CLI Behavior
- Failure conditions are explicit and non-ambiguous.
- Success output is human-readable and suitable for report capture.
