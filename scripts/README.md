# 🐬 Dolphin Framework Scripts

This directory contains installation and maintenance scripts for the Dolphin Framework.

## 📦 Installation Script

### `install.sh`
Automated installer for the Dolphin CLI.

**Usage:**
```bash
# Install latest version
curl -fsSL https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/install.sh | bash

# Install specific version
VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/install.sh | bash
```

**Features:**
- ✅ Automatic Go detection
- ✅ Smart PATH configuration
- ✅ Global binary installation (if possible)
- ✅ Shell profile updates
- ✅ Installation verification

## 🗑️ Uninstallation Script

### `uninstall.sh`
Complete removal script for the Dolphin CLI.

**Usage:**
```bash
# Automated uninstaller
curl -fsSL https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/uninstall.sh | bash
```

**Features:**
- ✅ Finds all dolphin installations
- ✅ Removes binaries from all locations
- ✅ Cleans Go module cache
- ✅ Removes configuration files
- ✅ Removes man pages and completions
- ✅ Optional project cleanup
- ✅ Safe operation with confirmations

## 🔧 Manual Installation/Uninstallation

### Install Dolphin CLI
```bash
# Using Go
go install github.com/mrhoseah/dolphin/cmd/dolphin@latest

# Or specific version
go install github.com/mrhoseah/dolphin/cmd/dolphin@v0.1.0
```

### Uninstall Dolphin CLI
```bash
# Remove binaries
sudo rm -f /usr/local/bin/dolphin
rm -f ~/bin/dolphin
rm -f $(go env GOPATH)/bin/dolphin

# Clean Go cache
go clean -modcache -cache

# Remove config (optional)
rm -rf ~/.dolphin
rm -rf ~/.config/dolphin
```

## 🚀 Quick Start After Installation

```bash
# Create a new project
dolphin new my-awesome-app --auth

# Navigate to project
cd my-awesome-app

# Install dependencies
go mod tidy

# Start development server
dolphin serve
```

## 📋 Requirements

- **Go 1.19+** - Required for building and running Dolphin
- **Git** - Required for cloning and updating
- **Bash** - Required for running installation scripts

## 🔍 Troubleshooting

### Installation Issues
```bash
# Check Go installation
go version

# Check PATH
echo $PATH

# Verify installation
which dolphin
dolphin --version
```

### Uninstallation Issues
```bash
# Find all dolphin installations
find /usr -name "dolphin" 2>/dev/null
find ~ -name "dolphin" 2>/dev/null

# Check Go bin directory
ls -la $(go env GOPATH)/bin/dolphin
```

## 📞 Support

- **GitHub Issues**: [Report problems](https://github.com/mrhoseah/dolphin/issues)
- **Documentation**: [Read the docs](https://github.com/mrhoseah/dolphin#readme)
- **Community**: Join our Discord server

## 📄 License

These scripts are part of the Dolphin Framework and are licensed under the MIT License.
