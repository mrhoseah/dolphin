package app

import (
	"dolphin/internal/app"
	"dolphin/internal/config"
	"dolphin/internal/database"
	"go.uber.org/zap"
)

// Application represents the main application instance
type Application = app.App

// New creates a new application instance
func New(cfg *config.Config, logger *zap.Logger, db *database.Manager) *Application {
	return app.New(cfg, logger, db)
}

