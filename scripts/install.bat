@echo off
REM Dolphin Framework Installer for Windows (Batch)
REM Usage: install.bat

echo 🐬 Installing Dolphin Framework...

REM Check if Go is installed
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Go is not installed. Please install Go first:
    echo    https://golang.org/dl/
    echo.
    echo After installing Go, restart your command prompt and run this script again.
    pause
    exit /b 1
)

echo ✅ Found Go:
go version

REM Get Go bin path
for /f "tokens=*" %%i in ('go env GOPATH') do set GOPATH=%%i
if "%GOPATH%"=="" set GOPATH=%USERPROFILE%\go
set GOBIN=%GOPATH%\bin

REM Create bin directory if it doesn't exist
if not exist "%GOBIN%" mkdir "%GOBIN%"

REM Check if already installed
if exist "%GOBIN%\dolphin.exe" (
    echo ⚠️  Dolphin is already installed at: %GOBIN%\dolphin.exe
    set /p REINSTALL="Do you want to reinstall? (y/N): "
    if /i not "%REINSTALL%"=="y" (
        echo Installation cancelled.
        pause
        exit /b 0
    )
)

REM Install Dolphin CLI
echo 📥 Installing Dolphin CLI...
set GOBIN=%GOBIN%
go install github.com/mrhoseah/dolphin/cmd/dolphin@latest

if %ERRORLEVEL% NEQ 0 (
    echo ❌ Installation failed!
    pause
    exit /b 1
)

REM Check if dolphin.exe was created
if exist "%GOBIN%\dolphin.exe" (
    echo ✅ Dolphin CLI installed successfully!
    echo    Location: %GOBIN%\dolphin.exe
) else (
    echo ❌ Installation completed but dolphin.exe not found!
    pause
    exit /b 1
)

REM Add to PATH if not already there
echo %PATH% | findstr /C:"%GOBIN%" >nul
if %ERRORLEVEL% NEQ 0 (
    echo 📝 Adding %GOBIN% to PATH...
    setx PATH "%PATH%;%GOBIN%" >nul
    echo ✅ Added to PATH. Please restart your terminal for changes to take effect.
) else (
    echo ✅ PATH already contains %GOBIN%
)

echo.
echo 🚀 Quick start:
echo    dolphin new my-app
echo    cd my-app
echo    dolphin serve
echo.
echo 📚 Visit http://localhost:8080 for your app
echo 📖 Visit http://localhost:8080/swagger for API docs
echo.
echo 💡 If 'dolphin' command is not found, restart your terminal.
echo.

pause

