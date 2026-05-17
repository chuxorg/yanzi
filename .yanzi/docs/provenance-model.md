# Provenance Model

## Purpose
Define how Yanzi preserves release and certification traceability.

## Append-Only Certification History
- Certification records are append-only.
- Historical outcomes are retained for auditability.
- Corrections are added as new records, not silent rewrites.

## Release Certification History
- Each release should reference its certification evidence.
- Evidence should include scenario outcomes, drift disposition, and reviewer notes.
- Promotion decisions should be reconstructable from stored artifacts.

## Operational Traceability
Traceability should connect:
- branch and PR context
- deterministic scenario execution
- normalized and expected snapshots
- certification reports
- release promotion decisions

## Deterministic Release Evidence
- Evidence must be human-readable and reproducible.
- Deterministic workflows are preferred over implicit automation state.
- PASS/WARN/FAIL outcomes should be explicit.

## Certification Artifact Retention
- Preserve certification reports and snapshot baselines.
- Preserve drift findings and governance rationale.
- Maintain discoverable structure for historical review.

## Future Direction (No Implementation Commitment)
- signed releases
- SBOM generation
- deterministic release manifests
- provenance bundles
