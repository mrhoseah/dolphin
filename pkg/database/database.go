package database

import (
	"github.com/mrhoseah/dolphin/internal/config"
	"github.com/mrhoseah/dolphin/internal/database"
	"gorm.io/gorm"
)

// Manager represents the database manager
type Manager = database.Manager

// New creates a new database manager instance
func New(cfg *config.DatabaseConfig) (*Manager, error) {
	return database.New(cfg)
}

// GetDB returns the underlying GORM database instance from the manager
func GetDB(db *Manager) *gorm.DB {
	return db.GetDB()
}

