# CAP-100 Phase 1 - Documentation Topology Alignment

## Scope

This phase aligns the public documentation topology so readers can tell where narrative content lives, where authoritative technical reference lives, and how the README, yanzi.sh, and GitHub Pages relate to each other.

Covered in this phase:
- README simplification
- documentation ownership boundaries
- yanzi.sh bridge-page navigation
- GitHub Pages technical-reference framing
- correction of critical stale topology references

## Non-Goals

- rewriting the entire documentation system
- moving all content into a single site
- changing product semantics
- changing CLI/runtime behavior
- introducing new docs infrastructure

## Documentation Ownership Rules

### yanzi.sh

The website owns:
- product narrative
- philosophy
- onboarding
- case studies
- whitepapers
- high-level runtime explanation

### GitHub Pages

GitHub Pages owns:
- CLI reference
- install docs
- quickstart
- API reference
- runtime reference
- operational specs
- architecture specs

### README

README owns:
- a brief project introduction
- install entrypoints
- minimal quickstart
- direct links to yanzi.sh and GitHub Pages

## Acceptance Criteria

- README is lightweight and does not duplicate the full docs tree
- ownership boundaries are explicit in a canonical topology document
- yanzi.sh points readers to GitHub Pages through a bridge page
- GitHub Pages is presented as the authoritative technical reference
- top-level navigation is less overloaded
- critical stale references to the old topology are corrected

## Deferred Documentation Work

- site-wide link audit across every narrative page
- deeper GitHub Pages information architecture cleanup
- any future subdomain consolidation for docs hosting
- page-by-page rewrite of older architecture notes that still describe the pre-alignment model

## Notes

This phase establishes the documentation split, but it does not finish every downstream cleanup. The main outcome is a clear and maintainable boundary between narrative and reference content.
