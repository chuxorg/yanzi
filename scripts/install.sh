#!/usr/bin/env sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname "$0")" && pwd)"
REPO_ROOT="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"
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

require_supported_go() {
  need_cmd go

  raw_version="$(go env GOVERSION 2>/dev/null || true)"
  if [ -z "$raw_version" ]; then
    raw_version="$(go version | awk '{print $3}')"
  fi
  raw_version="${raw_version#go}"

  major="$(printf '%s' "$raw_version" | awk -F. '{print $1}')"
  minor="$(printf '%s' "$raw_version" | awk -F. '{print $2}')"
  if [ -z "$major" ] || [ -z "$minor" ]; then
    fail "unable to determine Go version. Requires Go >= 1.25 for source installs."
  fi
  if [ "$major" -lt 1 ] || { [ "$major" -eq 1 ] && [ "$minor" -lt 25 ]; }; then
    fail "local source install requires Go >= 1.25. Found Go $raw_version."
  fi
}

maybe_update_path() {
  install_dir="$1"
  add_path="$2"

  in_path=false
  case ":$PATH:" in
    *":$install_dir:"*) in_path=true ;;
    *) in_path=false ;;
  esac

  if [ "$add_path" = true ]; then
    if [ "$in_path" = true ]; then
      echo "PATH already contains install directory."
      return
    fi

    shell_config=""
    export_line="export PATH=\"\$PATH:$install_dir\""
    case "${SHELL:-}" in
      *zsh*) shell_config="$HOME/.zshrc" ;;
      *bash*) shell_config="$HOME/.bashrc" ;;
      *fish*) shell_config="$HOME/.config/fish/config.fish" ;;
      *) shell_config="" ;;
    esac

    if [ -n "$shell_config" ]; then
      mkdir -p "$(dirname "$shell_config")"
      if [ -f "$shell_config" ] && grep -F "$export_line" "$shell_config" >/dev/null 2>&1; then
        echo "PATH already contains install directory."
      else
        printf "\n%s\n" "$export_line" >> "$shell_config"
        echo "PATH updated successfully."
      fi
      return
    fi

    echo "Yanzi was installed to $install_dir, but that directory is not in your PATH."
    echo "Add the following line to your shell config:"
    echo "$export_line"
    echo "Edit your shell profile (for example: ~/.profile)."
    return
  fi

  if [ "$in_path" = false ]; then
    echo "Yanzi was installed to $install_dir, but that directory is not in your PATH."
    echo "Add the following line to your shell config:"
    echo "export PATH=\"\$PATH:$install_dir\""
    case "${SHELL:-}" in
      *zsh*) echo "For zsh, edit: ~/.zshrc" ;;
      *bash*) echo "For bash, edit: ~/.bashrc" ;;
      *fish*) echo "For fish, edit: ~/.config/fish/config.fish" ;;
      *) echo "Edit your shell profile (for example: ~/.profile)." ;;
    esac
  fi
}

use_release_installer=false
if [ -n "${YANZI_INSTALL_RELEASES_API:-}" ] || [ -n "${YANZI_INSTALL_REPO:-}" ] || [ -n "${YANZI_INSTALL_FORCE_RELEASE:-}" ]; then
  use_release_installer=true
fi

if [ "$use_release_installer" = true ]; then
  exec sh "$SCRIPT_DIR/../install.sh" "$@"
fi

ADD_PATH=false
for arg in "$@"; do
  case "$arg" in
    --add-path) ADD_PATH=true ;;
    *) ;;
  esac
done

need_cmd uname
need_cmd mktemp
need_cmd chmod
need_cmd mv
need_cmd mkdir
need_cmd rm
need_cmd awk
need_cmd grep
require_supported_go

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
TMP_DIR="$(mktemp -d "$TMP_ROOT/yanzi-local-install.XXXXXX")"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

echo "Building Yanzi from local checkout..."
if ! (cd "$REPO_ROOT" && go build -o "$TMP_DIR/yanzi" ./cmd/yanzi); then
  fail "failed to build Yanzi from the local checkout."
fi
if [ ! -s "$TMP_DIR/yanzi" ]; then
  fail "local build did not produce a usable yanzi binary."
fi

chmod +x "$TMP_DIR/yanzi"
mv -f "$TMP_DIR/yanzi" "$INSTALL_DIR/yanzi"

INSTALLED_OUTPUT="$("$INSTALL_DIR/yanzi" --version 2>/dev/null || true)"
INSTALLED_VERSION="$(printf '%s\n' "$INSTALLED_OUTPUT" | awk '/^yanzi / { print; exit }')"
if [ -z "$INSTALLED_VERSION" ]; then
  fail "installed local binary did not report a version successfully."
fi

echo "Yanzi local checkout installed successfully."
echo "$INSTALLED_VERSION"
echo "Run: yanzi --help"
maybe_update_path "$INSTALL_DIR" "$ADD_PATH"
