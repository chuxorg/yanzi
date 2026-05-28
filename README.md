# Yanzi

[![QA Build](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml)
[![Release](https://github.com/chuxorg/yanzi/actions/workflows/release.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/release.yml)

Yanzi is a CLI for recording intent, storing context, creating checkpoints, and rehydrating AI-assisted work.
It stays local-first by default and exposes an optional shared runtime plus a deterministic technical reference.

## Install

macOS:

```bash
brew install chuxorg/yanzi/yanzi
```

macOS or Linux install script:

```bash
curl -sSL https://raw.githubusercontent.com/chuxorg/yanzi/master/install.sh | bash
```

## Quickstart

```bash
yanzi project create demo
yanzi project use demo
yanzi capture --author "Ada" --prompt "Summarize the current task" --response "Add distribution docs and validate examples."
yanzi checkpoint create --summary "Initial project state"
yanzi rehydrate --dry-run
```

## Documentation

- Product narrative and concepts: [yanzi.sh](https://yanzi.sh/)
- Authoritative technical reference: [GitHub Pages](https://chuxorg.github.io/yanzi/)
- Documentation topology: [docs/specs/documentation-topology.md](docs/specs/documentation-topology.md)

The technical reference covers CLI, API, install, runtime, and operational specifications. The website covers the narrative and onboarding layer.

## License

MIT
