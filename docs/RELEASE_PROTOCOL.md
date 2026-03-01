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
Bump version and tag
Create a Pull Request for merging the local development branch using the attached PR Protocol