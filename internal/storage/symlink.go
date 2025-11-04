package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateSymlink creates a symbolic link from public/storage to storage/app/public
// This is similar to Laravel's storage:link command
func CreateSymlink(publicPath, storagePath string) error {
	// Ensure storage/app/public exists
	storagePublicPath := filepath.Join(storagePath, "app", "public")
	if err := os.MkdirAll(storagePublicPath, 0755); err != nil {
		return fmt.Errorf("failed to create storage/app/public directory: %w", err)
	}

	// Path to symlink
	linkPath := filepath.Join(publicPath, "storage")

	// Remove existing symlink or file if it exists
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	}

	// Get absolute path for storage
	absStoragePath, err := filepath.Abs(storagePublicPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get absolute path for public directory
	absPublicPath, err := filepath.Abs(publicPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Calculate relative path from public to storage
	relPath, err := filepath.Rel(absPublicPath, absStoragePath)
	if err != nil {
		// If relative path calculation fails, use absolute path
		relPath = absStoragePath
	}

	// Create symlink
	if err := os.Symlink(relPath, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// RemoveSymlink removes the storage symlink
func RemoveSymlink(publicPath string) error {
	linkPath := filepath.Join(publicPath, "storage")

	// Check if it's a symlink
	info, err := os.Lstat(linkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Already doesn't exist
		}
		return err
	}

	// Only remove if it's a symlink
	if info.Mode()&os.ModeSymlink != 0 {
		return os.Remove(linkPath)
	}

	return fmt.Errorf("%s exists but is not a symlink", linkPath)
}

// SymlinkExists checks if the storage symlink exists
func SymlinkExists(publicPath string) bool {
	linkPath := filepath.Join(publicPath, "storage")

	info, err := os.Lstat(linkPath)
	if err != nil {
		return false
	}

	// Check if it's actually a symlink
	return info.Mode()&os.ModeSymlink != 0
}

// EnsureSymlink ensures the storage symlink exists, creating it if necessary
func EnsureSymlink(publicPath, storagePath string) error {
	if SymlinkExists(publicPath) {
		return nil // Already exists
	}

	return CreateSymlink(publicPath, storagePath)
}
