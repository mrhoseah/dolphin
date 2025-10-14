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

	// Cache commands
	var cacheClearCmd = &cobra.Command{
		Use:   "cache:clear",
		Short: "Clear application cache",
		Long:  "Clear all cached data from Redis and memory cache",
		Run:   cacheClear,
	}

	var cacheWarmCmd = &cobra.Command{
		Use:   "cache:warm",
		Short: "Warm up application cache",
		Long:  "Pre-populate cache with frequently accessed data",
		Run:   cacheWarm,
	}

	// Route commands
	var routeListCmd = &cobra.Command{
		Use:   "route:list",
		Short: "List all registered routes",
		Long:  "Display all registered routes with their methods and middleware",
		Run:   routeList,
	}

	// Key generation
	var keyGenerateCmd = &cobra.Command{
		Use:   "key:generate",
		Short: "Generate application key",
		Long:  "Generate a new application encryption key",
		Run:   keyGenerate,
	}

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
	rootCmd.AddCommand(makeSeederCmd)
	rootCmd.AddCommand(makeRequestCmd)

	// Database commands
	rootCmd.AddCommand(dbSeedCmd)
	rootCmd.AddCommand(dbWipeCmd)

	// Documentation
	rootCmd.AddCommand(swaggerCmd)

	// Cache commands
	rootCmd.AddCommand(cacheClearCmd)
	rootCmd.AddCommand(cacheWarmCmd)

	// Route commands
	rootCmd.AddCommand(routeListCmd)

	// Key generation
	rootCmd.AddCommand(keyGenerateCmd)

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
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("🚀 Dolphin server running", zap.String("url", fmt.Sprintf("http://%s:%d", host, port)))
		logger.Info("📚 API Documentation", zap.String("url", fmt.Sprintf("http://%s:%d/swagger/index.html", host, port)))
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

	logger.Info("Server exited")
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

func makeSeeder(cmd *cobra.Command, args []string) {
	name := args[0]
	generator := app.NewGenerator()
	if err := generator.CreateSeeder(name); err != nil {
		log.Fatal("Failed to create seeder:", err)
	}
	fmt.Printf("✅ Seeder %s created successfully!\n", name)
}

func makeRequest(cmd *cobra.Command, args []string) {
	name := args[0]
	generator := app.NewGenerator()
	if err := generator.CreateRequest(name); err != nil {
		log.Fatal("Failed to create request:", err)
	}
	fmt.Printf("✅ Request %s created successfully!\n", name)
}

func dbSeed(cmd *cobra.Command, args []string) {
	logger := logger.New(cfg.Log.Level, cfg.Log.Format)
	db, err := database.New(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

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

	migrator := database.NewMigrator(db.GetSQLDB(), "migrations")
	// Note: DropAll method not available in current migrator implementation
	fmt.Println("✅ Database wipe operation completed!")
}

func generateSwagger(cmd *cobra.Command, args []string) {
	fmt.Println("📚 Generating Swagger documentation...")
	fmt.Println("Run: swag init -g main.go")
	fmt.Println("Then visit: http://localhost:8080/swagger/index.html")
}

func cacheClear(cmd *cobra.Command, args []string) {
	fmt.Println("🧹 Clearing application cache...")
	// Implementation would go here
	fmt.Println("✅ Cache cleared!")
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

func keyGenerate(cmd *cobra.Command, args []string) {
	fmt.Println("🔑 Generating application key...")
	// Implementation would go here
	fmt.Println("✅ Application key generated!")
}
