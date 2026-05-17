# 1. Purpose

QA certification validates release readiness before promotion from `qa` to `master`.

Releases must be operationally reproducible using documented workflows and validation steps.

Certification reduces release ambiguity by applying consistent validation criteria and explicit outcomes.

Certification must validate real workflows, not theoretical assumptions.

Deterministic validation is required before release.

# 2. Certification Principles

- real workflow validation
- deterministic execution
- reproducible results
- explicit `PASS | WARN | FAIL` classification
- operational transparency
- human-readable reporting
- governance compliance validation

# 3. Certification Levels

## PASS

Definition:
- release acceptable
- no known blocking defects

Examples:
- build, CLI, and regression checks pass with no blocking findings
- documentation matches current behavior for validated workflows

## WARN

Definition:
- release acceptable with documented concerns
- non-blocking operational issues allowed

Examples:
- minor documentation gap with clear workaround
- non-critical usability issue that does not compromise deterministic behavior

## FAIL

Definition:
- release blocked
- deterministic behavior compromised
- regression present
- critical workflow broken

Examples:
- command output is inconsistent across equivalent runs
- required workflow cannot be completed from current release candidate
- export output is broken or unreadable for supported flows

# 4. Required Validation Areas

## Build Validation

Purpose:
- confirm release candidate can be built using documented build process

Minimum expectations:
- build completes successfully
- resulting binary/artifact is usable for intended validation steps

Failure examples:
- build process fails on required environment
- produced artifact cannot execute

## CLI Validation

Purpose:
- verify user-facing command behavior is correct and deterministic

Minimum expectations:
- key commands execute successfully
- help, argument, and error behavior is validated against expected behavior

Failure examples:
- command execution fails for supported usage
- equivalent invocations produce inconsistent results

## Workflow Validation

Purpose:
- verify documented engineering, QA, and release workflows are executable

Minimum expectations:
- workflow steps can be followed as written
- validation and failure-handling points are actionable

Failure examples:
- workflow requires missing implicit steps
- documented sequence cannot be completed

## Export Validation

Purpose:
- verify export/output behavior remains usable and deterministic

Minimum expectations:
- exports complete for supported scenarios
- exported outputs are readable and consistent

Failure examples:
- export command fails or produces malformed output
- export output format changes without documentation

## Documentation Validation

Purpose:
- ensure docs reflect actual release behavior

Minimum expectations:
- install and quickstart instructions are accurate
- workflow and example documentation match observed behavior

Failure examples:
- documentation references commands/flags that no longer work
- release notes omit significant behavior changes

## Regression Validation

Purpose:
- verify previously operational behavior remains operational

Minimum expectations:
- core prior workflows still function
- deterministic behavior remains intact

Failure examples:
- prior documented flow now fails without intentional deprecation
- previously stable behavior is inconsistent

## Governance Validation

Purpose:
- verify governance artifacts remain clear, explicit, and usable

Minimum expectations:
- required packs/rules/workflows are loadable and relevant
- governance guidance can be applied without ambiguity

Failure examples:
- pack loads irrelevant or conflicting guidance
- workflow or rule language is ambiguous in practice

## Release Artifact Validation

Purpose:
- verify release artifacts are complete and aligned with release candidate

Minimum expectations:
- expected artifacts are present and verifiable
- version/tag alignment is correct

Failure examples:
- missing required artifact
- artifact version does not match tagged release

# 5. CLI Validation Requirements

CLI certification expects validation of:
- command execution
- help output
- argument handling
- error handling
- deterministic output
- local mode behavior
- export functionality

Requirements:
- validation must use real CLI execution
- theoretical-only validation is insufficient

# 6. Documentation Validation

Documentation certification must validate:
- install instructions
- quickstart flows
- workflow documentation
- examples
- release notes
- governance docs

Documentation rules:
- docs must reflect actual behavior
- outdated docs are certification failures

# 7. Regression Validation

Regression certification must confirm:
- previous workflows remain operational
- exports remain readable
- deterministic behavior is preserved
- governance composition remains stable

# 8. Release Candidate Validation Flow

1. Feature PRs merged into `qa`.
2. QA validation executed.
3. `PASS | WARN | FAIL` report produced.
4. Human review.
5. `qa` to `master` decision.
6. Tag creation.
7. Release artifact verification.
8. Post-release validation.

# 9. QA Report Structure

Standard report sections:
- Scope
- Environment
- Validation Areas
- Findings
- `PASS | WARN | FAIL` Status
- Blocking Issues
- Recommendations
- Release Decision

# 10. Non-Goals

- QA certification is not autonomous release approval.
- Automation does not replace governance.
- CI systems do not replace validation discipline.
- Certification must fail when installed version, lineage, or install channel is ambiguous relative to candidate tag.
- Releases remain human-governed.

# 11. Future Direction

Possible future directions, without implementation commitments:
- automated regression harnesses
- snapshot testing
- deterministic export validation
- governance linting
- reproducible QA environments
