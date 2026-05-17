# Yanzi Release Protocol

## Branch Model

- All work must occur on feature branches.
- No direct commits to development or master.
- Each phase is completed on a feature branch.
- After phase completion:
  - Create PR -> merge into development.
- After QA validation:
  - Create PR -> merge development -> master.

## Versioning

- Version type (major/minor/patch) is decided by human before release.
- `VERSION` stores plain semver (`X.Y.Z`).
- Release automation resolves tag `v$(cat VERSION)` during master release flow.
- No tag reuse.
- No pseudo-versions.
- No replace directives in go.mod.

## Deterministic Lineage Requirements

- Candidate tag and candidate SHA must be recorded before certification.
- Release artifacts must embed version equal to candidate tag.
- `yanzi --version` must match candidate tag in distribution validation.
- Certification fails on lineage mismatch or channel ambiguity.

## QA Flow

1. Merge feature branch -> development.
2. QA build runs on merged PR event.
3. Validate QA checks.
4. Execute deterministic release certification (including distribution channels).
5. Merge development -> master only when certification is PASS or accepted WARN.
6. Release build runs on merged PR event.
7. GitHub release is created with production binaries and distribution archives.

## Homebrew Synchronization Contract

- Tap formula version and checksums must be synchronized during promotion.
- Release certification must verify Homebrew tap install behavior.
- RC policy must explicitly declare whether RC is published to tap; ambiguity is a gating failure.

## Release Process

Releases are generated from the CLI repository.

Version numbers are embedded in the binary using ldflags.

Example:

```bash
go build -ldflags "-X main.version=v1.1.0"
```

Artifacts are published as GitHub releases.
