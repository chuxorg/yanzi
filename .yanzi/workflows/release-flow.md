# Name

release-flow

# Purpose

Define release promotion and verification from QA-approved code.

# Preconditions

- QA approval is documented.
- Release and system rules are explicitly loaded.
- Target version is defined.

# Steps

1. Confirm QA approval requirements are satisfied.
2. Merge with disciplined promotion from `qa` to `master`.
3. Create and push the release tag for the target version.
4. Build and verify release artifacts.
5. Run deterministic release validation checks.
6. Publish release notes with artifact references.

# Validation

- Tag matches intended version and source commit.
- Artifacts are complete and verifiable.
- Release checks pass with reproducible results.

# Failure Handling

- If verification fails, halt release and investigate before publish.
- If critical issues are found post-tag, execute rollback expectations and document actions.
