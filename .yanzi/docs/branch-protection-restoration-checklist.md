# Branch Protection Restoration Checklist

Date: 2026-05-17
Target branch: `qa`

- [ ] Confirm `validate` workflow file exists on protected branch (`qa`) and default branch.
- [ ] Confirm workflow trigger includes PR events for target merge flows.
- [ ] Confirm `validate` status context is emitted by the expected GitHub App source.
- [ ] Re-enable required status checks for `qa` (`validate` at minimum).
- [ ] Verify required review policy and admin enforcement match governance intent.
- [ ] Run a PR dry-run against `qa` and confirm check gating works without manual overrides.
- [ ] Confirm release workflow trigger compatibility after PR merge.
- [ ] Re-run deterministic QA certification and confirm PASS under restored protections.
- [ ] Record restoration evidence in governance docs.
