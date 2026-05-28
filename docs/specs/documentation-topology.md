# Documentation Topology

Yanzi uses two public documentation surfaces with distinct responsibilities.

## Ownership Boundary

### `yanzi.sh`

The website owns the product narrative layer:

- product positioning
- philosophy
- operational concepts
- onboarding
- case studies
- whitepapers
- architecture overviews
- Packs and Seeds narrative
- high-level runtime explanation

The website should explain why Yanzi exists and how to approach it conceptually.

### GitHub Pages

The GitHub Pages docs site owns the authoritative technical reference layer:

- installation
- quickstart
- CLI reference
- API reference
- runtime reference
- operational specs
- architecture specs
- upgrade and migration guidance

The docs site should explain how Yanzi works in detail and how to operate it precisely.

### README

The repository README is only the landing page and navigation hub:

- short description
- install entrypoint
- minimal quickstart
- links to `yanzi.sh`
- links to GitHub Pages

The README must not duplicate the full technical manual.

## Source Of Truth Rules

1. Narrative and positioning live on `yanzi.sh`.
2. Reference and operational accuracy live on GitHub Pages.
3. README links outward to both surfaces.
4. When content spans both, the website gets the simplified explanation and the docs site gets the authoritative detail.
5. Technical reference pages on GitHub Pages should not re-state narrative content except where a short pointer is required.

## Current Boundary Notes

- `yanzi.sh` should link to GitHub Pages for technical depth.
- GitHub Pages should link back to `yanzi.sh` for narrative context.
- The release docs site is the authoritative source for CLI, API, runtime, and operational specifications after CAP-002.
