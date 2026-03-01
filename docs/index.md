# Yanzi CLI Docs

This site hosts operational documentation for the Yanzi CLI.

## Main Documents

- [Agent Bootstrap](AGENT_BOOTSTRAP.md)
- [Release Protocol](RELEASE_PROTOCOL.md)
- [Code Documentation Workflow](CODE_DOCUMENTATION.md)
- [Branch Protection](BRANCH_PROTECTION.md)

## API Docs

API docs are generated from Go code comments:

```bash
make docs-generate-api
```

Generated output is available under `docs/api/` and in `docs/API.md`.
