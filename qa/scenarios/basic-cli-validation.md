# Basic CLI Validation Scenario

## Objective
Validate baseline operational health for the Yanzi CLI through deterministic command execution of global help and version workflows.

## Environment Expectations
- Repository is available locally.
- `yanzi` binary exists at repository root (`./yanzi`).
- Local Bats runtime is installed via `qa/scripts/install-bats.sh`.
- Shell environment is POSIX-compatible (`bash`).

## Commands Executed
1. `./yanzi --version`
2. `./yanzi --help`
3. `qa/vendor/bin/bats qa/suites/smoke`

## Expected Results
- Both CLI commands return exit code `0`.
- `--version` output includes `yanzi` and version-identifying text.
- `--help` output includes `Usage` and `yanzi` references.
- No panic/crash signatures appear in output.

## Validation Criteria
- Smoke tests pass with no failures.
- Output checks are deterministic in content presence (not timing dependent).
- Test execution remains filesystem-first and inspectable.
