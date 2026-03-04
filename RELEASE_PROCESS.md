# Yanzi Release Process

## Overview
Yanzi is released from a single repository and follows a two-path merge model:

- QA Build path: merged PRs into `development`
- Release Build path: merged PR from `development` into `master`

## Versioning Rules
- Semantic versioning is required.
- `VERSION` stores plain semver (`X.Y.Z`).
- Production releases follow `vX.Y.Z`.
- Never reuse tags.

## QA Build Path
Triggered by GitHub Actions when a pull request is merged into `development`.

Expected checks:
- `go mod tidy` with dirty check on `go.mod`/`go.sum`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/yanzi`

## Release Build Path
Triggered by GitHub Actions when a pull request from `development` is merged into `master`.

Release workflow behavior:
- Reads `VERSION` and resolves release tag `v$(cat VERSION)`.
- Validates the tag format as `vX.Y.Z`.
- Builds binaries for:
  - linux/amd64
  - darwin/amd64
  - darwin/arm64
- Embeds version via:
  - `-ldflags "-X main.version=<tag>"`
- Creates a GitHub release and uploads built binaries.

## Operator Steps
1. Land feature work into `development` via PR.
2. Confirm QA build passes on merge.
3. Bump `VERSION` to next production version in `development`.
4. Merge `development` into `master` via PR.
5. Update CHANGELOG.md - if the file does not exsist, create it.
5. Confirm release workflow completes and GitHub release is created.
