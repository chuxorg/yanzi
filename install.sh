#!/usr/bin/env sh
set -eu

BASE_URL="https://github.com/chuxorg/yanzi/releases/latest/download"

OS_RAW="$(uname -s)"
ARCH_RAW="$(uname -m)"

case "$OS_RAW" in
  Darwin) OS="darwin" ;;
  Linux) OS="linux" ;;
  *)
    echo "Unsupported OS: $OS_RAW" >&2
    exit 1
    ;;
esac

case "$ARCH_RAW" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH_RAW" >&2
    exit 1
    ;;
esac

ARTIFACT="yanzi_${OS}_${ARCH}.tar.gz"
URL="$BASE_URL/$ARTIFACT"

if [ -w /usr/local/bin ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

ARCHIVE="$TMP_DIR/$ARTIFACT"

curl -fsSL "$URL" -o "$ARCHIVE"

tar -xzf "$ARCHIVE" -C "$TMP_DIR"

for bin in yanzi yanzi-emitter; do
  if [ ! -f "$TMP_DIR/$bin" ]; then
    echo "Missing binary in archive: $bin" >&2
    exit 1
  fi
  mv -f "$TMP_DIR/$bin" "$INSTALL_DIR/$bin"
  chmod +x "$INSTALL_DIR/$bin"
done

case ":$PATH:" in
  *":$INSTALL_DIR:"*)
    echo "Yanzi installed successfully."
    echo "Run: yanzi --help"
    ;;
  *)
    echo "Yanzi was installed to $INSTALL_DIR, but that directory is not in your PATH."
    echo "Add the following line to your shell config:"
    echo "export PATH=\"\$PATH:$INSTALL_DIR\""
    case "${SHELL:-}" in
      *zsh*)
        echo "For zsh, edit: ~/.zshrc"
        ;;
      *bash*)
        echo "For bash, edit: ~/.bashrc"
        ;;
      *)
        echo "Edit your shell profile (for example: ~/.profile)."
        ;;
    esac
    ;;
esac
