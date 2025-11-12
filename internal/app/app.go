package app

import (
	"database/sql"

	"github.com/mrhoseah/dolphin/internal/config"
	"github.com/mrhoseah/dolphin/internal/database"
	"go.uber.org/zap"
)

// App represents the main application instance
type App struct {
	config *config.Config
	logger *zap.Logger
	db     *database.Manager
}

// New creates a new application instance
func New(cfg *config.Config, logger *zap.Logger, db *database.Manager) *App {
	return &App{
		config: cfg,
		logger: logger,
		db:     db,
	}
}

// Config returns the application configuration
func (a *App) Config() *config.Config {
	return a.config
}

// Logger returns the application logger
func (a *App) Logger() *zap.Logger {
	return a.logger
}

// DB returns the database manager
func (a *App) DB() *database.Manager {
	return a.db
}

// GetDB returns the GORM database instance
func (a *App) GetDB() *sql.DB {
	return a.db.GetSQLDB()
}

// Close gracefully closes the application
func (a *App) Close() error {
	return a.db.Close()
}
