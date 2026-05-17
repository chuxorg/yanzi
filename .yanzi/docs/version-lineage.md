# Version Lineage Contract

## Runtime Identity Contract
`yanzi --version` is the operator-facing lineage identity for a built artifact.

For release artifacts, version output must match release tag identity (for example `v2.9.1-rc1` or `v2.9.1`).

## Embedding Requirements
- Build pipelines must inject release tag identity through linker flags.
- Untagged/local builds may report `dev` and are non-release artifacts.
- Certification must fail if runtime version identity does not match certified candidate tag.

## Distribution Contract
- Installer resolution must emit requested tag and resolved tag.
- Requested tag installs must fail on resolver mismatch.
- Homebrew formula version must reference the same promoted lineage.

## Provenance Requirements
Release lineage evidence must include:
- candidate tag
- candidate commit SHA
- artifact channel URLs
- runtime version output
- certification report and convergence status
