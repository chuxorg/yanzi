#!/usr/bin/env sh
set -eu

REPO="chuxorg/chux-yanzi-cli"
RELEASES_API="https://api.github.com/repos/$REPO/releases"

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

ASSET_BINARY="yanzi-${OS}-${ARCH}"
ASSET_TARBALL="yanzi_${OS}_${ARCH}.tar.gz"

URL=""
ASSET_KIND=""
for page in 1 2 3 4 5 6 7 8 9 10; do
  RELEASES_JSON="$(curl -fsSL -H "Accept: application/vnd.github+json" "$RELEASES_API?per_page=100&page=$page")"
  if [ "$RELEASES_JSON" = "[]" ]; then
    break
  fi

  ASSET_URLS="$(printf '%s\n' "$RELEASES_JSON" | awk -F'"' '/"browser_download_url":/ { print $4 }')"

  URL="$(printf '%s\n' "$ASSET_URLS" | grep "/$ASSET_BINARY$" | head -n 1 || true)"
  if [ -n "$URL" ]; then
    ASSET_KIND="binary"
    break
  fi

  URL="$(printf '%s\n' "$ASSET_URLS" | grep "/$ASSET_TARBALL$" | head -n 1 || true)"
  if [ -n "$URL" ]; then
    ASSET_KIND="tarball"
    break
  fi

  # Fallback for versioned/legacy artifact naming conventions.
  URL="$(printf '%s\n' "$ASSET_URLS" | grep -E "/yanzi[-_][^/]*[-_]${OS}[-_]${ARCH}$" | head -n 1 || true)"
  if [ -n "$URL" ]; then
    ASSET_KIND="binary"
    break
  fi

  URL="$(printf '%s\n' "$ASSET_URLS" | grep -E "/yanzi[^/]*_${OS}_${ARCH}[^/]*\\.tar\\.gz$" | head -n 1 || true)"
  if [ -n "$URL" ]; then
    ASSET_KIND="tarball"
    break
  fi
done

if [ -z "$URL" ]; then
  echo "Failed to find release asset for $OS/$ARCH in $REPO." >&2
  echo "The most recent releases may not contain binary assets for $OS/$ARCH." >&2
  exit 1
fi
VERSION="$(printf '%s\n' "$URL" | sed -n 's#.*releases/download/\([^/]*\)/.*#\1#p' | head -n 1)"
if [ -z "$VERSION" ]; then
  echo "Failed to parse release version from asset URL" >&2
  exit 1
fi

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

if [ "$ASSET_KIND" = "tarball" ]; then
  ARCHIVE="$TMP_DIR/$ASSET_TARBALL"
  curl -fsSL "$URL" -o "$ARCHIVE"

  tar -xzf "$ARCHIVE" -C "$TMP_DIR"

  if [ ! -f "$TMP_DIR/yanzi" ]; then
    echo "Missing binary in archive: yanzi" >&2
    exit 1
  fi

  mv -f "$TMP_DIR/yanzi" "$INSTALL_DIR/yanzi"
  chmod +x "$INSTALL_DIR/yanzi"
else
  curl -fsSL "$URL" -o "$INSTALL_DIR/yanzi"
  chmod +x "$INSTALL_DIR/yanzi"
fi

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
    echo "Yanzi $VERSION installed successfully."
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
