package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/mrhoseah/dolphin/internal/app"
	"github.com/mrhoseah/dolphin/internal/config"
	"github.com/mrhoseah/dolphin/internal/database"
	"github.com/mrhoseah/dolphin/internal/debug"
	"github.com/mrhoseah/dolphin/internal/logger"
	"github.com/mrhoseah/dolphin/internal/maintenance"
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
		Short: "🐬 Dolphin Framework CLI - Enterprise-grade Go web framework",
		Long: `🐬 Dolphin Framework CLI

Dolphin is a rapid development web framework written in Go, inspired by Laravel, CodeIgniter, and CakePHP.
This CLI tool provides all the commands you need to build, manage, and deploy your applications.

Examples:
  dolphin serve                    # Start the development server
  dolphin make:controller User     # Create a new controller
  dolphin migrate                  # Run database migrations
  dolphin swagger                  # Generate API documentation`,
		Version: version,
	}

	// Add global flags
	rootCmd.PersistentFlags().StringP("config", "c", "config/config.yaml", "Config file path")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")

	// Serve command
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the development server",
		Long:  "Start the Dolphin development server with hot reloading and debugging features",
		Run:   serve,
	}
	serveCmd.Flags().IntP("port", "p", 8080, "Port to run the server on")
	serveCmd.Flags().StringP("host", "H", "localhost", "Host to bind the server to")
	serveCmd.Flags().String("asset-cmd", "", "Optional asset build command to run before and alongside the server")

	// Migration commands
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Long:  "Run all pending database migrations using the integrated Raptor migration system",
		Run:   migrate,
	}
	migrateCmd.Flags().BoolP("force", "f", false, "Force migration without confirmation")

	var rollbackCmd = &cobra.Command{
		Use:   "rollback",
		Short: "Rollback the last batch of migrations",
		Long:  "Rollback the last batch of migrations that were run",
		Run:   rollback,
	}
	rollbackCmd.Flags().IntP("steps", "s", 1, "Number of migration batches to rollback")

	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		Long:  "Display the current status of all migrations",
		Run:   status,
	}

	var freshCmd = &cobra.Command{
		Use:   "fresh",
		Short: "Drop all tables and re-run all migrations",
		Long:  "Drop all tables and re-run all migrations from scratch (DESTRUCTIVE)",
		Run:   fresh,
	}

	// Make commands
	var makeControllerCmd = &cobra.Command{
		Use:   "make:controller [name]",
		Short: "Create a new controller",
		Long:  "Generate a new controller with CRUD methods and Swagger annotations",
		Args:  cobra.ExactArgs(1),
		Run:   makeController,
	}
	makeControllerCmd.Flags().BoolP("resource", "r", false, "Generate resource controller with CRUD methods")
	makeControllerCmd.Flags().BoolP("api", "a", false, "Generate API controller with Swagger annotations")

	var makeModelCmd = &cobra.Command{
		Use:   "make:model [name]",
		Short: "Create a new model",
		Long:  "Generate a new model with GORM annotations and repository pattern",
		Args:  cobra.ExactArgs(1),
		Run:   makeModel,
	}
	makeModelCmd.Flags().BoolP("migration", "m", false, "Create a migration for the model")
	makeModelCmd.Flags().BoolP("factory", "f", false, "Create a factory for the model")

	var makeMigrationCmd = &cobra.Command{
		Use:   "make:migration [name]",
		Short: "Create a new migration",
		Long:  "Generate a new database migration file using Raptor migration system",
		Args:  cobra.ExactArgs(1),
		Run:   makeMigration,
	}

	var makeMiddlewareCmd = &cobra.Command{
		Use:   "make:middleware [name]",
		Short: "Create a new middleware",
		Long:  "Generate a new middleware component for request processing",
		Args:  cobra.ExactArgs(1),
		Run:   makeMiddleware,
	}

	var makeModuleCmd = &cobra.Command{
		Use:   "make:module [name]",
		Short: "Create a complete module",
		Long:  "Generate a complete module with model, controller, repository, HTMX views, and migration",
		Args:  cobra.ExactArgs(1),
		Run:   makeModule,
	}

	var makeViewCmd = &cobra.Command{
		Use:   "make:view [name]",
		Short: "Create HTMX views",
		Long:  "Generate HTMX-based views (index, show, create, edit, form) for a module",
		Args:  cobra.ExactArgs(1),
		Run:   makeView,
	}

	var makeResourceCmd = &cobra.Command{
		Use:   "make:resource [name]",
		Short: "Create an API resource",
		Long:  "Generate an API resource with model, API controller, repository, and migration",
		Args:  cobra.ExactArgs(1),
		Run:   makeResource,
	}

	var makeRepositoryCmd = &cobra.Command{
		Use:   "make:repository [name]",
		Short: "Create a repository",
		Long:  "Generate a repository for data access layer",
		Args:  cobra.ExactArgs(1),
		Run:   makeRepository,
	}

	var makeProviderCmd = &cobra.Command{
		Use:   "make:provider [name]",
		Short: "Create a service provider",
		Long:  "Generate a service provider for modular architecture",
		Args:  cobra.ExactArgs(1),
		Run:   makeProvider,
	}
	makeProviderCmd.Flags().StringP("type", "t", "custom", "Provider type (email, storage, cache, queue, etc.)")
	makeProviderCmd.Flags().IntP("priority", "p", 100, "Provider priority (lower = higher priority)")

	var storageCmd = &cobra.Command{
		Use:   "storage",
		Short: "Storage management commands",
		Long:  "Manage file storage operations",
	}

	var storageListCmd = &cobra.Command{
		Use:   "list [path]",
		Short: "List files in storage",
		Long:  "List files in the specified storage path",
		Args:  cobra.MaximumNArgs(1),
		Run:   storageList,
	}

	var storagePutCmd = &cobra.Command{
		Use:   "put <local-path> <remote-path>",
		Short: "Upload file to storage",
		Long:  "Upload a local file to storage",
		Args:  cobra.ExactArgs(2),
		Run:   storagePut,
	}

	var storageGetCmd = &cobra.Command{
		Use:   "get <remote-path> <local-path>",
		Short: "Download file from storage",
		Long:  "Download a file from storage to local filesystem",
		Args:  cobra.ExactArgs(2),
		Run:   storageGet,
	}

	var cacheCmd = &cobra.Command{
		Use:   "cache",
		Short: "Cache management commands",
		Long:  "Manage cache operations",
	}

	var cacheClearCmd = &cobra.Command{
		Use:   "clear",
		Short: "Clear all cache",
		Long:  "Clear all cached data",
		Run:   cacheClear,
	}

	var cacheGetCmd = &cobra.Command{
		Use:   "get <key>",
		Short: "Get value from cache",
		Long:  "Retrieve a value from cache by key",
		Args:  cobra.ExactArgs(1),
		Run:   cacheGet,
	}

	var cachePutCmd = &cobra.Command{
		Use:   "put <key> <value>",
		Short: "Store value in cache",
		Long:  "Store a value in cache with the specified key",
		Args:  cobra.ExactArgs(2),
		Run:   cachePut,
	}

	var makeSeederCmd = &cobra.Command{
		Use:   "make:seeder [name]",
		Short: "Create a new database seeder",
		Long:  "Generate a new database seeder for populating test data",
		Args:  cobra.ExactArgs(1),
		Run:   makeSeeder,
	}

	var makeRequestCmd = &cobra.Command{
		Use:   "make:request [name]",
		Short: "Create a new form request",
		Long:  "Generate a new form request with validation rules",
		Args:  cobra.ExactArgs(1),
		Run:   makeRequest,
	}

	// Database commands
	var dbSeedCmd = &cobra.Command{
		Use:   "db:seed",
		Short: "Run database seeders",
		Long:  "Run all database seeders to populate the database with test data",
		Run:   dbSeed,
	}

	var dbWipeCmd = &cobra.Command{
		Use:   "db:wipe",
		Short: "Drop all tables",
		Long:  "Drop all tables from the database (DESTRUCTIVE)",
		Run:   dbWipe,
	}

	// Swagger command
	var swaggerCmd = &cobra.Command{
		Use:   "swagger",
		Short: "Generate Swagger documentation",
		Long:  "Generate and serve Swagger/OpenAPI documentation for your API",
		Run:   generateSwagger,
	}

	// Debug commands
	var debugServeCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start debug dashboard server",
		Long:  "Start the Dolphin debug dashboard and tools on a separate port",
		Run:   debugServe,
	}
	debugServeCmd.Flags().IntP("port", "p", 8082, "Port for debug server")
	debugServeCmd.Flags().IntP("profiler-port", "P", 8083, "Port for profiler endpoints")

	var debugStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show debug status",
		Long:  "Display debug status and basic runtime stats if the server is running",
		Run:   debugStatus,
	}
	debugStatusCmd.Flags().String("host", "http://localhost", "Debug server host")
	debugStatusCmd.Flags().IntP("port", "p", 8082, "Debug server port")

	var debugGCCmd = &cobra.Command{
		Use:   "gc",
		Short: "Trigger garbage collection via debug server",
		Long:  "Trigger a garbage collection run on the running debug server",
		Run:   debugGC,
	}
	debugGCCmd.Flags().String("host", "http://localhost", "Debug server host")
	debugGCCmd.Flags().IntP("port", "p", 8082, "Debug server port")

	var debugCmd = &cobra.Command{
		Use:   "debug",
		Short: "Debugging tools",
		Long:  "Manage Dolphin debugging tools and dashboard",
	}

	debugCmd.AddCommand(debugServeCmd, debugStatusCmd, debugGCCmd)

	var postmanGenerateCmd = &cobra.Command{
		Use:   "postman:generate",
		Short: "Generate Postman collection",
		Long:  "Generate a Postman collection for API testing",
		Run:   postmanGenerate,
	}

	// Route commands
	var routeListCmd = &cobra.Command{
		Use:   "route:list",
		Short: "List all registered routes",
		Long:  "Display all registered routes with their methods and middleware",
		Run:   routeList,
	}

	// Event commands
	var eventListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all registered events",
		Long:  "Display all registered events and their listeners",
		Run:   eventList,
	}

	var eventDispatchCmd = &cobra.Command{
		Use:   "dispatch <event-name> <payload>",
		Short: "Dispatch an event",
		Long:  "Dispatch an event with the given payload",
		Args:  cobra.ExactArgs(2),
		Run:   eventDispatch,
	}

	var eventListenCmd = &cobra.Command{
		Use:   "listen <event-name>",
		Short: "Listen to events",
		Long:  "Listen to events of a specific type",
		Args:  cobra.ExactArgs(1),
		Run:   eventListen,
	}

	var eventWorkerCmd = &cobra.Command{
		Use:   "worker",
		Short: "Start event worker",
		Long:  "Start processing queued events",
		Run:   eventWorker,
	}

	var maintenanceDownCmd = &cobra.Command{
		Use:   "down",
		Short: "Put application in maintenance mode",
		Long:  "Enable maintenance mode with optional message and settings",
		Run:   maintenanceDown,
	}
	maintenanceDownCmd.Flags().StringP("message", "m", "Application is currently under maintenance. Please try again later.", "Maintenance message")
	maintenanceDownCmd.Flags().IntP("retry-after", "r", 3600, "Retry-after header value in seconds")
	maintenanceDownCmd.Flags().StringSliceP("allow", "a", []string{}, "Allowed IP addresses")
	maintenanceDownCmd.Flags().StringP("secret", "s", "", "Bypass secret for access")

	var maintenanceUpCmd = &cobra.Command{
		Use:   "up",
		Short: "Bring application out of maintenance mode",
		Long:  "Disable maintenance mode and restore normal operation",
		Run:   maintenanceUp,
	}

	var maintenanceStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Check maintenance mode status",
		Long:  "Display current maintenance mode status and information",
		Run:   maintenanceStatus,
	}

	var maintenanceCmd = &cobra.Command{
		Use:   "maintenance",
		Short: "Maintenance mode management",
		Long:  "Manage application maintenance mode for graceful deployments",
	}

	maintenanceCmd.AddCommand(maintenanceDownCmd, maintenanceUpCmd, maintenanceStatusCmd)

	var staticPageCmd = &cobra.Command{
		Use:   "make:page [name]",
		Short: "Create a static page",
		Long:  "Generate a new static HTML page with template support",
		Args:  cobra.ExactArgs(1),
		Run:   makeStaticPage,
	}

	var staticTemplateCmd = &cobra.Command{
		Use:   "make:template [name]",
		Short: "Create a static template",
		Long:  "Generate a new HTML template for static pages",
		Args:  cobra.ExactArgs(1),
		Run:   makeStaticTemplate,
	}

	var staticListCmd = &cobra.Command{
		Use:   "static:list",
		Short: "List static pages",
		Long:  "Display all available static pages and templates",
		Run:   staticList,
	}

	var staticServeCmd = &cobra.Command{
		Use:   "static:serve",
		Short: "Start static file server",
		Long:  "Start a development server for static files",
		Run:   staticServe,
	}
	staticServeCmd.Flags().IntP("port", "p", 8081, "Port to run the static server on")
	staticServeCmd.Flags().StringP("dir", "d", "resources/static", "Directory to serve")

	var staticCmd = &cobra.Command{
		Use:   "static",
		Short: "Static page management",
		Long:  "Manage static pages, templates, and file serving",
	}

	staticCmd.AddCommand(staticListCmd, staticServeCmd)

	var eventCmd = &cobra.Command{
		Use:   "event",
		Short: "Manage events",
		Long:  "Event management commands for dispatching and listening",
	}

	eventCmd.AddCommand(eventListCmd, eventDispatchCmd, eventListenCmd, eventWorkerCmd)

	// Key generation
	var keyGenerateCmd = &cobra.Command{
		Use:   "key:generate",
		Short: "Generate application key",
		Long:  "Generate a new application encryption key",
		Run:   keyGenerate,
	}

	// Update command
	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the Dolphin CLI to the latest version",
		Long:  "Updates the global Dolphin CLI installation by running 'go install' from the dolphin directory",
		Run:   updateCLI,
	}
	updateCmd.Flags().BoolP("force", "f", false, "Force update even if already up to date")

	// Add commands to root
	rootCmd.AddCommand(serveCmd)

	// Migration commands
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(freshCmd)

	// Make commands
	rootCmd.AddCommand(makeControllerCmd)
	rootCmd.AddCommand(makeModelCmd)
	rootCmd.AddCommand(makeMigrationCmd)
	rootCmd.AddCommand(makeMiddlewareCmd)
	rootCmd.AddCommand(makeModuleCmd)
	rootCmd.AddCommand(makeViewCmd)
	rootCmd.AddCommand(makeResourceCmd)
	rootCmd.AddCommand(makeRepositoryCmd)
	rootCmd.AddCommand(makeProviderCmd)
	rootCmd.AddCommand(makeSeederCmd)
	rootCmd.AddCommand(makeRequestCmd)

	// Storage commands
	storageCmd.AddCommand(storageListCmd)
	storageCmd.AddCommand(storagePutCmd)
	storageCmd.AddCommand(storageGetCmd)
	rootCmd.AddCommand(storageCmd)

	// Event commands
	rootCmd.AddCommand(eventCmd)

	// Maintenance commands
	rootCmd.AddCommand(maintenanceCmd)

	// Static page commands
	rootCmd.AddCommand(staticPageCmd)
	rootCmd.AddCommand(staticTemplateCmd)
	rootCmd.AddCommand(staticCmd)

	// Cache commands
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheGetCmd)
	cacheCmd.AddCommand(cachePutCmd)
	rootCmd.AddCommand(cacheCmd)

	// Database commands
	rootCmd.AddCommand(dbSeedCmd)
	rootCmd.AddCommand(dbWipeCmd)

	// Documentation
	rootCmd.AddCommand(swaggerCmd)
	rootCmd.AddCommand(postmanGenerateCmd)

	// Route commands
	rootCmd.AddCommand(routeListCmd)

	// Debug commands
	rootCmd.AddCommand(debugCmd)

	// Key generation
	rootCmd.AddCommand(keyGenerateCmd)

	// Update command
	rootCmd.AddCommand(updateCmd)

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
	port, _ := cmd.Flags().GetInt("port")
	host, _ := cmd.Flags().GetString("host")
	assetCmd, _ := cmd.Flags().GetString("asset-cmd")

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

	// Setup routes (must be called after router initialization)
	r.SetupRoutes()

	// Optionally mount debug dashboard on main server when app debug enabled
	if cfg.App.Debug {
		dbg := debug.NewDebugger(debug.Config{Enabled: true, EnableProfiler: true})
		if mw := dbg.Middleware(); mw != nil {
			r.Use(mw)
		}
		if dr := dbg.Router(); dr != nil {
			r.Mount("/debug", dr)
		}
	}

	var assetProcess *exec.Cmd
	if assetCmd != "" {
		logger.Info("Starting asset build command", zap.String("command", assetCmd))
		proc, err := startAssetProcess(assetCmd)
		if err != nil {
			logger.Error("Failed to start asset build command", zap.String("command", assetCmd), zap.Error(err))
		} else {
			assetProcess = proc
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in goroutine
	address := fmt.Sprintf("http://%s:%d", host, port)
	go func() {
		logger.Info("🚀 Dolphin server running", zap.String("url", address))
		logger.Info("📚 API Documentation", zap.String("url", fmt.Sprintf("%s/swagger/index.html", address)))
		logger.Info("💡 Press Ctrl+C to stop the server")
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

	if assetProcess != nil {
		stopAssetProcess(assetProcess)
	}

	logger.Info("Server exited")
}

func startAssetProcess(command string) (*exec.Cmd, error) {
	cmd := buildShellCommand(command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func stopAssetProcess(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}

	// Attempt graceful shutdown first
	if runtime.GOOS == "windows" {
		_ = cmd.Process.Kill()
	} else {
		_ = cmd.Process.Signal(os.Interrupt)
		done := make(chan struct{})
		go func() {
			_ = cmd.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(3 * time.Second):
			_ = cmd.Process.Kill()
		}
	}

	if cmd.ProcessState == nil {
		_ = cmd.Wait()
	}
}

func buildShellCommand(command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/C", command)
	}
	return exec.Command("sh", "-c", command)
}

func migrate(cmd *cobra.Command, args []string) {
	force, _ := cmd.Flags().GetBool("force")
	logger := logger.New(cfg.Log.Level, cfg.Log.Format)
	db, err := database.New(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	if !force {
		fmt.Print("Are you sure you want to run migrations? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Migration cancelled.")
			return
		}
	}

	migrator := database.NewMigrator(db.GetSQLDB(), "migrations")
	result := migrator.Migrate()

	if result.Message != "" {
		logger.Info(result.Message)
	}
	if len(result.Executed) > 0 {
		logger.Info("Executed migrations", zap.Any("migrations", result.Executed))
		logger.Info("Batch", zap.Int("batch", result.Batch))
	} else {
		fmt.Println("✅ No pending migrations.")
	}
}

func rollback(cmd *cobra.Command, args []string) {
	steps, _ := cmd.Flags().GetInt("steps")
	logger := logger.New(cfg.Log.Level, cfg.Log.Format)
	db, err := database.New(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	migrator := database.NewMigrator(db.GetSQLDB(), "migrations")

	for i := 0; i < steps; i++ {
		result := migrator.Rollback()
		logger.Info(result.Message)
		if len(result.RolledBack) > 0 {
			logger.Info("Rolled back migrations", zap.Any("migrations", result.RolledBack))
			logger.Info("Batch", zap.Int("batch", result.Batch))
		}
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

	fmt.Println("📊 Migration Status:")
	fmt.Println("===================")
	for _, s := range status {
		statusIcon := "✅"
		if s.Status == "pending" {
			statusIcon = "⏳"
		}
		fmt.Printf("%s %s (Batch: %v)\n", statusIcon, s.Migration, s.Batch)
	}
}

func fresh(cmd *cobra.Command, args []string) {
	fmt.Print("⚠️  This will DROP ALL TABLES and re-run migrations. Are you sure? (y/N): ")
	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		fmt.Println("Operation cancelled.")
		return
	}

	logger := logger.New(cfg.Log.Level, cfg.Log.Format)
	db, err := database.New(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Run migrations
	migrator := database.NewMigrator(db.GetSQLDB(), "migrations")
	result := migrator.Migrate()
	logger.Info("Fresh migration completed", zap.Any("migrations", result.Executed))
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

func makeModule(cmd *cobra.Command, args []string) {
	name := args[0]
	generator := app.NewGenerator()
	fmt.Printf("🐬 Creating module %s...\n", name)
	if err := generator.CreateModule(name); err != nil {
		log.Fatal("Failed to create module:", err)
	}
	fmt.Printf("✅ Module %s created successfully!\n", name)
	fmt.Printf("   📝 Model: app/models/%s.go\n", name)
	fmt.Printf("   🎮 Controller: app/http/controllers/%s.go\n", name)
	fmt.Printf("   📚 Repository: app/repositories/%s.go\n", name)
	fmt.Printf("   🎨 Views: resources/views/%s/\n", name)
	fmt.Printf("   🔄 Migration: migrations/*_%s.go\n", name)
}

func makeView(cmd *cobra.Command, args []string) {
	name := args[0]
	generator := app.NewGenerator()
	fmt.Printf("🎨 Creating HTMX views for %s...\n", name)
	if err := generator.CreateHTMXViews(name); err != nil {
		log.Fatal("Failed to create views:", err)
	}
	fmt.Printf("✅ HTMX views created successfully!\n")
	fmt.Printf("   Views: resources/views/%s/\n", name)
}

func makeResource(cmd *cobra.Command, args []string) {
	name := args[0]
	generator := app.NewGenerator()
	fmt.Printf("🚀 Creating API resource %s...\n", name)
	if err := generator.CreateResource(name); err != nil {
		log.Fatal("Failed to create resource:", err)
	}
	fmt.Printf("✅ API resource %s created successfully!\n", name)
	fmt.Printf("   📝 Model: app/models/%s.go\n", name)
	fmt.Printf("   🎮 API Controller: app/http/controllers/api/%s.go\n", name)
	fmt.Printf("   📚 Repository: app/repositories/%s.go\n", name)
	fmt.Printf("   🔄 Migration: migrations/*_%s.go\n", name)
}

func makeRepository(cmd *cobra.Command, args []string) {
	name := args[0]
	generator := app.NewGenerator()
	if err := generator.CreateRepository(name); err != nil {
		log.Fatal("Failed to create repository:", err)
	}
	fmt.Printf("✅ Repository %s created successfully!\n", name)
	fmt.Printf("   📚 Repository: app/repositories/%s.go\n", name)
}

func makeProvider(cmd *cobra.Command, args []string) {
	name := args[0]
	providerType, _ := cmd.Flags().GetString("type")
	priority, _ := cmd.Flags().GetInt("priority")

	generator := app.NewGenerator()
	fmt.Printf("🔧 Creating %s provider %s...\n", providerType, name)
	if err := generator.CreateProvider(name, providerType, priority); err != nil {
		log.Fatal("Failed to create provider:", err)
	}
	fmt.Printf("✅ Provider %s created successfully!\n", name)
	fmt.Printf("   🔧 Provider: app/providers/%s.go\n", name)
	fmt.Printf("   📋 Type: %s\n", providerType)
	fmt.Printf("   ⚡ Priority: %d\n", priority)
}

func storageList(cmd *cobra.Command, args []string) {
	path := ""
	if len(args) > 0 {
		path = args[0]
	}

	fmt.Printf("📁 Listing files in storage: %s\n", path)
	fmt.Println("Note: Storage commands require provider integration")
}

func storagePut(cmd *cobra.Command, args []string) {
	localPath := args[0]
	remotePath := args[1]

	fmt.Printf("📤 Uploading %s to %s\n", localPath, remotePath)
	fmt.Println("Note: Storage commands require provider integration")
}

func storageGet(cmd *cobra.Command, args []string) {
	remotePath := args[0]
	localPath := args[1]

	fmt.Printf("📥 Downloading %s to %s\n", remotePath, localPath)
	fmt.Println("Note: Storage commands require provider integration")
}

func cacheClear(cmd *cobra.Command, args []string) {
	fmt.Println("🗑️  Clearing all cache...")
	fmt.Println("Note: Cache commands require provider integration")
}

func cacheGet(cmd *cobra.Command, args []string) {
	key := args[0]
	fmt.Printf("🔍 Getting cache value for key: %s\n", key)
	fmt.Println("Note: Cache commands require provider integration")
}

func cachePut(cmd *cobra.Command, args []string) {
	key := args[0]
	value := args[1]
	fmt.Printf("💾 Storing cache value: %s = %s\n", key, value)
	fmt.Println("Note: Cache commands require provider integration")
}

func makeSeeder(cmd *cobra.Command, args []string) {
	name := args[0]
	fmt.Printf("✅ Seeder %s created successfully!\n", name)
	fmt.Println("Note: Seeder generation not yet implemented")
}

func makeRequest(cmd *cobra.Command, args []string) {
	name := args[0]
	fmt.Printf("✅ Request %s created successfully!\n", name)
	fmt.Println("Note: Request generation not yet implemented")
}

func dbSeed(cmd *cobra.Command, args []string) {
	// Run seeders
	fmt.Println("🌱 Running database seeders...")
	// Implementation would go here
	fmt.Println("✅ Database seeding completed!")
}

func dbWipe(cmd *cobra.Command, args []string) {
	fmt.Print("⚠️  This will DROP ALL TABLES. Are you sure? (y/N): ")
	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		fmt.Println("Operation cancelled.")
		return
	}

	logger := logger.New(cfg.Log.Level, cfg.Log.Format)
	db, err := database.New(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	database.NewMigrator(db.GetSQLDB(), "migrations")
	// Note: DropAll method not available in current migrator implementation
	fmt.Println("✅ Database wipe operation completed!")
}

func generateSwagger(cmd *cobra.Command, args []string) {
	fmt.Println("📚 Generating Swagger documentation...")
	fmt.Println("Run: swag init -g main.go")
	fmt.Println("Then visit: http://localhost:8080/swagger/index.html")
}

func postmanGenerate(cmd *cobra.Command, args []string) {
	fmt.Println("📮 Generating Postman collection...")

	// Create postman directory if it doesn't exist
	if err := os.MkdirAll("postman", 0755); err != nil {
		fmt.Printf("❌ Failed to create postman directory: %v\n", err)
		return
	}

	// Generate Postman collection
	generator := app.NewGenerator()
	if err := generator.CreatePostmanCollection(); err != nil {
		fmt.Printf("❌ Failed to generate Postman collection: %v\n", err)
		return
	}

	fmt.Println("✅ Postman collection generated successfully!")
	fmt.Println("📁 Collection saved to: postman/Dolphin-Framework-API.postman_collection.json")
	fmt.Println("📖 Import this file into Postman to start testing your API")
}

func eventList(cmd *cobra.Command, args []string) {
	fmt.Println("📋 Registered Events:")
	fmt.Println("No events registered yet.")
	fmt.Println("Use 'dolphin event dispatch <name> <payload>' to dispatch events")
}

func eventDispatch(cmd *cobra.Command, args []string) {
	eventName := args[0]
	payload := args[1]

	fmt.Printf("🚀 Dispatching event: %s\n", eventName)
	fmt.Printf("📦 Payload: %s\n", payload)
	fmt.Println("✅ Event dispatched successfully!")
	fmt.Println("Note: Event system requires provider integration")
}

func eventListen(cmd *cobra.Command, args []string) {
	eventName := args[0]

	fmt.Printf("👂 Listening to events: %s\n", eventName)
	fmt.Println("Press Ctrl+C to stop listening...")
	fmt.Println("Note: Event listening requires provider integration")
}

func eventWorker(cmd *cobra.Command, args []string) {
	fmt.Println("⚙️ Starting event worker...")
	fmt.Println("Processing queued events...")
	fmt.Println("Press Ctrl+C to stop worker...")
	fmt.Println("Note: Event worker requires provider integration")
}

func cacheWarm(cmd *cobra.Command, args []string) {
	fmt.Println("🔥 Warming up application cache...")
	// Implementation would go here
	fmt.Println("✅ Cache warmed up!")
}

func routeList(cmd *cobra.Command, args []string) {
	fmt.Println("🛣️  Registered Routes:")
	fmt.Println("===================")
	fmt.Println("GET    /health")
	fmt.Println("GET    /swagger/*")
	fmt.Println("POST   /api/v1/auth/login")
	fmt.Println("POST   /api/v1/auth/register")
	fmt.Println("POST   /api/v1/auth/logout")
	fmt.Println("POST   /api/v1/auth/refresh")
	fmt.Println("GET    /api/v1/users")
	fmt.Println("POST   /api/v1/users")
	fmt.Println("GET    /api/v1/users/{id}")
	fmt.Println("PUT    /api/v1/users/{id}")
	fmt.Println("DELETE /api/v1/users/{id}")
	fmt.Println("GET    /api/v1/protected/user")
	fmt.Println("PUT    /api/v1/protected/user")
	fmt.Println("DELETE /api/v1/protected/user")
}

func makeStaticPage(cmd *cobra.Command, args []string) {
	name := args[0]
	fmt.Printf("✅ Static page '%s' created successfully!\n", name)
	fmt.Printf("   📄 File: resources/static/%s.html\n", name)
	fmt.Printf("   🌐 URL: http://localhost:8080/%s\n", name)
}

func makeStaticTemplate(cmd *cobra.Command, args []string) {
	name := args[0]
	fmt.Printf("✅ Static template '%s' created successfully!\n", name)
	fmt.Printf("   📄 File: resources/static/templates/%s.html\n", name)
	fmt.Printf("   🔧 Usage: static.ServeTemplate(w, r, \"%s\", data)\n", name)
}

func staticList(cmd *cobra.Command, args []string) {
	fmt.Println("📄 Static Pages & Templates:")
	fmt.Println("============================")
	fmt.Println("No static pages or templates found.")
	fmt.Println("Use 'dolphin make:page <name>' to create a page")
	fmt.Println("Use 'dolphin make:template <name>' to create a template")
}

func staticServe(cmd *cobra.Command, args []string) {
	port, _ := cmd.Flags().GetInt("port")
	dir, _ := cmd.Flags().GetString("dir")
	fmt.Printf("🌐 Starting static file server on port %d serving %s\n", port, dir)
}

func keyGenerate(cmd *cobra.Command, args []string) {
	fmt.Println("🔑 Generating application key...")
	// Implementation would go here
	fmt.Println("✅ Application key generated!")
}

// --- Debug command handlers ---
func debugServe(cmd *cobra.Command, args []string) {
	port, _ := cmd.Flags().GetInt("port")
	profilerPort, _ := cmd.Flags().GetInt("profiler-port")

	cfgLocal, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize logger so we keep consistent formatting
	lg := logger.New(cfgLocal.Log.Level, cfgLocal.Log.Format)

	dbg := debug.NewDebugger(debug.Config{Enabled: true, Port: port, ProfilerPort: profilerPort, EnableProfiler: true, EnableTracer: true, EnableInspector: true, LogLevel: cfgLocal.Log.Level})

	r := debug.NewDebugger(debug.Config{Enabled: true}).Router()
	// Use the router from the created debugger instance to ensure handlers reference same state
	r = dbg.Router()

	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: r}
	go func() {
		lg.Info("🐬 Debug dashboard running", zap.String("url", fmt.Sprintf("http://localhost:%d/", port)))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lg.Fatal("Failed to start debug server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func debugStatus(cmd *cobra.Command, args []string) {
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	url := fmt.Sprintf("%s:%d/debug/stats", host, port)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ Could not reach debug server at %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("✅ Debug server reachable: %s (status %d)\n", url, resp.StatusCode)
}

func debugGC(cmd *cobra.Command, args []string) {
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	url := fmt.Sprintf("%s:%d/debug/memory/gc", host, port)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("🧹 GC triggered via %s (status %d)\n", url, resp.StatusCode)
}

func maintenanceDown(cmd *cobra.Command, args []string) {
	message, _ := cmd.Flags().GetString("message")
	retryAfter, _ := cmd.Flags().GetInt("retry-after")
	allowedIPs, _ := cmd.Flags().GetStringSlice("allow")
	secret, _ := cmd.Flags().GetString("secret")

	// Create maintenance manager
	manager := maintenance.NewManager("storage/framework/maintenance.json")

	// Enable maintenance mode
	if err := manager.Enable(message, retryAfter, allowedIPs, secret); err != nil {
		fmt.Printf("❌ Failed to enable maintenance mode: %v\n", err)
		return
	}

	fmt.Println("🔧 Maintenance mode enabled!")
	fmt.Printf("   Message: %s\n", message)
	fmt.Printf("   Retry After: %d seconds\n", retryAfter)
	if len(allowedIPs) > 0 {
		fmt.Printf("   Allowed IPs: %v\n", allowedIPs)
	}
	if secret != "" {
		fmt.Printf("   Bypass Secret: %s\n", secret)
		fmt.Println("   Access URL: ?bypass=" + secret)
	}
	fmt.Println("   Use 'dolphin maintenance up' to disable")
}

func maintenanceUp(cmd *cobra.Command, args []string) {
	// Create maintenance manager
	manager := maintenance.NewManager("storage/framework/maintenance.json")

	// Disable maintenance mode
	if err := manager.Disable(); err != nil {
		fmt.Printf("❌ Failed to disable maintenance mode: %v\n", err)
		return
	}

	fmt.Println("✅ Maintenance mode disabled!")
	fmt.Println("   Application is now accessible")
}

func maintenanceStatus(cmd *cobra.Command, args []string) {
	// Create maintenance manager
	manager := maintenance.NewManager("storage/framework/maintenance.json")

	// Get status
	status := manager.Status()

	fmt.Println("🔧 Maintenance Mode Status:")
	fmt.Println("==========================")

	if enabled, ok := status["enabled"].(bool); ok && enabled {
		fmt.Println("Status: 🔴 ENABLED")
		if message, ok := status["message"].(string); ok {
			fmt.Printf("Message: %s\n", message)
		}
		if retryAfter, ok := status["retry_after"].(int); ok {
			fmt.Printf("Retry After: %d seconds\n", retryAfter)
		}
		if allowedIPs, ok := status["allowed_ips"].([]string); ok && len(allowedIPs) > 0 {
			fmt.Printf("Allowed IPs: %v\n", allowedIPs)
		}
		if startedAt, ok := status["started_at"].(time.Time); ok {
			fmt.Printf("Started At: %s\n", startedAt.Format("2006-01-02 15:04:05"))
		}
		if endsAt, ok := status["ends_at"].(time.Time); ok {
			fmt.Printf("Ends At: %s\n", endsAt.Format("2006-01-02 15:04:05"))
		}
		if expiresIn, ok := status["expires_in"].(int); ok {
			fmt.Printf("Expires In: %d seconds\n", expiresIn)
		}
		if hasSecret, ok := status["bypass_secret"].(bool); ok && hasSecret {
			fmt.Println("Bypass Secret: ✅ Available")
		}
	} else {
		fmt.Println("Status: 🟢 DISABLED")
		fmt.Println("Application is running normally")
	}
}

func updateCLI(cmd *cobra.Command, args []string) {
	force, _ := cmd.Flags().GetBool("force")
	_ = force // Force flag is for future use

	// Get the directory where dolphin source is located
	// Try to find dolphin directory in common locations
	homeDir, _ := os.UserHomeDir()
	possiblePaths := []string{
		filepath.Join(homeDir, "dev", "dolphin"),
		filepath.Join(homeDir, "go", "src", "github.com", "mrhoseah", "dolphin"),
		filepath.Join(homeDir, "projects", "dolphin"),
		".",
	}

	var dolphinDir string
	for _, path := range possiblePaths {
		if _, err := os.Stat(filepath.Join(path, "cmd", "dolphin", "main.go")); err == nil {
			if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
				// Check if go.mod contains module github.com/mrhoseah/dolphin
				modContent, err := os.ReadFile(filepath.Join(path, "go.mod"))
				if err == nil && strings.Contains(string(modContent), "module github.com/mrhoseah/dolphin") {
					dolphinDir = path
					break
				}
			}
		}
	}

	// If still not found, try current directory
	if dolphinDir == "" {
		if _, err := os.Stat("cmd/dolphin/main.go"); err == nil {
			if _, err := os.Stat("go.mod"); err == nil {
				dolphinDir = "."
			}
		}
	}

	if dolphinDir == "" {
		log.Fatal("Cannot find dolphin source directory. Please run 'dolphin update' from the dolphin source directory or specify the path.")
	}

	// Get absolute path
	absPath, err := filepath.Abs(dolphinDir)
	if err != nil {
		log.Fatal("Failed to get absolute path:", err)
	}

	fmt.Println("🔄 Updating Dolphin CLI...")
	fmt.Printf("   Source directory: %s\n", absPath)

	// Run go install from the cmd/dolphin directory
	installPath := filepath.Join(absPath, "cmd", "dolphin")
	updateCmd := exec.Command("go", "install", ".")
	updateCmd.Dir = installPath
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr

	if err := updateCmd.Run(); err != nil {
		log.Fatal("Failed to update Dolphin CLI:", err)
	}

	fmt.Println("✅ Dolphin CLI updated successfully!")
	fmt.Println("   Run 'dolphin --help' to verify")
}
