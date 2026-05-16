# Validation Checklist

- [ ] Installation command and source are documented.
- [ ] `yanzi --version` exits `0` and includes deterministic identity markers.
- [ ] `command -v yanzi` resolves before uninstall.
- [ ] `yanzi --help` exits `0` with usage output.
- [ ] Uninstall/removal command exits `0`.
- [ ] `command -v yanzi` fails or returns empty after removal.
- [ ] PASS/WARN/FAIL outcome is recorded with evidence.
