# QA Reports and Certification History

## Reporting Policy
Certification reports are append-only and become part of operational provenance for each release.

## Provenance Requirements
- Reports preserve what was validated, how it was validated, and who performed validation.
- Historical certification records remain available for audit and retrospective analysis.
- Reports should remain human-readable and operationally useful.

## Record Retention Direction
- Each release retains its own certification history.
- New findings are added, not backfilled over prior certification outcomes.

## Example Future Structure
```text
qa/reports/
└── vX.Y.Z/
    ├── certification-report.md
    ├── convergence-validation.md
    ├── findings.md
    └── snapshots/
```
