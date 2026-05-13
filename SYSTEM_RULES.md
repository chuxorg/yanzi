# Yanzi AI System Rules

These rules govern all AI-assisted development within the Yanzi project.
They are mandatory and must be followed exactly.

---

# 1. Execution Workflow

## Pre-Task Requirements

Before executing any development task:

* Ignore all changes in:

  * `YANZI_LOG.*`
  * issue tracking files

* Confirm:

  * Current branch is a **feature branch**
  * Feature branch is created from `development`
  * NOT on `development` or `master`

---

## Task Definition

A **task** is a single, discrete unit of work as defined by the execution prompt.

---

## During Task Execution

After completing each task:

1. Stage changes
2. Commit with clear message
3. Push to remote branch

---

## Task Completion Requirements

Before pushing any changes:

* All builds MUST succeed
* All functionality must have associated unit tests.
* All unit tests MUST pass
* All new or modified testable code MUST include unit tests
* Code MUST follow:

  * `docs/CODE_DOCUMENTATION.md`

Failure in any of the above blocks the push.

---

## Phase Completion

At the end of a phase:

* Create PR → `development`
* DO NOT create tags
* Human explicitly determines phase completion

---

# 2. Git Safety Rules (CRITICAL)

The following actions are strictly prohibited:

* NEVER commit directly to `development`
* NEVER commit directly to `master`
* NEVER create or push tags before merge
* NEVER bypass tests or build failures
* NEVER rewrite history on shared branches

---

# 3. Destructive Action Protection (CRITICAL)

* NEVER perform destructive operations without explicit human instruction

Examples:

* Dropping databases
* Deleting production data
* Overwriting critical files
* Removing infrastructure

If uncertain → STOP and request clarification

---

# 4. Stability Guarantees

## Command Stability

* `capture`, `verify`, and `chain` are stable
* Behavior must remain consistent across minor and patch releases

## Flag Compatibility

* Flags remain backward compatible within a major version
* Deprecation must occur before removal
* Breaking changes only allowed in major versions

## Versioning

* Follows semantic versioning: `MAJOR.MINOR.PATCH`

  * MAJOR → breaking changes
  * MINOR → backward-compatible features
  * PATCH → fixes and documentation

---

# 5. Release Process

## Overview

Two-path merge model:

* QA Build → merge into `development`
* Release Build → merge `development` → `master`

---

## QA Build Path

Triggered on merge into `development`

Required checks:

* `go mod tidy` (clean state required)
* `go vet ./...`
* `go test ./...`
* `go build ./cmd/yanzi`

All must pass.

---

## Release Build Path

Triggered when:

* PR from `development` → `master` is merged, OR
* Tag `vX.Y.Z` is pushed

---

## Release Rules

* `VERSION` contains `X.Y.Z`
* Tags must be `vX.Y.Z`
* NEVER reuse tags

---

## Operator Steps

1. Merge feature work into `development`
2. Confirm QA build passes
3. Update `CHANGELOG.md` (create if missing)
4. Bump `VERSION`
5. Create PR: `development` → `master`
6. Merge PR
7. Confirm release workflow completes
8. Verify GitHub release artifacts

---

## Build Output

* linux/amd64
* darwin/amd64
* darwin/arm64

Version embedded via:

* `-ldflags "-X main.version=<tag>"`

---

# 6. Security Reporting

If a vulnerability is discovered:

* Report via GitHub Security Advisories
* Do NOT disclose publicly until resolved

Response target:

* Acknowledge within 3 business days

---

# 7. Rule Authority

These rules are:

* Canonical
* Version-controlled
* Enforced via AI execution discipline

If a rule conflicts with a prompt:

→ The rule takes precedence
