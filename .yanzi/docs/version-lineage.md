# Version Lineage

## Purpose

Version lineage ensures released artifacts, runtime version output, and promotion records point to the same release identity.

## Deterministic Lineage Rules

- Release tags are immutable.
- A certified candidate tag maps to one commit SHA.
- Release artifacts must be built with embedded version equal to the certified tag.
- Runtime `yanzi --version` must match the release tag used for build.
- QA certification fails if runtime version diverges from candidate tag.

## Build and Verification Discipline

- Build release artifacts using explicit ldflags version injection.
- Record candidate tag and candidate SHA in certification evidence.
- Validate installed runtime version after every distribution-channel install.
- Treat install-channel ambiguity as certification failure.

## Operational Contract

- Candidate lineage is proven by: tag, commit SHA, artifact source URL, and runtime version output.
- Missing or mismatched lineage evidence blocks release promotion.
