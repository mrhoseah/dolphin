#!/usr/bin/env bash
set -euo pipefail

# Dolphin installer
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/install.sh | bash
#   VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/install.sh | bash

info() { echo -e "\033[1;32m[INFO]\033[0m $*"; }
warn() { echo -e "\033[1;33m[WARN]\033[0m $*"; }
err()  { echo -e "\033[1;31m[ERR ]\033[0m $*" >&2; }

VERSION="${VERSION:-main}"

detect_os() {
  case "$(uname -s)" in
    Linux*)   echo linux ;;
    Darwin*)  echo darwin ;;
    CYGWIN*|MINGW*|MSYS*) echo windows ;;
    *)        echo linux ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo amd64 ;;
    aarch64|arm64) echo arm64 ;;
    armv7l|armv6l) echo arm ;;
    *) echo amd64 ;;
  esac
}

try_download_binary() {
  OS="$(detect_os)"
  ARCH="$(detect_arch)"
  ASSET_NAME="dolphin_${OS}_${ARCH}.tar.gz"
  URL="https://github.com/mrhoseah/dolphin/releases/latest/download/${ASSET_NAME}"

  info "Checking for prebuilt binary (${OS}/${ARCH})..."
  TMP_DIR="$(mktemp -d)"
  TAR_PATH="$TMP_DIR/${ASSET_NAME}"

  if curl -fsSL -o "$TAR_PATH" "$URL" ; then
    info "Downloading prebuilt binary..."
    mkdir -p "$TMP_DIR/bin"
    tar -xzf "$TAR_PATH" -C "$TMP_DIR/bin" || { rm -rf "$TMP_DIR"; return 1; }
    if [ -f "$TMP_DIR/bin/dolphin" ] || [ -f "$TMP_DIR/bin/dolphin.exe" ]; then
      BIN_SRC="$TMP_DIR/bin/dolphin"
      [ -f "$TMP_DIR/bin/dolphin.exe" ] && BIN_SRC="$TMP_DIR/bin/dolphin.exe"
      mkdir -p "$GOBIN"
      cp "$BIN_SRC" "$GOBIN/dolphin" || { rm -rf "$TMP_DIR"; return 1; }
      rm -rf "$TMP_DIR"
      return 0
    fi
  fi

  rm -rf "$TMP_DIR"
  return 1
}

command -v go >/dev/null 2>&1 || { err "Go is required. Install Go (1.21+) and re-run."; exit 1; }

GOPATH="$(go env GOPATH)"
GOBIN="$(go env GOBIN)"
if [ -z "${GOBIN}" ]; then
  GOBIN="$GOPATH/bin"
fi

# 1) Try prebuilt binary first
if try_download_binary ; then
  info "Installed prebuilt binary to $GOBIN/dolphin"
else
  info "Installing Dolphin CLI ($VERSION) via go install..."

  # 2) Try go install; if it fails, fall back to source build
  set +e
  GOPROXY=direct GOSUMDB=off go install "github.com/mrhoseah/dolphin/cmd/dolphin@${VERSION}"
  INSTALL_STATUS=$?
  set -e

  if [ $INSTALL_STATUS -ne 0 ]; then
    warn "go install failed. Falling back to source build (no extra steps needed)."
    command -v git >/dev/null 2>&1 || { err "git is required for fallback install. Please install git and retry."; exit 1; }

    TMP_DIR="$(mktemp -d)"
    REPO_DIR="$TMP_DIR/dolphin"
    info "Cloning source into $REPO_DIR ..."
    git clone --depth=1 https://github.com/mrhoseah/dolphin "$REPO_DIR" >/dev/null 2>&1 || {
      err "Failed to clone repository for fallback install."; rm -rf "$TMP_DIR"; exit 1;
    }

    info "Building Dolphin CLI from source..."
    (
      cd "$REPO_DIR"
      go build -o "$GOBIN/dolphin" ./cmd/dolphin || { err "Source build failed."; rm -rf "$TMP_DIR"; exit 1; }
    )

    # Cleanup temp
    rm -rf "$TMP_DIR"
  fi
fi

BIN_SRC="$GOBIN/dolphin"
if [ ! -f "$BIN_SRC" ]; then
  err "dolphin binary not found at $BIN_SRC after install"
  exit 1
fi

# Try to place into /usr/local/bin if possible for global availability
TARGET="/usr/local/bin/dolphin"
if [ -w "/usr/local/bin" ]; then
  info "Copying dolphin to $TARGET"
  cp "$BIN_SRC" "$TARGET"
else
  if sudo -n true >/dev/null 2>&1; then
    info "Copying dolphin to $TARGET (sudo)"
    sudo cp "$BIN_SRC" "$TARGET"
  else
    warn "/usr/local/bin not writable. Keeping dolphin at $BIN_SRC"
  fi
fi

# Ensure PATH contains GOBIN or /usr/local/bin
if ! command -v dolphin >/dev/null 2>&1; then
  warn "dolphin not found on PATH yet. Updating shell profile."
  PROFILE="$HOME/.bashrc"
  if [ -n "${ZSH_VERSION:-}" ]; then PROFILE="$HOME/.zshrc"; fi
  {
    echo "# Added by Dolphin installer"
    echo "export PATH=\"$GOBIN:\$PATH\""
  } >> "$PROFILE"
  # shellcheck disable=SC1090
  . "$PROFILE" || true
fi

if command -v dolphin >/dev/null 2>&1; then
  info "Installation complete: $(command -v dolphin)"
  info "Run: dolphin --help"
  echo
  info "🐬 Dolphin Framework installed successfully!"
  info "Quick start:"
  info "  dolphin new my-app --auth    # Create new project with auth"
  info "  dolphin serve                # Start development server"
  info "  dolphin --help               # See all commands"
  echo
  info "To uninstall dolphin, run:"
  info "  curl -fsSL https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/uninstall.sh | bash"
else
  warn "Installation finished, but dolphin not on PATH. Add this to your shell profile:"
  echo "export PATH=\"$GOBIN:\$PATH\""
fi


