# Snapshot Normalization Rules

## Objective
Normalization removes volatile noise so snapshot comparison preserves deterministic operational meaning.

## Philosophy
Normalization exists to preserve certification signal quality. It should remove only non-deterministic variability while keeping operationally meaningful behavior exact.

## Normalize These Volatile Fields
### Timestamps
Normalize absolute timestamps when they are generated at runtime and not the direct subject of validation.
Reason: timestamps vary per execution and can hide real behavioral drift.

### ULIDs / Generated IDs
Normalize generated identifiers when identity values are not the contract being certified.
Reason: generated IDs are expected to vary even when behavior is correct.

### Temporary Paths
Normalize ephemeral filesystem paths (`/tmp/...`, run-specific directories).
Reason: host/runtime path allocation varies by environment.

### Environment-Specific Values
Normalize hostnames, user home paths, and machine-specific prefixes.
Reason: environment differences should not invalidate workflow certification.

## Keep These Exact
- command names and subcommands
- exit status semantics
- required structural fields in exports
- logical ordering when deterministic ordering is part of behavior
- key status markers used for PASS/WARN/FAIL decisions

## Why Normalization Is Necessary
Without normalization, volatile runtime values create false drift and reduce trust in certification results. With scoped normalization, reviewers can focus on true operational regressions.

## Guardrails
- Do not normalize away errors, warnings, or governance markers.
- Do not normalize fields that define release identity (for example: version string under validation).
- Do not normalize ordering when deterministic ordering is required by scenario.

## Example Normalization Intent
- Replace runtime timestamp values with `<TIMESTAMP>` token.
- Replace ULID-like values with `<ID>` token when ID value is not under test.
- Replace ephemeral paths with `<TMP_PATH>` token.

Normalization must be documented per scenario so reviewers can verify that operational meaning is preserved.
