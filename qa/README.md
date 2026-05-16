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

## Directory Structure
- `fixtures/`: static inputs used by deterministic scenarios
- `reports/`: generated QA run outputs and summaries
- `scenarios/`: operational validation playbooks
- `scripts/`: setup and execution helper scripts
- `snapshots/`: baseline output contracts (future expansion)
- `suites/`: executable Bats test suites
- `tmp/`: temporary files generated during QA runs
- `vendor/`: locally installed Bats runtime and helper libraries

## Workflow-Oriented Testing
Tests are organized around operator-visible CLI workflows rather than isolated implementation details. The suite emphasizes reproducible command execution and contract validation.

## Unit Testing vs Operational Certification
- Unit tests verify internal logic in isolation.
- This QA harness certifies end-to-end operational behavior for release readiness.
