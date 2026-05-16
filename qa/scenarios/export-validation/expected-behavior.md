# Expected Behavior

## Snapshot Expectations
- Export files are generated at deterministic paths.
- File presence and non-empty content are required.
- Output reflects same logical dataset across formats.

## Deterministic Formatting Expectations
- Markdown contains stable section structure and headings.
- HTML contains deterministic document scaffolding and expected content markers.
- JSON is parseable, structurally consistent, and field-stable for identical inputs.

## Export Consistency Expectations
- Equivalent records appear across markdown, HTML, and JSON exports.
- Ordering remains deterministic when identical filters and order settings are used.
- Repeated export operations with identical inputs produce equivalent semantic outputs.

## Append-Only Expectations
- Certification snapshots are append-only once approved.
- Baseline changes require explicit review notes and release-context rationale.
