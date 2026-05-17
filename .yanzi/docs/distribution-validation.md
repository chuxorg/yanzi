# Distribution Validation Philosophy

## Purpose
Define deterministic validation expectations for release distribution paths.

Distribution paths must be QA validated. Install behavior is operationally significant.
Version drift is a release integrity issue.

## Homebrew Validation
- Validate install success and executable availability.
- Validate `yanzi --version` matches release expectation.
- Validate uninstall behavior and cleanup expectations.

## APT Validation
- Validate package install and path correctness.
- Validate runtime version output.
- Validate upgrade behavior between released versions.
- Validate uninstall/remove behavior.

## Binary Release Validation
- Validate artifact presence and expected naming.
- Validate executable startup, help output, and version correctness.
- Validate checksums against release records.

## Upgrade Behavior Validation
- Validate that upgrades preserve expected operational behavior.
- Validate no unexpected regressions in core workflows.
- Validate migration/recovery behavior where applicable.

## Governance Expectations
- Distribution verification evidence is included in release review.
- Drift across distribution channels requires explicit disposition.
- Unverified distribution paths block trusted release promotion.
