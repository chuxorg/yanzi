# QA Execution Guide

## QA Branch Workflow
- Start QA work from `qa/*` branches derived from completed feature scope.
- QA branches may add tests, fixtures, snapshots, and docs validation artifacts.
- QA branches must avoid production redesign and unrelated feature work.
- Open PRs from local QA working branches into `qa` for staged validation.

## Local QA Commands
Minimum validation:
```bash
go test ./...
go vet ./...
go build ./cmd/yanzi
```

Run E2E suite only:
```bash
go test ./test/e2e -v
```

Run docs/link validation:
```bash
go run ./test/qa/docvalidate --root .
```

## Full Regression Validation
```bash
go test ./... -count=1
go vet ./...
go build ./cmd/yanzi
go test ./test/e2e -v -count=1
go run ./test/qa/docvalidate --root .
```

Expected outcomes:
- deterministic pass/fail by exit code
- reproducible snapshots for markdown/html export contracts
- explicit QA logs for docs/link validation

## Failure Categories
- `FAIL/showstopper`: regression, broken contract, broken command behavior, failed required validation.
- `WARN/deferred`: non-blocking issue with documented risk and follow-up tracking.
- `enhancement`: quality improvement outside blocking scope.

## Snapshot Policy
- Snapshot files live in `test/snapshots/`.
- Update snapshots intentionally only when output contract changes are reviewed.
- To update snapshots:
```bash
UPDATE_SNAPSHOTS=1 go test ./test/e2e -run TestYanziProjectCaptureCheckpointExportAndRehydrate -v
```
- Re-run without `UPDATE_SNAPSHOTS` to ensure stable comparison.

## Guidance For AI QA Agents
- Always isolate `HOME` and workspace during CLI E2E execution.
- Treat export outputs as contracts; compare normalized snapshots to reduce time/hash noise.
- Never modify production behavior solely to satisfy test assertions.
- Record exact command, exit code, and contract evidence for each failure.
