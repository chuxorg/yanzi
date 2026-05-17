# Open PR Inventory (Deterministic Convergence)

Date: 2026-05-17
Branch analyzed: `feature/release-convergence-consolidation`
Scope repositories:
- `chuxorg/yanzi`
- `chuxorg/chux-yanzi-qa` (requested as `yanzi-cli-qa`; canonical slug discovered via org inventory)
- `chuxorg/homebrew-yanzi`

## `chuxorg/yanzi`

| PR | Title | Source -> Target | Operational Purpose | Related Release Phase | Governance Relevance | QA Relevance | Provenance Relevance | Current Status |
|---|---|---|---|---|---|---|---|---|
| #118 | fix: resolve deterministic release convergence for v2.9.1-rc1 | `feature/release-convergence-resolution` -> `qa` | Consolidate convergence semantics across governance/QA/distribution docs | Convergence stabilization | High | High | High | Open, blocked on review |
| #117 | docs+qa: deterministic release promotion semantics and convergence governance | `feature/deterministic-release-promotion` -> `qa` | Define deterministic promotion contract | Promotion semantics | High | Medium | High | Open, blocked on review, checks pass |
| #116 | chore: deterministic distribution convergence hardening | `feature/distribution-convergence-v1` -> `qa` | Align distribution checkpoints and branch governance | Distribution hardening | High | Medium | Medium | Open, blocked on review |
| #115 | qa: v2.9.1-rc1 deterministic release certification evidence | `feature/release-cert-v2.9.1-rc1` -> `qa` | Add RC certification artifacts and evidence | RC certification | Medium | High | High | Open, blocked on review, checks pass |
| #114 | ci/docs: implement GitHub governance and release hardening v1 | `feature/github-hardening-v1` -> `qa` | CI + governance controls | Governance hardening | High | Medium | Medium | Open, blocked on review, checks pass |
| #113 | docs: add release integrity and operational trust architecture | `feature/release-integrity-model` -> `qa` | Release integrity model | Governance foundation | High | Low | Medium | Open, blocked on review, checks pass |
| #112 | qa: establish certified deterministic snapshot baseline | `feature/certified-snapshot-baseline` -> `qa` | Snapshot baseline for deterministic QA | Snapshot baseline | Medium | High | High | Open, blocked on review, checks pass |
| #111 | qa: implement deterministic scenario execution foundation | `feature/deterministic-scenario-execution` -> `qa` | Deterministic scenario execution foundation | QA foundation | Medium | High | Medium | Open, blocked on review, checks pass |
| #110 | qa: add snapshot certification architecture | `feature/snapshot-certification-model` -> `qa` | Snapshot certification architecture | QA certification architecture | Medium | High | High | Open, blocked on review, checks pass |
| #109 | qa: add deterministic operational QA scenarios | `feature/qa-scenarios-v1` -> `qa` | Scenario inventory for QA | QA scenarios | Medium | High | Medium | Open, blocked on review, checks pass |
| #108 | docs: add deterministic QA certification model | `feature/qa-certification-model` -> `qa` | QA certification model (earlier iteration) | QA model evolution | Medium | High | Medium | Open, blocked on review |
| #107 | docs: add governance validation process and observation tracking | `feature/governance-validation` -> `qa` | Governance validation process | Governance iteration | Medium | Medium | Medium | Open, blocked on review |
| #106 | docs: add canonical governance artifacts and initial packs | `feature/canonical-governance-artifacts` -> `qa` | Canonical artifact definitions | Governance iteration | High | Low | Medium | Open, blocked on review |
| #105 | docs: add canonical governance artifact specifications | `feature/governance-artifact-specs` -> `qa` | Governance artifact spec baseline | Governance iteration | High | Low | Medium | Open, blocked on review |
| #104 | docs: add governance ontology for deterministic context composition | `feature/governance-ontology` -> `qa` | Governance ontology baseline | Governance iteration | High | Low | Low | Open, blocked on review |
| #103 | chore: reset governance structure for deterministic context composition | `feature/governance-reset` -> `qa` | Reset governance structure | Governance reset | High | Low | Low | Open, blocked on review |
| #102 | Implement initial Yanzi QA framework and regression foundation | `qa-framework-process-implementation` -> `qa` | Initial QA framework baseline | QA framework base | Medium | High | Medium | Open, blocked on review, checks pass |
| #101 | Formalize machine-readable contracts and metadata guidance | `feature/phase9-operational-contracts` -> `development` | Development-phase contract hardening | Dev phase stream | Medium | Low | Low | Open, clean |
| #100 | Add continuity observability and status surfaces | `feature/phase8-continuity-observability` -> `development` | Development observability iteration | Dev phase stream | Low | Medium | Low | Open, clean |
| #99 | Phase 7: harden local SQLite runtime and contention behavior | `feature/phase7-sqlite-operational-hardening` -> `development` | Runtime hardening | Dev phase stream | Low | Medium | Low | Open, clean |
| #98 | Phase 6: harden continuity semantics and protocol integrity | `feature/phase6-continuity-hardening` -> `development` | Continuity protocol hardening | Dev phase stream | Low | Medium | Low | Open, clean |
| #97 | Phase 5: demo readiness and operational polish | `feature/phase5-demo-readiness` -> `development` | Product polish | Dev phase stream | Low | Low | Low | Open, clean |
| #96 | chore(deps): bump modernc.org/sqlite from 1.50.0 to 1.50.1 | `dependabot/go_modules/modernc.org/sqlite-1.50.1` -> `master` | Dependency patch | Maintenance stream | Low | Low | Low | Open, blocked on review |
| #95 | Phase 4: AI-agent ergonomics and workflow usability | `feature/phase4-agent-ergonomics` -> `development` | Agent ergonomics | Dev phase stream | Low | Low | Low | Open, clean |
| #94 | [codex] Phase 3: harden install flow and align docs | `feature/phase3-install-reliability` -> `development` | Install flow hardening | Dev phase stream | Low | Medium | Low | Open (draft), clean |

## `chuxorg/chux-yanzi-qa`

No open PRs.

## `chuxorg/homebrew-yanzi`

| PR | Title | Source -> Target | Operational Purpose | Related Release Phase | Governance Relevance | QA Relevance | Provenance Relevance | Current Status |
|---|---|---|---|---|---|---|---|---|
| #1 | fix: align Homebrew RC artifact lineage | `feature/release-convergence-resolution` -> `master` | Align formula lineage with RC convergence state | Distribution convergence | High | Medium | High | Open, clean |

## Operational Relevance Classification

| Repo | PR | Classification | Rationale |
|---|---|---|---|
| `yanzi` | #118 | Active Release Convergence; Candidate for Promotion | End-of-line convergence branch resolving RC deterministic release semantics. |
| `yanzi` | #117 | Governance Evolution; Candidate for Consolidation | Promotion governance contract feeds canonical convergence but is not final lineage tip. |
| `yanzi` | #116 | Distribution Hardening; Candidate for Consolidation | Distribution alignment work is required input to convergence but superseded by #118 semantics. |
| `yanzi` | #115 | QA Certification; Snapshot/Provenance; Candidate for Consolidation | RC certification evidence is required proof artifact lineage. |
| `yanzi` | #114 | Governance Evolution; Candidate for Consolidation | CI/governance controls are prerequisite controls feeding canonical convergence state. |
| `yanzi` | #113 | Governance Evolution; Candidate for Consolidation | Release integrity model contributes governance substrate for final convergence docs. |
| `yanzi` | #112 | Snapshot/Provenance; Superseded Operational State; Candidate for Closure | Baseline evidence superseded by RC-specific evidence and final convergence branch. |
| `yanzi` | #111 | QA Certification; Superseded Operational State; Candidate for Closure | Foundational scenario execution no longer the active promotable tip. |
| `yanzi` | #110 | Snapshot/Provenance; Superseded Operational State; Candidate for Closure | Certification architecture iteration superseded by later baseline + RC evidence chain. |
| `yanzi` | #109 | QA Certification; Superseded Operational State; Candidate for Closure | Scenario inventory is an intermediate state, retained for provenance only. |
| `yanzi` | #108 | QA Certification; Superseded Operational State; Candidate for Closure | Earlier QA model replaced by richer snapshot/certification lineage. |
| `yanzi` | #107 | Governance Evolution; Superseded Operational State; Candidate for Closure | Validation iteration absorbed by later convergence governance material. |
| `yanzi` | #106 | Governance Evolution; Superseded Operational State; Candidate for Closure | Artifact packaging iteration superseded by convergence-resolution outputs. |
| `yanzi` | #105 | Governance Evolution; Superseded Operational State; Candidate for Closure | Specification baseline now superseded by integrated governance lineage. |
| `yanzi` | #104 | Governance Evolution; Superseded Operational State; Candidate for Closure | Ontology baseline retained as provenance, not active convergence candidate. |
| `yanzi` | #103 | Governance Evolution; Superseded Operational State; Candidate for Closure | Reset phase was transitional and is no longer the canonical operational state. |
| `yanzi` | #102 | QA Certification; Experimental/Intermediate | Foundational QA framework; valuable provenance but not current RC convergence head. |
| `yanzi` | #101 | Experimental/Intermediate | Development-stream contracts work, outside current `qa` release lineage. |
| `yanzi` | #100 | Experimental/Intermediate | Development-stream observability work, separate from RC convergence path. |
| `yanzi` | #99 | Experimental/Intermediate | Development runtime hardening branch not in current QA promotable line. |
| `yanzi` | #98 | Experimental/Intermediate | Development continuity hardening branch not in current QA promotable line. |
| `yanzi` | #97 | Experimental/Intermediate | Development polish branch, not part of current convergence gate. |
| `yanzi` | #96 | Experimental/Intermediate | Maintenance dependency patch to `master`, operationally orthogonal to RC convergence. |
| `yanzi` | #95 | Experimental/Intermediate | Development ergonomics stream outside current release promotion lineage. |
| `yanzi` | #94 | Experimental/Intermediate | Draft development install hardening stream outside current QA convergence gate. |
| `homebrew-yanzi` | #1 | Distribution Hardening; Active Release Convergence; Candidate for Promotion | Distribution repo counterpart to canonical convergence branch; required for installer lineage alignment. |
