# Yanzi

[![QA Build](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/ci.yml)
[![Release](https://github.com/chuxorg/yanzi/actions/workflows/release.yml/badge.svg)](https://github.com/chuxorg/yanzi/actions/workflows/release.yml)

Yanzi is a CLI for recording intent, storing context, creating checkpoints, and rehydrating AI-assisted work.
The project has two documentation layers:

- [yanzi.sh](https://yanzi.sh/) for product narrative and onboarding
- [GitHub Pages](https://chuxorg.github.io/yanzi/) for authoritative technical reference

## Quick Start

```bash
yanzi project create my-project
yanzi project use my-project
yanzi capture --author "Ada" --prompt "hello" --response "world"
yanzi checkpoint create --summary "first boundary"
yanzi rehydrate --dry-run
```

## Install

macOS:

```bash
brew install chuxorg/yanzi/yanzi
```

macOS or Linux install script:

```bash
curl -sSL https://raw.githubusercontent.com/chuxorg/yanzi/master/install.sh | bash
```

Windows:

1. Download `yanzi-windows-amd64.zip` from the latest release.
2. Extract `yanzi.exe`.
3. Add the extract directory to `PATH`.

## Documentation

- [yanzi.sh](https://yanzi.sh/) - product narrative and onboarding
- [GitHub Pages](https://chuxorg.github.io/yanzi/) - authoritative technical reference
- [Documentation topology](docs/specs/documentation-topology.md) - ownership boundaries and link rules

Homebrew upgrades depend on the tap formula being refreshed. If `brew upgrade yanzi` does not move you to the latest release immediately, use the install script above.

## License

MIT
