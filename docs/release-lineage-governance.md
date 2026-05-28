# Release Lineage Governance

## Purpose
Define operational semantics for release-candidate lineage and stable production lineage.

## RC Lineage Semantics

1. RC tags (for example `v2.9.1-rc1`) are certification candidates.
2. RC lineage is append-only provenance and must never be rewritten.
3. RC artifacts remain consumable only by explicit request (`--version=<rc-tag>` or explicit artifact URL).

## Stable Lineage Semantics

1. Stable tags (for example `v2.9.1`) are official production lineage.
2. Default distribution channels must resolve stable lineage only.
3. Stable promotion is a deterministic transition from certified RC lineage to official release lineage.

## RC to Stable Promotion Semantics

1. Certification completes on RC lineage.
2. Official stable tag is promoted and assets are published.
3. Distribution channels are normalized:
   - Homebrew formula version/url/checksum
   - installer default version tag
   - release documentation
4. Lightweight post-release convergence validation confirms default channel correctness.

## Normalization Expectations

1. Stable/default paths must never resolve RC artifacts after promotion.
2. RC install behavior must require explicit version request.
3. Historical certification evidence remains unchanged and traceable.

## Provenance Continuity Requirements

1. No historical certification rewrite.
2. No RC provenance deletion.
3. Promotion actions must be traceable in commits and validation reports.
