# Workflow

## Phased Execution Model

Yanzi is designed for work that proceeds in discrete phases. Each phase has a clear start, a set of captures, and a completion gate.

**Phase start:**
1. Create or select a project: `yanzi project use <name>`
2. Read SYSTEM_RULES.md.
3. Review recent history: `yanzi list --limit 20`
4. Rehydrate if needed: `yanzi rehydrate`

**During a phase:**
- Record decisions as captures.
- Create checkpoints at meaningful boundaries.
- Do not commit directly to `development` or `master`.

**Phase completion:**
- All tasks are captured.
- A final checkpoint is created.
- The log is exported: `yanzi export --format markdown`
- A PR is opened to `development`.
- The human reviews and merges.

---

## Branch and PR Flow

```
master
  └── development
        └── feature/<short-description>
```

- All new work starts from `development`.
- Feature branches are named `feature/<short-description>`.
- PRs target `development`, not `master`.
- `master` is updated only via release PRs from `development`.
- Tags are created only after merge.

---

## Capturing Decisions

Captures form the audit trail. Record decisions, not noise. A good capture has:

- A clear prompt (the question or task).
- A clear response (the decision or outcome).
- An accurate author field.
- Metadata when relevant (`--meta phase=2 --meta component=auth`).

```sh
yanzi capture \
  --author "Ada" \
  --prompt "Should we use ULIDs or UUIDs for intent IDs?" \
  --response "ULIDs — they are sortable by time, which simplifies chain queries." \
  --meta component=library
```

---

## Checkpoints as Rehydration Anchors

A checkpoint marks a stable, known-good state. When context is lost, `yanzi rehydrate` walks back to the most recent checkpoint and replays captures forward from there.

Create checkpoints:
- Before large refactors.
- After completing a feature.
- At phase boundaries.

```sh
yanzi checkpoint create --summary "Auth module complete, all tests passing"
```

---

## Export and External Systems

At the end of a phase, export and archive:

```sh
yanzi export --format json
yanzi export --format html
```

JSON output can be piped to external systems for analysis, archiving, or compliance:

```sh
yanzi export --format json > audit/phase-2.json
```
