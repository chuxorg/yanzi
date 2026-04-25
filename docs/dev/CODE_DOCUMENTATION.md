# Yanzi Code Documentation

Yanzi uses `gomarkdoc` to generate API documentation directly from Go code comments.

## Tooling
- Library/tool: `github.com/princjef/gomarkdoc` (pinned at `v1.1.0`)
- Output: `docs/API.md`
- TOC: generated as the `## Index` section in output

## Commands
Generate docs:

```sh
make docs
```

Verify docs are up to date:

```sh
make docs-check
```

## Commenting Standard
- Exported types, functions, constants, and variables must have doc comments.
- Keep comments concise and behavior-focused so generated docs stay readable.
