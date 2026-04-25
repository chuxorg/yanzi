# Issues

## Export fails when no active project is set after project creation

- Repro:
  - `./yanzi project create "alpha"`
  - `./yanzi capture --author "Ada" --prompt "Hello 2" --response "World"`
  - `./yanzi export --format html`
- Observed:
  - Export fails with `no active project set`
- Expected:
  - Either `project create` should make the new project active automatically, or the CLI should make the required `project use "alpha"` step more obvious before capture/export workflows.
- Notes:
  - `capture` succeeded even though no active project was set, which makes the later export failure easy to miss and confusing in normal usage.
