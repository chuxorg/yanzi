# Install

Yanzi is local-first. It stores data on the machine where it runs.

No background service is required for local mode.

Release artifacts are published at:

- https://github.com/chuxorg/yanzi/releases

## macOS

```bash
brew install chuxorg/yanzi/yanzi
```

## Linux

Download the `.deb` release artifact, then install it:

```bash
sudo dpkg -i yanzi_*.deb
```

The Debian package installs the binary to `/usr/local/bin/yanzi`.

## Windows

1. Download `yanzi-windows-amd64.zip` from the latest release.
2. Extract `yanzi.exe`.
3. Add the extract directory to `PATH`.
4. Open a new terminal.

## Verify

Confirm that the CLI is installed and reachable:

```bash
yanzi --version
```

## Full Technical Docs

- https://chuxorg.github.io/yanzi/
