package config

import (
	"dolphin/internal/config"
)

// Config represents application configuration
type Config = config.Config

// Load loads configuration from file
func Load() (*Config, error) {
	return config.Load()
}

