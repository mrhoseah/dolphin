# Dolphin Framework Installer for Windows (PowerShell)
# Usage: powershell -ExecutionPolicy Bypass -File install.ps1

Write-Host "🐬 Installing Dolphin Framework..." -ForegroundColor Cyan

# Check if Go is installed
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Go is not installed. Please install Go first:" -ForegroundColor Red
    Write-Host "   https://golang.org/dl/" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "After installing Go, restart your terminal and run this script again." -ForegroundColor Yellow
    exit 1
}

$goVersion = go version
Write-Host "✅ Found Go: $goVersion" -ForegroundColor Green

# Determine installation directory
$goBin = go env GOPATH
if ([string]::IsNullOrWhiteSpace($goBin)) {
    $goBin = "$env:USERPROFILE\go"
}
$goBinPath = "$goBin\bin"

# Create bin directory if it doesn't exist
if (-not (Test-Path $goBinPath)) {
    New-Item -ItemType Directory -Path $goBinPath -Force | Out-Null
    Write-Host "📁 Created directory: $goBinPath" -ForegroundColor Green
}

# Check if already installed
if (Test-Path "$goBinPath\dolphin.exe") {
    Write-Host "⚠️  Dolphin is already installed at: $goBinPath\dolphin.exe" -ForegroundColor Yellow
    $response = Read-Host "Do you want to reinstall? (y/N)"
    if ($response -ne "y" -and $response -ne "Y") {
        Write-Host "Installation cancelled." -ForegroundColor Yellow
        exit 0
    }
}

# Install Dolphin CLI
Write-Host "📥 Installing Dolphin CLI..." -ForegroundColor Cyan
$env:GOBIN = $goBinPath
go install github.com/mrhoseah/dolphin/cmd/dolphin@latest

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Installation failed!" -ForegroundColor Red
    exit 1
}

# Check if dolphin.exe was created
if (Test-Path "$goBinPath\dolphin.exe") {
    Write-Host "✅ Dolphin CLI installed successfully!" -ForegroundColor Green
    Write-Host "   Location: $goBinPath\dolphin.exe" -ForegroundColor Gray
} else {
    Write-Host "❌ Installation completed but dolphin.exe not found!" -ForegroundColor Red
    exit 1
}

# Add to PATH if not already there
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$goBinPath*") {
    Write-Host "📝 Adding $goBinPath to PATH..." -ForegroundColor Cyan
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$goBinPath", "User")
    Write-Host "✅ Added to PATH. Please restart your terminal for changes to take effect." -ForegroundColor Green
} else {
    Write-Host "✅ PATH already contains $goBinPath" -ForegroundColor Green
}

# Verify installation
Write-Host ""
Write-Host "🔍 Verifying installation..." -ForegroundColor Cyan
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
if (Get-Command dolphin -ErrorAction SilentlyContinue) {
    $dolphinVersion = dolphin --version 2>&1
    Write-Host "✅ Dolphin is ready!" -ForegroundColor Green
    Write-Host "   Version: $dolphinVersion" -ForegroundColor Gray
} else {
    Write-Host "⚠️  Dolphin installed but not on PATH yet." -ForegroundColor Yellow
    Write-Host "   Please restart your terminal or run:" -ForegroundColor Yellow
    Write-Host "   `$env:Path += `";$goBinPath`"" -ForegroundColor Cyan
}

Write-Host ""
Write-Host "🚀 Quick start:" -ForegroundColor Cyan
Write-Host "   dolphin new my-app" -ForegroundColor White
Write-Host "   cd my-app" -ForegroundColor White
Write-Host "   dolphin serve" -ForegroundColor White
Write-Host ""
Write-Host "📚 Visit http://localhost:8080 for your app" -ForegroundColor Gray
Write-Host "📖 Visit http://localhost:8080/swagger for API docs" -ForegroundColor Gray
Write-Host ""
Write-Host "💡 If 'dolphin' command is not found, restart your terminal." -ForegroundColor Yellow

