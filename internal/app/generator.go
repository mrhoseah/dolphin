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

// CreateAuth generates complete authentication setup (like Laravel Breeze)
func (g *Generator) CreateAuth() error {
	// Create views directory structure
	viewsDir := "views"
	authDir := filepath.Join(viewsDir, "auth")
	layoutsDir := filepath.Join(viewsDir, "layouts")

	if err := os.MkdirAll(authDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(layoutsDir, 0755); err != nil {
		return err
	}

	// Create controllers directory
	controllersDir := "app/http/controllers"
	if err := os.MkdirAll(controllersDir, 0755); err != nil {
		return err
	}

	// Create bootstrap directory for routes
	bootstrapDir := "bootstrap"
	if err := os.MkdirAll(bootstrapDir, 0755); err != nil {
		return err
	}

	// 1. Create authentication views
	if err := g.createAuthView("login", authDir); err != nil {
		return err
	}

	if err := g.createAuthView("register", authDir); err != nil {
		return err
	}

	if err := g.createAuthView("forgot-password", authDir); err != nil {
		return err
	}

	if err := g.createAuthView("reset-password", authDir); err != nil {
		return err
	}

	// 2. Create or update base layout with auth navigation
	baseLayoutPath := filepath.Join(layoutsDir, "base.fin.html")
	if err := g.createOrUpdateBaseLayout(baseLayoutPath); err != nil {
		return err
	}

	// 3. Create User model (required for authentication)
	modelsDir := "app/models"
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return err
	}
	userModelPath := filepath.Join(modelsDir, "user.go")
	if err := g.createUserModel(userModelPath); err != nil {
		return err
	}

	// 4. Generate AuthController with web methods
	authControllerPath := filepath.Join(controllersDir, "auth_controller.go")
	if err := g.createAuthController(authControllerPath); err != nil {
		return err
	}

	// 5. Generate auth routes file
	authRoutesPath := filepath.Join(bootstrapDir, "auth_routes.go")
	if err := g.createAuthRoutes(authRoutesPath); err != nil {
		return err
	}

	// 6. Create users migration (required for authentication)
	if err := g.createUsersMigration(); err != nil {
		return err
	}

	// 7. Create essential pages (home, dashboard)
	pagesDir := filepath.Join(viewsDir, "pages")
	if err := os.MkdirAll(pagesDir, 0755); err != nil {
		return err
	}

	// Create home page
	homePagePath := filepath.Join(pagesDir, "home.fin.html")
	if err := g.createHomePage(homePagePath); err != nil {
		return err
	}

	// Create dashboard page
	dashboardPagePath := filepath.Join(pagesDir, "dashboard.fin.html")
	if err := g.createDashboardPage(dashboardPagePath); err != nil {
		return err
	}

	// 8. Create error pages
	errorsDir := filepath.Join(viewsDir, "errors")
	if err := os.MkdirAll(errorsDir, 0755); err != nil {
		return err
	}

	// Create error page
	errorPagePath := filepath.Join(errorsDir, "error.fin.html")
	if err := g.createErrorPage(errorPagePath); err != nil {
		return err
	}

	return nil
}

// createAuthView creates a specific authentication view
func (g *Generator) createAuthView(viewType, authDir string) error {
	filename := fmt.Sprintf("%s.fin.html", viewType)
	filepath := filepath.Join(authDir, filename)
	content := g.generateAuthViewContent(viewType)
	return os.WriteFile(filepath, []byte(content), 0644)
}

// createBaseLayout creates the base layout template
func (g *Generator) createBaseLayout(layoutsDir string) error {
	filename := "base.fin.html"
	filepath := filepath.Join(layoutsDir, filename)
	content := g.generateBaseLayoutContent()
	return os.WriteFile(filepath, []byte(content), 0644)
}

// createOrUpdateBaseLayout creates or updates base layout with auth navigation
func (g *Generator) createOrUpdateBaseLayout(filepath string) error {
	// Always create/update with auth-enabled layout
	content := g.generateBaseLayoutWithAuth()
	return os.WriteFile(filepath, []byte(content), 0644)
}

// createAuthController generates AuthController with web methods
func (g *Generator) createAuthController(filepath string) error {
	// Try to detect module name from go.mod
	moduleName := "app" // default
	if modContent, err := os.ReadFile("go.mod"); err == nil {
		lines := strings.Split(string(modContent), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "module ") {
				moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
				break
			}
		}
	}
	content := g.generateAuthControllerContent(moduleName)
	return os.WriteFile(filepath, []byte(content), 0644)
}

// createAuthRoutes generates auth routes file
func (g *Generator) createAuthRoutes(filepath string) error {
	// Try to detect module name from go.mod
	moduleName := "app" // default
	if modContent, err := os.ReadFile("go.mod"); err == nil {
		lines := strings.Split(string(modContent), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "module ") {
				moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
				break
			}
		}
	}

	content := g.generateAuthRoutesContent(moduleName)
	return os.WriteFile(filepath, []byte(content), 0644)
}

// createHomePage creates the home page view
func (g *Generator) createHomePage(filepath string) error {
	content := g.generateHomePageContent()
	return os.WriteFile(filepath, []byte(content), 0644)
}

// createDashboardPage creates the dashboard page view
func (g *Generator) createDashboardPage(filepath string) error {
	content := g.generateDashboardPageContent()
	return os.WriteFile(filepath, []byte(content), 0644)
}

// createErrorPage creates the error page view
func (g *Generator) createErrorPage(filepath string) error {
	content := g.generateErrorPageContent()
	return os.WriteFile(filepath, []byte(content), 0644)
}

// createUserModel creates the User model for authentication
func (g *Generator) createUserModel(filepath string) error {
	content := g.generateUserModelContent()
	return os.WriteFile(filepath, []byte(content), 0644)
}

// createUsersMigration creates a timestamped users migration file under database/migrations
func (g *Generator) createUsersMigration() error {
	// Ensure migrations directory exists
	migrationsDir := filepath.Join("database", "migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return err
	}

	// Detect module name for imports
	moduleName := "app"
	if modContent, err := os.ReadFile("go.mod"); err == nil {
		lines := strings.Split(string(modContent), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "module ") {
				moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
				break
			}
		}
	}

	// Timestamped filename
	ts := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_create_users_table.go", ts)
	path := filepath.Join(migrationsDir, filename)

	content := g.generateUsersMigrationContent(moduleName)
	return os.WriteFile(path, []byte(content), 0644)
}

// CreateHTMXViews generates HTMX-based views for a module
func (g *Generator) CreateHTMXViews(name string) error {
	viewsDir := fmt.Sprintf("views/pages/%s", strings.ToLower(name))
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
// Enforces .fin.html extension
func (g *Generator) createHTMXView(name, viewType, viewsDir string) error {
	filename := fmt.Sprintf("%s.fin.html", viewType)
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
        // log.Println("Request:", r.Method, r.URL.Path)
		
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

// CreateNewApp creates a new Dolphin application with the specified frontend stack
func (g *Generator) CreateNewApp(appName, frontend string, includeAuth bool) error {
	// Create app directory
	if err := os.MkdirAll(appName, 0755); err != nil {
		return fmt.Errorf("failed to create app directory: %w", err)
	}

	// Change to app directory for file creation
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(appName); err != nil {
		return fmt.Errorf("failed to change to app directory: %w", err)
	}

	// Create basic directory structure
	dirs := []string{
		"app/models",
		"app/http/controllers",
		"app/repositories",
		"bootstrap",
		"config",
		"database/migrations",
		"views/layouts",
		"views/pages",
		"public",
		"storage/app/public",
		"assets",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create go.mod
	if err := g.createGoMod(appName); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}

	// Create main.go
	if err := g.createMainGo(appName); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}

	// Create config files
	if err := g.createConfigFiles(); err != nil {
		return fmt.Errorf("failed to create config files: %w", err)
	}

	// Setup frontend based on choice
	switch frontend {
	case "react":
		if err := g.setupReactFrontend(); err != nil {
			return fmt.Errorf("failed to setup React frontend: %w", err)
		}
	case "vue":
		if err := g.setupVueFrontend(); err != nil {
			return fmt.Errorf("failed to setup Vue frontend: %w", err)
		}
	case "fin":
		if err := g.setupFinFrontend(); err != nil {
			return fmt.Errorf("failed to setup Fin frontend: %w", err)
		}
	}

	// Create base layout
	if err := g.createBaseLayoutForNewApp(frontend); err != nil {
		return fmt.Errorf("failed to create base layout: %w", err)
	}

	// Create home page
	if err := g.createHomePageForNewApp(frontend); err != nil {
		return fmt.Errorf("failed to create home page: %w", err)
	}

	// Create bootstrap routes (pass includeAuth flag)
	if err := g.createBootstrapRoutes(includeAuth, appName); err != nil {
		return fmt.Errorf("failed to create bootstrap routes: %w", err)
	}

	// Include auth if requested
	if includeAuth {
		if err := g.CreateAuth(); err != nil {
			return fmt.Errorf("failed to create auth: %w", err)
		}
	}

	return nil
}

// setupReactFrontend sets up React frontend
func (g *Generator) setupReactFrontend() error {
	// Create assets directories
	dirs := []string{"assets/ts", "assets/js", "assets/css"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create package.json
	if err := g.createReactPackageJson(); err != nil {
		return err
	}

	// Create TypeScript config
	if err := g.createTsConfig(); err != nil {
		return err
	}

	// Create Tailwind config
	if err := g.createTailwindConfig(); err != nil {
		return err
	}

	// Create PostCSS config
	if err := g.createPostCSSConfig(); err != nil {
		return err
	}

	// Create basic React entry point
	if err := g.createReactEntryPoint(); err != nil {
		return err
	}

	// Create CSS file
	if err := g.createAppCSS(); err != nil {
		return err
	}

	return nil
}

// setupVueFrontend sets up Vue.js frontend
func (g *Generator) setupVueFrontend() error {
	// Create assets directories
	dirs := []string{"assets/js", "assets/css"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create package.json for Vue
	if err := g.createVuePackageJson(); err != nil {
		return err
	}

	// Create Tailwind config
	if err := g.createTailwindConfig(); err != nil {
		return err
	}

	// Create PostCSS config
	if err := g.createPostCSSConfig(); err != nil {
		return err
	}

	// Create basic Vue entry point
	if err := g.createVueEntryPoint(); err != nil {
		return err
	}

	// Create CSS file
	if err := g.createAppCSS(); err != nil {
		return err
	}

	return nil
}

// setupFinFrontend sets up Fin templates with vanilla JS
func (g *Generator) setupFinFrontend() error {
	// Create assets directories
	dirs := []string{"assets/js", "assets/css"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create package.json for vanilla JS
	if err := g.createFinPackageJson(); err != nil {
		return err
	}

	// Create Tailwind config
	if err := g.createTailwindConfig(); err != nil {
		return err
	}

	// Create PostCSS config
	if err := g.createPostCSSConfig(); err != nil {
		return err
	}

	// Create vanilla JS entry point
	if err := g.createVanillaJSEntryPoint(); err != nil {
		return err
	}

	// Create CSS file
	if err := g.createAppCSS(); err != nil {
		return err
	}

	return nil
}

// createConfigFiles creates default application configuration files
func (g *Generator) createConfigFiles() error {
	// Ensure config directory exists
	if err := os.MkdirAll("config", 0755); err != nil {
		return err
	}

	content := `app:
  name: "Dolphin App"
  env: "development"
  debug: true
  url: "http://localhost:8080"

server:
  host: "localhost"
  port: 8080
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 120

database:
  driver: "sqlite"
  database: "app.db"
  host: ""
  port: 0
  username: ""
  password: ""
  ssl_mode: "disable"

log:
  level: "debug"
  format: "json"
  output: "stdout"
`
	return os.WriteFile("config/config.yaml", []byte(content), 0644)
}
