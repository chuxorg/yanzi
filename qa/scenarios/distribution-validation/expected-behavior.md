# Expected Behavior

- Pinned installer resolves exactly to requested candidate tag.
- Installed runtime version includes requested candidate identity.
- Uninstall removes executable from command path.
- Reinstall restores executable with same candidate lineage.
- Homebrew channel lineage either matches candidate or is explicitly classified as drift.

## PASS/WARN/FAIL Discipline
- PASS: Installer and Homebrew channels agree on candidate lineage.
- WARN: Non-blocking lag is detected with explicit remediation path and channel constraints.
- FAIL: Version or lineage mismatch exists for governed install paths.
