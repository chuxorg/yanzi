# Expected Behavior

## Release Certification Expectations
- Published artifacts match expected version and checksum records.
- Installed binary reports exact release version.
- Distribution endpoints provide complete release set.
- Installer and Homebrew resolve the same certified candidate lineage.
- Core install and smoke commands execute without crashes.
- Documentation examples remain accurate for released behavior.

## Operational Trust Expectations
- Operators can reproduce release validation with explicit commands.
- Validation evidence is human-readable and archived.
- Certification status is unambiguous and auditable.

## Validation Requirements
- Any checksum mismatch is FAIL.
- Any version mismatch between artifact and runtime output is FAIL.
- Any governed channel lineage mismatch is FAIL.
- Documentation mismatch is WARN or FAIL based on operator impact, with rationale recorded.
