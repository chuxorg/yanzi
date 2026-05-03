# Changelog

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
