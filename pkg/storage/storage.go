package storage

import (
	"github.com/mrhoseah/dolphin/internal/storage"
)

// CreateSymlink creates a symbolic link from public/storage to storage/app/public
func CreateSymlink(publicPath, storagePath string) error {
	return storage.CreateSymlink(publicPath, storagePath)
}

// SymlinkExists checks if the storage symlink exists
func SymlinkExists(publicPath string) bool {
	return storage.SymlinkExists(publicPath)
}

// RemoveSymlink removes the storage symlink
func RemoveSymlink(publicPath string) error {
	return storage.RemoveSymlink(publicPath)
}

// EnsureSymlink ensures the storage symlink exists
func EnsureSymlink(publicPath, storagePath string) error {
	return storage.EnsureSymlink(publicPath, storagePath)
}

