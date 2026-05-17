# Distribution Validation Scenario

## Objective
Validate deterministic release-channel convergence across installer, Homebrew, and artifact lineage checks for a certified candidate.

## Inputs
- Candidate tag (required)
- Optional expected SHA for governance recording

## Deterministic Workflow
1. Install candidate using pinned installer tag.
2. Verify runtime version matches candidate tag.
3. Validate uninstall/reinstall behavior.
4. Query Homebrew channel lineage and compare with candidate.
5. Classify PASS/WARN/FAIL and record operational findings.

## Certification Boundary
This scenario validates operational distribution convergence and release lineage integrity. It does not validate CI orchestration internals.
