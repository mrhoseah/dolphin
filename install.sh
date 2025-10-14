#!/bin/bash

# Dolphin Framework Installer
echo "🐬 Installing Dolphin Framework..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first:"
    echo "   https://golang.org/dl/"
    exit 1
fi

# Clone and build
echo "📥 Downloading Dolphin Framework..."
git clone https://github.com/mrhoseah/dolphin.git /tmp/dolphin
cd /tmp/dolphin

echo "🔨 Building CLI tool..."
go build -o dolphin ./cmd/cli

echo "📦 Installing CLI tool..."
sudo mv dolphin /usr/local/bin/

echo "🧹 Cleaning up..."
rm -rf /tmp/dolphin

echo "✅ Dolphin Framework installed successfully!"
echo ""
echo "🚀 Quick start:"
echo "   dolphin new my-app"
echo "   cd my-app"
echo "   dolphin serve"
echo ""
echo "📚 Visit http://localhost:8080 for your app"
echo "📖 Visit http://localhost:8080/swagger for API docs"
