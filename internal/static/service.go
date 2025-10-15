package static

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Service handles static page serving with templating
type Service struct {
	templates map[string]*template.Template
	mu        sync.RWMutex
	baseDir   string
	fs        embed.FS
	cache     map[string]*CachedPage
	cacheMu   sync.RWMutex
	cacheTTL  time.Duration
}

// CachedPage represents a cached static page
type CachedPage struct {
	Content   []byte
	Timestamp time.Time
	TTL       time.Duration
}

// PageData represents data to be passed to templates
type PageData struct {
	Title       string
	Description string
	Keywords    string
	Author      string
	Data        map[string]interface{}
	Meta        map[string]string
	Assets      map[string]string
}

// Config holds static service configuration
type Config struct {
	BaseDir     string
	TemplateDir string
	CacheTTL    time.Duration
	EnableCache bool
	EmbedFS     embed.FS
}

// NewService creates a new static service
func NewService(config Config) *Service {
	if config.BaseDir == "" {
		config.BaseDir = "resources/static"
	}
	if config.TemplateDir == "" {
		config.TemplateDir = "resources/views"
	}
	if config.CacheTTL == 0 {
		config.CacheTTL = 5 * time.Minute
	}

	return &Service{
		templates: make(map[string]*template.Template),
		baseDir:   config.BaseDir,
		fs:        config.EmbedFS,
		cache:     make(map[string]*CachedPage),
		cacheTTL:  config.CacheTTL,
	}
}

// LoadTemplates loads all templates from the template directory
func (s *Service) LoadTemplates() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear existing templates
	s.templates = make(map[string]*template.Template)

	// Load templates from filesystem
	if err := s.loadTemplatesFromFS(); err != nil {
		return fmt.Errorf("failed to load templates from filesystem: %w", err)
	}

	// Load templates from embedded FS if available
	if s.fs != (embed.FS{}) {
		if err := s.loadTemplatesFromEmbed(); err != nil {
			return fmt.Errorf("failed to load templates from embed: %w", err)
		}
	}

	return nil
}

// loadTemplatesFromFS loads templates from filesystem
func (s *Service) loadTemplatesFromFS() error {
	templateDir := s.baseDir + "/templates"
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		// Create template directory if it doesn't exist
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			return err
		}
		return nil
	}

	return filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			name := strings.TrimPrefix(path, templateDir+"/")
			name = strings.TrimSuffix(name, ".html")

			tmpl, err := template.ParseFiles(path)
			if err != nil {
				return fmt.Errorf("failed to parse template %s: %w", path, err)
			}

			s.templates[name] = tmpl
		}

		return nil
	})
}

// loadTemplatesFromEmbed loads templates from embedded filesystem
func (s *Service) loadTemplatesFromEmbed() error {
	return fs.WalkDir(s.fs, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".html") {
			name := strings.TrimPrefix(path, "templates/")
			name = strings.TrimSuffix(name, ".html")

			content, err := s.fs.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read embedded template %s: %w", path, err)
			}

			tmpl, err := template.New(name).Parse(string(content))
			if err != nil {
				return fmt.Errorf("failed to parse embedded template %s: %w", path, err)
			}

			s.templates[name] = tmpl
		}

		return nil
	})
}

// Render renders a template with data
func (s *Service) Render(templateName string, data PageData) ([]byte, error) {
	s.mu.RLock()
	tmpl, exists := s.templates[templateName]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("template %s not found", templateName)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return []byte(buf.String()), nil
}

// ServePage serves a static page with optional templating
func (s *Service) ServePage(w http.ResponseWriter, r *http.Request, pageName string, data PageData) error {
	// Check cache first
	if s.isCacheEnabled() {
		if cached := s.getCachedPage(pageName); cached != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(cached.Content)
			return nil
		}
	}

	// Try to render template first
	if content, err := s.Render(pageName, data); err == nil {
		// Cache the rendered content
		if s.isCacheEnabled() {
			s.setCachedPage(pageName, content)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
		return nil
	}

	// Fallback to static file
	return s.ServeStaticFile(w, r, pageName+".html")
}

// ServeStaticFile serves a static file
func (s *Service) ServeStaticFile(w http.ResponseWriter, r *http.Request, filename string) error {
	filePath := filepath.Join(s.baseDir, filename)

	// Check if file exists in filesystem
	if _, err := os.Stat(filePath); err == nil {
		http.ServeFile(w, r, filePath)
		return nil
	}

	// Check embedded filesystem
	if s.fs != (embed.FS{}) {
		if content, err := s.fs.ReadFile(filename); err == nil {
			w.Header().Set("Content-Type", s.getContentType(filename))
			w.Write(content)
			return nil
		}
	}

	return fmt.Errorf("static file %s not found", filename)
}

// CreatePage creates a new static page
func (s *Service) CreatePage(name, content string) error {
	filePath := filepath.Join(s.baseDir, name+".html")

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write page content
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write page: %w", err)
	}

	// Reload templates
	return s.LoadTemplates()
}

// CreateTemplate creates a new template
func (s *Service) CreateTemplate(name, content string) error {
	templatePath := filepath.Join(s.baseDir, "templates", name+".html")

	// Ensure template directory exists
	dir := filepath.Dir(templatePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	// Write template content
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	// Reload templates
	return s.LoadTemplates()
}

// ListPages returns a list of available pages
func (s *Service) ListPages() ([]string, error) {
	var pages []string

	// List files from filesystem
	if err := filepath.Walk(s.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			relPath, _ := filepath.Rel(s.baseDir, path)
			name := strings.TrimSuffix(relPath, ".html")
			pages = append(pages, name)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// List templates
	s.mu.RLock()
	for name := range s.templates {
		pages = append(pages, "template:"+name)
	}
	s.mu.RUnlock()

	return pages, nil
}

// DeletePage deletes a static page
func (s *Service) DeletePage(name string) error {
	filePath := filepath.Join(s.baseDir, name+".html")
	return os.Remove(filePath)
}

// DeleteTemplate deletes a template
func (s *Service) DeleteTemplate(name string) error {
	templatePath := filepath.Join(s.baseDir, "templates", name+".html")
	return os.Remove(templatePath)
}

// getCachedPage retrieves a cached page
func (s *Service) getCachedPage(name string) *CachedPage {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	cached, exists := s.cache[name]
	if !exists {
		return nil
	}

	// Check if cache has expired
	if time.Since(cached.Timestamp) > cached.TTL {
		return nil
	}

	return cached
}

// setCachedPage stores a page in cache
func (s *Service) setCachedPage(name string, content []byte) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.cache[name] = &CachedPage{
		Content:   content,
		Timestamp: time.Now(),
		TTL:       s.cacheTTL,
	}
}

// isCacheEnabled checks if caching is enabled
func (s *Service) isCacheEnabled() bool {
	return s.cacheTTL > 0
}

// getContentType returns the appropriate content type for a file
func (s *Service) getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	default:
		return "text/plain"
	}
}

// ClearCache clears the page cache
func (s *Service) ClearCache() {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.cache = make(map[string]*CachedPage)
}

// GetCacheStats returns cache statistics
func (s *Service) GetCacheStats() map[string]interface{} {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	return map[string]interface{}{
		"cached_pages":  len(s.cache),
		"cache_ttl":     s.cacheTTL.String(),
		"cache_enabled": s.isCacheEnabled(),
	}
}
