---
# Install

## Problem

Yanzi must run locally across different environments with minimal setup.

Release artifacts are published at:

- https://github.com/chuxorg/yanzi/releases

## macOS (Homebrew Tap)

Install from the Yanzi tap:

```bash
brew tap chuxorg/yanzi
brew install chuxorg/yanzi/yanzi
```

For deterministic release validation, verify the installed version:

```bash
yanzi --version
```

## macOS / Linux Install Script (Deterministic)

Install a pinned candidate tag:

```bash
curl -sSL https://raw.githubusercontent.com/chuxorg/yanzi/main/install.sh | bash -s -- --version v2.9.1-rc1
```

The script resolves the release from an explicit tag when provided and emits provenance:

- requested tag
- resolved tag
- asset URL
- installed runtime version

The installer fails if requested/resolved/installed versions do not match.

## Windows

Download `yanzi-windows-amd64.zip` from the target release tag.

Steps:

1. Extract the archive.
2. Open a terminal in the extracted directory.
3. Run `yanzi.exe --version`.

## Direct Binary

Download the appropriate release artifact from the target release tag:

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

## Deterministic RC Validation Flow

1. Select candidate tag and SHA.
2. Install with pinned tag (`install.sh --version <tag>`).
3. Verify runtime version equals candidate tag.
4. Validate Homebrew tap version path explicitly.
5. Record PASS/WARN/FAIL in QA certification report.

## Notes

- Yanzi runs locally by default.
- No services or infrastructure are required.
- HTTP mode is optional and not required for most workflows.
- Homebrew distribution must be synchronized with release governance before promotion.
