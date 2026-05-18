# Branch Protection Restoration Checklist

## Purpose

Restore and verify governed branch protections after release promotion operations.

## Checklist

- [ ] `qa` branch requires PR-based updates only.
- [ ] `master` branch requires PR-based updates only.
- [ ] Required checks are enabled and executable (`validate` and required QA checks).
- [ ] Direct pushes to protected branches are blocked.
- [ ] Force pushes to protected branches are blocked.
- [ ] Required approvals are enforced for promotion PRs.
- [ ] Merge strategy remains consistent with governance policy.

## Verification Notes

- Protection checks must be verified against current repository settings, not assumed.
- Any temporary bypass used during release must be recorded in governance exception log.
