# Yanzi QA Harness

## QA Philosophy
This QA harness focuses on deterministic operational validation of real CLI behavior.

Core principles:
- filesystem-first execution
- deterministic checks
- inspectable artifacts
- human-readable workflows
- explicit, simple commands

## What This QA Layer Validates
- Real workflows executed through the actual `yanzi` binary
- Release-facing behavior for core command paths
- Documentation accuracy for command usage and expected outcomes
- Deterministic operational behavior under repeat execution

## What This Is Not
- Not CI/CD orchestration
- Not browser automation
- Not Selenium-style testing

## Bats Usage
Install local Bats dependencies:

```bash
qa/scripts/install-bats.sh
```

Run smoke suite:

```bash
qa/vendor/bin/bats qa/suites/smoke
```

## Certified Snapshot Baselines
Certified baselines are governance-approved deterministic snapshots stored under `qa/snapshots/<scenario>/expected/`.

These baselines are operational truth surfaces for workflow certification, not disposable test artifacts. Snapshot updates require explicit human review and recertification.

## Operational Provenance
Certification evidence is append-only and human-reviewable:
- scenario outputs (`actual/`)
- normalized comparison artifacts (`normalized/`)
- certified baselines (`expected/`)
- certification reports in `qa/reports/`

## Human-Governed Certification
Baseline promotion is manual and explicit. Reviewers must verify operational meaning, drift classification, and normalization integrity before approving any expected snapshot updates.

## Deterministic Certification Philosophy
Certification emphasizes reproducible command execution and stable operator-visible behavior. PASS/WARN/FAIL outcomes are assigned through governance-driven review, not hidden automation.

## Directory Structure
- `certification/`: certification governance and review procedures
- `execution/`: explicit deterministic scenario execution scripts
- `fixtures/`: static inputs used by deterministic scenarios
- `reports/`: generated QA run outputs and certification summaries
- `scenarios/`: operational validation playbooks
- `scripts/`: setup and execution helper scripts
- `snapshots/`: baseline output contracts and comparison artifacts
- `suites/`: executable Bats test suites
- `tmp/`: temporary files generated during QA runs
- `vendor/`: locally installed Bats runtime and helper libraries

## Workflow-Oriented Testing
Tests are organized around operator-visible CLI workflows rather than isolated implementation details. The suite emphasizes reproducible command execution and contract validation.

## Unit Testing vs Operational Certification
- Unit tests verify internal logic in isolation.
- This QA harness certifies end-to-end operational behavior for release readiness.
