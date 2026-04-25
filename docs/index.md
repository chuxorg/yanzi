# Yanzi CLI Docs

This site hosts operational documentation for the Yanzi CLI.

## Main Documents

- [Tutorial](./tutorial.md)
- [Agent Bootstrap](./agent-bootstrap.md)
- [Release Protocol](./dev/RELEASE_PROTOCOL.md))
- [Code Documentation Workflow](./dev/CODE_DOCUMENTATION.md)
- [Branch Protection](./dev/BRANCH_PROTECTION.md)

## Getting Started

If you are using Yanzi with an AI coding agent, begin with the seed prompt in:

- [AI_AGENT_SEED.md](/Users/developer/projects/chuxorg/chux-yanzi-cli/prompts/AI_AGENT_SEED.md)

Then continue with the [Tutorial](./tutorial.md).

## API Docs

API docs are generated from Go code comments:

```bash
make docs-generate-api
```

Generated output is available under `docs/api/` and in `docs/API.md`.
