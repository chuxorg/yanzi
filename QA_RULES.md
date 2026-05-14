# QA_RULES

## PASS / FAIL / WARN
- Define test outcomes using PASS, FAIL, and WARN classifications.
- WARN indicates risk requiring explicit acknowledgment.

## Regression Expectations
- Prevent known-good behavior from regressing.
- Reproductions and regression checks should be documented.

## Snapshot Validation
- Snapshot changes require intent review, not blind acceptance.
- Unexpected snapshot drift should be investigated before approval.

## Documentation Validation
- User-facing and operational docs must match implemented behavior.
- Material behavior changes require documentation updates.

## Release Certification Philosophy
- Certification is evidence-based, traceable, and risk-aware.
- Unresolved critical issues block release readiness.

## QA Escalation Process
- Escalate ambiguous or high-risk findings promptly.
- Record escalation context, impact, and recommended disposition.
