package main

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
		Short: "Dolphin Framework - Enterprise-grade Go web framework",
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
