# Snapshot Certification Architecture

## Purpose
Snapshots provide deterministic operational certification evidence for Yanzi CLI workflows. They help verify that user-visible behavior, exports, and recovery flows remain stable and trustworthy across releases.

## Deterministic Certification Philosophy
Snapshots are used to certify operational meaning, not to freeze implementation internals. Certification relies on reproducible command execution, deterministic expectations, and human-reviewable evidence.

Snapshots validate:
- operational meaning of CLI workflows
- deterministic workflow behavior under identical inputs
- export consistency across markdown, HTML, and JSON
- CLI behavior stability for key commands
- reproducible operational outputs suitable for release trust decisions

Snapshots do not target:
- brittle cosmetic-only string matching
- implementation-detail lock-in
- non-operational formatting trivia

## Operational Evidence Concepts
Snapshot evidence should answer:
- What command or workflow was executed?
- What output was expected?
- What output was observed?
- What normalization was applied and why?
- What certification decision (PASS/WARN/FAIL) resulted?

Evidence is operational provenance and must remain readable, traceable, and reviewable.

## Snapshot Lifecycle
1. Define deterministic scenario command set.
2. Capture raw operational output.
3. Normalize volatile fields to preserve operational meaning.
4. Compare normalized actual output to normalized expected baseline.
5. Classify drift and assign PASS/WARN/FAIL.
6. Record rationale and retain artifacts as append-only evidence.

## Snapshot Candidate Definitions
### Markdown Exports
Deterministic validation matters because markdown is a primary operator-readable release artifact. Certification value: proves narrative exports remain stable and usable for audits.

### HTML Exports
Deterministic validation matters because HTML output is a distribution-facing rendering format. Certification value: confirms stable report structure and content continuity.

### JSON Exports
Deterministic validation matters because JSON is consumed by downstream tooling and reviewers. Certification value: verifies machine-readable contract stability.

### Help Output
Deterministic validation matters because help text is the operational entrypoint. Certification value: confirms command discoverability and documentation alignment.

### Version Output
Deterministic validation matters because version output anchors release identity. Certification value: ensures runtime binary truth matches release claims.

### Project Listings
Deterministic validation matters because operators rely on listing output to inspect state. Certification value: validates consistent visibility of project inventory.

### Checkpoint Listings
Deterministic validation matters because checkpoints define recovery anchors. Certification value: certifies recovery references remain stable and queryable.

### Rehydration Outputs
Deterministic validation matters because rehydration is operational recovery. Certification value: proves deterministic state reconstruction behavior.

## Snapshot Storage Philosophy
Expected future structure:

```text
qa/snapshots/
└── scenario-name/
    ├── expected/
    ├── actual/
    └── normalized/
```

Storage direction:
- Evidence is append-only certification history.
- Snapshots contribute operational provenance for each validated release.
- Artifacts support release traceability from scenario to decision.
- Structure supports reproducible workflow verification by independent operators.

## Drift Classification
### Expected Drift
Meaning: intentional, approved behavior change.
Certification impact: can remain PASS with explicit rationale.
Example: documented new export section added by release scope.

### Non-Deterministic Drift
Meaning: output varies across identical inputs without intentional change.
Certification impact: WARN or FAIL depending on operational risk.
Example: unstable ordering of listing entries between runs.

### Regression Drift
Meaning: previously certified behavior degraded or broke.
Certification impact: FAIL.
Example: export omits previously present required fields.

### Formatting Drift
Meaning: output format changed without semantic contract break.
Certification impact: WARN by default; FAIL if it harms operational consumption.
Example: heading style changes but required content remains intact.

### Governance Drift
Meaning: certified baseline changed without required approval path.
Certification impact: FAIL until governance requirements are met.
Example: expected snapshot modified without release rationale.

### Operational Drift
Meaning: workflow-level behavior changed in a way that affects operator outcomes.
Certification impact: WARN or FAIL based on impact severity.
Example: checkpoint creation still works but output no longer exposes required identifiers.

## Certification Expectations
- Snapshot evidence supports PASS/WARN/FAIL discipline.
- Deterministic outputs improve release trustworthiness.
- Certification evidence remains human-reviewable.
- Snapshot review and baseline updates remain governance-driven.

## Non-Goals
- Snapshots are not pixel-perfect UI testing.
- Snapshots are not implementation lock-in mechanisms.
- Snapshots are not orchestration systems.
- Snapshots do not replace operational scenario review.

## Future Direction
Possible future areas (no implementation commitment in this phase):
- automated snapshot comparison helpers
- export certification tooling layers
- deterministic release manifest generation
- release provenance bundles
- governance-integrated certification workflows
