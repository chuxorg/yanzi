# CAP-100 Phase 1 — Documentation Topology Alignment

## Scope

CAP-100 Phase 1 establishes a clear, maintainable documentation architecture for Yanzi by defining explicit ownership boundaries across its three documentation surfaces: `yanzi.sh`, GitHub Pages, and the repository README.

This phase resolves inconsistencies introduced during rapid feature development (CAP-001, CAP-002) where documentation authority became unclear, navigation became overloaded, and technical reference content was mixed with product narrative.

## Documentation Ownership Rules

### `yanzi.sh`

Owns the product narrative layer:

- product positioning and philosophy
- operational concepts and onboarding
- case studies and whitepapers
- architecture overviews
- Packs and Seeds narrative
- high-level runtime explanation

The website explains *why* Yanzi exists and how to approach it conceptually.

### GitHub Pages (`chuxorg.github.io/yanzi`)

Owns the authoritative technical reference layer:

- installation
- quickstart
- CLI reference
- API reference
- runtime reference
- operational specs and architecture specs
- upgrade and migration guidance
- documentation topology

The docs site explains *how* Yanzi works in detail and how to operate it precisely.

### README

The repository README is only the landing page and navigation hub:

- short description
- install entrypoint
- minimal quickstart
- links to `yanzi.sh`
- links to GitHub Pages

The README must not become the full technical manual or a giant duplicated TOC.

## Non-Goals

This phase does not include:

- rewriting the content of existing technical docs
- migrating `yanzi.sh` content (separate repository)
- redesigning the HTML export format
- generating new API reference content
- adding new technical documentation beyond topology alignment

## Acceptance Criteria

- Documentation ownership boundaries are explicit in `docs/specs/documentation-topology.md`.
- README simplified to landing page and navigation hub.
- GitHub Pages (`docs/`) positioned as authoritative technical reference with clear header pointer to `yanzi.sh`.
- All modified GitHub Pages pages include a canonical pointer to `yanzi.sh` for narrative context.
- MkDocs navigation is restructured into sections (Getting Started, Reference, Specifications, Operational Docs) to reduce top-level navigation overload.
- `docs/specs/documentation-topology.md` is added to MkDocs navigation under Specifications.
- `go build ./cmd/yanzi` passes (no Go changes in this phase — confirming no regressions).
- MkDocs build succeeds (`mkdocs build`).

## Implementation Summary

### Files Changed

| File | Change |
|---|---|
| `README.md` | Simplified to landing page: short description, install, minimal quickstart, documentation links |
| `docs/index.md` | Updated header to "Yanzi Technical Reference", added `yanzi.sh` pointer, reorganized Start Here section |
| `docs/install.md` | Added canonical pointer to `yanzi.sh` for narrative context |
| `docs/quickstart.md` | Added canonical pointer to `yanzi.sh` for narrative context |
| `docs/api/index.md` | Added canonical pointer to `yanzi.sh` for narrative context |
| `docs/specs/documentation-topology.md` | Created: defines ownership boundaries and source-of-truth rules |
| `mkdocs.yml` | Restructured navigation into logical sections; added documentation topology entry |

### Audit Findings

**README before CAP-100:**
- Mixed quickstart, UI overview, Docs section with GitHub Pages links only (no yanzi.sh link)
- Used `main` branch in install script URL (incorrect — branch is `master`)
- Contained extended prose that duplicated the website

**GitHub Pages before CAP-100:**
- `docs/index.md` did not identify itself as a technical reference
- Technical pages contained no canonical pointer to yanzi.sh
- Navigation was a flat list at top-level causing horizontal overflow on narrow displays
- `docs/specs/documentation-topology.md` did not exist

**Cross-link behavior before CAP-100:**
- README linked only to GitHub Pages, not to yanzi.sh
- Technical pages did not link back to yanzi.sh for narrative context
- No canonical definition of which surface owns which content

### Navigation Changes

MkDocs navigation restructured from a flat list into labeled sections:

- **Getting Started** — Install, Quickstart, Problem, How It Works, Use Yanzi
- **Reference** — CLI, API, Rehydrate, UI, AI Seed
- **Specifications** — Documentation Topology, Architecture, Release Lineage Governance, Agent Bootstrap
- **Operational Docs** — Release Protocol, Code Documentation, Branch Protection, Use Cases
- **Roles** — Release Steward
- **Generated API** — CLI Package, Internal Packages, Combined API

## Deferred Documentation Work

The following items are identified but deferred to CAP-100 Phase 2 or later:

- **`yanzi.sh` cross-link alignment** — updating the website to consistently link into GitHub Pages for technical depth requires access to the `chuxorg/yanzi.sh` repository.
- **Stale technical content audit** — pages like `architecture.md`, `how-it-works.md`, `use-yanzi.md`, and use-case pages have not been reviewed for accuracy against v2.10.0 behavior.
- **CLI reference completeness** — `docs/cli.md` has not been validated against the full 22-command surface added through CAP-002.
- **API reference alignment** — `docs/api/index.md` describes the HTTP API at a high level; endpoint-level reference for the CAP-002 REST API is not yet documented.
- **Generated API docs** — `docs/api/cmd.md`, `docs/api/internal.md`, and `docs/API.md` are generated references whose regeneration cadence is not yet defined.
- **Navigation deep-link validation** — sidebar links to `dev/` subdocs and generated API pages have not been validated end-to-end in a deployed environment.

## Recommendation for CAP-100 Phase 2

Phase 2 should focus on:

1. Auditing and updating stale technical content against v2.10.0 (CLI reference, architecture, API).
2. Aligning `yanzi.sh` to link into GitHub Pages for technical depth (requires `chuxorg/yanzi.sh` access).
3. Establishing a docs-generation cadence for the generated API reference.
4. Validating all navigation links in a deployed GitHub Pages environment.
