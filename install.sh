#!/usr/bin/env sh
set -eu

REPO="${YANZI_INSTALL_REPO:-chuxorg/yanzi}"
RELEASES_API="${YANZI_INSTALL_RELEASES_API:-https://api.github.com/repos/$REPO/releases/latest}"
PRIMARY_INSTALL_DIR="${YANZI_INSTALL_PRIMARY_DIR:-/usr/local/bin}"

fail() {
  echo "Install failed: $1" >&2
  exit 1
}

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    fail "required tool '$1' is not available. Install it and rerun the installer."
  fi
}

download_file() {
  label="$1"
  url="$2"
  output="$3"

  echo "Downloading $label..."
  if ! curl -fL --silent --show-error --retry 3 --retry-delay 1 "$url" -o "$output"; then
    fail "unable to download $label from $url. Check network access and try again."
  fi
  if [ ! -f "$output" ]; then
    fail "$label download did not create $output."
  fi
  if [ ! -s "$output" ]; then
    fail "$label download was empty."
  fi
}

ensure_writable_dir() {
  target_dir="$1"

  if ! mkdir -p "$target_dir" 2>/dev/null; then
    fail "install directory $target_dir could not be created. Choose a writable location or adjust permissions."
  fi

  probe="$target_dir/.yanzi-write-test.$$"
  if ! : >"$probe" 2>/dev/null; then
    fail "install directory $target_dir is not writable. Choose a writable location or rerun with appropriate permissions."
  fi
  rm -f "$probe"
}

extract_version() {
  release_json_path="$1"
  asset_url="$2"
  version="$(awk -F'"' '/"tag_name":/ { print $4; exit }' "$release_json_path")"
  if [ -n "$version" ]; then
    printf '%s\n' "$version"
    return 0
  fi

  version="$(printf '%s\n' "$asset_url" | sed -n 's#.*releases/download/\([^/]*\)/.*#\1#p' | head -n 1)"
  if [ -z "$version" ]; then
    fail "failed to parse release version from asset URL."
  fi
  printf '%s\n' "$version"
}

ADD_PATH=false
for arg in "$@"; do
  case "$arg" in
    --add-path) ADD_PATH=true ;;
    *) ;;
  esac
done

need_cmd curl
need_cmd uname
need_cmd mktemp
need_cmd chmod
need_cmd grep
need_cmd awk
need_cmd sed
need_cmd head
need_cmd mv
need_cmd mkdir
need_cmd rm

OS_RAW="${YANZI_INSTALL_OS:-$(uname -s)}"
ARCH_RAW="${YANZI_INSTALL_ARCH:-$(uname -m)}"

case "$OS_RAW" in
  Darwin) OS="darwin" ;;
  Linux) OS="linux" ;;
  *)
    fail "unsupported OS: $OS_RAW. Supported operating systems are macOS and Linux."
    ;;
esac

case "$ARCH_RAW" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    fail "unsupported architecture: $ARCH_RAW. Supported architectures are amd64 and arm64."
    ;;
esac

ASSET_BINARY="yanzi-${OS}-${ARCH}"
ASSET_TARBALL="yanzi_${OS}_${ARCH}.tar.gz"

if [ -n "${YANZI_INSTALL_DIR:-}" ]; then
  INSTALL_DIR="$YANZI_INSTALL_DIR"
else
  INSTALL_DIR="$HOME/.local/bin"
  if [ -d "$PRIMARY_INSTALL_DIR" ] && [ -w "$PRIMARY_INSTALL_DIR" ]; then
    INSTALL_DIR="$PRIMARY_INSTALL_DIR"
  fi
fi

ensure_writable_dir "$INSTALL_DIR"

TMP_ROOT="${TMPDIR:-/tmp}"
if [ -n "${YANZI_INSTALL_TMPDIR_ROOT:-}" ]; then
  TMP_ROOT="$YANZI_INSTALL_TMPDIR_ROOT"
fi
TMP_DIR="$(mktemp -d "$TMP_ROOT/yanzi-install.XXXXXX")"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

RELEASE_JSON_PATH="$TMP_DIR/release.json"
download_file "release metadata" "$RELEASES_API" "$RELEASE_JSON_PATH"

ASSET_URLS="$(awk -F'"' '/"browser_download_url":/ { print $4 }' "$RELEASE_JSON_PATH")"
if [ -z "$ASSET_URLS" ]; then
  fail "release metadata did not include any downloadable assets."
fi

URL="$(printf '%s\n' "$ASSET_URLS" | grep "/$ASSET_BINARY$" | head -n 1 || true)"
ASSET_KIND=""
if [ -n "$URL" ]; then
  ASSET_KIND="binary"
fi

if [ -z "$URL" ]; then
  URL="$(printf '%s\n' "$ASSET_URLS" | grep "/$ASSET_TARBALL$" | head -n 1 || true)"
  if [ -n "$URL" ]; then
    ASSET_KIND="tarball"
  fi
fi

if [ -z "$URL" ]; then
  URL="$(printf '%s\n' "$ASSET_URLS" | grep -E "/yanzi[-_][^/]*[-_]${OS}[-_]${ARCH}$" | head -n 1 || true)"
  if [ -n "$URL" ]; then
    ASSET_KIND="binary"
  fi
fi

if [ -z "$URL" ]; then
  URL="$(printf '%s\n' "$ASSET_URLS" | grep -E "/yanzi[^/]*_${OS}_${ARCH}[^/]*\\.tar\\.gz$" | head -n 1 || true)"
  if [ -n "$URL" ]; then
    ASSET_KIND="tarball"
  fi
fi

if [ -z "$URL" ]; then
  fail "failed to find a release asset for $OS/$ARCH in $REPO."
fi

VERSION="$(extract_version "$RELEASE_JSON_PATH" "$URL")"

if [ "$ASSET_KIND" = "tarball" ]; then
  need_cmd tar
  ARCHIVE="$TMP_DIR/$ASSET_TARBALL"
  download_file "release archive" "$URL" "$ARCHIVE"

  if ! tar -xzf "$ARCHIVE" -C "$TMP_DIR"; then
    fail "failed to extract release archive $ARCHIVE."
  fi

  if [ ! -f "$TMP_DIR/yanzi" ]; then
    fail "release archive did not contain the yanzi binary."
  fi
  if [ ! -s "$TMP_DIR/yanzi" ]; then
    fail "release archive contained an empty yanzi binary."
  fi

  mv -f "$TMP_DIR/yanzi" "$INSTALL_DIR/yanzi"
  chmod +x "$INSTALL_DIR/yanzi"
else
  TARGET_BINARY="$TMP_DIR/yanzi"
  download_file "release binary" "$URL" "$TARGET_BINARY"
  chmod +x "$TARGET_BINARY"
  mv -f "$TARGET_BINARY" "$INSTALL_DIR/yanzi"
fi

INSTALLED_OUTPUT="$("$INSTALL_DIR/yanzi" --version 2>/dev/null || true)"
INSTALLED_VERSION="$(printf '%s\n' "$INSTALLED_OUTPUT" | awk '/^yanzi / { print; exit }')"
if [ -z "$INSTALLED_VERSION" ]; then
  fail "installed binary did not report a version successfully."
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
    echo "$INSTALLED_VERSION"
    echo "Run: yanzi --help"
  else
    echo "Yanzi was installed to $INSTALL_DIR, but that directory is not in your PATH."
    echo "$INSTALLED_VERSION"
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
