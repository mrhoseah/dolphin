package template

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TemplateType represents the type of template
type TemplateType int

const (
	TypeLayout TemplateType = iota
	TypePartial
	TypePage
	TypeComponent
	TypeEmail
	TypeSMS
)

func (tt TemplateType) String() string {
	switch tt {
	case TypeLayout:
		return "layout"
	case TypePartial:
		return "partial"
	case TypePage:
		return "page"
	case TypeComponent:
		return "component"
	case TypeEmail:
		return "email"
	case TypeSMS:
		return "sms"
	default:
		return "unknown"
	}
}

// Config represents template engine configuration
type Config struct {
	// Template directories
	LayoutsDir    string `yaml:"layouts_dir" json:"layouts_dir"`
	PartialsDir   string `yaml:"partials_dir" json:"partials_dir"`
	PagesDir      string `yaml:"pages_dir" json:"pages_dir"`
	ComponentsDir string `yaml:"components_dir" json:"components_dir"`
	EmailsDir     string `yaml:"emails_dir" json:"emails_dir"`

	// Template settings
	Extension      string `yaml:"extension" json:"extension"`
	AutoReload     bool   `yaml:"auto_reload" json:"auto_reload"`
	CacheTemplates bool   `yaml:"cache_templates" json:"cache_templates"`

	// Layout settings
	DefaultLayout string `yaml:"default_layout" json:"default_layout"`
	LayoutVar     string `yaml:"layout_var" json:"layout_var"`

	// Helper settings
	EnableHelpers bool `yaml:"enable_helpers" json:"enable_helpers"`

	// Security settings
	EscapeHTML     bool     `yaml:"escape_html" json:"escape_html"`
	TrustedOrigins []string `yaml:"trusted_origins" json:"trusted_origins"`

	// Performance settings
	MaxCacheSize int           `yaml:"max_cache_size" json:"max_cache_size"`
	CacheExpiry  time.Duration `yaml:"cache_expiry" json:"cache_expiry"`

	// Logging
	EnableLogging  bool `yaml:"enable_logging" json:"enable_logging"`
	VerboseLogging bool `yaml:"verbose_logging" json:"verbose_logging"`
}

// DefaultConfig returns default template engine configuration
func DefaultConfig() *Config {
	return &Config{
		LayoutsDir:     "ui/views/layouts",
		PartialsDir:    "ui/views/partials",
		PagesDir:       "ui/views/pages",
		ComponentsDir:  "ui/views/components",
		EmailsDir:      "ui/views/emails",
		Extension:      ".html",
		AutoReload:     true,
		CacheTemplates: true,
		DefaultLayout:  "base",
		LayoutVar:      "layout",
		EnableHelpers:  true,
		EscapeHTML:     true,
		TrustedOrigins: []string{},
		MaxCacheSize:   1000,
		CacheExpiry:    24 * time.Hour,
		EnableLogging:  true,
		VerboseLogging: false,
	}
}

// Template represents a compiled template
type Template struct {
	Name         string             `json:"name"`
	Type         TemplateType       `json:"type"`
	Path         string             `json:"path"`
	Content      string             `json:"content"`
	Compiled     *template.Template `json:"-"`
	LastModified time.Time          `json:"last_modified"`
	Size         int64              `json:"size"`
	Hash         string             `json:"hash"`
	Blocks       map[string]string  `json:"blocks,omitempty"`
	Extends      string             `json:"extends,omitempty"`
	Includes     []string           `json:"includes,omitempty"`
}

// TemplateData represents data passed to templates
type TemplateData map[string]interface{}

// HelperFunc represents a template helper function
type HelperFunc func(args ...interface{}) (interface{}, error)

// Engine represents the template engine
type Engine struct {
	config *Config
	logger *zap.Logger

	// Template storage
	templates  map[string]*Template
	layouts    map[string]*Template
	partials   map[string]*Template
	pages      map[string]*Template
	components map[string]*Template
	emails     map[string]*Template

	// Helper functions
	helpers map[string]HelperFunc

	// Cache
	cache map[string]*Template

	// File watcher
	watcher *TemplateWatcher

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewEngine creates a new template engine
func NewEngine(config *Config, logger *zap.Logger) (*Engine, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create directories if they don't exist
	dirs := []string{
		config.LayoutsDir,
		config.PartialsDir,
		config.PagesDir,
		config.ComponentsDir,
		config.EmailsDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	engine := &Engine{
		config:     config,
		logger:     logger,
		templates:  make(map[string]*Template),
		layouts:    make(map[string]*Template),
		partials:   make(map[string]*Template),
		pages:      make(map[string]*Template),
		components: make(map[string]*Template),
		emails:     make(map[string]*Template),
		helpers:    make(map[string]HelperFunc),
		cache:      make(map[string]*Template),
	}

	// Register default helpers
	engine.registerDefaultHelpers()

	// Load templates
	if err := engine.LoadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	// Start file watcher if auto-reload is enabled
	if config.AutoReload {
		watcher, err := NewTemplateWatcher(engine, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create template watcher: %w", err)
		}
		engine.watcher = watcher
		go watcher.Watch()
	}

	return engine, nil
}

// LoadTemplates loads all templates from directories
func (e *Engine) LoadTemplates() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clear existing templates
	e.templates = make(map[string]*Template)
	e.layouts = make(map[string]*Template)
	e.partials = make(map[string]*Template)
	e.pages = make(map[string]*Template)
	e.components = make(map[string]*Template)
	e.emails = make(map[string]*Template)

	// Load templates from each directory
	directories := map[string]TemplateType{
		e.config.LayoutsDir:    TypeLayout,
		e.config.PartialsDir:   TypePartial,
		e.config.PagesDir:      TypePage,
		e.config.ComponentsDir: TypeComponent,
		e.config.EmailsDir:     TypeEmail,
	}

	for dir, templateType := range directories {
		if err := e.loadTemplatesFromDir(dir, templateType); err != nil {
			return fmt.Errorf("failed to load templates from %s: %w", dir, err)
		}
	}

	if e.config.EnableLogging && e.logger != nil {
		e.logger.Info("Templates loaded successfully",
			zap.Int("total", len(e.templates)),
			zap.Int("layouts", len(e.layouts)),
			zap.Int("partials", len(e.partials)),
			zap.Int("pages", len(e.pages)),
			zap.Int("components", len(e.components)),
			zap.Int("emails", len(e.emails)))
	}

	return nil
}

// loadTemplatesFromDir loads templates from a specific directory
func (e *Engine) loadTemplatesFromDir(dir string, templateType TemplateType) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check file extension
		if !strings.HasSuffix(path, e.config.Extension) {
			return nil
		}

		// Load template
		template, err := e.loadTemplate(path, templateType)
		if err != nil {
			if e.config.EnableLogging && e.logger != nil {
				e.logger.Warn("Failed to load template",
					zap.String("path", path),
					zap.Error(err))
			}
			return nil // Continue loading other templates
		}

		// Store template
		e.templates[template.Name] = template

		// Store in type-specific map
		switch templateType {
		case TypeLayout:
			e.layouts[template.Name] = template
		case TypePartial:
			e.partials[template.Name] = template
		case TypePage:
			e.pages[template.Name] = template
		case TypeComponent:
			e.components[template.Name] = template
		case TypeEmail:
			e.emails[template.Name] = template
		}

		return nil
	})
}

// loadTemplate loads a single template
func (e *Engine) loadTemplate(path string, templateType TemplateType) (*Template, error) {
	// Read template content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Generate template name from path
	name := e.generateTemplateName(path, templateType)

	// Create template
	tmpl := &Template{
		Name:         name,
		Type:         templateType,
		Path:         path,
		Content:      string(content),
		LastModified: info.ModTime(),
		Size:         info.Size(),
		Hash:         e.generateHash(string(content)),
	}

	// Compile template
	if err := e.compileTemplate(tmpl); err != nil {
		return nil, fmt.Errorf("failed to compile template %s: %w", name, err)
	}

	return tmpl, nil
}

// generateTemplateName generates a template name from path
func (e *Engine) generateTemplateName(path string, templateType TemplateType) string {
	// Get relative path from template directory
	var baseDir string
	switch templateType {
	case TypeLayout:
		baseDir = e.config.LayoutsDir
	case TypePartial:
		baseDir = e.config.PartialsDir
	case TypePage:
		baseDir = e.config.PagesDir
	case TypeComponent:
		baseDir = e.config.ComponentsDir
	case TypeEmail:
		baseDir = e.config.EmailsDir
	}

	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		relPath = path
	}

	// Remove extension and convert to template name
	name := strings.TrimSuffix(relPath, e.config.Extension)
	name = strings.ReplaceAll(name, string(filepath.Separator), ".")

	return name
}

// generateHash generates a hash for template content
func (e *Engine) generateHash(content string) string {
	// Simple hash implementation
	hash := 0
	for _, char := range content {
		hash = hash*31 + int(char)
	}
	return fmt.Sprintf("%x", hash)
}

// compileTemplate compiles a template with helpers
func (e *Engine) compileTemplate(tmpl *Template) error {
	// Create template with helpers
	funcMap := template.FuncMap{}

	// Add helper functions
	if e.config.EnableHelpers {
		for name, helper := range e.helpers {
			funcMap[name] = helper
		}
	}

	// Compile template
	compiled, err := template.New(tmpl.Name).Funcs(funcMap).Parse(tmpl.Content)
	if err != nil {
		return err
	}

	tmpl.Compiled = compiled
	return nil
}

// Render renders a template with data
func (e *Engine) Render(name string, data TemplateData) (string, error) {
	e.mu.RLock()
	tmpl, exists := e.templates[name]
	e.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("template %s not found", name)
	}

	// Check if template needs recompilation
	if e.config.AutoReload && e.needsRecompilation(tmpl) {
		if err := e.reloadTemplate(tmpl); err != nil {
			if e.config.EnableLogging && e.logger != nil {
				e.logger.Warn("Failed to reload template",
					zap.String("template", name),
					zap.Error(err))
			}
		}
	}

	// Render template
	var buf bytes.Buffer
	if err := tmpl.Compiled.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", name, err)
	}

	return buf.String(), nil
}

// RenderWithLayout renders a template with a layout
func (e *Engine) RenderWithLayout(pageName, layoutName string, data TemplateData) (string, error) {
	// Get page template
	e.mu.RLock()
	page, exists := e.pages[pageName]
	e.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("page template %s not found", pageName)
	}

	// Use default layout if not specified
	if layoutName == "" {
		layoutName = e.config.DefaultLayout
	}

	// Get layout template
	e.mu.RLock()
	layout, exists := e.layouts[layoutName]
	e.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("layout template %s not found", layoutName)
	}

	// Add page content to data
	data[e.config.LayoutVar] = page.Content

	// Render layout with page content
	return e.Render(layout.Name, data)
}

// RenderPartial renders a partial template
func (e *Engine) RenderPartial(name string, data TemplateData) (string, error) {
	e.mu.RLock()
	partial, exists := e.partials[name]
	e.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("partial template %s not found", name)
	}

	return e.Render(partial.Name, data)
}

// RenderComponent renders a component template
func (e *Engine) RenderComponent(name string, data TemplateData) (string, error) {
	e.mu.RLock()
	component, exists := e.components[name]
	e.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("component template %s not found", name)
	}

	return e.Render(component.Name, data)
}

// RenderEmail renders an email template
func (e *Engine) RenderEmail(name string, data TemplateData) (string, error) {
	e.mu.RLock()
	email, exists := e.emails[name]
	e.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("email template %s not found", name)
	}

	return e.Render(email.Name, data)
}

// needsRecompilation checks if a template needs recompilation
func (e *Engine) needsRecompilation(tmpl *Template) bool {
	// Check if file has been modified
	info, err := os.Stat(tmpl.Path)
	if err != nil {
		return false
	}

	return info.ModTime().After(tmpl.LastModified)
}

// reloadTemplate reloads a template
func (e *Engine) reloadTemplate(tmpl *Template) error {
	// Read updated content
	content, err := os.ReadFile(tmpl.Path)
	if err != nil {
		return err
	}

	// Update template
	tmpl.Content = string(content)
	tmpl.LastModified = time.Now()
	tmpl.Hash = e.generateHash(string(content))

	// Recompile template
	return e.compileTemplate(tmpl)
}

// RegisterHelper registers a template helper function
func (e *Engine) RegisterHelper(name string, helper HelperFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.helpers[name] = helper
}

// GetTemplate returns a template by name
func (e *Engine) GetTemplate(name string) (*Template, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	tmpl, exists := e.templates[name]
	return tmpl, exists
}

// GetTemplatesByType returns templates by type
func (e *Engine) GetTemplatesByType(templateType TemplateType) map[string]*Template {
	e.mu.RLock()
	defer e.mu.RUnlock()

	switch templateType {
	case TypeLayout:
		result := make(map[string]*Template)
		for name, tmpl := range e.layouts {
			result[name] = tmpl
		}
		return result
	case TypePartial:
		result := make(map[string]*Template)
		for name, tmpl := range e.partials {
			result[name] = tmpl
		}
		return result
	case TypePage:
		result := make(map[string]*Template)
		for name, tmpl := range e.pages {
			result[name] = tmpl
		}
		return result
	case TypeComponent:
		result := make(map[string]*Template)
		for name, tmpl := range e.components {
			result[name] = tmpl
		}
		return result
	case TypeEmail:
		result := make(map[string]*Template)
		for name, tmpl := range e.emails {
			result[name] = tmpl
		}
		return result
	default:
		return make(map[string]*Template)
	}
}

// GetAllTemplates returns all templates
func (e *Engine) GetAllTemplates() map[string]*Template {
	e.mu.RLock()
	defer e.mu.RUnlock()

	templates := make(map[string]*Template)
	for name, tmpl := range e.templates {
		templates[name] = tmpl
	}
	return templates
}

// Stop stops the template engine
func (e *Engine) Stop() error {
	if e.watcher != nil {
		e.watcher.Stop()
	}
	return nil
}

// ConfigFromEnv creates config from environment variables
func ConfigFromEnv() *Config {
	config := DefaultConfig()

	// Override with environment variables if present
	// This would typically read from environment variables
	// For now, return the default config
	return config
}
