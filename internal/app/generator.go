package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Generator handles code generation for scaffolding
type Generator struct{}

// NewGenerator creates a new generator instance
func NewGenerator() *Generator {
	return &Generator{}
}

// CreateModule generates a complete module with model, controller, repository, and HTMX views
func (g *Generator) CreateModule(name string) error {
	// Create model
	if err := g.CreateModel(name); err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Create controller
	if err := g.CreateController(name); err != nil {
		return fmt.Errorf("failed to create controller: %w", err)
	}

	// Create repository
	if err := g.CreateRepository(name); err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	// Create HTMX views
	if err := g.CreateHTMXViews(name); err != nil {
		return fmt.Errorf("failed to create HTMX views: %w", err)
	}

	// Create migration
	if err := g.CreateMigration(name); err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	return nil
}

// CreateResource generates a complete API resource with CRUD operations
func (g *Generator) CreateResource(name string) error {
	// Create model
	if err := g.CreateModel(name); err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Create API controller
	if err := g.CreateAPIController(name); err != nil {
		return fmt.Errorf("failed to create API controller: %w", err)
	}

	// Create repository
	if err := g.CreateRepository(name); err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	// Create migration
	if err := g.CreateMigration(name); err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	return nil
}

// CreateHTMXViews generates HTMX-based views for a module
func (g *Generator) CreateHTMXViews(name string) error {
	viewsDir := fmt.Sprintf("resources/views/%s", strings.ToLower(name))
	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		return err
	}

	// Create index view
	if err := g.createHTMXView(name, "index", viewsDir); err != nil {
		return err
	}

	// Create show view
	if err := g.createHTMXView(name, "show", viewsDir); err != nil {
		return err
	}

	// Create create view
	if err := g.createHTMXView(name, "create", viewsDir); err != nil {
		return err
	}

	// Create edit view
	if err := g.createHTMXView(name, "edit", viewsDir); err != nil {
		return err
	}

	// Create form partial
	if err := g.createHTMXView(name, "form", viewsDir); err != nil {
		return err
	}

	return nil
}

// CreateRepository generates a repository for data access
func (g *Generator) CreateRepository(name string) error {
	repositoriesDir := "app/repositories"
	if err := os.MkdirAll(repositoriesDir, 0755); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.go", strings.ToLower(name))
	filepath := filepath.Join(repositoriesDir, filename)
	content := g.generateRepositoryContent(name)

	return os.WriteFile(filepath, []byte(content), 0644)
}

// CreateAPIController generates an API-specific controller
func (g *Generator) CreateAPIController(name string) error {
	controllersDir := "app/http/controllers/api"
	if err := os.MkdirAll(controllersDir, 0755); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.go", strings.ToLower(name))
	filepath := filepath.Join(controllersDir, filename)
	content := g.generateAPIControllerContent(name)

	return os.WriteFile(filepath, []byte(content), 0644)
}

// CreatePostmanCollection generates a Postman collection for API testing
func (g *Generator) CreatePostmanCollection() error {
	// Ensure postman directory exists
	postmanDir := "postman"
	if err := os.MkdirAll(postmanDir, 0755); err != nil {
		return err
	}

	filename := "Dolphin-Framework-API.postman_collection.json"
	filepath := filepath.Join(postmanDir, filename)
	content := g.generatePostmanCollectionContent()

	return os.WriteFile(filepath, []byte(content), 0644)
}

// CreateProvider generates a service provider
func (g *Generator) CreateProvider(name, providerType string, priority int) error {
	providersDir := "app/providers"
	if err := os.MkdirAll(providersDir, 0755); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.go", strings.ToLower(name))
	filepath := filepath.Join(providersDir, filename)
	content := g.generateProviderContent(name, providerType, priority)

	return os.WriteFile(filepath, []byte(content), 0644)
}

// createHTMXView creates a specific HTMX view
func (g *Generator) createHTMXView(name, viewType, viewsDir string) error {
	filename := fmt.Sprintf("%s.html", viewType)
	filepath := filepath.Join(viewsDir, filename)
	content := g.generateHTMXViewContent(name, viewType)
	return os.WriteFile(filepath, []byte(content), 0644)
}

// CreateController generates a new controller
func (g *Generator) CreateController(name string) error {
	// Ensure controllers directory exists
	controllersDir := "app/http/controllers"
	if err := os.MkdirAll(controllersDir, 0755); err != nil {
		return err
	}

	// Generate controller filename
	filename := fmt.Sprintf("%s.go", strings.ToLower(name))
	filepath := filepath.Join(controllersDir, filename)

	// Generate controller content
	content := g.generateControllerContent(name)

	return os.WriteFile(filepath, []byte(content), 0644)
}

// CreateModel generates a new model
func (g *Generator) CreateModel(name string) error {
	// Ensure models directory exists
	modelsDir := "app/models"
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return err
	}

	// Generate model filename
	filename := fmt.Sprintf("%s.go", strings.ToLower(name))
	filepath := filepath.Join(modelsDir, filename)

	// Generate model content
	content := g.generateModelContent(name)

	return os.WriteFile(filepath, []byte(content), 0644)
}

// CreateMigration generates a new migration
func (g *Generator) CreateMigration(name string) error {
	// Ensure migrations directory exists
	migrationsDir := "migrations"
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return err
	}

	// Generate migration filename with timestamp
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s.go", timestamp, strings.ToLower(name))
	filepath := filepath.Join(migrationsDir, filename)

	// Generate migration content
	content := g.generateMigrationContent(name)

	return os.WriteFile(filepath, []byte(content), 0644)
}

// CreateMiddleware generates a new middleware
func (g *Generator) CreateMiddleware(name string) error {
	// Ensure middleware directory exists
	middlewareDir := "app/http/middleware"
	if err := os.MkdirAll(middlewareDir, 0755); err != nil {
		return err
	}

	// Generate middleware filename
	filename := fmt.Sprintf("%s.go", strings.ToLower(name))
	filepath := filepath.Join(middlewareDir, filename)

	// Generate middleware content
	content := g.generateMiddlewareContent(name)

	return os.WriteFile(filepath, []byte(content), 0644)
}

// generateControllerContent creates controller template
func (g *Generator) generateControllerContent(name string) string {
	lowerName := strings.ToLower(name)
	return `package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// ` + name + ` handles ` + lowerName + ` related requests
type ` + name + ` struct{}

// New` + name + ` creates a new ` + name + ` controller
func New` + name + `() *` + name + ` {
	return &` + name + `{}
}

// Index handles GET /` + lowerName + `
func (c *` + name + `) Index(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]interface{}{
		"message": "List of ` + lowerName + `",
		"data":    []interface{}{},
	})
}

// Show handles GET /` + lowerName + `/{id}
func (c *` + name + `) Show(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	render.JSON(w, r, map[string]interface{}{
		"message": "Show ` + lowerName + `",
		"id":      id,
		"data":    map[string]interface{}{},
	})
}

// Store handles POST /` + lowerName + `
func (c *` + name + `) Store(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]interface{}{
		"message": "` + lowerName + ` created successfully",
		"data":    map[string]interface{}{},
	})
}

// Update handles PUT /` + lowerName + `/{id}
func (c *` + name + `) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	render.JSON(w, r, map[string]interface{}{
		"message": "` + name + ` updated successfully",
		"id":      id,
		"data":    map[string]interface{}{},
	})
}

// Destroy handles DELETE /` + lowerName + `/{id}
func (c *` + name + `) Destroy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	render.JSON(w, r, map[string]interface{}{
		"message": "` + lowerName + ` deleted successfully",
		"id":      id,
	})
}`
}

// generateModelContent creates model template
func (g *Generator) generateModelContent(name string) string {
	return fmt.Sprintf(`package models

import (
	"time"
	"gorm.io/gorm"
)

// %s represents a %s model
type %s struct {
	ID        uint           `+"`gorm:\"primarykey\"`"+`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `+"`gorm:\"index\"`"+`
	
	// Add your fields here
	// Name string `+"`gorm:\"not null\"`"+`
	// Email string `+"`gorm:\"uniqueIndex\"`"+`
}

// TableName returns the table name for the %s model
func (%s) TableName() string {
	return "%s"
}

// BeforeCreate is called before creating a new record
func (m *%s) BeforeCreate(tx *gorm.DB) error {
	// Add any pre-create logic here
	return nil
}

// BeforeUpdate is called before updating a record
func (m *%s) BeforeUpdate(tx *gorm.DB) error {
	// Add any pre-update logic here
	return nil
}

// BeforeDelete is called before deleting a record
func (m *%s) BeforeDelete(tx *gorm.DB) error {
	// Add any pre-delete logic here
	return nil
}
`, name, strings.ToLower(name), name, strings.ToLower(name), name, strings.ToLower(name), name, name, name)
}

// generateMigrationContent creates migration template
func (g *Generator) generateMigrationContent(name string) string {
	return fmt.Sprintf(`package migrations

import (
	raptor "github.com/mrhoseah/raptor/core"
)

// %s represents the %s migration
type %s struct{}

// Name returns the migration name
func (m *%s) Name() string {
	return "%s"
}

// Up runs the migration
func (m *%s) Up(s raptor.Schema) error {
	// Add your migration logic here
	// Example: Create a table
	// return s.CreateTable("%s", []string{"id", "name", "email", "created_at"})
	
	return nil
}

// Down rolls back the migration
func (m *%s) Down(s raptor.Schema) error {
	// Add your rollback logic here
	// Example: Drop a table
	// return s.DropTable("%s")
	
	return nil
}
`, name, strings.ToLower(name), name, name, strings.ToLower(name), name, strings.ToLower(name), name, strings.ToLower(name))
}

// generateMiddlewareContent creates middleware template
func (g *Generator) generateMiddlewareContent(name string) string {
	return fmt.Sprintf(`package middleware

import (
	"net/http"
)

// %s middleware
func %s(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add your middleware logic here
		
		// Example: Add custom header
		// w.Header().Set("X-Custom-Header", "value")
		
		// Example: Log request
		// log.Printf("Request: %s %s", r.Method, r.URL.Path)
		
		// Example: Authentication check
		// if !isAuthenticated(r) {
		//     http.Error(w, "Unauthorized", http.StatusUnauthorized)
		//     return
		// }
		
		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}
`, name, name)
}
