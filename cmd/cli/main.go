package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "dolphin",
		Short: "🐬 Dolphin Framework CLI - Enterprise-grade Go web framework",
		Long: `🐬 Dolphin Framework CLI

Dolphin is a rapid development web framework written in Go, inspired by Laravel, CodeIgniter, and CakePHP.
This CLI tool provides all the commands you need to build, manage, and deploy your applications.

Examples:
  dolphin new my-app              # Create a new Dolphin application
  dolphin serve                   # Start the development server
  dolphin make:controller User     # Create a new controller
  dolphin migrate                 # Run database migrations
  dolphin swagger                 # Generate API documentation`,
		Version: version,
	}

	// New project command
	var newCmd = &cobra.Command{
		Use:   "new [name]",
		Short: "Create a new Dolphin application",
		Long:  "Create a new Dolphin application with all necessary files and structure",
		Args:  cobra.ExactArgs(1),
		Run:   createNewProject,
	}
	newCmd.Flags().BoolP("force", "f", false, "Overwrite existing directory")

	// List available commands
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all available commands",
		Long:  "Display all available Dolphin CLI commands with descriptions",
		Run:   listCommands,
	}

	// Version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display the current version of Dolphin Framework CLI",
		Run:   showVersion,
	}

	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func createNewProject(cmd *cobra.Command, args []string) {
	projectName := args[0]
	force, _ := cmd.Flags().GetBool("force")

	// Check if directory exists
	if _, err := os.Stat(projectName); err == nil && !force {
		fmt.Printf("❌ Directory '%s' already exists. Use --force to overwrite.\n", projectName)
		return
	}

	fmt.Printf("🚀 Creating new Dolphin application: %s\n", projectName)

	// Create project structure
	createProjectStructure(projectName)

	fmt.Printf("✅ Dolphin application '%s' created successfully!\n", projectName)
	fmt.Printf("\n📋 Next steps:\n")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Printf("  go mod tidy\n")
	fmt.Printf("  dolphin serve\n")
	fmt.Printf("\n📚 Documentation: https://github.com/mrhoseah/dolphin\n")
}

func createProjectStructure(projectName string) {
	// Create main directories
	dirs := []string{
		projectName,
		filepath.Join(projectName, "app"),
		filepath.Join(projectName, "app", "http"),
		filepath.Join(projectName, "app", "http", "controllers"),
		filepath.Join(projectName, "app", "http", "middleware"),
		filepath.Join(projectName, "app", "http", "requests"),
		filepath.Join(projectName, "app", "models"),
		filepath.Join(projectName, "app", "repositories"),
		filepath.Join(projectName, "app", "services"),
		filepath.Join(projectName, "app", "seeders"),
		filepath.Join(projectName, "config"),
		filepath.Join(projectName, "database"),
		filepath.Join(projectName, "database", "migrations"),
		filepath.Join(projectName, "database", "seeders"),
		filepath.Join(projectName, "public"),
		filepath.Join(projectName, "resources"),
		filepath.Join(projectName, "resources", "views"),
		filepath.Join(projectName, "resources", "assets"),
		filepath.Join(projectName, "storage"),
		filepath.Join(projectName, "storage", "logs"),
		filepath.Join(projectName, "storage", "cache"),
		filepath.Join(projectName, "tests"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("❌ Failed to create directory %s: %v\n", dir, err)
			return
		}
	}

	// Create go.mod file
	goModContent := fmt.Sprintf(`module %s

go 1.21

require (
	github.com/mrhoseah/dolphin v1.0.0
)`, projectName)

	if err := os.WriteFile(filepath.Join(projectName, "go.mod"), []byte(goModContent), 0644); err != nil {
		fmt.Printf("❌ Failed to create go.mod: %v\n", err)
		return
	}

	// Create main.go file
	mainGoContent := `package main

// @title Dolphin Framework API
// @version 1.0
// @description Enterprise-grade Go web framework API documentation
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.dolphin-framework.com/support
// @contact.email support@dolphin-framework.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrhoseah/dolphin/internal/app"
	"github.com/mrhoseah/dolphin/internal/config"
	"github.com/mrhoseah/dolphin/internal/database"
	"github.com/mrhoseah/dolphin/internal/logger"
	"github.com/mrhoseah/dolphin/internal/router"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	version = "1.0.0"
	cfg     *config.Config
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "dolphin",
		Short: "🐬 Dolphin Framework - Enterprise-grade Go web framework",
		Long:  "Dolphin is a rapid development web framework written in Go, inspired by Laravel, CodeIgniter, and CakePHP.",
	}

	// Serve command
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the development server",
		Run:   serve,
	}

	// Migration commands
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Run:   migrate,
	}

	var rollbackCmd = &cobra.Command{
		Use:   "rollback",
		Short: "Rollback the last batch of migrations",
		Run:   rollback,
	}

	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		Run:   status,
	}

	// Make commands
	var makeControllerCmd = &cobra.Command{
		Use:   "make:controller [name]",
		Short: "Create a new controller",
		Args:  cobra.ExactArgs(1),
		Run:   makeController,
	}

	var makeModelCmd = &cobra.Command{
		Use:   "make:model [name]",
		Short: "Create a new model",
		Args:  cobra.ExactArgs(1),
		Run:   makeModel,
	}

	var makeMigrationCmd = &cobra.Command{
		Use:   "make:migration [name]",
		Short: "Create a new migration",
		Args:  cobra.ExactArgs(1),
		Run:   makeMigration,
	}

	var makeMiddlewareCmd = &cobra.Command{
		Use:   "make:middleware [name]",
		Short: "Create a new middleware",
		Args:  cobra.ExactArgs(1),
		Run:   makeMiddleware,
	}

	// Swagger command
	var swaggerCmd = &cobra.Command{
		Use:   "swagger",
		Short: "Generate Swagger documentation",
		Run:   generateSwagger,
	}

	// Add commands
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(makeControllerCmd)
	rootCmd.AddCommand(makeModelCmd)
	rootCmd.AddCommand(makeMigrationCmd)
	rootCmd.AddCommand(makeMiddlewareCmd)
	rootCmd.AddCommand(swaggerCmd)

	// Initialize configuration
	var err error
	cfg, err = config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func serve(cmd *cobra.Command, args []string) {
	// Initialize logger
	logger := logger.New(cfg.Log.Level, cfg.Log.Format)

	// Initialize database
	db, err := database.New(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Initialize application
	app := app.New(cfg, logger, db)

	// Initialize router
	r := router.New(app)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("🚀 Dolphin server running at http://localhost:8080")
		logger.Info("📚 API Documentation available at http://localhost:8080/swagger/index.html")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func migrate(cmd *cobra.Command, args []string) {
	logger := logger.New(cfg.Log.Level, cfg.Log.Format)
	db, err := database.New(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	migrator := database.NewMigrator(db.GetSQLDB(), "migrations")
	result := migrator.Migrate()
	
	if result.Message != "" {
		logger.Info(result.Message)
	}
	if len(result.Executed) > 0 {
		logger.Info("Executed migrations", zap.Any("migrations", result.Executed))
		logger.Info("Batch", zap.Int("batch", result.Batch))
	}
}

func rollback(cmd *cobra.Command, args []string) {
	logger := logger.New(cfg.Log.Level, cfg.Log.Format)
	db, err := database.New(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	migrator := database.NewMigrator(db.GetSQLDB(), "migrations")
	result := migrator.Rollback()
	
	logger.Info(result.Message)
	if len(result.RolledBack) > 0 {
		logger.Info("Rolled back migrations", zap.Any("migrations", result.RolledBack))
		logger.Info("Batch", zap.Int("batch", result.Batch))
	}
}

func status(cmd *cobra.Command, args []string) {
	logger := logger.New(cfg.Log.Level, cfg.Log.Format)
	db, err := database.New(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	migrator := database.NewMigrator(db.GetSQLDB(), "migrations")
	status := migrator.Status()
	
	logger.Info("Migration Status:")
	for _, s := range status {
		logger.Info("Migration status", zap.String("migration", s.Migration), zap.String("status", s.Status), zap.Any("batch", s.Batch))
	}
}

func makeController(cmd *cobra.Command, args []string) {
	name := args[0]
	generator := app.NewGenerator()
	if err := generator.CreateController(name); err != nil {
		log.Fatal("Failed to create controller:", err)
	}
	fmt.Printf("✅ Controller %s created successfully!\n", name)
}

func makeModel(cmd *cobra.Command, args []string) {
	name := args[0]
	generator := app.NewGenerator()
	if err := generator.CreateModel(name); err != nil {
		log.Fatal("Failed to create model:", err)
	}
	fmt.Printf("✅ Model %s created successfully!\n", name)
}

func makeMigration(cmd *cobra.Command, args []string) {
	name := args[0]
	generator := app.NewGenerator()
	if err := generator.CreateMigration(name); err != nil {
		log.Fatal("Failed to create migration:", err)
	}
	fmt.Printf("✅ Migration %s created successfully!\n", name)
}

func makeMiddleware(cmd *cobra.Command, args []string) {
	name := args[0]
	generator := app.NewGenerator()
	if err := generator.CreateMiddleware(name); err != nil {
		log.Fatal("Failed to create middleware:", err)
	}
	fmt.Printf("✅ Middleware %s created successfully!\n", name)
}

func generateSwagger(cmd *cobra.Command, args []string) {
	fmt.Println("📚 Generating Swagger documentation...")
	fmt.Println("Run: swag init -g main.go")
	fmt.Println("Then visit: http://localhost:8080/swagger/index.html")
}
`

	if err := os.WriteFile(filepath.Join(projectName, "main.go"), []byte(mainGoContent), 0644); err != nil {
		fmt.Printf("❌ Failed to create main.go: %v\n", err)
		return
	}

	// Create config file
	configContent := `server:
  port: 8080
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 120

database:
  driver: "sqlite"
  host: "localhost"
  port: 3306
  database: "dolphin.db"
  username: ""
  password: ""
  charset: "utf8mb4"
  parse_time: true
  loc: "Local"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600

jwt:
  secret: "your-secret-key-change-this-in-production"
  expires_in: 24

cache:
  driver: "memory"
  redis:
    host: "localhost"
    port: 6379
    password: ""
    db: 0

session:
  driver: "cookie"
  lifetime: 120
  encrypt: true
  same_site: "lax"

log:
  level: "info"
  format: "json"
`

	if err := os.WriteFile(filepath.Join(projectName, "config", "config.yaml"), []byte(configContent), 0644); err != nil {
		fmt.Printf("❌ Failed to create config file: %v\n", err)
		return
	}

	// Create .env file
	envContent := `APP_NAME=Dolphin App
APP_ENV=local
APP_DEBUG=true
APP_URL=http://localhost:8080

DB_CONNECTION=sqlite
DB_HOST=localhost
DB_PORT=3306
DB_DATABASE=dolphin.db
DB_USERNAME=
DB_PASSWORD=

JWT_SECRET=your-secret-key-change-this-in-production
JWT_EXPIRES_IN=24

CACHE_DRIVER=memory
SESSION_DRIVER=cookie
`

	if err := os.WriteFile(filepath.Join(projectName, ".env"), []byte(envContent), 0644); err != nil {
		fmt.Printf("❌ Failed to create .env file: %v\n", err)
		return
	}

	// Create README
	readmeContent := fmt.Sprintf(`# %s

A Dolphin Framework application.

## Getting Started

1. Install dependencies:
   `+"```"+`bash
   go mod tidy
   `+"```"+`

2. Start the development server:
   `+"```"+`bash
   dolphin serve
   `+"```"+`

3. Visit http://localhost:8080

## Available Commands

- `+"`"+`dolphin serve`+"`"+` - Start the development server
- `+"`"+`dolphin migrate`+"`"+` - Run database migrations
- `+"`"+`dolphin make:controller [name]`+"`"+` - Create a new controller
- `+"`"+`dolphin make:model [name]`+"`"+` - Create a new model
- `+"`"+`dolphin make:migration [name]`+"`"+` - Create a new migration
- `+"`"+`dolphin swagger`+"`"+` - Generate API documentation

## Documentation

Visit https://github.com/mrhoseah/dolphin for complete documentation.
`, projectName)

	if err := os.WriteFile(filepath.Join(projectName, "README.md"), []byte(readmeContent), 0644); err != nil {
		fmt.Printf("❌ Failed to create README: %v\n", err)
		return
	}
}

func listCommands(cmd *cobra.Command, args []string) {
	fmt.Println("🐬 Dolphin Framework CLI Commands")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("📁 Project Management:")
	fmt.Println("  dolphin new [name]           Create a new Dolphin application")
	fmt.Println("  dolphin list                 List all available commands")
	fmt.Println("  dolphin version              Show version information")
	fmt.Println()
	fmt.Println("🚀 Development:")
	fmt.Println("  dolphin serve                Start the development server")
	fmt.Println("  dolphin serve --port 3000    Start server on specific port")
	fmt.Println()
	fmt.Println("🗄️  Database:")
	fmt.Println("  dolphin migrate              Run database migrations")
	fmt.Println("  dolphin migrate --force       Force migration without confirmation")
	fmt.Println("  dolphin rollback              Rollback the last batch of migrations")
	fmt.Println("  dolphin rollback --steps 3   Rollback multiple batches")
	fmt.Println("  dolphin status               Show migration status")
	fmt.Println("  dolphin fresh                Drop all tables and re-run migrations")
	fmt.Println("  dolphin db:seed              Run database seeders")
	fmt.Println("  dolphin db:wipe              Drop all tables")
	fmt.Println()
	fmt.Println("🔨 Code Generation:")
	fmt.Println("  dolphin make:controller User     Create a new controller")
	fmt.Println("  dolphin make:controller User --resource --api  Create resource controller")
	fmt.Println("  dolphin make:model User          Create a new model")
	fmt.Println("  dolphin make:model User --migration --factory  Create model with migration")
	fmt.Println("  dolphin make:migration create_users_table  Create a new migration")
	fmt.Println("  dolphin make:middleware Auth     Create a new middleware")
	fmt.Println("  dolphin make:seeder UserSeeder   Create a new seeder")
	fmt.Println("  dolphin make:request UserRequest Create a new form request")
	fmt.Println()
	fmt.Println("📚 Documentation:")
	fmt.Println("  dolphin swagger               Generate Swagger documentation")
	fmt.Println()
	fmt.Println("💾 Cache:")
	fmt.Println("  dolphin cache:clear           Clear application cache")
	fmt.Println("  dolphin cache:warm            Warm up application cache")
	fmt.Println()
	fmt.Println("🛣️  Routes:")
	fmt.Println("  dolphin route:list            List all registered routes")
	fmt.Println()
	fmt.Println("🔑 Security:")
	fmt.Println("  dolphin key:generate          Generate application key")
	fmt.Println()
	fmt.Println("For more information, visit: https://github.com/mrhoseah/dolphin")
}

func showVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("🐬 Dolphin Framework CLI v%s\n", version)
	fmt.Println("Built with ❤️  using Go")
	fmt.Println("https://github.com/mrhoseah/dolphin")
}
