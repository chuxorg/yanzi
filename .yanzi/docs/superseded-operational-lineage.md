# Superseded Operational Lineage

Date: 2026-05-17

Operational rule applied: keep append-only provenance, but stop promoting intermediate branches once their intent is subsumed by a higher-order convergence branch.

## Superseded / Reduced-Priority PR Lineage

| Repo | PR / Branch | Why Superseded | Replacing Lineage | Recommendation |
|---|---|---|---|---|
| `yanzi` | #103 `feature/governance-reset` | Foundational reset intent is incorporated into later governance docs and convergence packages | `feature/release-convergence-resolution` (PR #118) + #117/#114 | Preserve as provenance only; close PR |
| `yanzi` | #104 `feature/governance-ontology` | Ontology concepts are upstream inputs, not final promotable state | PR #118 consolidated governance state | Preserve as provenance only; close PR |
| `yanzi` | #105 `feature/governance-artifact-specs` | Spec iteration superseded by canonical + release-integrity/release-promotion docs | PR #118 with #113/#117 lineage | Preserve as provenance only; close PR |
| `yanzi` | #106 `feature/canonical-governance-artifacts` | Intermediate artifact packaging superseded by convergence-resolution bundle | PR #118 | Preserve as provenance only; close PR |
| `yanzi` | #107 `feature/governance-validation` | Validation process now embedded in convergence governance model | PR #118 + existing governance docs | Preserve as provenance only; close PR |
| `yanzi` | #108 `feature/qa-certification-model` | Early certification model superseded by full snapshot/certification RC chain | #110 -> #112 -> #115 -> #118 | Preserve as provenance only; close PR |
| `yanzi` | #109 `feature/qa-scenarios-v1` | Scenario layer is now an upstream component, not convergence tip | #111/#112/#115/#118 | Preserve as provenance only; close PR |
| `yanzi` | #110 `feature/snapshot-certification-model` | Architecture superseded by certified baseline and RC evidence | #112/#115/#118 | Preserve as provenance only; close PR |
| `yanzi` | #111 `feature/deterministic-scenario-execution` | Superseded by combined certification + convergence lineage | #112/#115/#118 | Preserve as provenance only; close PR |
| `yanzi` | #112 `feature/certified-snapshot-baseline` | Baseline superseded by RC-specific certification and final convergence branch | #115/#118 | Preserve as provenance only; close PR |
| `yanzi` | #113 `feature/release-integrity-model` | Integrity model content folded into final convergence docs | #118 | Candidate consolidation into #118; then close |
| `yanzi` | #114 `feature/github-hardening-v1` | Hardening controls are prerequisites; convergence branch already references completed hardening | #118 | Candidate consolidation into #118; then close |
| `yanzi` | #115 `feature/release-cert-v2.9.1-rc1` | Certification evidence is required input, but not final canonical branch by itself | #118 | Candidate consolidation into #118; keep open until explicitly confirmed absorbed |
| `yanzi` | #116 `feature/distribution-convergence-v1` | Distribution convergence intent superseded by explicit convergence-resolution lineage across repos | #118 + homebrew PR #1 | Candidate consolidation into #118; then close |
| `yanzi` | #117 `feature/deterministic-release-promotion` | Promotion semantics are upstream to convergence resolution and should not compete as parallel tip | #118 | Candidate consolidation into #118; then close |
| `yanzi` | #94-#101 dev-phase PR stream to `development` | Separate historical development stream; not part of immediate `qa` promotable lineage | None for current RC convergence | Preserve as separate stream; do not force-close in this convergence phase |
| `yanzi` | #96 dependabot patch to `master` | Maintenance patch outside RC governance lineage | None for this RC lineage | Keep independent; not a closure target for convergence |

## Homebrew Operational State

| Repo | PR / Branch | Why Superseded? | Replacing Lineage | Recommendation |
|---|---|---|---|---|
| `homebrew-yanzi` | (none currently superseded) | Only one active convergence PR | N/A | Keep PR #1 as active distribution lineage contributor |

## QA Repo Operational State

| Repo | PR / Branch | Why Superseded? | Replacing Lineage | Recommendation |
|---|---|---|---|---|
| `chux-yanzi-qa` | No open PRs | No competing lineage states detected | N/A | Preserve repository state as provenance baseline |
