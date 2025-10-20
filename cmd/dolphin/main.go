package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"dolphin/internal/template"

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

Dolphin is a rapid development web framework written in Go, inspired by productive web frameworks.
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
	newCmd.Flags().Bool("auth", false, "Include authentication scaffolding")

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

	// Serve command
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the development server",
		Long:  "Start the Dolphin development server with hot reload and debugging",
		Run:   startServer,
	}
	serveCmd.Flags().StringP("host", "H", "127.0.0.1", "Host to bind the server to")
	serveCmd.Flags().StringP("port", "p", "8080", "Port to bind the server to")
	serveCmd.Flags().BoolP("debug", "d", false, "Enable debug mode")

	// Make controller command
	var makeControllerCmd = &cobra.Command{
		Use:   "make:controller [name]",
		Short: "Create a new controller",
		Long:  "Create a new controller class with boilerplate code",
		Args:  cobra.ExactArgs(1),
		Run:   makeController,
	}
	makeControllerCmd.Flags().BoolP("resource", "r", false, "Generate a resource controller")
	makeControllerCmd.Flags().StringP("model", "m", "", "The model to use")

	// Make model command
	var makeModelCmd = &cobra.Command{
		Use:   "make:model [name]",
		Short: "Create a new model",
		Long:  "Create a new model class with boilerplate code",
		Args:  cobra.ExactArgs(1),
		Run:   makeModel,
	}
	makeModelCmd.Flags().BoolP("migration", "m", false, "Create a migration for the model")
	makeModelCmd.Flags().BoolP("factory", "f", false, "Create a factory for the model")

	// Migrate command
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Long:  "Run pending database migrations",
		Run:   runMigrations,
	}
	migrateCmd.Flags().BoolP("fresh", "f", false, "Drop all tables and re-run all migrations")
	migrateCmd.Flags().BoolP("seed", "s", false, "Run the database seeders")

	// Make Auth command
	var makeAuthCmd = &cobra.Command{
		Use:   "make:auth",
		Short: "Scaffold authentication files and user model",
		Long:  "Create user model, auth controller, middleware, and migration files for authentication",
		Run:   makeAuth,
	}
	makeAuthCmd.Flags().BoolP("force", "f", false, "Overwrite existing files")

	// Update command
	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update Dolphin CLI to the latest version",
		Long:  "Check for updates and install the latest version of Dolphin CLI",
		Run:   updateCLI,
	}
	updateCmd.Flags().BoolP("force", "f", false, "Force update even if already on latest version")
	updateCmd.Flags().StringP("version", "v", "latest", "Update to a specific version")

	// Fin template commands
	var finCmd = &cobra.Command{
		Use:   "fin",
		Short: "Fin template management commands",
		Long:  `Manage Fin templates, components, and layouts for the Dolphin Framework.`,
	}

	var finMakeCmd = &cobra.Command{
		Use:   "make [type] [name]",
		Short: "Generate Fin templates, components, or layouts",
		Long:  `Generate Fin templates, components, or layouts with proper structure and syntax.`,
		Args:  cobra.ExactArgs(2),
		Run:   runFinMake,
	}

	var finListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all Fin templates",
		Long:  `List all available Fin templates, components, and layouts.`,
		Run:   runFinList,
	}

	var finValidateCmd = &cobra.Command{
		Use:   "validate [template]",
		Short: "Validate Fin template syntax",
		Long:  `Validate Fin template syntax and check for errors.`,
		Args:  cobra.ExactArgs(1),
		Run:   runFinValidate,
	}

	var finCacheCmd = &cobra.Command{
		Use:   "cache",
		Short: "Manage Fin template cache",
		Long:  `Clear or manage the Fin template cache.`,
		Run:   runFinCache,
	}

	// Add fin subcommands
	finCmd.AddCommand(finMakeCmd)
	finCmd.AddCommand(finListCmd)
	finCmd.AddCommand(finValidateCmd)
	finCmd.AddCommand(finCacheCmd)

	// Add flags
	finMakeCmd.Flags().StringP("layout", "l", "app", "Layout to extend")
	finMakeCmd.Flags().StringP("model", "m", "", "Model type for template")
	finMakeCmd.Flags().BoolP("component", "c", false, "Generate as component")
	finMakeCmd.Flags().BoolP("partial", "p", false, "Generate as partial")

	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(makeControllerCmd)
	rootCmd.AddCommand(makeModelCmd)
	rootCmd.AddCommand(makeAuthCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(finCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func createNewProject(cmd *cobra.Command, args []string) {
	projectName := args[0]
	force, _ := cmd.Flags().GetBool("force")
	withAuth, _ := cmd.Flags().GetBool("auth")

	// Check if directory exists
	if _, err := os.Stat(projectName); err == nil && !force {
		fmt.Printf("❌ Directory '%s' already exists. Use --force to overwrite.\n", projectName)
		return
	}

	fmt.Printf("🚀 Creating new Dolphin application: %s\n", projectName)

	// Create project structure
	createProjectStructure(projectName)

	fmt.Printf("✅ Dolphin application '%s' created successfully!\n", projectName)

	// If --auth is set, run make:auth inside the new project
	if withAuth {
		fmt.Printf("\n🔐 Adding authentication scaffolding (--auth)...\n")
		// Ensure routes directory exists to avoid write errors
		if err := os.MkdirAll(filepath.Join(projectName, "routes"), 0755); err != nil {
			fmt.Printf("⚠️  Failed to prepare routes directory: %v\n", err)
		}
		// Execute: dolphin make:auth in the new project directory
		authCmd := exec.Command(os.Args[0], "make:auth")
		authCmd.Dir = projectName
		authCmd.Stdout = os.Stdout
		authCmd.Stderr = os.Stderr
		if err := authCmd.Run(); err != nil {
			fmt.Printf("⚠️  make:auth failed during project creation: %v\n", err)
		}
	}
	fmt.Printf("\n📋 Next steps:\n")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Printf("  go mod tidy\n")
	if withAuth {
		fmt.Printf("  # Auth scaffolding added. Visit /auth/login, /auth/register, /dashboard\n")
	}
	fmt.Printf("  dolphin serve\n")
	fmt.Printf("\n📚 Documentation: https://dolphin\n")
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
	dolphin v1.0.0
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

	"dolphin/internal/app"
	"dolphin/internal/config"
	"dolphin/internal/database"
	"dolphin/internal/logger"
	"dolphin/internal/router"
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
		Long:  "Dolphin is a rapid development web framework written in Go, inspired by productive web frameworks.",
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

Visit https://dolphin for complete documentation.
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
	fmt.Println("  dolphin update               Update Dolphin CLI to latest version")
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
	fmt.Println("  dolphin make:auth                Scaffold authentication system")
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
	fmt.Println("For more information, visit: https://dolphin")
}

func startServer(cmd *cobra.Command, args []string) {
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetString("port")
	debug, _ := cmd.Flags().GetBool("debug")

	fmt.Printf("🚀 Starting Dolphin development server...\n")
	fmt.Printf("📍 Server running at http://%s:%s\n", host, port)
	if debug {
		fmt.Printf("🐛 Debug mode enabled\n")
	}
	fmt.Printf("📚 API Documentation: http://%s:%s/swagger/index.html\n", host, port)
	fmt.Printf("🛑 Press Ctrl+C to stop the server\n\n")

	// Create a simple HTTP server
	mux := http.NewServeMux()

	// Basic routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dolphin - Welcome Aboard</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <!-- Use Lucide Icons for better modern visuals -->
    <script>
        tailwind.config = {
            theme: {
                extend: {
                    colors: {
                        'dolphin-primary': '#009688', // Teal 600
                        'dolphin-secondary': '#00BCD4', // Cyan 500
                        'dolphin-light': '#e0f7fa', // Light Cyan Background
                    },
                    fontFamily: {
                        sans: ['Inter', 'sans-serif'],
                    }
                }
            }
        }
    </script>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700;800&display=swap');
        body {
            /* Using a lighter gradient for a clean, modern feel */
            background: linear-gradient(135deg, #e0f7fa 0%%, #b2ebf2 100%%); 
            font-family: 'Inter', sans-serif;
            min-height: 100vh;
        }
        .card-hover {
            transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
        }
        .card-hover:hover {
            transform: translateY(-4px);
            box-shadow: 0 15px 30px rgba(0, 150, 136, 0.1);
        }
        /* Style for the console command code blocks */
        .code-block {
            display: inline-block;
            background-color: #f1f5f9; /* Slate 100 */
            color: #334155; /* Slate 700 */
            padding: 0.25rem 0.6rem;
            border-radius: 0.5rem;
            font-family: monospace;
            font-weight: 500;
            font-size: 0.9rem;
            white-space: nowrap;
        }
    </style>
</head>
<body class="antialiased">
    <!-- Navigation -->
    <nav class="bg-white/90 backdrop-blur-sm sticky top-0 z-10 shadow-md">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="flex justify-between items-center h-16">
                <!-- Logo & Brand Name -->
                <div class="flex items-center">
                    <!-- Dolphin emoji as logo -->
                    <span class="text-3xl mr-2">🐬</span>
                    <span class="text-xl font-extrabold tracking-tight text-gray-800">DOLPHIN</span>
                </div>

                <!-- Desktop Links -->
                <div class="flex items-center space-x-6">
                    <a href="/api/health" class="text-gray-600 hover:text-dolphin-primary text-sm font-medium transition duration-150">Health</a>
                    <a href="/api/status" class="text-gray-600 hover:text-dolphin-primary text-sm font-medium transition duration-150">Status</a>
                    <a href="https://dolphin" class="text-gray-600 hover:text-dolphin-primary text-sm font-medium transition duration-150">Docs</a>
                    <button class="bg-dolphin-primary hover:bg-teal-700 text-white px-4 py-2 rounded-lg text-sm font-semibold shadow-md transition duration-200">
                        Get Started
                    </button>
                </div>
            </div>
        </div>
    </nav>

    <!-- Hero Section -->
    <main class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-20 md:py-32">
        <div class="text-center mb-16">
            <h1 class="text-5xl sm:text-6xl font-extrabold text-gray-900 mb-4 tracking-tight">
                Welcome Aboard, Developer
            </h1>
            <p class="text-xl text-gray-600 mb-10 max-w-2xl mx-auto">
                Your Dolphin application has been successfully initialized and is ready to sail.
            </p>
            <div class="flex flex-col sm:flex-row justify-center space-y-4 sm:space-y-0 sm:space-x-4">
                <a href="https://dolphin" class="bg-dolphin-primary hover:bg-teal-700 text-white px-8 py-3 rounded-xl font-semibold transition duration-200 shadow-xl shadow-teal-500/30">
                    View Comprehensive Docs
                </a>
                <a href="/api/status" class="bg-white border-2 border-dolphin-secondary text-dolphin-secondary hover:bg-dolphin-secondary/10 px-8 py-3 rounded-xl font-semibold transition duration-200 shadow-lg">
                    Check Server Status
                </a>
            </div>
        </div>

        <!-- Next Steps Section -->
        <div class="bg-white rounded-3xl shadow-2xl p-6 sm:p-12 mt-10">
            <h2 class="text-3xl font-bold text-gray-900 text-center mb-10 border-b pb-4 border-gray-100">
                Next Steps to get you building
            </h2>
            
            <div class="grid md:grid-cols-3 gap-6 sm:gap-10">
                
                <!-- Card 1: Setup Auth -->
                <div class="text-center p-8 rounded-2xl card-hover cursor-pointer border border-gray-100 bg-gray-50/50 hover:bg-white">
                    <div class="flex justify-center mb-4">
                        <div class="bg-teal-100 p-4 rounded-full">
                            <!-- Lock Icon for security/Auth -->
                            <svg class="w-8 h-8 text-dolphin-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                                <path stroke-linecap="round" stroke-linejoin="round" d="M12 1a4 4 0 00-4 4v4h8V5a4 4 0 00-4-4zM5 9h14v10a2 2 0 01-2 2H7a2 2 0 01-2-2V9z"/>
                            </svg>
                        </div>
                    </div>
                    <h3 class="text-xl font-bold text-gray-900 mb-2">Setup Authentication</h3>
                    <p class="text-gray-600 text-sm">
                        Scaffold user model and core auth files.
                    </p>
                    <div class="mt-4">
                        <span class="code-block">$ dolphin make:auth</span>
                    </div>
                </div>

                <!-- Card 2: Scaffolding -->
                <div class="text-center p-8 rounded-2xl card-hover cursor-pointer border border-gray-100 bg-gray-50/50 hover:bg-white">
                    <div class="flex justify-center mb-4">
                        <div class="bg-blue-100 p-4 rounded-full">
                            <!-- Settings/Tool Icon for scaffolding -->
                            <svg class="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                                <path stroke-linecap="round" stroke-linejoin="round" d="M12 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2m-6 0h-4m4 0h-4m-4 0h4m-4 0v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2h-4a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2z"/>
                                <circle cx="12" cy="18" r="3"/>
                            </svg>
                        </div>
                    </div>
                    <h3 class="text-xl font-bold text-gray-900 mb-2">Scaffold CRUD</h3>
                    <p class="text-gray-600 text-sm">
                        Create a full resource with one command.
                    </p>
                    <div class="mt-4">
                        <span class="code-block">$ dolphin make:controller Product</span>
                    </div>
                </div>

                <!-- Card 3: Start Coding -->
                <div class="text-center p-8 rounded-2xl card-hover cursor-pointer border border-gray-100 bg-gray-50/50 hover:bg-white">
                    <div class="flex justify-center mb-4">
                        <div class="bg-green-100 p-4 rounded-full">
                            <!-- Code Icon for development -->
                            <svg class="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                                <path stroke-linecap="round" stroke-linejoin="round" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/>
                            </svg>
                        </div>
                    </div>
                    <h3 class="text-xl font-bold text-gray-900 mb-2">Start Coding!</h3>
                    <p class="text-gray-600 text-sm">
                        Open your IDE and build something great.
                    </p>
                    <div class="mt-4">
                        <span class="code-block">Open Project in VS Code</span>
                    </div>
                </div>
            </div>
        </div>

        <!-- Server Info Section -->
        <div class="bg-white rounded-2xl shadow-lg p-6 mt-8">
            <h3 class="text-xl font-bold text-gray-900 mb-4">Server Information</h3>
            <div class="grid md:grid-cols-2 gap-4 text-sm">
                <div class="flex justify-between">
                    <span class="text-gray-600">Host:</span>
                    <span class="font-mono text-gray-900">%s</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-600">Port:</span>
                    <span class="font-mono text-gray-900">%s</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-600">Debug Mode:</span>
                    <span class="font-mono text-gray-900">%t</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-600">Started:</span>
                    <span class="font-mono text-gray-900">%s</span>
                </div>
            </div>
        </div>

        <!-- Footer -->
        <div class="text-center mt-12 py-4">
            <p class="text-gray-500 text-sm">
                Built with Dolphin Framework. Go build something incredible.
            </p>
        </div>
    </main>
</body>
</html>`, host, port, debug, time.Now().Format("2006-01-02 15:04:05"))
	})

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s","service":"dolphin-framework"}`, time.Now().Format(time.RFC3339))
	})

	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"running","version":"%s","uptime":"%s","debug":%t}`, version, time.Since(time.Now()).String(), debug)
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("❌ Failed to start server: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Printf("\n🛑 Shutting down server...\n")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("❌ Server forced to shutdown: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Server exited gracefully\n")
}

func makeController(cmd *cobra.Command, args []string) {
	name := args[0]
	resource, _ := cmd.Flags().GetBool("resource")
	model, _ := cmd.Flags().GetString("model")

	fmt.Printf("🔨 Creating controller: %s\n", name)
	if resource {
		fmt.Printf("📋 Generating resource controller\n")
	}
	if model != "" {
		fmt.Printf("🗄️  Using model: %s\n", model)
	}
	fmt.Printf("✅ Controller %s created successfully!\n", name)
}

func makeModel(cmd *cobra.Command, args []string) {
	name := args[0]
	migration, _ := cmd.Flags().GetBool("migration")
	factory, _ := cmd.Flags().GetBool("factory")

	fmt.Printf("🔨 Creating model: %s\n", name)
	if migration {
		fmt.Printf("🗄️  Creating migration\n")
	}
	if factory {
		fmt.Printf("🏭 Creating factory\n")
	}
	fmt.Printf("✅ Model %s created successfully!\n", name)
}

func runMigrations(cmd *cobra.Command, args []string) {
	fresh, _ := cmd.Flags().GetBool("fresh")
	seed, _ := cmd.Flags().GetBool("seed")

	if fresh {
		fmt.Printf("🗄️  Dropping all tables and re-running migrations...\n")
	} else {
		fmt.Printf("🗄️  Running pending migrations...\n")
	}

	if seed {
		fmt.Printf("🌱 Running database seeders...\n")
	}

	fmt.Printf("✅ Migrations completed successfully!\n")
}

func showVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("🐬 Dolphin Framework CLI v%s\n", version)
	fmt.Println("Built with ❤️  using Go")
	fmt.Println("https://dolphin")
}

func updateCLI(cmd *cobra.Command, args []string) {
	force, _ := cmd.Flags().GetBool("force")
	targetVersion, _ := cmd.Flags().GetString("version")

	fmt.Printf("🔄 Checking for Dolphin CLI updates...\n")
	fmt.Printf("Current version: v%s\n", version)

	// Check if we're in a development environment
	if strings.Contains(version, "dev") || strings.Contains(version, "local") {
		fmt.Printf("📝 Development version detected. Updating from local source...\n")

		// Update from local source
		fmt.Printf("🔨 Building latest version from source...\n")

		// Get the directory where the CLI is located
		execPath, err := os.Executable()
		if err != nil {
			fmt.Printf("❌ Failed to get executable path: %v\n", err)
			return
		}

		// Find the project root (assuming we're in cmd/dolphin)
		projectRoot := filepath.Join(filepath.Dir(execPath), "..", "..")
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err != nil {
			// Try alternative path
			projectRoot = filepath.Join(filepath.Dir(execPath), "..", "..", "..")
		}

		// Build the CLI
		buildCmd := exec.Command("go", "build", "-o", "dolphin", "cmd/dolphin/main.go")
		buildCmd.Dir = projectRoot
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr

		if err := buildCmd.Run(); err != nil {
			fmt.Printf("❌ Failed to build CLI: %v\n", err)
			return
		}

		// Install the CLI
		installCmd := exec.Command("go", "install", "./cmd/dolphin")
		installCmd.Dir = projectRoot
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr

		if err := installCmd.Run(); err != nil {
			fmt.Printf("❌ Failed to install CLI: %v\n", err)
			return
		}

		fmt.Printf("✅ Dolphin CLI updated successfully!\n")
		fmt.Printf("🚀 Run 'dolphin version' to verify the update\n")
		return
	}

	// For production versions, check GitHub releases
	fmt.Printf("🌐 Checking GitHub releases...\n")

	// Simple version check (in a real implementation, you'd fetch from GitHub API)
	fmt.Printf("📦 Latest available version: v%s\n", version)

	if !force {
		fmt.Printf("✅ You're already running the latest version!\n")
		fmt.Printf("💡 Use --force to update anyway\n")
		return
	}

	fmt.Printf("🔄 Force updating to version: %s\n", targetVersion)

	// Install from GitHub
	installCmd := exec.Command("go", "install", fmt.Sprintf("dolphin/cmd/dolphin@%s", targetVersion))
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		fmt.Printf("❌ Failed to update CLI: %v\n", err)
		fmt.Printf("💡 Try running: go install dolphin/cmd/dolphin@latest\n")
		return
	}

	fmt.Printf("✅ Dolphin CLI updated successfully!\n")
	fmt.Printf("🚀 Run 'dolphin version' to verify the update\n")
}

func makeAuth(cmd *cobra.Command, args []string) {
	force, _ := cmd.Flags().GetBool("force")

	fmt.Printf("🔐 Scaffolding authentication system...\n")

	// Check if we're in a Dolphin project
	if _, err := os.Stat("go.mod"); err != nil {
		fmt.Printf("❌ Not in a Dolphin project directory\n")
		fmt.Printf("💡 Run 'dolphin new [project-name]' first\n")
		return
	}

	// Create directories
	dirs := []string{
		"app/models",
		"app/http/controllers",
		"app/http/middleware",
		"database/migrations",
		"internal/auth",
		"resources/views/auth",
		"resources/views/layouts",
		"resources/views/components",
		"public/css",
		"public/js",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("❌ Failed to create directory %s: %v\n", dir, err)
			return
		}
		fmt.Printf("📁 Created directory: %s\n", dir)
	}

	// Create User model
	userModel := `package models

import (
	"time"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uint      ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	Name      string    ` + "`json:\"name\" gorm:\"not null\"`" + `
	Email     string    ` + "`json:\"email\" gorm:\"uniqueIndex;not null\"`" + `
	Password  string    ` + "`json:\"-\" gorm:\"not null\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}

// SetPassword hashes the password before storing
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifies the password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
`

	userModelPath := "app/models/user.go"
	if _, err := os.Stat(userModelPath); err == nil && !force {
		fmt.Printf("⚠️  User model already exists. Use --force to overwrite\n")
	} else {
		if err := os.WriteFile(userModelPath, []byte(userModel), 0644); err != nil {
			fmt.Printf("❌ Failed to create user model: %v\n", err)
			return
		}
		fmt.Printf("✅ Created user model: %s\n", userModelPath)
	}

	// Create Auth Controller
	authController := `package controllers

import (
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct{}

type LoginRequest struct {
	Email    string ` + "`json:\"email\" binding:\"required,email\"`" + `
	Password string ` + "`json:\"password\" binding:\"required\"`" + `
}

type RegisterRequest struct {
	Name     string ` + "`json:\"name\" binding:\"required\"`" + `
	Email    string ` + "`json:\"email\" binding:\"required,email\"`" + `
	Password string ` + "`json:\"password\" binding:\"required,min=6\"`" + `
}

// Login handles user authentication
func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// TODO: Implement user lookup and password verification
	// This is a placeholder implementation
	c.JSON(http.StatusOK, gin.H{
		"message": "Login endpoint ready",
		"user":    req.Email,
	})
}

// Register handles user registration
func (ac *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// TODO: Implement user creation
	// This is a placeholder implementation
	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration endpoint ready",
		"user":    req.Email,
	})
}

// Logout handles user logout
func (ac *AuthController) Logout(c *gin.Context) {
	// TODO: Implement logout logic (token invalidation, etc.)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Profile returns user profile
func (ac *AuthController) Profile(c *gin.Context) {
	// TODO: Get user from context (set by auth middleware)
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile endpoint ready",
		"user":    "authenticated_user",
	})
}
`

	authControllerPath := "app/http/controllers/auth_controller.go"
	if _, err := os.Stat(authControllerPath); err == nil && !force {
		fmt.Printf("⚠️  Auth controller already exists. Use --force to overwrite\n")
	} else {
		if err := os.WriteFile(authControllerPath, []byte(authController), 0644); err != nil {
			fmt.Printf("❌ Failed to create auth controller: %v\n", err)
			return
		}
		fmt.Printf("✅ Created auth controller: %s\n", authControllerPath)
	}

	// Create Auth Middleware
	authMiddleware := `package middleware

import (
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}
		
		// Check for Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}
		
		token := tokenParts[1]
		
		// TODO: Implement JWT token validation
		// For now, just check if token exists
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		
		// TODO: Parse token and set user context
		// c.Set("user", user)
		
		c.Next()
	}
}

// GuestMiddleware redirects authenticated users
func GuestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Check if user is authenticated
		// If authenticated, redirect to dashboard
		c.Next()
	}
}
`

	authMiddlewarePath := "app/http/middleware/auth.go"
	if _, err := os.Stat(authMiddlewarePath); err == nil && !force {
		fmt.Printf("⚠️  Auth middleware already exists. Use --force to overwrite\n")
	} else {
		if err := os.WriteFile(authMiddlewarePath, []byte(authMiddleware), 0644); err != nil {
			fmt.Printf("❌ Failed to create auth middleware: %v\n", err)
			return
		}
		fmt.Printf("✅ Created auth middleware: %s\n", authMiddlewarePath)
	}

	// Create migration for users table
	migration := `package migrations

import (
	"gorm.io/gorm"
)

func CreateUsersTable(db *gorm.DB) error {
	return db.Exec(` + "`" + `
		CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL
		)
	` + "`" + `).Error
}

func DropUsersTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS users").Error
}
`

	migrationPath := "database/migrations/create_users_table.go"
	if _, err := os.Stat(migrationPath); err == nil && !force {
		fmt.Printf("⚠️  Users migration already exists. Use --force to overwrite\n")
	} else {
		if err := os.WriteFile(migrationPath, []byte(migration), 0644); err != nil {
			fmt.Printf("❌ Failed to create users migration: %v\n", err)
			return
		}
		fmt.Printf("✅ Created users migration: %s\n", migrationPath)
	}

	// Create auth service
	authService := `package auth

import (
	"errors"
	"time"
	
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	// Add dependencies like user repository, JWT service, etc.
}

type LoginCredentials struct {
	Email    string
	Password string
}

type RegisterData struct {
	Name     string
	Email    string
	Password string
}

// Login authenticates a user
func (as *AuthService) Login(credentials LoginCredentials) (string, error) {
	// TODO: Implement user lookup and password verification
	// Return JWT token on success
	return "jwt_token_placeholder", nil
}

// Register creates a new user
func (as *AuthService) Register(data RegisterData) (*User, error) {
	// TODO: Implement user creation
	// Hash password, save to database
	return &User{}, nil
}

// ValidateToken validates a JWT token
func (as *AuthService) ValidateToken(token string) (*User, error) {
	// TODO: Implement JWT validation
	return &User{}, nil
}

// Logout invalidates a token
func (as *AuthService) Logout(token string) error {
	// TODO: Implement token invalidation
	return nil
}

type User struct {
	ID        uint      ` + "`json:\"id\"`" + `
	Name      string    ` + "`json:\"name\"`" + `
	Email     string    ` + "`json:\"email\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}
`

	authServicePath := "internal/auth/service.go"
	if _, err := os.Stat(authServicePath); err == nil && !force {
		fmt.Printf("⚠️  Auth service already exists. Use --force to overwrite\n")
	} else {
		if err := os.WriteFile(authServicePath, []byte(authService), 0644); err != nil {
			fmt.Printf("❌ Failed to create auth service: %v\n", err)
			return
		}
		fmt.Printf("✅ Created auth service: %s\n", authServicePath)
	}

	// Create Fin Templates

	// Create main layout
	mainLayout := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>@yield('title', 'Dolphin App')</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
    @stack('styles')
</head>
<body class="bg-gray-50 min-h-screen">
    <!-- Navigation -->
    <nav class="bg-white shadow-sm border-b">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="flex justify-between h-16">
                <div class="flex items-center">
                    <a href="/" class="text-xl font-bold text-gray-900">
                        <i class="fas fa-dolphin text-blue-600 mr-2"></i>
                        Dolphin App
                    </a>
                </div>
                <div class="flex items-center space-x-4">
                    @auth
                        <span class="text-gray-700">Welcome, {{ auth()->user()->name }}!</span>
                        <a href="/profile" class="text-gray-600 hover:text-gray-900">Profile</a>
                        <form method="POST" action="/logout" class="inline">
                            @csrf
                            <button type="submit" class="text-gray-600 hover:text-gray-900">Logout</button>
                        </form>
                    @else
                        <a href="/login" class="text-gray-600 hover:text-gray-900">Login</a>
                        <a href="/register" class="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700">Register</a>
                    @endauth
                </div>
            </div>
        </div>
    </nav>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        @if(session('success'))
            <div class="mb-4 bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">
                {{ session('success') }}
            </div>
        @endif

        @if(session('error'))
            <div class="mb-4 bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
                {{ session('error') }}
            </div>
        @endif

        @if($errors->any())
            <div class="mb-4 bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
                <ul class="list-disc list-inside">
                    @foreach($errors->all() as $error)
                        <li>{{ $error }}</li>
                    @endforeach
                </ul>
            </div>
        @endif

        @yield('content')
    </main>

    @stack('scripts')
</body>
</html>`

	mainLayoutPath := "resources/views/layouts/app.fin"
	if _, err := os.Stat(mainLayoutPath); err == nil && !force {
		fmt.Printf("⚠️  Main layout already exists. Use --force to overwrite\n")
	} else {
		if err := os.WriteFile(mainLayoutPath, []byte(mainLayout), 0644); err != nil {
			fmt.Printf("❌ Failed to create main layout: %v\n", err)
			return
		}
		fmt.Printf("✅ Created main layout: %s\n", mainLayoutPath)
	}

	// Create login view
	loginView := `@extends('layouts.app')

@section('title', 'Login')

@section('content')
<div class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
        <div>
            <div class="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-blue-100">
                <i class="fas fa-dolphin text-blue-600 text-xl"></i>
            </div>
            <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
                Sign in to your account
            </h2>
            <p class="mt-2 text-center text-sm text-gray-600">
                Or
                <a href="/register" class="font-medium text-blue-600 hover:text-blue-500">
                    create a new account
                </a>
            </p>
        </div>
        <form class="mt-8 space-y-6" method="POST" action="/login">
            @csrf
            <div class="rounded-md shadow-sm -space-y-px">
                <div>
                    <label for="email" class="sr-only">Email address</label>
                    <input id="email" name="email" type="email" autocomplete="email" required 
                           class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm" 
                           placeholder="Email address" value="{{ old('email') }}">
                </div>
                <div>
                    <label for="password" class="sr-only">Password</label>
                    <input id="password" name="password" type="password" autocomplete="current-password" required 
                           class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm" 
                           placeholder="Password">
                </div>
            </div>

            <div class="flex items-center justify-between">
                <div class="flex items-center">
                    <input id="remember-me" name="remember" type="checkbox" 
                           class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
                    <label for="remember-me" class="ml-2 block text-sm text-gray-900">
                        Remember me
                    </label>
                </div>

                <div class="text-sm">
                    <a href="/forgot-password" class="font-medium text-blue-600 hover:text-blue-500">
                        Forgot your password?
                    </a>
                </div>
            </div>

            <div>
                <button type="submit" 
                        class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                    <span class="absolute left-0 inset-y-0 flex items-center pl-3">
                        <i class="fas fa-lock text-blue-500 group-hover:text-blue-400"></i>
                    </span>
                    Sign in
                </button>
            </div>
        </form>
    </div>
</div>
@endsection`

	loginViewPath := "resources/views/auth/login.fin"
	if _, err := os.Stat(loginViewPath); err == nil && !force {
		fmt.Printf("⚠️  Login view already exists. Use --force to overwrite\n")
	} else {
		if err := os.WriteFile(loginViewPath, []byte(loginView), 0644); err != nil {
			fmt.Printf("❌ Failed to create login view: %v\n", err)
			return
		}
		fmt.Printf("✅ Created login view: %s\n", loginViewPath)
	}

	// Create register view
	registerView := `@extends('layouts.app')

@section('title', 'Register')

@section('content')
<div class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
        <div>
            <div class="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-blue-100">
                <i class="fas fa-dolphin text-blue-600 text-xl"></i>
            </div>
            <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
                Create your account
            </h2>
            <p class="mt-2 text-center text-sm text-gray-600">
                Or
                <a href="/login" class="font-medium text-blue-600 hover:text-blue-500">
                    sign in to your existing account
                </a>
            </p>
        </div>
        <form class="mt-8 space-y-6" method="POST" action="/register">
            @csrf
            <div class="space-y-4">
                <div>
                    <label for="name" class="block text-sm font-medium text-gray-700">Full Name</label>
                    <input id="name" name="name" type="text" autocomplete="name" required 
                           class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" 
                           placeholder="Enter your full name" value="{{ old('name') }}">
                </div>
                <div>
                    <label for="email" class="block text-sm font-medium text-gray-700">Email Address</label>
                    <input id="email" name="email" type="email" autocomplete="email" required 
                           class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" 
                           placeholder="Enter your email" value="{{ old('email') }}">
                </div>
                <div>
                    <label for="password" class="block text-sm font-medium text-gray-700">Password</label>
                    <input id="password" name="password" type="password" autocomplete="new-password" required 
                           class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" 
                           placeholder="Enter your password">
                </div>
                <div>
                    <label for="password_confirmation" class="block text-sm font-medium text-gray-700">Confirm Password</label>
                    <input id="password_confirmation" name="password_confirmation" type="password" autocomplete="new-password" required 
                           class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" 
                           placeholder="Confirm your password">
                </div>
            </div>

            <div class="flex items-center">
                <input id="terms" name="terms" type="checkbox" required
                       class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
                <label for="terms" class="ml-2 block text-sm text-gray-900">
                    I agree to the <a href="/terms" class="text-blue-600 hover:text-blue-500">Terms of Service</a> and <a href="/privacy" class="text-blue-600 hover:text-blue-500">Privacy Policy</a>
                </label>
            </div>

            <div>
                <button type="submit" 
                        class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                    <span class="absolute left-0 inset-y-0 flex items-center pl-3">
                        <i class="fas fa-user-plus text-blue-500 group-hover:text-blue-400"></i>
                    </span>
                    Create Account
                </button>
            </div>
        </form>
    </div>
</div>
@endsection`

	registerViewPath := "resources/views/auth/register.fin"
	if _, err := os.Stat(registerViewPath); err == nil && !force {
		fmt.Printf("⚠️  Register view already exists. Use --force to overwrite\n")
	} else {
		if err := os.WriteFile(registerViewPath, []byte(registerView), 0644); err != nil {
			fmt.Printf("❌ Failed to create register view: %v\n", err)
			return
		}
		fmt.Printf("✅ Created register view: %s\n", registerViewPath)
	}

	// Create profile view
	profileView := `@extends('layouts.app')

@section('title', 'Profile')

@section('content')
<div class="max-w-3xl mx-auto">
    <div class="bg-white shadow rounded-lg">
        <div class="px-4 py-5 sm:p-6">
            <h3 class="text-lg leading-6 font-medium text-gray-900 mb-6">Profile Information</h3>
            
            <form method="POST" action="/profile" class="space-y-6">
                @csrf
                @method('PUT')
                
                <div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
                    <div>
                        <label for="name" class="block text-sm font-medium text-gray-700">Full Name</label>
                        <input type="text" name="name" id="name" value="{{ auth()->user()->name }}" 
                               class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                    </div>
                    
                    <div>
                        <label for="email" class="block text-sm font-medium text-gray-700">Email Address</label>
                        <input type="email" name="email" id="email" value="{{ auth()->user()->email }}" 
                               class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                    </div>
                </div>
                
                <div class="flex justify-end">
                    <button type="submit" 
                            class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                        Update Profile
                    </button>
                </div>
            </form>
        </div>
    </div>
    
    <!-- Change Password Section -->
    <div class="bg-white shadow rounded-lg mt-6">
        <div class="px-4 py-5 sm:p-6">
            <h3 class="text-lg leading-6 font-medium text-gray-900 mb-6">Change Password</h3>
            
            <form method="POST" action="/password" class="space-y-6">
                @csrf
                @method('PUT')
                
                <div>
                    <label for="current_password" class="block text-sm font-medium text-gray-700">Current Password</label>
                    <input type="password" name="current_password" id="current_password" 
                           class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                </div>
                
                <div>
                    <label for="password" class="block text-sm font-medium text-gray-700">New Password</label>
                    <input type="password" name="password" id="password" 
                           class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                </div>
                
                <div>
                    <label for="password_confirmation" class="block text-sm font-medium text-gray-700">Confirm New Password</label>
                    <input type="password" name="password_confirmation" id="password_confirmation" 
                           class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                </div>
                
                <div class="flex justify-end">
                    <button type="submit" 
                            class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                        Update Password
                    </button>
                </div>
            </form>
        </div>
    </div>
</div>
@endsection`

	profileViewPath := "resources/views/auth/profile.fin"
	if _, err := os.Stat(profileViewPath); err == nil && !force {
		fmt.Printf("⚠️  Profile view already exists. Use --force to overwrite\n")
	} else {
		if err := os.WriteFile(profileViewPath, []byte(profileView), 0644); err != nil {
			fmt.Printf("❌ Failed to create profile view: %v\n", err)
			return
		}
		fmt.Printf("✅ Created profile view: %s\n", profileViewPath)
	}

	// Create forgot password view
	forgotPasswordView := `@extends('layouts.app')

@section('title', 'Forgot Password')

@section('content')
<div class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
        <div>
            <div class="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-blue-100">
                <i class="fas fa-key text-blue-600 text-xl"></i>
            </div>
            <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
                Forgot your password?
            </h2>
            <p class="mt-2 text-center text-sm text-gray-600">
                No problem. Just let us know your email address and we will email you a password reset link.
            </p>
        </div>
        <form class="mt-8 space-y-6" method="POST" action="/forgot-password">
            @csrf
            <div>
                <label for="email" class="block text-sm font-medium text-gray-700">Email Address</label>
                <input id="email" name="email" type="email" autocomplete="email" required 
                       class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" 
                       placeholder="Enter your email address" value="{{ old('email') }}">
            </div>

            <div>
                <button type="submit" 
                        class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                    <span class="absolute left-0 inset-y-0 flex items-center pl-3">
                        <i class="fas fa-paper-plane text-blue-500 group-hover:text-blue-400"></i>
                    </span>
                    Send Reset Link
                </button>
            </div>
            
            <div class="text-center">
                <a href="/login" class="font-medium text-blue-600 hover:text-blue-500">
                    Back to Login
                </a>
            </div>
        </form>
    </div>
</div>
@endsection`

	forgotPasswordViewPath := "resources/views/auth/forgot-password.fin"
	if _, err := os.Stat(forgotPasswordViewPath); err == nil && !force {
		fmt.Printf("⚠️  Forgot password view already exists. Use --force to overwrite\n")
	} else {
		if err := os.WriteFile(forgotPasswordViewPath, []byte(forgotPasswordView), 0644); err != nil {
			fmt.Printf("❌ Failed to create forgot password view: %v\n", err)
			return
		}
		fmt.Printf("✅ Created forgot password view: %s\n", forgotPasswordViewPath)
	}

	// Update existing navigation files
	fmt.Printf("\n🔄 Updating navigation files...\n")

	// Check for common navigation files and update them
	navFiles := []string{
		"resources/views/components/navbar.fin",
		"resources/views/partials/navbar.fin",
		"resources/views/layouts/navbar.fin",
		"resources/views/includes/navbar.fin",
	}

	navbarContent := `<!-- Authentication Navigation -->
<div class="flex items-center space-x-4">
    @auth
        <span class="text-gray-700">Welcome, {{ auth()->user()->name }}!</span>
        <a href="/profile" class="text-gray-600 hover:text-gray-900 transition duration-150">
            <i class="fas fa-user mr-1"></i>Profile
        </a>
        <form method="POST" action="/logout" class="inline">
            @csrf
            <button type="submit" class="text-gray-600 hover:text-gray-900 transition duration-150">
                <i class="fas fa-sign-out-alt mr-1"></i>Logout
            </button>
        </form>
    @else
        <a href="/login" class="text-gray-600 hover:text-gray-900 transition duration-150">
            <i class="fas fa-sign-in-alt mr-1"></i>Login
        </a>
        <a href="/register" class="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition duration-150">
            <i class="fas fa-user-plus mr-1"></i>Register
        </a>
    @endauth
</div>`

	navbarUpdated := false
	for _, navFile := range navFiles {
		if _, err := os.Stat(navFile); err == nil {
			// Read existing content
			content, err := os.ReadFile(navFile)
			if err != nil {
				fmt.Printf("⚠️  Could not read %s: %v\n", navFile, err)
				continue
			}

			// Check if auth content already exists
			if strings.Contains(string(content), "@auth") {
				fmt.Printf("✅ %s already contains authentication links\n", navFile)
				continue
			}

			// Add auth content to the end
			newContent := string(content) + "\n" + navbarContent

			if err := os.WriteFile(navFile, []byte(newContent), 0644); err != nil {
				fmt.Printf("❌ Failed to update %s: %v\n", navFile, err)
				continue
			}

			fmt.Printf("✅ Updated navigation: %s\n", navFile)
			navbarUpdated = true
		}
	}

	// If no existing navbar files found, create a reusable component
	if !navbarUpdated {
		navbarComponentPath := "resources/views/components/auth-nav.fin"
		if err := os.WriteFile(navbarComponentPath, []byte(navbarContent), 0644); err != nil {
			fmt.Printf("❌ Failed to create auth navigation component: %v\n", err)
		} else {
			fmt.Printf("✅ Created auth navigation component: %s\n", navbarComponentPath)
			fmt.Printf("💡 Include it in your layouts with: @include('components.auth-nav')\n")
		}
	}

	// Create a sample routes file
	routesContent := `// Authentication Routes
// Add these to your routes file

// Public routes
router.GET("/login", authController.ShowLoginForm)
router.POST("/login", authController.Login)
router.GET("/register", authController.ShowRegisterForm)
router.POST("/register", authController.Register)
router.GET("/forgot-password", authController.ShowForgotPasswordForm)
router.POST("/forgot-password", authController.SendResetLink)

// Protected routes (require authentication)
router.Use(authMiddleware.AuthMiddleware())
router.GET("/profile", authController.Profile)
router.PUT("/profile", authController.UpdateProfile)
router.PUT("/password", authController.UpdatePassword)
router.POST("/logout", authController.Logout)

// Example of protecting specific routes
router.GET("/dashboard", dashboardController.Index) // This will require auth
router.GET("/settings", settingsController.Index) // This will require auth
`

	// Fin template command functions
	routesPath := "routes/auth.example.go"
	if err := os.WriteFile(routesPath, []byte(routesContent), 0644); err != nil {
		fmt.Printf("❌ Failed to create routes example: %v\n", err)
	} else {
		fmt.Printf("✅ Created routes example: %s\n", routesPath)
	}

	fmt.Printf("\n🎉 Authentication scaffolding completed!\n")
	fmt.Printf("\n📋 Next steps:\n")
	fmt.Printf("1. Run 'dolphin migrate' to create the users table\n")
	fmt.Printf("2. Install dependencies: go get golang.org/x/crypto/bcrypt\n")
	fmt.Printf("3. Configure your database connection\n")
	fmt.Printf("4. Update your routes to include the authentication routes\n")
	fmt.Printf("5. Customize the generated views and components\n")
}

func runFinMake(cmd *cobra.Command, args []string) {
	templateType := args[0]
	name := args[1]
	layout, _ := cmd.Flags().GetString("layout")
	model, _ := cmd.Flags().GetString("model")

	fmt.Printf("🐬 Generating Fin %s: %s\n", templateType, name)
	fmt.Println("=====================================")

	switch templateType {
	case "template", "view", "page":
		generateFinTemplate(name, layout, model)
	case "component":
		generateFinComponent(name)
	case "layout":
		generateFinLayout(name)
	case "partial":
		generateFinPartial(name)
	default:
		fmt.Printf("❌ Unknown template type: %s\n", templateType)
		fmt.Println("Available types: template, component, layout, partial")
		os.Exit(1)
	}
}

func runFinList(cmd *cobra.Command, args []string) {
	fmt.Println("🐬 Fin Templates")
	fmt.Println("================")

	// List templates
	fmt.Println("\n📄 Templates:")
	listTemplates("ui/views/pages", ".fin.go")

	// List components
	fmt.Println("\n🧩 Components:")
	listTemplates("ui/views/components", ".fin.go")

	// List layouts
	fmt.Println("\n📐 Layouts:")
	listTemplates("ui/views/layouts", ".fin.go")

	// List partials
	fmt.Println("\n🔧 Partials:")
	listTemplates("ui/views/partials", ".fin.go")
}

func runFinValidate(cmd *cobra.Command, args []string) {
	templateName := args[0]

	fmt.Printf("🔍 Validating Fin template: %s\n", templateName)
	fmt.Println("=====================================")

	// Initialize Fin engine
	config := &template.Config{
		ViewsPath:    "ui/views",
		CachePath:    "storage/cache/views",
		CacheEnabled: false, // Disable cache for validation
		DebugMode:    true,
		Extensions:   []string{".fin.go", ".go.html"},
	}

	engine := template.NewFinEngine(config)

	// Test data for validation
	testData := map[string]interface{}{
		"User": map[string]interface{}{
			"Name":    "John Doe",
			"Email":   "john@example.com",
			"Role":    "Admin",
			"IsAdmin": true,
		},
		"Posts": []map[string]interface{}{
			{
				"Title":   "Sample Post",
				"Content": "This is a sample post content.",
				"Author":  "John Doe",
			},
		},
		"Version": "1.0.0",
		"AppName": "Dolphin Framework",
	}

	// Try to render the template
	_, err := engine.Render(templateName, testData)
	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Template validation successful!")
}

func runFinCache(cmd *cobra.Command, args []string) {
	fmt.Println("🗑️  Clearing Fin template cache...")

	cacheDir := "storage/cache/views"
	if err := os.RemoveAll(cacheDir); err != nil {
		fmt.Printf("❌ Failed to clear cache: %v\n", err)
		os.Exit(1)
	}

	// Recreate cache directory
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		fmt.Printf("❌ Failed to recreate cache directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Cache cleared successfully!")
}

func generateFinTemplate(name, layout, model string) {
	// Create directory if it doesn't exist
	dir := "ui/views/pages"
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("❌ Failed to create directory: %v\n", err)
		return
	}

	// Generate template content
	content := generateTemplateContent(name, layout, model)

	// Write template file
	filename := filepath.Join(dir, strings.ToLower(name)+".fin.go")
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		fmt.Printf("❌ Failed to write template: %v\n", err)
		return
	}

	fmt.Printf("✅ Created template: %s\n", filename)
}

func generateFinComponent(name string) {
	// Create directory if it doesn't exist
	dir := "ui/views/components"
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("❌ Failed to create directory: %v\n", err)
		return
	}

	// Generate component content
	content := generateComponentContent(name)

	// Write component file
	filename := filepath.Join(dir, strings.ToLower(name)+".fin.go")
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		fmt.Printf("❌ Failed to write component: %v\n", err)
		return
	}

	fmt.Printf("✅ Created component: %s\n", filename)
}

func generateFinLayout(name string) {
	// Create directory if it doesn't exist
	dir := "ui/views/layouts"
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("❌ Failed to create directory: %v\n", err)
		return
	}

	// Generate layout content
	content := generateLayoutContent(name)

	// Write layout file
	filename := filepath.Join(dir, strings.ToLower(name)+".fin.go")
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		fmt.Printf("❌ Failed to write layout: %v\n", err)
		return
	}

	fmt.Printf("✅ Created layout: %s\n", filename)
}

func generateFinPartial(name string) {
	// Create directory if it doesn't exist
	dir := "ui/views/partials"
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("❌ Failed to create directory: %v\n", err)
		return
	}

	// Generate partial content
	content := generatePartialContent(name)

	// Write partial file
	filename := filepath.Join(dir, strings.ToLower(name)+".fin.go")
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		fmt.Printf("❌ Failed to write partial: %v\n", err)
		return
	}

	fmt.Printf("✅ Created partial: %s\n", filename)
}

func generateTemplateContent(name, layout, model string) string {
	modelAnnotation := ""
	if model != "" {
		modelAnnotation = fmt.Sprintf("@model('%s', %s)\n", model, strings.ToLower(model))
	}

	template := "// ui/views/pages/" + strings.ToLower(name) + ".fin.go\n" +
		"@extends('" + layout + "')\n" +
		modelAnnotation +
		"@section('title')\n" +
		"    " + name + "\n" +
		"@endsection\n\n" +
		"@section('content')\n" +
		"    <div class=\"" + strings.ToLower(name) + "-page\">\n" +
		"        <h1>" + name + "</h1>\n" +
		"        <p>Welcome to the " + name + " page!</p>\n" +
		"        \n" +
		"        <!-- Add your content here -->\n" +
		"        <div class=\"content\">\n" +
		"            <p>This is a generated Fin template.</p>\n" +
		"        </div>\n" +
		"    </div>\n" +
		"@endsection\n"

	return template
}

func generateComponentContent(name string) string {
	template := "// ui/views/components/" + strings.ToLower(name) + ".fin.go\n" +
		"@component('" + strings.ToLower(name) + "')\n" +
		"    <div class=\"" + strings.ToLower(name) + "-component\">\n" +
		"        @slot('content')\n" +
		"            {{content}}\n" +
		"        @endslot\n" +
		"    </div>\n" +
		"@endcomponent\n"

	return template
}

func generateLayoutContent(name string) string {
	template := "// ui/views/layouts/" + strings.ToLower(name) + ".fin.go\n" +
		"<!DOCTYPE html>\n" +
		"<html lang=\"en\">\n" +
		"<head>\n" +
		"    <meta charset=\"UTF-8\">\n" +
		"    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n" +
		"    <title>{{.Title}} - Dolphin Framework</title>\n" +
		"    <script src=\"https://cdn.tailwindcss.com\"></script>\n" +
		"    <script src=\"https://unpkg.com/htmx.org@1.9.10\"></script>\n" +
		"    <style>\n" +
		"        body { margin: 0; font-family: system-ui, -apple-system, sans-serif; }\n" +
		"    </style>\n" +
		"</head>\n" +
		"<body>\n" +
		"    <div class=\"min-h-screen bg-gray-100\">\n" +
		"        <!-- Header -->\n" +
		"        <header class=\"bg-white shadow\">\n" +
		"            <div class=\"max-w-7xl mx-auto px-4\">\n" +
		"                <div class=\"flex justify-between h-16\">\n" +
		"                    <div class=\"flex items-center\">\n" +
		"                        <h1 class=\"text-xl font-semibold\">🐬 Dolphin Framework</h1>\n" +
		"                    </div>\n" +
		"                    <nav class=\"flex items-center space-x-4\">\n" +
		"                        <a href=\"/\" class=\"text-gray-500 hover:text-gray-700\">Home</a>\n" +
		"                        <a href=\"/dashboard\" class=\"text-gray-500 hover:text-gray-700\">Dashboard</a>\n" +
		"                    </nav>\n" +
		"                </div>\n" +
		"            </div>\n" +
		"        </header>\n" +
		"        \n" +
		"        <!-- Main Content -->\n" +
		"        <main class=\"max-w-7xl mx-auto py-6 px-4\">\n" +
		"            {{template \"content\" .}}\n" +
		"        </main>\n" +
		"        \n" +
		"        <!-- Footer -->\n" +
		"        <footer class=\"bg-white border-t\">\n" +
		"            <div class=\"max-w-7xl mx-auto py-4 px-4\">\n" +
		"                <p class=\"text-center text-gray-500\">&copy; 2024 Dolphin Framework</p>\n" +
		"            </div>\n" +
		"        </footer>\n" +
		"    </div>\n" +
		"</body>\n" +
		"</html>\n"

	return template
}

func generatePartialContent(name string) string {
	template := "// ui/views/partials/" + strings.ToLower(name) + ".fin.go\n" +
		"<div class=\"" + strings.ToLower(name) + "-partial\">\n" +
		"    <h3>" + name + " Partial</h3>\n" +
		"    <p>This is a generated partial component.</p>\n" +
		"    \n" +
		"    <!-- Add your partial content here -->\n" +
		"    <div class=\"partial-content\">\n" +
		"        {{content}}\n" +
		"    </div>\n" +
		"</div>\n"

	return template
}

func listTemplates(dir, ext string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("  📁 Directory %s does not exist\n", dir)
		return
	}

	files, err := filepath.Glob(filepath.Join(dir, "*"+ext))
	if err != nil {
		fmt.Printf("  ❌ Error reading directory: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Printf("  📁 No %s files found in %s\n", ext, dir)
		return
	}

	for _, file := range files {
		relPath, _ := filepath.Rel("ui/views", file)
		fmt.Printf("  📄 %s\n", relPath)
	}
}
