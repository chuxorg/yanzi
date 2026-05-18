# Name

distribution-governance.seed

# Purpose

Provide deterministic guidance for distribution channel convergence validation.

# Guidance

- Validate installer, direct binary, and Homebrew lineage before promotion.
- Block promotion on version mismatch across channels.
- Record propagation timing and cache observations as operational evidence.
- Treat channel ambiguity as a release integrity risk.

# Constraints

- No hidden automation assumptions.
- No runtime orchestration behavior.
- No references to other artifacts.
