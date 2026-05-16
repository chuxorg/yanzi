# Expected Behavior

## Expected Outputs
- Installation completes with exit code `0`.
- `yanzi --version` returns exit code `0` and includes `yanzi` and a version identifier.
- `command -v yanzi` resolves to expected install location.
- `yanzi --help` returns exit code `0` and usage text.
- Uninstall/removal completes with exit code `0`.
- Post-removal `command -v yanzi` returns no result.

## PASS / WARN / FAIL Discipline
- PASS: All required commands succeed and cleanup is complete.
- WARN: Core commands pass but non-critical residue remains (for example: optional cache directory) with documented reason.
- FAIL: Any required command fails, CLI crashes, or binary remains unexpectedly callable after removal.

## Examples
- PASS example: `--version` and `--help` both return `0`; executable absent after uninstall.
- WARN example: executable removed, but empty optional metadata directory persists.
- FAIL example: uninstall returns `0` but `yanzi --version` still executes from stale path.
