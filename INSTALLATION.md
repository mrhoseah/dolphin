# 🐬 Dolphin Framework - Installation Guide

Complete installation guide for **Windows**, **macOS**, and **Linux**.

## 📋 Prerequisites

- **Go 1.19+** - [Download Go](https://golang.org/dl/)
- **Git** - Usually pre-installed, or [download here](https://git-scm.com/downloads)

## 🪟 Windows Installation

### Option 1: PowerShell (Recommended)

1. **Open PowerShell as Administrator** (Right-click → Run as Administrator)

2. **Run the installer**:
```powershell
# Download and run installer
irm https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/install.ps1 | iex
```

Or download and run manually:
```powershell
# Download the script
Invoke-WebRequest -Uri https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/install.ps1 -OutFile install.ps1

# Run it
powershell -ExecutionPolicy Bypass -File install.ps1
```

### Option 2: Command Prompt (Batch)

1. **Open Command Prompt as Administrator**

2. **Download and run**:
```batch
# Download (using PowerShell)
powershell -Command "Invoke-WebRequest -Uri https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/install.bat -OutFile install.bat"

# Run
install.bat
```

### Option 3: Manual Installation (Windows)

1. **Install Go** from [https://golang.org/dl/](https://golang.org/dl/)

2. **Open PowerShell or Command Prompt** and run:
```powershell
# Install Dolphin
go install github.com/mrhoseah/dolphin/cmd/dolphin@latest

# Add Go bin to PATH (if not already added)
$env:Path += ";$env:USERPROFILE\go\bin"
```

3. **Add to PATH permanently**:
   - Open **System Properties** → **Environment Variables**
   - Add `%USERPROFILE%\go\bin` to User PATH
   - Restart your terminal

4. **Verify installation**:
```powershell
dolphin --version
```

## 🍎 macOS Installation

### Option 1: Automated Installer (Recommended)

```bash
# Download and run installer
curl -fsSL https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/install-mac.sh | bash
```

Or download and run manually:
```bash
# Download
curl -O https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/install-mac.sh

# Make executable
chmod +x install-mac.sh

# Run
./install-mac.sh
```

### Option 2: Using Homebrew

If you have Homebrew installed:

```bash
# Install Go (if not already installed)
brew install go

# Install Dolphin
go install github.com/mrhoseah/dolphin/cmd/dolphin@latest

# Add to PATH (add to ~/.zshrc or ~/.bash_profile)
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

### Option 3: Manual Installation (macOS)

1. **Install Go**:
   - Download from [https://golang.org/dl/](https://golang.org/dl/)
   - Or use Homebrew: `brew install go`

2. **Install Dolphin**:
```bash
go install github.com/mrhoseah/dolphin/cmd/dolphin@latest
```

3. **Add to PATH**:
```bash
# For Zsh (default on macOS Catalina+)
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc

# For Bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bash_profile
source ~/.bash_profile
```

4. **Verify installation**:
```bash
dolphin --version
```

## 🐧 Linux Installation

### Option 1: Automated Installer (Recommended)

```bash
# Download and run installer
curl -fsSL https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/install.sh | bash
```

Or download and run manually:
```bash
# Download
curl -O https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/install.sh

# Make executable
chmod +x install.sh

# Run
./install.sh
```

### Option 2: Manual Installation (Linux)

1. **Install Go**:
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# Fedora/RHEL
sudo dnf install golang

# Arch Linux
sudo pacman -S go
```

2. **Install Dolphin**:
```bash
go install github.com/mrhoseah/dolphin/cmd/dolphin@latest
```

3. **Add to PATH**:
```bash
# Add to ~/.bashrc or ~/.zshrc
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc
```

4. **Verify installation**:
```bash
dolphin --version
```

## 🔍 Verify Installation

After installation, verify it works:

```bash
# Check version
dolphin --version

# See all commands
dolphin --help

# Test creating a project
dolphin new test-project
cd test-project
dolphin serve
```

## 🚀 Quick Start

Once installed, create your first project:

```bash
# Create a new project
dolphin new my-awesome-app
# Or with authentication
dolphin new my-awesome-app --auth

# Navigate to project
cd my-awesome-app

# Install dependencies
go mod tidy

# Start development server
dolphin serve

# Visit your app
# Windows: start http://localhost:8080
# macOS: open http://localhost:8080
# Linux: xdg-open http://localhost:8080
```

## 🔧 Troubleshooting

### "dolphin: command not found"

**Windows:**
- Restart your terminal/PowerShell
- Verify PATH contains `%USERPROFILE%\go\bin`
- Check: `echo %PATH%`

**macOS/Linux:**
- Restart your terminal
- Verify PATH: `echo $PATH`
- Check Go bin location: `go env GOPATH`
- Add to PATH: `export PATH="$PATH:$(go env GOPATH)/bin"`

### "Go is not installed"

1. Download Go from [https://golang.org/dl/](https://golang.org/dl/)
2. Install following the official instructions
3. Restart your terminal
4. Verify: `go version`

### "Permission denied" (Linux/macOS)

```bash
# Make script executable
chmod +x install.sh

# Or install with sudo (if needed)
sudo ./install.sh
```

### Windows PowerShell Execution Policy

If you get an execution policy error:

```powershell
# Set execution policy (as Administrator)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Or run with bypass
powershell -ExecutionPolicy Bypass -File install.ps1
```

### PATH Issues

**Windows:**
- Open **System Properties** → **Environment Variables**
- Edit **User PATH** variable
- Add: `%USERPROFILE%\go\bin`
- Restart terminal

**macOS/Linux:**
- Edit `~/.zshrc` (Zsh) or `~/.bashrc` (Bash)
- Add: `export PATH="$PATH:$(go env GOPATH)/bin"`
- Reload: `source ~/.zshrc` or `source ~/.bashrc`

## 📦 Installation Locations

### Default Locations

- **Windows**: `%USERPROFILE%\go\bin\dolphin.exe`
- **macOS**: `~/go/bin/dolphin`
- **Linux**: `~/go/bin/dolphin`

### Custom Go Path

If you have a custom `GOPATH`:
- Dolphin will be installed in `$GOPATH/bin/dolphin`
- Make sure this path is in your `PATH` environment variable

## 🔄 Updating Dolphin

To update to the latest version:

```bash
go install github.com/mrhoseah/dolphin/cmd/dolphin@latest
```

Or use the update command:

```bash
dolphin update
```

## 🗑️ Uninstalling Dolphin

### Windows

```powershell
# Remove binary
Remove-Item "$env:USERPROFILE\go\bin\dolphin.exe" -ErrorAction SilentlyContinue

# Remove from PATH (manually via System Properties)
```

### macOS/Linux

```bash
# Remove binary
rm -f $(go env GOPATH)/bin/dolphin

# Or use uninstaller
curl -fsSL https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/uninstall.sh | bash
```

## 📚 Next Steps

- [Quick Start Guide](../README.md#quick-start)
- [Documentation](https://dolphin-docs.netlify.app/)
- [Examples](../examples/)
- [CLI Commands](../README.md#-cli-commands-dolphin-cli)

## 🆘 Need Help?

- Check the [Troubleshooting](#-troubleshooting) section above
- Visit [GitHub Issues](https://github.com/mrhoseah/dolphin/issues)
- Read the [Documentation](https://dolphin-docs.netlify.app/)

---

**Happy coding! 🐬**

