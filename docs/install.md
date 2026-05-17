# Install

## Problem

Yanzi must run locally across different environments with minimal setup.

Release artifacts are published at:

- https://github.com/chuxorg/yanzi/releases

## macOS (Homebrew - Recommended)

Install using Homebrew:

```bash
brew install chuxorg/yanzi/yanzi
```

## macOS / Linux Install Script

Install the latest release directly from GitHub:

```bash
curl -sSL https://raw.githubusercontent.com/chuxorg/yanzi/main/install.sh | bash
```

For deterministic certification or RC installs, pin an explicit tag:

```bash
curl -sSL https://raw.githubusercontent.com/chuxorg/yanzi/main/install.sh | bash -s -- --version vX.Y.Z-rcN
```

The script emits `requested_tag`, `resolved_tag`, and `asset_url` so operators can verify lineage continuity.

## Windows

Download `yanzi-windows-amd64.zip` from the latest release.

Steps:

1. Extract the archive.
2. Open a terminal in the extracted directory.
3. Run `yanzi.exe --version`.

## Direct Binary

Download the appropriate release artifact from the latest release:

- `yanzi-darwin-arm64`
- `yanzi-darwin-amd64`
- `yanzi-linux-amd64`
- `yanzi-windows-amd64.zip` containing `yanzi.exe`

For macOS and Linux:

```bash
chmod +x yanzi
./yanzi --version
```

For Windows, extract the zip and run:

```powershell
.\yanzi.exe --version
```

## Verify Installation

Run:

```bash
yanzi --version
```

Expected:

```text
yanzi vX.Y.Z
```

## Notes

- Yanzi runs locally by default.
- No services or infrastructure are required.
- HTTP mode is optional and not required for most workflows.
- Homebrew upgrades depend on tap formula refresh timing.
- Governed release validation requires verifying Homebrew lineage and installer lineage against the same certified tag before promotion.
