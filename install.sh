#!/usr/bin/env sh
set -eu

BASE_URL="https://github.com/chuxorg/yanzi/releases/latest/download"

ADD_PATH=false
for arg in "$@"; do
  case "$arg" in
    --add-path) ADD_PATH=true ;;
    *) ;;
  esac
done

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

in_path=false
case ":$PATH:" in
  *":$INSTALL_DIR:"*) in_path=true ;;
  *) in_path=false ;;
esac

if [ "$ADD_PATH" = true ]; then
  if [ "$in_path" = true ]; then
    echo "PATH already contains install directory."
  else
    SHELL_CONFIG=""
    EXPORT_LINE="export PATH=\"\$PATH:$INSTALL_DIR\""
    case "${SHELL:-}" in
      *zsh*) SHELL_CONFIG="$HOME/.zshrc" ;;
      *bash*) SHELL_CONFIG="$HOME/.bashrc" ;;
      *fish*) SHELL_CONFIG="$HOME/.config/fish/config.fish" ;;
      *) SHELL_CONFIG="" ;;
    esac

    if [ -n "$SHELL_CONFIG" ]; then
      mkdir -p "$(dirname "$SHELL_CONFIG")"
      if [ -f "$SHELL_CONFIG" ] && grep -F "$EXPORT_LINE" "$SHELL_CONFIG" >/dev/null 2>&1; then
        echo "PATH already contains install directory."
      else
        printf "\n%s\n" "$EXPORT_LINE" >> "$SHELL_CONFIG"
        echo "PATH updated successfully."
      fi
    else
      echo "Yanzi was installed to $INSTALL_DIR, but that directory is not in your PATH."
      echo "Add the following line to your shell config:"
      echo "$EXPORT_LINE"
      echo "Edit your shell profile (for example: ~/.profile)."
    fi
  fi
else
  if [ "$in_path" = true ]; then
    echo "Yanzi installed successfully."
    echo "Run: yanzi --help"
  else
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
      *fish*)
        echo "For fish, edit: ~/.config/fish/config.fish"
        ;;
      *)
        echo "Edit your shell profile (for example: ~/.profile)."
        ;;
    esac
  fi
fi
