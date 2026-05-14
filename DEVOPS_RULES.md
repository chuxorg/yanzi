# DEVOPS_RULES

## Release Process
- Releases follow a documented, repeatable process.
- Release inputs and outputs must be traceable.

## Deployment Approvals
- Production deployments require explicit approval gates.
- Approval ownership and audit trail must be clear.

## Release Validation
- Validate release candidates before production rollout.
- Validation evidence should be retained for review.

## Tagging Rules
- Tags should be consistent, immutable, and semantically meaningful.
- Release tags must map to a single authoritative commit.

## Artifact Verification
- Verify build artifacts for integrity and expected contents.
- Artifact provenance must be attributable.

## Rollback Philosophy
- Every release should have a practical rollback path.
- Rollback criteria and execution ownership should be pre-defined.
