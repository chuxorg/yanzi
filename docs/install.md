# Install

## Problem

Yanzi must run locally across different environments with minimal setup.

Release artifacts are published at:

- https://github.com/chuxorg/yanzi/releases

Requires Go >= 1.25 for source builds. Ubuntu LTS images often ship an older Go toolchain, so prefer the release installer or install a newer Go from go.dev before building from source.

## macOS (Homebrew - Recommended)

Install using Homebrew:

```bash
brew install chuxorg/yanzi/yanzi
```

## macOS / Linux Install Script

Install the latest release directly from the canonical repository `chuxorg/yanzi`:

```bash
curl -fL -o /tmp/yanzi-install.sh https://raw.githubusercontent.com/chuxorg/yanzi/main/install.sh
test -s /tmp/yanzi-install.sh
sh /tmp/yanzi-install.sh
```

The installer now:

- checks for required shell tools before downloading anything
- validates supported OS and architecture explicitly
- downloads into temp files before execution
- fails if metadata or release assets are missing or empty
- installs `yanzi` into `/usr/local/bin` when writable or `~/.local/bin` otherwise

If you already cloned the repository, you can run the same installer locally:

```bash
./scripts/install.sh
```

The repo-local wrapper installs the current checkout instead of the latest published release. That keeps local branch testing deterministic, but it requires Go 1.25 or newer.

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
ASSET=yanzi-darwin-arm64 # replace with the downloaded asset name
mv "$ASSET" yanzi
chmod +x yanzi
./yanzi --version
```

For Windows, extract the zip and run:

```powershell
.\yanzi.exe --version
```

## Source Build

If you need a local build instead of a release artifact:

```bash
go version
go install github.com/chuxorg/yanzi/cmd/yanzi@latest
```

`go version` must report Go 1.25 or newer.

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
- Homebrew upgrades depend on the tap formula being refreshed. If `brew upgrade yanzi` does not install the latest release yet, use the install script instead.
