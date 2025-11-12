package config

import (
	"github.com/mrhoseah/dolphin/internal/config"
)

// Config represents application configuration
type Config = config.Config

// Load loads configuration from file
func Load() (*Config, error) {
	return config.Load()
}

