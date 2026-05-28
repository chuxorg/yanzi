# Changelog

## [2.10.0] - 2026-05-27

### Added

- CAP-002 Operational API and runtime foundation
  - project and checkpoint endpoints
  - artifact capture and read endpoints
  - verification and chain endpoints
  - export endpoints for markdown, JSON, and HTML
  - deterministic rehydration endpoint
  - optional shared runtime via `yanzi serve`

### Notes

- CLI primacy remains intact
- local-first behavior remains the default operating model
- no breaking CLI behavior was introduced
- the release preserves deterministic operational semantics

## [2.9.0] - 2026-05-09

### Added

- Interactive HTML UI for export
  - Collapsible artifacts
  - Prompt previews
  - Search
  - Copy actions
  - Sticky header
  - Checkpoint visualization

### Improved

- Timeline readability
- Export usability

### Notes

- No breaking changes
- No new dependencies
- Fully offline-compatible

## [2.8.0]

- Added Packs (portable context bundles)
- Added automatic project binding per directory
- Added retrieval filtering (`--type`, `--meta`, `--fields`, `--order`, `--limit`)
- Improved export behavior (agent-ready by default)
- Introduced Use Cases documentation
- Adopted Material for MkDocs with dark mode
- Completed documentation overhaul

## [2.7.0]

- Added Debian (.deb) distribution
- Added Windows (.zip) distribution
- Rewrote documentation for clarity and accuracy
- Introduced website-ready documentation structure

## [2.6.1]

- Fixed Homebrew formula to correctly install v2.6.0 binaries
- Improved installation reliability
- Added Debian `.deb` and Windows `.zip` release artifacts

## [2.6.0]

### Commands
- Added `bootstrap` command for ordered context loading from `.yanzi/bootstrap.yaml`
- Added message channel wrappers: `message send`, `message list`, `message pull`
- Added `export --format claude-context` for direct prompt injection workflows
- Added `rehydrate --dry-run` preview mode
- Added `types list --json` and type alias support including `governance -> process_rule`

### Documentation
- README updated with agent usability release highlights

## [2.5.0]

### Commands
- `rules` command: `rules add`, `rules list`, `rules export` — wraps rule/policy files as structured context captures
- `delete` command — soft-delete (tombstone) intents and artifacts by ID with optional `--cascade` and `--force` flags
- `restore` command — remove tombstone from a previously deleted record
- Profile support: `--profile <name>` flag on `capture`, `list`, `rules add`, `rules list`, `rules export`, and `export`
- Composed rules export: `rules export --format markdown --compose` — concatenates all matching rule captures in priority order
- `--include-deleted` flag on `list`, `rules list`, and `export` to include tombstoned records

### Documentation
- Documentation site added under `/docs` with GitHub Pages deployment
- `docs/index.md` — overview and core concepts
- `docs/quickstart.md` — install and first capture
- `docs/cli.md` — full command reference
- `docs/rules.md` — SYSTEM_RULES governance model
- `docs/ai-seed.md` — AI agent seed prompt
- `docs/workflow.md` — phased execution and branch model
- AI agent seed prompt (`prompts/AI_AGENT_SEED.md`) expanded with composed export guidance and full workflow steps
- README updated with GitHub Pages link and quickstart snippet

## [2.0.0]

- Major version bump
- Cross-platform support (Windows, macOS, Linux)
- Local-first SQLite storage
- Captures, checkpoints, intent and context artifacts
- Export formats: markdown, JSON, HTML
- HTML export with timeline viewer and search
- Project management commands
- Rehydrate command for context reconstruction
- Chain and verify commands for hash integrity
