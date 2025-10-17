#!/bin/bash

# 🐬 Dolphin Framework Uninstaller
# This script completely removes Dolphin CLI from your system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to find all dolphin binaries
find_dolphin_binaries() {
    local binaries=()
    
    # Common locations where dolphin might be installed
    local locations=(
        "/usr/local/bin/dolphin"
        "/usr/bin/dolphin"
        "$HOME/bin/dolphin"
        "$HOME/.local/bin/dolphin"
        "$(go env GOPATH)/bin/dolphin"
        "$(go env GOBIN)/dolphin"
    )
    
    for location in "${locations[@]}"; do
        if [[ -f "$location" ]]; then
            binaries+=("$location")
        fi
    done
    
    # Also search in PATH
    if command_exists dolphin; then
        local dolphin_path=$(which dolphin)
        if [[ -f "$dolphin_path" ]]; then
            binaries+=("$dolphin_path")
        fi
    fi
    
    # Remove duplicates
    printf '%s\n' "${binaries[@]}" | sort -u
}

# Function to remove dolphin from PATH
remove_from_path() {
    local shell_config=""
    local dolphin_path=""
    
    # Detect shell and config file
    if [[ "$SHELL" == *"zsh"* ]]; then
        shell_config="$HOME/.zshrc"
    elif [[ "$SHELL" == *"bash"* ]]; then
        shell_config="$HOME/.bashrc"
    else
        shell_config="$HOME/.profile"
    fi
    
    # Find dolphin path in config
    if [[ -f "$shell_config" ]]; then
        dolphin_path=$(grep -o 'export PATH=.*dolphin[^:]*' "$shell_config" 2>/dev/null || true)
        
        if [[ -n "$dolphin_path" ]]; then
            print_status "Found dolphin PATH entry in $shell_config"
            print_warning "You may need to manually remove this line from $shell_config:"
            echo "  $dolphin_path"
        fi
    fi
}

# Function to clean Go module cache
clean_go_cache() {
    if command_exists go; then
        print_status "Cleaning Go module cache..."
        go clean -modcache -cache 2>/dev/null || true
        print_success "Go cache cleaned"
    else
        print_warning "Go not found, skipping cache cleanup"
    fi
}

# Function to remove dolphin projects (optional)
remove_projects() {
    local projects_dir="$HOME/dolphin-projects"
    
    if [[ -d "$projects_dir" ]]; then
        print_warning "Found dolphin projects directory: $projects_dir"
        read -p "Do you want to remove all dolphin projects? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$projects_dir"
            print_success "Dolphin projects removed"
        else
            print_status "Keeping dolphin projects"
        fi
    fi
}

# Function to remove configuration files
remove_config() {
    local config_dirs=(
        "$HOME/.dolphin"
        "$HOME/.config/dolphin"
        "/etc/dolphin"
    )
    
    for config_dir in "${config_dirs[@]}"; do
        if [[ -d "$config_dir" ]]; then
            print_status "Removing config directory: $config_dir"
            rm -rf "$config_dir"
            print_success "Config removed: $config_dir"
        fi
    done
}

# Function to remove man pages
remove_man_pages() {
    local man_pages=(
        "/usr/local/share/man/man1/dolphin.1"
        "/usr/share/man/man1/dolphin.1"
    )
    
    for man_page in "${man_pages[@]}"; do
        if [[ -f "$man_page" ]]; then
            print_status "Removing man page: $man_page"
            sudo rm -f "$man_page" 2>/dev/null || rm -f "$man_page"
            print_success "Man page removed: $man_page"
        fi
    done
}

# Function to remove completion scripts
remove_completions() {
    local completion_dirs=(
        "/usr/local/share/bash-completion/completions"
        "/usr/share/bash-completion/completions"
        "$HOME/.local/share/bash-completion/completions"
    )
    
    for completion_dir in "${completion_dirs[@]}"; do
        local completion_file="$completion_dir/dolphin"
        if [[ -f "$completion_file" ]]; then
            print_status "Removing completion script: $completion_file"
            sudo rm -f "$completion_file" 2>/dev/null || rm -f "$completion_file"
            print_success "Completion script removed: $completion_file"
        fi
    done
}

# Main uninstall function
main() {
    echo "🐬 Dolphin Framework Uninstaller"
    echo "================================="
    echo
    
    # Check if running as root
    if [[ $EUID -eq 0 ]]; then
        print_warning "Running as root. This will remove dolphin system-wide."
        read -p "Continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_status "Uninstall cancelled"
            exit 0
        fi
    fi
    
    # Find all dolphin binaries
    print_status "Searching for dolphin installations..."
    local binaries=($(find_dolphin_binaries))
    
    if [[ ${#binaries[@]} -eq 0 ]]; then
        print_warning "No dolphin installations found"
        exit 0
    fi
    
    # Show found installations
    print_status "Found dolphin installations:"
    for binary in "${binaries[@]}"; do
        echo "  - $binary"
    done
    echo
    
    # Confirm removal
    read -p "Do you want to remove all dolphin installations? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Uninstall cancelled"
        exit 0
    fi
    
    # Remove binaries
    print_status "Removing dolphin binaries..."
    for binary in "${binaries[@]}"; do
        if [[ -f "$binary" ]]; then
            print_status "Removing: $binary"
            sudo rm -f "$binary" 2>/dev/null || rm -f "$binary"
            print_success "Removed: $binary"
        fi
    done
    
    # Remove other components
    remove_config
    remove_man_pages
    remove_completions
    remove_from_path
    clean_go_cache
    
    # Optional: remove projects
    remove_projects
    
    echo
    print_success "🐬 Dolphin Framework has been completely removed!"
    echo
    print_status "What was removed:"
    echo "  ✅ Dolphin CLI binaries"
    echo "  ✅ Configuration files"
    echo "  ✅ Man pages"
    echo "  ✅ Completion scripts"
    echo "  ✅ Go module cache"
    echo
    print_status "Manual cleanup (if needed):"
    echo "  - Check your shell config file for dolphin PATH entries"
    echo "  - Remove any dolphin projects you no longer need"
    echo
    print_status "To reinstall dolphin, run:"
    echo "  go install github.com/mrhoseah/dolphin/cmd/dolphin@latest"
    echo
}

# Run main function
main "$@"
