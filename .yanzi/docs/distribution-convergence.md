# Distribution Convergence Governance

## Objective
Define deterministic convergence rules so all governed install paths resolve to the same certified release lineage.

## Governed Channels
- GitHub Releases (artifact source of record)
- install.sh (channel resolver)
- Homebrew tap formula
- apt/deb distribution artifacts
- direct binary artifacts

## Required Lineage Agreement
All governed channels must agree on:
- release tag
- candidate commit lineage
- runtime version identity (`yanzi --version`)
- channel artifact provenance

## Propagation Expectations
1. Candidate certified first.
2. Promotion-approved release tag published.
3. Artifacts published and checksummed.
4. Installer and Homebrew updated to same release lineage.
5. Convergence validation executed and recorded.

## Certification Timing
- Certification occurs on immutable candidate state.
- Promotion eligibility evaluated only after convergence validation completes.
- Post-publication validation is required before declaring converged release state.

## Drift Handling
- Lineage mismatch: FAIL, promotion blocked.
- Channel lag with explicit operator impact: WARN or FAIL based on trust risk.
- Snapshot regression: FAIL.
- Documentation mismatch affecting install correctness: WARN/FAIL.

## Deterministic Rule
All governed install paths must resolve to the same certified lineage before promotion.
