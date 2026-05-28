# Documentation Topology

Yanzi has two public documentation layers plus a lightweight README landing page.

## Ownership Boundaries

### yanzi.sh

The website is the product and narrative layer.

It should contain:
- product narrative
- philosophy and operating principles
- onboarding
- case studies
- whitepapers
- architecture overviews
- Packs and Seeds narrative
- high-level runtime explanation

It should not duplicate the authoritative technical reference.

### GitHub Pages

GitHub Pages is the authoritative technical reference.

It should contain:
- CLI reference
- install docs
- quickstart
- API reference
- runtime reference
- operational specs
- architecture specs
- upgrade and migration docs

### README

README is the lightweight landing page and navigation hub.

It should contain:
- a short project description
- a minimal quickstart
- install entrypoints
- direct links to yanzi.sh and GitHub Pages
- a link to this topology document

It should not become a full technical manual or duplicate the full docs tree.

## Source Of Truth Rules

1. If a document explains why Yanzi exists or how it should be understood, it belongs on yanzi.sh.
2. If a document defines commands, flags, endpoints, schemas, or operational behavior, it belongs on GitHub Pages.
3. If the README needs to mention a deeper topic, it should link to the canonical location instead of reproducing it.
4. When a topic exists in both places, yanzi.sh gets the concise narrative version and GitHub Pages gets the exact reference version.

## Link Rules

- yanzi.sh should link to GitHub Pages for technical depth.
- GitHub Pages should link back to yanzi.sh for product context where useful.
- README should link to both sites and avoid duplicating either one.

## Deferred Work

- A fuller information architecture pass for GitHub Pages
- A matching website-side cross-link audit for all narrative pages
- Any future docs site consolidation if the hosting model changes
