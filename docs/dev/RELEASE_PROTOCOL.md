# Yanzi Release Protocol

## Branch Model

- All work must occur on feature branches.
- No direct commits to development or master.
- Each phase is completed on a feature branch.
- After phase completion:
  - Create PR → merge into development.
- After QA validation:
  - Create PR → merge development → master.

## Versioning

- Version type (major/minor/patch) is decided by human before release.
- `VERSION` stores plain semver (`X.Y.Z`).
- Release automation resolves tag `v$(cat VERSION)` during master release flow.
- No tag reuse.
- No pseudo-versions.
- No replace directives in go.mod.

## QA Flow

1. Merge feature branch → development.
2. QA build runs on merged PR event.
3. Validate QA checks.
4. Merge development → master.
5. Release build runs on merged PR event.
6. GitHub release is created with production binaries.

---

Stage, Commit, and Push any and all changes

Create a Pull Request for merging the local development branch using the attached PR Protocol

# Release Process

Releases are generated from the CLI repository.

Bump version and tag 

Tagging the CLI triggers GitHub Actions which build binaries for supported platforms.

Version numbers are embedded in the binary using ldflags.

Example:

```go build -ldflags "-X main.version=v1.1.0"```

Artifacts are published as GitHub releases.

The Yanzi repository provides documentation and entry points to these releases.
---