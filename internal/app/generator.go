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
	return fmt.Sprintf(`package controllers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// %s handles %s related requests
type %s struct{}

// New%s creates a new %s controller
func New%s() *%s {
	return &%s{}
}

// Index handles GET /%s
func (c *%s) Index(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]interface{}{
		"message": "List of %s",
		"data":    []interface{}{},
	})
}

// Show handles GET /%s/{id}
func (c *%s) Show(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	render.JSON(w, r, map[string]interface{}{
		"message": "Show %s",
		"id":      id,
		"data":    map[string]interface{}{},
	})
}

// Store handles POST /%s
func (c *%s) Store(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]interface{}{
		"message": "%s created successfully",
		"data":    map[string]interface{}{},
	})
}

// Update handles PUT /%s/{id}
func (c *%s) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	render.JSON(w, r, map[string]interface{}{
		"message": "%s updated successfully",
		"id":      id,
		"data":    map[string]interface{}{},
	})
}

// Destroy handles DELETE /%s/{id}
func (c *%s) Destroy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	render.JSON(w, r, map[string]interface{}{
		"message": "%s deleted successfully",
		"id":      id,
	})
}
`, 
		name, strings.ToLower(name), name, name, name, name, name, name, 
		strings.ToLower(name), name, strings.ToLower(name), strings.ToLower(name), 
		name, strings.ToLower(name), strings.ToLower(name), name, strings.ToLower(name), 
		name, strings.ToLower(name), name, strings.ToLower(name), name, strings.ToLower(name), 
		name, strings.ToLower(name), name, strings.ToLower(name), name, strings.ToLower(name))
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
	ID        uint           ` + "`gorm:\"primarykey\"`" + `
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt ` + "`gorm:\"index\"`" + `
	
	// Add your fields here
	// Name string ` + "`gorm:\"not null\"`" + `
	// Email string ` + "`gorm:\"uniqueIndex\"`" + `
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
