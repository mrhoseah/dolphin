#!/bin/bash

# Dolphin Framework Installer for macOS
# Usage: bash install-mac.sh

set -e

echo "🐬 Installing Dolphin Framework on macOS..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first:"
    echo "   Option 1: Download from https://golang.org/dl/"
    echo "   Option 2: Install via Homebrew: brew install go"
    echo ""
    exit 1
fi

GO_VERSION=$(go version)
echo "✅ Found Go: $GO_VERSION"

# Get Go environment
GOPATH=$(go env GOPATH)
if [ -z "$GOPATH" ]; then
    GOPATH="$HOME/go"
fi

GOBIN="$GOPATH/bin"

# Create bin directory if it doesn't exist
mkdir -p "$GOBIN"
echo "📁 Using directory: $GOBIN"

# Check if already installed
if command -v dolphin &> /dev/null; then
    DOLPHIN_PATH=$(command -v dolphin)
    echo "⚠️  Dolphin is already installed at: $DOLPHIN_PATH"
    read -p "Do you want to reinstall? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Installation cancelled."
        exit 0
    fi
fi

# Install Dolphin CLI
echo "📥 Installing Dolphin CLI..."
export GOBIN="$GOBIN"
go install github.com/mrhoseah/dolphin/cmd/dolphin@latest

if [ $? -ne 0 ]; then
    echo "❌ Installation failed!"
    exit 1
fi

# Verify installation
if [ -f "$GOBIN/dolphin" ]; then
    echo "✅ Dolphin CLI installed successfully!"
    echo "   Location: $GOBIN/dolphin"
    
    # Make it executable
    chmod +x "$GOBIN/dolphin"
else
    echo "❌ Installation completed but dolphin binary not found!"
    exit 1
fi

# Add to PATH if not already there
SHELL_PROFILE=""
if [ -n "$ZSH_VERSION" ]; then
    SHELL_PROFILE="$HOME/.zshrc"
elif [ -n "$BASH_VERSION" ]; then
    SHELL_PROFILE="$HOME/.bash_profile"
    if [ ! -f "$SHELL_PROFILE" ]; then
        SHELL_PROFILE="$HOME/.bashrc"
    fi
fi

if [ -n "$SHELL_PROFILE" ]; then
    if ! grep -q "$GOBIN" "$SHELL_PROFILE" 2>/dev/null; then
        echo "📝 Adding $GOBIN to PATH in $SHELL_PROFILE..."
        {
            echo ""
            echo "# Added by Dolphin installer"
            echo "export PATH=\"\$PATH:$GOBIN\""
        } >> "$SHELL_PROFILE"
        echo "✅ Added to PATH. Please restart your terminal or run: source $SHELL_PROFILE"
    else
        echo "✅ PATH already contains $GOBIN"
    fi
fi

# Verify installation
echo ""
echo "🔍 Verifying installation..."
export PATH="$PATH:$GOBIN"
if command -v dolphin &> /dev/null; then
    DOLPHIN_VERSION=$(dolphin --version 2>&1 || echo "installed")
    echo "✅ Dolphin is ready!"
    echo "   Version: $DOLPHIN_VERSION"
else
    echo "⚠️  Dolphin installed but not on PATH yet."
    echo "   Please restart your terminal or run: export PATH=\"\$PATH:$GOBIN\""
fi

echo ""
echo "🚀 Quick start:"
echo "   dolphin new my-app"
echo "   cd my-app"
echo "   dolphin serve"
echo ""
echo "📚 Visit http://localhost:8080 for your app"
echo "📖 Visit http://localhost:8080/swagger for API docs"
echo ""
echo "💡 If 'dolphin' command is not found, restart your terminal."

