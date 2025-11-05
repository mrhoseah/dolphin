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
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"dolphin/internal/app"
	"dolphin/internal/cli"
	"dolphin/internal/config"
	"dolphin/internal/database"
	"dolphin/internal/logger"
	"dolphin/internal/router"
	"dolphin/internal/storage"

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

	var makeAuthCmd = &cobra.Command{
		Use:   "make:auth",
		Short: "Generate authentication views and pages",
		Run:   makeAuth,
	}

	// Shadcn UI setup command
	var makeShadcnCmd = &cobra.Command{
		Use:   "make:shadcn",
		Short: "Set up shadcn/ui for React-based Dolphin apps",
		Long:  "Sets up shadcn/ui component library similar to Laravel's integration. Configures components.json, Tailwind CSS, and necessary dependencies.",
		Run:   makeShadcn,
	}

	// Swagger command
	var swaggerCmd = &cobra.Command{
		Use:   "swagger",
		Short: "Generate Swagger documentation",
		Run:   generateSwagger,
	}

	// Storage symlink command
	var storageLinkCmd = &cobra.Command{
		Use:   "storage:link",
		Short: "Create a symbolic link from public/storage to storage/app/public",
		Run:   storageLink,
	}

	// Client generation command
	var makeClientCmd = &cobra.Command{
		Use:   "make:client [language]",
		Short: "Generate type-safe API client from OpenAPI spec",
		Long:  "Generate type-safe API client (go or typescript) from OpenAPI spec",
		Args:  cobra.ExactArgs(1),
		Run:   makeClient,
	}
	makeClientCmd.Flags().StringP("spec", "s", "swagger.json", "OpenAPI spec file path")
	makeClientCmd.Flags().StringP("output", "o", "generated", "Output directory")

	// Update command
	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the Dolphin CLI to the latest version",
		Long:  "Updates the global Dolphin CLI installation by running 'go install .' from the dolphin directory",
		Run:   updateCLI,
	}
	updateCmd.Flags().BoolP("force", "f", false, "Force update even if already up to date")

	// Add commands
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(makeControllerCmd)
	rootCmd.AddCommand(makeModelCmd)
	rootCmd.AddCommand(makeMigrationCmd)
	rootCmd.AddCommand(makeMiddlewareCmd)
	rootCmd.AddCommand(makeAuthCmd)
	rootCmd.AddCommand(makeShadcnCmd)
	rootCmd.AddCommand(swaggerCmd)
	rootCmd.AddCommand(storageLinkCmd)
	rootCmd.AddCommand(makeClientCmd)
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
	// Check if we're in an application directory (has main.go with serve command)
	// If so, delegate to the application's serve command
	if isApplicationDirectory() {
		// Run the application's serve command
		runApplicationServe()
		return
	}

	// Otherwise, run the framework's default server
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

func makeAuth(cmd *cobra.Command, args []string) {
	generator := app.NewGenerator()
	if err := generator.CreateAuth(); err != nil {
		log.Fatal("Failed to create auth views:", err)
	}
	fmt.Println("✅ Authentication views and pages created successfully!")
	fmt.Println("   Created views/auth/login.fin.html")
	fmt.Println("   Created views/auth/register.fin.html")
	fmt.Println("   Created views/auth/forgot-password.fin.html")
	fmt.Println("   Created views/auth/reset-password.fin.html")
}

func makeShadcn(cmd *cobra.Command, args []string) {
	generator := app.NewGenerator()
	if err := generator.CreateShadcnUI(); err != nil {
		log.Fatal("Failed to set up shadcn/ui:", err)
	}
	fmt.Println("✅ shadcn/ui setup completed successfully!")
	fmt.Println("")
	fmt.Println("📦 Created files:")
	fmt.Println("   - components.json (shadcn configuration)")
	fmt.Println("   - lib/utils.ts (utility functions)")
	fmt.Println("   - components/ui/button.tsx (example component)")
	fmt.Println("   - tsconfig.paths.json (path aliases)")
	fmt.Println("")
	fmt.Println("🎨 Updated files:")
	fmt.Println("   - tailwind.config.js (shadcn theme)")
	fmt.Println("   - assets/css/app.css (CSS variables)")
	fmt.Println("")
	fmt.Println("📝 Next steps:")
	fmt.Println("   1. Install dependencies: npm install")
	fmt.Println("   2. Install shadcn dependencies:")
	fmt.Println("      npm install clsx tailwind-merge tailwindcss-animate")
	fmt.Println("   3. Add components:")
	fmt.Println("      npx shadcn@latest add button")
	fmt.Println("      npx shadcn@latest add card")
	fmt.Println("      # See more at: https://ui.shadcn.com/docs/components")
	fmt.Println("")
	fmt.Println("📚 Documentation: https://ui.shadcn.com/docs/installation/laravel")
}

func generateSwagger(cmd *cobra.Command, args []string) {
	fmt.Println("📚 Generating Swagger documentation...")
	fmt.Println("Run: swag init -g main.go")
	fmt.Println("Then visit: http://localhost:8080/swagger/index.html")
}

func storageLink(cmd *cobra.Command, args []string) {
	logger := logger.New(cfg.Log.Level, cfg.Log.Format)

	publicPath := "public"
	storagePath := "storage"

	if err := storage.CreateSymlink(publicPath, storagePath); err != nil {
		logger.Fatal("Failed to create storage symlink", zap.Error(err))
	}

	fmt.Println("✅ Storage symlink created successfully!")
	fmt.Println("   public/storage → storage/app/public")
}

func makeClient(cmd *cobra.Command, args []string) {
	language := args[0]
	if language != "go" && language != "typescript" {
		log.Fatal("Language must be 'go' or 'typescript'")
	}

	specPath, _ := cmd.Flags().GetString("spec")
	outputDir, _ := cmd.Flags().GetString("output")

	if err := cli.GenerateClientCommand(specPath, outputDir, language); err != nil {
		log.Fatal("Failed to generate client:", err)
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
		filepath.Join(homeDir, "go", "src", "dolphin"),
		filepath.Join(homeDir, "projects", "dolphin"),
		".",
	}

	var dolphinDir string
	for _, path := range possiblePaths {
		if _, err := os.Stat(filepath.Join(path, "main.go")); err == nil {
			if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
				// Check if go.mod contains module dolphin
				modContent, err := os.ReadFile(filepath.Join(path, "go.mod"))
				if err == nil && strings.Contains(string(modContent), "module dolphin") {
					dolphinDir = path
					break
				}
			}
		}
	}

	// If still not found, try current directory
	if dolphinDir == "" {
		if _, err := os.Stat("main.go"); err == nil {
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

	// Run go install
	updateCmd := exec.Command("go", "install", ".")
	updateCmd.Dir = absPath
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr

	if err := updateCmd.Run(); err != nil {
		log.Fatal("Failed to update Dolphin CLI:", err)
	}

	fmt.Println("✅ Dolphin CLI updated successfully!")
	fmt.Println("   Run 'dolphin --help' to verify")
}

// isApplicationDirectory checks if the current directory is an application directory
// An application directory has:
// 1. A main.go file
// 2. A go.mod file with a module name that's not "dolphin"
func isApplicationDirectory() bool {
	// Check if main.go exists
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		return false
	}

	// Check if go.mod exists
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return false
	}

	// Read go.mod to check module name
	modContent, err := os.ReadFile("go.mod")
	if err != nil {
		return false
	}

	modContentStr := string(modContent)
	// Check if it's NOT the dolphin framework module
	// Framework module is "dolphin" or "module dolphin"
	if strings.Contains(modContentStr, "module dolphin") {
		return false
	}

	// Check if main.go has a serve command (look for "serve" function or cobra command)
	mainContent, err := os.ReadFile("main.go")
	if err != nil {
		return false
	}

	mainContentStr := string(mainContent)
	// Check if it has a serve function or cobra serve command
	hasServe := strings.Contains(mainContentStr, "func serve") ||
		strings.Contains(mainContentStr, "serveCmd") ||
		strings.Contains(mainContentStr, "Use:   \"serve\"")

	return hasServe
}

// runApplicationServe runs the application's serve command
func runApplicationServe() {
	fmt.Println("🐬 Detected application directory. Running application server...")
	fmt.Println("   (Use 'go run main.go serve' directly for more control)")
	fmt.Println()

	// Run: go run main.go serve
	serveCmd := exec.Command("go", "run", "main.go", "serve")
	serveCmd.Stdin = os.Stdin
	serveCmd.Stdout = os.Stdout
	serveCmd.Stderr = os.Stderr

	if err := serveCmd.Run(); err != nil {
		log.Fatalf("Failed to run application server: %v", err)
	}
}
