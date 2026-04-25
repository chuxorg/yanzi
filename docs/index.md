# Yanzi CLI Docs

This site hosts operational documentation for the Yanzi CLI.

## Main Documents

- [Agent Bootstrap](./agent-bootstrap.md)
- [Release Protocol](./dev/RELEASE_PROTOCOL.md))
- [Code Documentation Workflow](./dev/CODE_DOCUMENTATION.md)
- [Branch Protection](./dev/BRANCH_PROTECTION.md)

## API Docs

API docs are generated from Go code comments:

```bash
make docs-generate-api
```

Generated output is available under `docs/api/` and in `docs/API.md`.
