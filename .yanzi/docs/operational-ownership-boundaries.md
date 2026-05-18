# Operational Ownership Boundaries

## Purpose

Define bounded mutation authority across operational roles so governance remains deterministic and provenance-safe.

## Architect

May mutate:
- system-level governance model documents
- architecture and role boundary definitions
- artifact taxonomy and dependency rules

May not mutate:
- QA certification outcomes
- release promotion decisions
- historical certification provenance

## Engineer

May mutate:
- implementation code and related docs
- engineering workflow artifacts
- engineering validation notes

May not mutate:
- QA final certification status
- release convergence decisions
- release promotion authority records

## QA Engineer

May mutate:
- QA scenarios, checklists, and certification reports
- PASS | WARN | FAIL findings and blocking issues
- QA validation evidence artifacts

May not mutate:
- release promotion decision records after handoff
- release artifact publication metadata
- lineage authority decisions owned by Release Engineer

## Release Engineer

May mutate:
- release governance docs and release workflows
- convergence validation outputs and release certification aggregation
- release promotion recommendation records
- distribution lineage verification artifacts

May not mutate:
- underlying feature implementation to force readiness
- QA raw test evidence without explicit QA ownership handoff
- prior provenance history (must remain append-only)

## Provenance Ownership

- QA owns behavioral certification provenance.
- Release Engineer owns promotion/convergence provenance.
- Shared evidence remains append-only and cross-referenceable.

## Certification Ownership

- QA owns PASS | WARN | FAIL certification disposition.
- Release Engineer consumes QA disposition to determine promotion eligibility.

## Release Ownership

- Release Engineer owns governed promotion authority and final promotable determination.

Bounded operational authority prevents agent drift and provenance corruption.
