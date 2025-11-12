package template

import (
	"github.com/mrhoseah/dolphin/internal/template"
)

// FinTemplateEngine represents the Fin template engine interface
type FinTemplateEngine = template.FinTemplateEngine

// Config represents template engine configuration
type Config = template.Config

// NewFinEngine creates a new Fin template engine
func NewFinEngine(cfg *Config) FinTemplateEngine {
	return template.NewFinEngine(cfg)
}

