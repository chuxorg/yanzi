# Yanzi QA Rules

## QA Philosophy
- Deterministic validation: QA outcomes must be repeatable from a clean environment with explicit inputs.
- Regression prevention: every discovered defect should become a test, fixture, or snapshot contract where practical.
- Release certification mindset: QA evaluates release readiness, not just test pass rates.
- Exports as behavioral contracts: markdown/json/html exports are treated as externally visible contracts.
- Documentation claim validation: QA verifies that documentation statements and examples match real CLI behavior.

## Branch Strategy
- `feature/*`: implementation branches for product changes.
- `qa/*`: validation branches derived from completed feature branches. These can evolve independently for QA artifacts.
- `development`: integration branch for active feature development.
- `master`: stable release branch.

Lifecycle:
1. Product work lands in `feature/*` and merges toward `development`.
2. QA starts from that completed scope in `qa/*`.
3. QA adds tests, fixtures, snapshots, and validation scripts without redesigning production behavior.
4. QA findings either block merge (showstopper) or are categorized for deferment.
5. Certified scope merges from QA branch per release governance.

## QA Responsibilities
- Unit testing and targeted bug-regression tests.
- CLI end-to-end validation across core user flows.
- Snapshot verification for deterministic exports.
- Migration validation for schema and startup behavior.
- Install/uninstall validation for local developer workflows.
- Documentation validation for command accuracy and examples.
- Link validation for internal markdown references and URL formatting.

## PASS | FAIL | WARN Semantics
- `PASS`: requirement validated with deterministic evidence.
- `FAIL`: requirement not met or behavior regressed.
- `WARN`: non-blocking risk, ambiguity, or incomplete coverage.
- `showstopper`: release-blocking failure requiring correction before merge.
- `deferred issue`: accepted non-showstopper tracked for later fix.
- `enhancement`: quality improvement request that is outside blocking QA scope.

## QA Restrictions
QA may:
- add tests
- add fixtures
- add snapshots
- improve docs
- improve validation scripts/tooling

QA may NOT:
- redesign shipped features
- add unrelated functionality
- silently alter behavior to make tests pass

## Release Certification Concept
Release certification will evolve into a formal checklist with required evidence bundles:
- deterministic test run logs
- snapshot verification results
- migration/install validation outcomes
- documentation and link validation reports
- categorized defect list (`showstopper`, `deferred issue`, `enhancement`)

A release is certifiable only when showstoppers are resolved and remaining WARN/deferred items are explicitly accepted.
