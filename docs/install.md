# Install

Release artifacts are published at:

- https://github.com/chuxorg/yanzi/releases

## macOS

```bash
brew install chuxorg/yanzi/yanzi
yanzi --version
```

Current Homebrew formula check on May 5, 2026 showed `stable 2.6.0`. The repository `VERSION` file is `2.7.0`, so formula and source may differ until the next release publish.

## Linux

Download the `.deb` release artifact, then install it:

```bash
sudo dpkg -i yanzi_*.deb
yanzi --version
```

The Debian package installs the binary to `/usr/local/bin/yanzi`.

## Windows

1. Download `yanzi-windows-amd64.zip` from the latest release.
2. Extract `yanzi.exe`.
3. Add the extract directory to `PATH`.
4. Open a new terminal and run `yanzi.exe --version`.

## Full Technical Docs

- https://chuxorg.github.io/yanzi/
