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

command -v go >/dev/null 2>&1 || { err "Go is required. Install Go (1.21+) and re-run."; exit 1; }

GOPATH="$(go env GOPATH)"
GOBIN="$(go env GOBIN)"
if [ -z "${GOBIN}" ]; then
  GOBIN="$GOPATH/bin"
fi

info "Installing Dolphin CLI ($VERSION) via go install..."
GOPROXY=direct GOSUMDB=off go install "github.com/mrhoseah/dolphin/cmd/dolphin@${VERSION}"

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


