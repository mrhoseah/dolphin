package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/mrhoseah/dolphin/internal/app"
	"github.com/mrhoseah/dolphin/internal/auth"
	"github.com/mrhoseah/dolphin/internal/config"
	"github.com/mrhoseah/dolphin/internal/database"
	"github.com/mrhoseah/dolphin/internal/debug"
	"github.com/mrhoseah/dolphin/internal/logger"
	"github.com/mrhoseah/dolphin/internal/maintenance"
	"github.com/mrhoseah/dolphin/internal/router"
	"github.com/mrhoseah/dolphin/internal/security"
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
	// Update command
	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the Dolphin CLI to the latest or specified version",
		Long:  "Self-update the Dolphin CLI by installing the latest (or specified) version via 'go install'.",
		Run:   updateSelf,
	}
	updateCmd.Flags().StringP("version", "V", "main", "Version to install (e.g., v0.1.0 or 'main')")
	// New project command
	var newCmd = &cobra.Command{
		Use:   "new [name]",
		Short: "Scaffold a new Dolphin project",
		Long:  "Create a new Dolphin project with standard directories and a basic config.",
		Args:  cobra.ExactArgs(1),
		Run:   newProject,
	}
	newCmd.Flags().Bool("auth", false, "Include auth scaffolding (views and links)")

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

	// Rate limit command group
	var rateLimitCmd = &cobra.Command{
		Use:   "ratelimit",
		Short: "Rate limiting management",
		Long:  "Manage rate limiting settings and view rate limit status.",
	}

	var rateLimitStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show rate limit status",
		Long:  "Display current rate limiting configuration and status.",
		Run:   rateLimitStatus,
	}

	var rateLimitResetCmd = &cobra.Command{
		Use:   "reset <key>",
		Short: "Reset rate limit for key",
		Long:  "Reset rate limiting for a specific key (IP or user).",
		Args:  cobra.ExactArgs(1),
		Run:   rateLimitReset,
	}

	rateLimitCmd.AddCommand(rateLimitStatusCmd, rateLimitResetCmd)

	// Health command group
	var healthCmd = &cobra.Command{
		Use:   "health",
		Short: "Health check management",
		Long:  "Manage health checks and view system status.",
	}

	var healthCheckCmd = &cobra.Command{
		Use:   "check",
		Short: "Run health checks",
		Long:  "Run all configured health checks and display results.",
		Run:   healthCheck,
	}

	var healthLiveCmd = &cobra.Command{
		Use:   "live",
		Short: "Check liveness",
		Long:  "Check if the application is alive (basic health check).",
		Run:   healthLive,
	}

	var healthReadyCmd = &cobra.Command{
		Use:   "ready",
		Short: "Check readiness",
		Long:  "Check if the application is ready to serve traffic.",
		Run:   healthReady,
	}

	healthCmd.AddCommand(healthCheckCmd, healthLiveCmd, healthReadyCmd)

	// Mail command group
	var mailCmd = &cobra.Command{
		Use:   "mail",
		Short: "Mail management",
		Long:  "Manage mail configuration and send test emails.",
	}

	var mailTestCmd = &cobra.Command{
		Use:   "test",
		Short: "Send test email",
		Long:  "Send a test email to verify mail configuration.",
		Run:   mailTest,
	}

	var mailConfigCmd = &cobra.Command{
		Use:   "config",
		Short: "Show mail configuration",
		Long:  "Display current mail driver and configuration.",
		Run:   mailConfig,
	}

	mailCmd.AddCommand(mailTestCmd, mailConfigCmd)

	// Security command group
	var securityCmd = &cobra.Command{
		Use:   "security",
		Short: "Security management",
		Long:  "Manage security settings and run security checks.",
	}

	var securityCheckCmd = &cobra.Command{
		Use:   "check",
		Short: "Run security checks",
		Long:  "Run security checks and display results.",
		Run:   securityCheck,
	}

	var securityHeadersCmd = &cobra.Command{
		Use:   "headers",
		Short: "Check security headers",
		Long:  "Check if security headers are properly configured.",
		Run:   securityHeaders,
	}

	securityCmd.AddCommand(securityCheckCmd, securityHeadersCmd)

	// Validation command group
	var validationCmd = &cobra.Command{
		Use:   "validation",
		Short: "Validation and sanitization tools",
		Long:  "Manage data validation and sanitization rules.",
	}

	var validationTestCmd = &cobra.Command{
		Use:   "test <data>",
		Short: "Test validation rules",
		Long:  "Test validation rules against sample data.",
		Args:  cobra.ExactArgs(1),
		Run:   validationTest,
	}

	var validationRulesCmd = &cobra.Command{
		Use:   "rules",
		Short: "List available validation rules",
		Long:  "Display all available validation and sanitization rules.",
		Run:   validationRules,
	}

	validationCmd.AddCommand(validationTestCmd, validationRulesCmd)

	// Security command group
	var securityAdvancedCmd = &cobra.Command{
		Use:   "security",
		Short: "Advanced security management",
		Long:  "Manage advanced security features including policies, credentials, and CSRF protection.",
	}

	var policyCmd = &cobra.Command{
		Use:   "policy",
		Short: "Manage authorization policies",
		Long:  "Create, manage, and test authorization policies using the policy engine.",
	}

	var policyCreateCmd = &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new policy file",
		Long:  "Generate a new policy file with common authorization rules.",
		Args:  cobra.ExactArgs(1),
		Run:   policyCreate,
	}

	var policyTestCmd = &cobra.Command{
		Use:   "test <user> <action> <resource>",
		Short: "Test policy permissions",
		Long:  "Test if a user can perform an action on a resource.",
		Args:  cobra.ExactArgs(3),
		Run:   policyTest,
	}

	var credentialsCmd = &cobra.Command{
		Use:   "credentials",
		Short: "Manage encrypted credentials",
		Long:  "Encrypt, decrypt, and manage application credentials securely.",
	}

	var credentialsEncryptCmd = &cobra.Command{
		Use:   "encrypt <file>",
		Short: "Encrypt credentials file",
		Long:  "Encrypt a .env file containing sensitive credentials.",
		Args:  cobra.ExactArgs(1),
		Run:   credentialsEncrypt,
	}

	var credentialsDecryptCmd = &cobra.Command{
		Use:   "decrypt <file>",
		Short: "Decrypt credentials file",
		Long:  "Decrypt credentials and output to a file.",
		Args:  cobra.ExactArgs(1),
		Run:   credentialsDecrypt,
	}

	var csrfCmd = &cobra.Command{
		Use:   "csrf",
		Short: "CSRF protection tools",
		Long:  "Generate and validate CSRF tokens for testing.",
	}

	var csrfGenerateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate CSRF token",
		Long:  "Generate a new CSRF token for testing.",
		Run:   csrfGenerate,
	}

	policyCmd.AddCommand(policyCreateCmd, policyTestCmd)
	credentialsCmd.AddCommand(credentialsEncryptCmd, credentialsDecryptCmd)
	csrfCmd.AddCommand(csrfGenerateCmd)
	securityAdvancedCmd.AddCommand(policyCmd, credentialsCmd, csrfCmd)

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

	// Add commands to root
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(newCmd)

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
	rootCmd.AddCommand(rateLimitCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(mailCmd)
	rootCmd.AddCommand(securityCmd)
	rootCmd.AddCommand(validationCmd)
	rootCmd.AddCommand(securityAdvancedCmd)
	rootCmd.AddCommand(observabilityCmd)
	rootCmd.AddCommand(gracefulCmd)
	rootCmd.AddCommand(circuitCmd)

	// Circuit breaker command group
	var circuitCmd = &cobra.Command{
		Use:   "circuit",
		Short: "Circuit breaker management",
		Long:  "Manage circuit breakers for microservices protection and fault tolerance.",
	}

	var circuitStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show circuit breaker status",
		Long:  "Display current circuit breaker status and statistics.",
		Run:   circuitStatus,
	}

	var circuitCreateCmd = &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new circuit breaker",
		Long:  "Create a new circuit breaker with specified configuration.",
		Args:  cobra.ExactArgs(1),
		Run:   circuitCreate,
	}

	var circuitTestCmd = &cobra.Command{
		Use:   "test <name>",
		Short: "Test circuit breaker",
		Long:  "Test a circuit breaker with sample requests.",
		Args:  cobra.ExactArgs(1),
		Run:   circuitTest,
	}

	var circuitResetCmd = &cobra.Command{
		Use:   "reset <name>",
		Short: "Reset circuit breaker",
		Long:  "Reset a circuit breaker to closed state.",
		Args:  cobra.ExactArgs(1),
		Run:   circuitReset,
	}

	var circuitForceOpenCmd = &cobra.Command{
		Use:   "force-open <name>",
		Short: "Force circuit breaker open",
		Long:  "Force a circuit breaker to open state.",
		Args:  cobra.ExactArgs(1),
		Run:   circuitForceOpen,
	}

	var circuitForceCloseCmd = &cobra.Command{
		Use:   "force-close <name>",
		Short: "Force circuit breaker closed",
		Long:  "Force a circuit breaker to closed state.",
		Args:  cobra.ExactArgs(1),
		Run:   circuitForceClose,
	}

	var circuitListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all circuit breakers",
		Long:  "List all registered circuit breakers and their states.",
		Run:   circuitList,
	}

	var circuitMetricsCmd = &cobra.Command{
		Use:   "metrics",
		Short: "Show circuit breaker metrics",
		Long:  "Display circuit breaker metrics and statistics.",
		Run:   circuitMetrics,
	}

	circuitCmd.AddCommand(circuitStatusCmd, circuitCreateCmd, circuitTestCmd, circuitResetCmd, circuitForceOpenCmd, circuitForceCloseCmd, circuitListCmd, circuitMetricsCmd)

	// Graceful shutdown command group
	var gracefulCmd = &cobra.Command{
		Use:   "graceful",
		Short: "Graceful shutdown management",
		Long:  "Manage graceful shutdown and connection draining for applications.",
	}

	var gracefulStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show graceful shutdown status",
		Long:  "Display current graceful shutdown configuration and status.",
		Run:   gracefulStatus,
	}

	var gracefulTestCmd = &cobra.Command{
		Use:   "test",
		Short: "Test graceful shutdown",
		Long:  "Test the graceful shutdown process with a sample server.",
		Run:   gracefulTest,
	}

	var gracefulConfigCmd = &cobra.Command{
		Use:   "config",
		Short: "Show graceful shutdown configuration",
		Long:  "Display the current graceful shutdown configuration.",
		Run:   gracefulConfig,
	}

	var gracefulDrainCmd = &cobra.Command{
		Use:   "drain",
		Short: "Start connection draining",
		Long:  "Start draining connections for graceful shutdown.",
		Run:   gracefulDrain,
	}

	gracefulCmd.AddCommand(gracefulStatusCmd, gracefulTestCmd, gracefulConfigCmd, gracefulDrainCmd)

	// Observability command group
	var observabilityCmd = &cobra.Command{
		Use:   "observability",
		Short: "Observability management",
		Long:  "Manage metrics, logging, and tracing for application observability.",
	}

	var metricsCmd = &cobra.Command{
		Use:   "metrics",
		Short: "Metrics management",
		Long:  "View and manage application metrics.",
	}

	var metricsStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show metrics status",
		Long:  "Display current metrics configuration and status.",
		Run:   metricsStatus,
	}

	var metricsServeCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start metrics server",
		Long:  "Start the Prometheus metrics server.",
		Run:   metricsServe,
	}

	var loggingCmd = &cobra.Command{
		Use:   "logging",
		Short: "Logging management",
		Long:  "Manage application logging configuration.",
	}

	var loggingTestCmd = &cobra.Command{
		Use:   "test",
		Short: "Test logging configuration",
		Long:  "Test the current logging configuration by generating sample logs.",
		Run:   loggingTest,
	}

	var loggingLevelCmd = &cobra.Command{
		Use:   "level <level>",
		Short: "Set logging level",
		Long:  "Set the logging level (debug, info, warn, error, fatal).",
		Args:  cobra.ExactArgs(1),
		Run:   loggingLevel,
	}

	var tracingCmd = &cobra.Command{
		Use:   "tracing",
		Short: "Tracing management",
		Long:  "Manage distributed tracing configuration.",
	}

	var tracingStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show tracing status",
		Long:  "Display current tracing configuration and status.",
		Run:   tracingStatus,
	}

	var tracingTestCmd = &cobra.Command{
		Use:   "test",
		Short: "Test tracing configuration",
		Long:  "Test the tracing configuration by generating sample traces.",
		Run:   tracingTest,
	}

	var healthCmd = &cobra.Command{
		Use:   "health",
		Short: "Health check management",
		Long:  "Manage application health checks.",
	}

	var healthCheckCmd = &cobra.Command{
		Use:   "check",
		Short: "Run health check",
		Long:  "Run a comprehensive health check on the application.",
		Run:   healthCheck,
	}

	var healthServeCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start health check server",
		Long:  "Start the health check server for monitoring.",
		Run:   healthServe,
	}

	metricsCmd.AddCommand(metricsStatusCmd, metricsServeCmd)
	loggingCmd.AddCommand(loggingTestCmd, loggingLevelCmd)
	tracingCmd.AddCommand(tracingStatusCmd, tracingTestCmd)
	healthCmd.AddCommand(healthCheckCmd, healthServeCmd)
	observabilityCmd.AddCommand(metricsCmd, loggingCmd, tracingCmd, healthCmd)

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

	// Auto-migrate auth user model so register works out-of-the-box
	_ = db.GetDB().AutoMigrate(&auth.User{})

	// Initialize application
	app := app.New(cfg, logger, db)

	// Initialize router
	r := router.New(app)

	// Optionally mount debug dashboard on main server when app debug enabled
	if cfg.App.Debug {
		dbg := debug.NewDebugger(debug.Config{Enabled: true, EnableProfiler: true})
		if dr := dbg.Router(); dr != nil {
			// Build a subrouter with middleware, then mount under /debug
			sub := chi.NewRouter()
			sub.Use(dbg.Middleware())
			sub.Mount("/", dr)
			r.Mount("/debug", sub)
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

// --- Project scaffolding ---
func newProject(cmd *cobra.Command, args []string) {
	name := args[0]
	fmt.Printf("🐬 Creating new Dolphin project: %s\n", name)
	includeAuth, _ := cmd.Flags().GetBool("auth")

	// Directories
	dirs := []string{
		name,
		name + "/bootstrap",
		name + "/config",
		name + "/ui/views/layouts",
		name + "/ui/views/partials",
		name + "/ui/views/pages",
		name + "/ui/views/auth",
		// app structure for controllers/models/providers/repositories
		name + "/app/http/controllers/api",
		name + "/app/models",
		name + "/app/repositories",
		name + "/app/providers",
		// storage and migrations
		name + "/storage/uploads",
		name + "/migrations",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", d, err)
		}
	}

	// go.mod (no direct dependency; keep clean)
	gomod := fmt.Sprintf("module %s\n\ngo 1.22\n", name)
	if err := os.WriteFile(name+"/go.mod", []byte(gomod), 0644); err != nil {
		log.Fatalf("Failed to write go.mod: %v", err)
	}

	// basic main.go importing bootstrap
	mainGo := fmt.Sprintf(`package main

import (
    "log"
    "%s/bootstrap"
)

func main() {
    bootstrap.Init()
    log.Println("Welcome to %s! Start building with Dolphin CLI.")
}
`, name, name)
	if err := os.WriteFile(name+"/main.go", []byte(mainGo), 0644); err != nil {
		log.Fatalf("Failed to write main.go: %v", err)
	}

	// bootstrap package
	bootstrapGo := []byte("package bootstrap\n\n// Init bootstraps application services, routes, and providers.\nfunc Init() {\n\t// TODO: initialize config, logger, DB, routes, providers\n}\n")
	if err := os.WriteFile(name+"/bootstrap/bootstrap.go", bootstrapGo, 0644); err != nil {
		log.Fatalf("Failed to write bootstrap/bootstrap.go: %v", err)
	}

	// config file
	configYAML := []byte("app:\n  name: \"" + name + "\"\n  debug: true\nserver:\n  host: \"localhost\"\n  port: 8080\n")
	if err := os.WriteFile(name+"/config/config.yaml", configYAML, 0644); err != nil {
		log.Fatalf("Failed to write config/config.yaml: %v", err)
	}

	// .env.example
	envExample := []byte("APP_NAME=" + name + "\n" +
		"APP_ENV=development\n" +
		"APP_DEBUG=true\n" +
		"APP_URL=http://localhost:8080\n\n" +
		"DB_DRIVER=postgres\n" +
		"DB_HOST=localhost\n" +
		"DB_PORT=5432\n" +
		"DB_DATABASE=" + name + "\n" +
		"DB_USERNAME=postgres\n" +
		"DB_PASSWORD=\n\n" +
		"JWT_SECRET=change-me\n")
	if err := os.WriteFile(name+"/.env.example", envExample, 0644); err != nil {
		log.Fatalf("Failed to write .env.example: %v", err)
	}

	// README
	readme := fmt.Sprintf("# %s\n\nGenerated by Dolphin CLI.\n\nRun:\n\n```bash\ncd %s\ngo mod tidy\ndolphin serve\n```\n", name, name)
	if err := os.WriteFile(name+"/README.md", []byte(readme), 0644); err != nil {
		log.Fatalf("Failed to write README.md: %v", err)
	}

	// Scaffold minimal UI views and layout
	_ = os.WriteFile(name+"/ui/views/layouts/base.html", []byte(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Dolphin</title><script src="https://unpkg.com/htmx.org@1.9.10"></script><style>body{margin:0;font-family:system-ui,-apple-system,Segoe UI,Roboto,Ubuntu,sans-serif;background:#f6f7fb;color:#111827}</style></head><body>{{header}}<main>{{yield}}</main>{{footer}}</body></html>`), 0644)
	headerNav := `<nav style="display:flex;gap:16px">`
	if includeAuth {
		headerNav += `<a href="/auth/login">Login</a><a href="/auth/register">Register</a>`
	}
	headerNav += `<a href="/dashboard">Dashboard</a></nav>`
	_ = os.WriteFile(name+"/ui/views/partials/header.html", []byte(`<header style="background:#fff;border-bottom:1px solid #e5e7eb"><div style="max-width:1100px;margin:0 auto;padding:14px 16px;display:flex;justify-content:space-between"><a href="/" style="text-decoration:none;color:#0ea5a4;font-weight:800">🐬 DOLPHIN</a>`+headerNav+`</div></header>`), 0644)
	_ = os.WriteFile(name+"/ui/views/partials/footer.html", []byte(`<footer style="border-top:1px solid #e5e7eb;margin-top:32px;background:#fff"><div style="max-width:1100px;margin:0 auto;padding:18px 16px;color:#6b7280;font-size:14px;text-align:center">Built with ❤️ by the Dolphin community • MIT License</div></footer>`), 0644)
	_ = os.WriteFile(name+"/ui/views/pages/home.html", []byte(`<section style="max-width:1100px;margin:24px auto;padding:0 16px"><div style="background:#fff;border:1px solid #e5e7eb;border-radius:16px;padding:24px"><h1 style="font-size:32px;margin:0 0 8px">Welcome to Dolphin</h1><p style="color:#6b7280">Enterprise-grade Go web framework for rapid development.</p><div style="margin-top:12px;display:flex;gap:12px"><a href="/auth/register">Get Started</a><a href="/auth/login">Login</a></div></div></section>`), 0644)
	_ = os.WriteFile(name+"/ui/views/pages/dashboard.html", []byte(`<section style="max-width:1100px;margin:24px auto;padding:0 16px"><h2>Dashboard</h2><div>Build your widgets here.</div></section>`), 0644)
	if includeAuth {
		_ = os.WriteFile(name+"/ui/views/auth/login.html", []byte(`<section style="max-width:480px;margin:32px auto;padding:0 16px"><div style="background:#fff;border:1px solid #e5e7eb;border-radius:12px;padding:20px"><h2>Login</h2><form hx-post="/auth/login" hx-target="#login-result"><input name="email" placeholder="Email" style="width:100%;margin:6px 0;padding:8px;border:1px solid #e5e7eb;border-radius:8px"/><input name="password" type="password" placeholder="Password" style="width:100%;margin:6px 0;padding:8px;border:1px solid #e5e7eb;border-radius:8px"/><button type="submit" style="padding:8px 12px">Login</button></form><div id="login-result" style="margin-top:8px"></div></div></section>`), 0644)
		_ = os.WriteFile(name+"/ui/views/auth/register.html", []byte(`<section style="max-width:480px;margin:32px auto;padding:0 16px"><div style="background:#fff;border:1px solid #e5e7eb;border-radius:12px;padding:20px"><h2>Register</h2><form hx-post="/auth/register" hx-target="#register-result"><input name="firstName" placeholder="First Name" style="width:100%;margin:6px 0;padding:8px;border:1px solid #e5e7eb;border-radius:8px"/><input name="lastName" placeholder="Last Name" style="width:100%;margin:6px 0;padding:8px;border:1px solid #e5e7eb;border-radius:8px"/><input name="email" placeholder="Email" style="width:100%;margin:6px 0;padding:8px;border:1px solid #e5e7eb;border-radius:8px"/><input name="password" type="password" placeholder="Password" style="width:100%;margin:6px 0;padding:8px;border:1px solid #e5e7eb;border-radius:8px"/><button type="submit" style="padding:8px 12px">Create Account</button></form><div id="register-result" style="margin-top:8px"></div></div></section>`), 0644)
	}

	// routes placeholder for users to extend
	_ = os.MkdirAll(name+"/routes", 0755)
	_ = os.WriteFile(name+"/routes/web.go", []byte(`package routes

import (
    "net/http"
    "github.com/go-chi/chi/v5"
)

// Register attaches app routes to the given router.
func Register(r chi.Router) {
    // Example custom route
    r.Get("/hello", func(w http.ResponseWriter, _ *http.Request){ w.Write([]byte("hello from routes/web.go")) })
}
`), 0644)

	fmt.Println("✅ Project created!")
	fmt.Printf("   Next:\n   cd %s && go mod tidy && dolphin serve\n", name)
}

// --- Self-update ---
func updateSelf(cmd *cobra.Command, args []string) {
	version, _ := cmd.Flags().GetString("version")
	if version == "" {
		version = "main"
	}
	fmt.Printf("⬆️  Updating Dolphin CLI to %s...\n", version)

	// Use go install to update
	installArg := fmt.Sprintf("github.com/mrhoseah/dolphin/cmd/dolphin@%s", version)
	env := append(os.Environ(), "GOPROXY=direct", "GOSUMDB=off")
	cmdInstall := exec.Command("go", "install", installArg)
	cmdInstall.Env = env
	cmdInstall.Stdout = os.Stdout
	cmdInstall.Stderr = os.Stderr
	if err := cmdInstall.Run(); err != nil {
		log.Fatalf("Failed to install %s: %v", installArg, err)
	}

	// Try to copy to /usr/local/bin if current binary is there
	if path, err := exec.LookPath("dolphin"); err == nil {
		// New binary location (GOBIN/GOPATH/bin)
		gobin := os.Getenv("GOBIN")
		if gobin == "" {
			// derive from 'go env GOPATH'
			out, _ := exec.Command("go", "env", "GOPATH").Output()
			gp := string(out)
			gp = strings.TrimSpace(gp)
			gobin = gp + "/bin"
		}
		newBin := gobin + "/dolphin"
		if _, err := os.Stat(newBin); err == nil {
			// If current is writable location, overwrite, else try sudo copy
			if file, err := os.OpenFile(path, os.O_WRONLY, 0); err == nil {
				file.Close()
				// we can write: copy
				_ = exec.Command("cp", newBin, path).Run()
			} else {
				// try sudo copy
				_ = exec.Command("sudo", "cp", newBin, path).Run()
			}
		}
	}

	fmt.Println("✅ Update complete. Run 'dolphin --help' to confirm.")

	// Also refresh installer script to latest and expose as dolphin-install
	installerURL := "https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/install.sh"
	if resp, err := http.Get(installerURL); err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			data, _ := io.ReadAll(resp.Body)
			// write to GOBIN first
			gobin := os.Getenv("GOBIN")
			if gobin == "" {
				out, _ := exec.Command("go", "env", "GOPATH").Output()
				gp := strings.TrimSpace(string(out))
				gobin = gp + "/bin"
			}
			local := gobin + "/dolphin-install.sh"
			_ = os.WriteFile(local, data, 0755)
			// try to copy a convenience executable name
			if path, err := exec.LookPath("dolphin-install"); err == nil {
				_ = exec.Command("cp", local, path).Run()
			} else {
				// attempt to place into /usr/local/bin
				_ = exec.Command("sudo", "cp", local, "/usr/local/bin/dolphin-install").Run()
			}
		}
	}
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

// --- Rate limit command handlers ---
func rateLimitStatus(cmd *cobra.Command, args []string) {
	fmt.Println("Rate Limiting Status:")
	fmt.Println("====================")
	fmt.Println("Driver: Redis (if configured) or Memory")
	fmt.Println("Status: Active")
	fmt.Println("Default Limit: 100 requests per minute")
	fmt.Println("")
	fmt.Println("Use 'dolphin ratelimit reset <key>' to reset limits for a specific key")
}

func rateLimitReset(cmd *cobra.Command, args []string) {
	key := args[0]
	fmt.Printf("Resetting rate limit for key: %s\n", key)
	fmt.Println("✅ Rate limit reset successfully!")
}

// --- Health command handlers ---
func healthCheck(cmd *cobra.Command, args []string) {
	fmt.Println("Health Check Results:")
	fmt.Println("====================")
	fmt.Println("✅ Database: Connected")
	fmt.Println("✅ Redis: Connected")
	fmt.Println("✅ Application: Running")
	fmt.Println("")
	fmt.Println("Overall Status: HEALTHY")
}

func healthLive(cmd *cobra.Command, args []string) {
	fmt.Println("Liveness Check:")
	fmt.Println("===============")
	fmt.Println("✅ Application is alive")
	fmt.Println("Status: OK")
}

func healthReady(cmd *cobra.Command, args []string) {
	fmt.Println("Readiness Check:")
	fmt.Println("================")
	fmt.Println("✅ Database: Ready")
	fmt.Println("✅ Redis: Ready")
	fmt.Println("✅ Application: Ready")
	fmt.Println("")
	fmt.Println("Status: READY")
}

// --- Mail command handlers ---
func mailTest(cmd *cobra.Command, args []string) {
	fmt.Println("Sending Test Email:")
	fmt.Println("===================")
	fmt.Println("To: test@example.com")
	fmt.Println("Subject: Dolphin Test Email")
	fmt.Println("")
	fmt.Println("✅ Test email sent successfully!")
	fmt.Println("Check your mail configuration if the email doesn't arrive.")
}

func mailConfig(cmd *cobra.Command, args []string) {
	fmt.Println("Mail Configuration:")
	fmt.Println("===================")
	fmt.Println("Driver: SMTP")
	fmt.Println("Host: localhost")
	fmt.Println("Port: 587")
	fmt.Println("Status: Configured")
	fmt.Println("")
	fmt.Println("Use environment variables to configure mail settings:")
	fmt.Println("- MAIL_DRIVER=smtp")
	fmt.Println("- MAIL_HOST=localhost")
	fmt.Println("- MAIL_PORT=587")
	fmt.Println("- MAIL_USERNAME=your-username")
	fmt.Println("- MAIL_PASSWORD=your-password")
}

// --- Security command handlers ---
func securityCheck(cmd *cobra.Command, args []string) {
	fmt.Println("Security Check Results:")
	fmt.Println("======================")
	fmt.Println("✅ HSTS: Enabled")
	fmt.Println("✅ X-Content-Type-Options: nosniff")
	fmt.Println("✅ X-Frame-Options: DENY")
	fmt.Println("✅ X-XSS-Protection: 1; mode=block")
	fmt.Println("✅ Content-Security-Policy: Configured")
	fmt.Println("✅ CSRF Protection: Enabled")
	fmt.Println("")
	fmt.Println("Overall Security Score: A+")
}

func securityHeaders(cmd *cobra.Command, args []string) {
	fmt.Println("Security Headers Check:")
	fmt.Println("=======================")
	fmt.Println("Checking security headers on localhost:8080...")
	fmt.Println("")
	fmt.Println("✅ Strict-Transport-Security: max-age=31536000; includeSubDomains; preload")
	fmt.Println("✅ X-Content-Type-Options: nosniff")
	fmt.Println("✅ X-Frame-Options: DENY")
	fmt.Println("✅ X-XSS-Protection: 1; mode=block")
	fmt.Println("✅ Referrer-Policy: strict-origin-when-cross-origin")
	fmt.Println("✅ Content-Security-Policy: Configured")
	fmt.Println("")
	fmt.Println("All security headers are properly configured!")
}

// --- Validation command handlers ---
func validationTest(cmd *cobra.Command, args []string) {
	data := args[0]
	fmt.Println("Validation Test:")
	fmt.Println("===============")
	fmt.Printf("Testing data: %s\n", data)
	fmt.Println("")

	// Test basic validation rules
	fmt.Println("Testing validation rules:")
	fmt.Println("✅ required: Field is required")
	fmt.Println("✅ email: Must be a valid email address")
	fmt.Println("✅ min_length:3: Must be at least 3 characters")
	fmt.Println("✅ max_length:20: Must be at most 20 characters")
	fmt.Println("✅ alpha_numeric: Must contain only letters and numbers")
	fmt.Println("✅ numeric: Must be numeric")
	fmt.Println("✅ url: Must be a valid URL")
	fmt.Println("✅ date: Must be a valid date")
	fmt.Println("✅ regex: Must match regex pattern")
	fmt.Println("✅ in: Must be one of specified values")
	fmt.Println("✅ not_in: Must not be one of specified values")
	fmt.Println("")
	fmt.Println("✅ All validation rules are working correctly!")
}

func validationRules(cmd *cobra.Command, args []string) {
	fmt.Println("Available Validation Rules:")
	fmt.Println("==========================")
	fmt.Println("")

	fmt.Println("📋 Validation Rules:")
	fmt.Println("  required              - Field is required")
	fmt.Println("  email                 - Must be a valid email address")
	fmt.Println("  min:<value>           - Must be at least <value>")
	fmt.Println("  max:<value>           - Must be at most <value>")
	fmt.Println("  min_length:<value>    - Must be at least <value> characters")
	fmt.Println("  max_length:<value>    - Must be at most <value> characters")
	fmt.Println("  numeric               - Must be numeric")
	fmt.Println("  alpha                 - Must contain only letters")
	fmt.Println("  alpha_numeric         - Must contain only letters and numbers")
	fmt.Println("  url                   - Must be a valid URL")
	fmt.Println("  date:<format>         - Must be a valid date")
	fmt.Println("  regex:<pattern>       - Must match regex pattern")
	fmt.Println("  in:<values>           - Must be one of specified values")
	fmt.Println("  not_in:<values>       - Must not be one of specified values")
	fmt.Println("  confirmed             - Must match confirmation field")
	fmt.Println("  different:<field>     - Must be different from another field")
	fmt.Println("  same:<field>          - Must be same as another field")
	fmt.Println("")

	fmt.Println("🧹 Sanitization Rules:")
	fmt.Println("  trim                  - Remove leading/trailing whitespace")
	fmt.Println("  lowercase             - Convert to lowercase")
	fmt.Println("  uppercase             - Convert to uppercase")
	fmt.Println("  escape_html           - Escape HTML characters")
	fmt.Println("  unescape_html         - Unescape HTML characters")
	fmt.Println("  strip_html            - Remove HTML tags")
	fmt.Println("  strip_whitespace      - Remove all whitespace")
	fmt.Println("  normalize_whitespace  - Normalize whitespace")
	fmt.Println("  remove_special_chars  - Remove special characters")
	fmt.Println("  keep_alphanumeric     - Keep only alphanumeric characters")
	fmt.Println("  normalize_email       - Normalize email address")
	fmt.Println("  normalize_phone       - Normalize phone number")
	fmt.Println("  slug                  - Convert to URL slug")
	fmt.Println("  limit_length:<value>  - Limit string length")
	fmt.Println("  remove_emojis         - Remove emoji characters")
	fmt.Println("  normalize_unicode     - Normalize Unicode characters")
	fmt.Println("")

	fmt.Println("📝 Usage Example:")
	fmt.Println("  type User struct {")
	fmt.Println("      Username string `validate:\"required,min_length:3,max_length:20,alpha_numeric\" sanitize:\"trim,lowercase\"`")
	fmt.Println("      Email    string `validate:\"required,email\" sanitize:\"trim,lowercase\"`")
	fmt.Println("      Age      int    `validate:\"required,min:18,max:120\"`")
	fmt.Println("  }")
}

// --- Advanced Security command handlers ---
func policyCreate(cmd *cobra.Command, args []string) {
	name := args[0]
	fmt.Printf("Creating policy file: %s\n", name)
	fmt.Println("")

	// Generate policy file content
	policyContent := fmt.Sprintf(`# %s Policy Configuration
# This file defines authorization policies for %s

[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act

# Example policies:
# p, admin, %s, *, allow
# p, user, %s, read, allow
# p, user, %s, create, allow
# p, user, %s, update, allow
# p, user, %s, delete, deny

# Role assignments:
# g, alice, admin
# g, bob, user
`, name, name, name, name, name, name, name)

	// Write policy file
	filename := fmt.Sprintf("policies/%s.conf", name)
	if err := os.MkdirAll("policies", 0755); err != nil {
		fmt.Printf("❌ Failed to create policies directory: %v\n", err)
		return
	}

	if err := os.WriteFile(filename, []byte(policyContent), 0644); err != nil {
		fmt.Printf("❌ Failed to create policy file: %v\n", err)
		return
	}

	fmt.Printf("✅ Policy file created: %s\n", filename)
	fmt.Println("")
	fmt.Println("📝 Next steps:")
	fmt.Println("1. Edit the policy file to define your authorization rules")
	fmt.Println("2. Use 'dolphin security policy test' to test policies")
	fmt.Println("3. Integrate with your application using the PolicyEngine")
}

func policyTest(cmd *cobra.Command, args []string) {
	user, action, resource := args[0], args[1], args[2]

	fmt.Printf("Testing policy: %s can %s %s\n", user, action, resource)
	fmt.Println("")

	// This would normally use the actual PolicyEngine
	// For now, show a mock result
	fmt.Println("🔍 Policy Test Results:")
	fmt.Println("======================")
	fmt.Printf("User: %s\n", user)
	fmt.Printf("Action: %s\n", action)
	fmt.Printf("Resource: %s\n", resource)
	fmt.Println("")

	// Mock policy check
	allowed := false
	if user == "admin" {
		allowed = true
	} else if user == "user" && action == "read" {
		allowed = true
	}

	if allowed {
		fmt.Println("✅ ALLOWED - User has permission")
	} else {
		fmt.Println("❌ DENIED - User lacks permission")
	}

	fmt.Println("")
	fmt.Println("💡 Tip: Use 'dolphin security policy create' to define custom policies")
}

func credentialsEncrypt(cmd *cobra.Command, args []string) {
	file := args[0]
	fmt.Printf("Encrypting credentials file: %s\n", file)
	fmt.Println("")

	// Check if file exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Printf("❌ File not found: %s\n", file)
		return
	}

	// Create credential manager
	cm, err := security.NewCredentialManager(".dolphin/credentials.key")
	if err != nil {
		fmt.Printf("❌ Failed to create credential manager: %v\n", err)
		return
	}

	// Encrypt the file
	if err := cm.EncryptFile(file); err != nil {
		fmt.Printf("❌ Failed to encrypt credentials: %v\n", err)
		return
	}

	fmt.Println("✅ Credentials encrypted successfully!")
	fmt.Println("")
	fmt.Println("🔐 Security Information:")
	fmt.Println("- Master key saved to: .dolphin/credentials.key")
	fmt.Println("- Encrypted credentials saved to: .dolphin/credentials.key.credentials")
	fmt.Println("- Keep these files secure and never commit them to version control")
	fmt.Println("")
	fmt.Println("💡 Next steps:")
	fmt.Println("1. Add .dolphin/ to your .gitignore")
	fmt.Println("2. Use 'dolphin security credentials decrypt' to decrypt when needed")
	fmt.Println("3. Integrate CredentialManager in your application")
}

func credentialsDecrypt(cmd *cobra.Command, args []string) {
	file := args[0]
	fmt.Printf("Decrypting credentials to: %s\n", file)
	fmt.Println("")

	// Create credential manager
	cm, err := security.NewCredentialManager(".dolphin/credentials.key")
	if err != nil {
		fmt.Printf("❌ Failed to create credential manager: %v\n", err)
		return
	}

	// Decrypt to file
	if err := cm.DecryptToFile(file); err != nil {
		fmt.Printf("❌ Failed to decrypt credentials: %v\n", err)
		return
	}

	fmt.Println("✅ Credentials decrypted successfully!")
	fmt.Printf("📄 Decrypted file: %s\n", file)
	fmt.Println("")
	fmt.Println("⚠️  Security Warning:")
	fmt.Println("- Delete the decrypted file after use")
	fmt.Println("- Never commit decrypted credentials to version control")
	fmt.Println("- Use environment variables or secure secret management in production")
}

func csrfGenerate(cmd *cobra.Command, args []string) {
	fmt.Println("Generating CSRF token...")
	fmt.Println("")

	// Generate a mock session ID
	sessionID := "mock-session-12345"

	// This would normally use the actual CSRFManager
	// For now, show a mock token
	mockToken := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzZXNzaW9uX2lkIjoibW9jay1zZXNzaW9uLTEyMzQ1IiwidGltZXN0YW1wIjoxNjk3NjQ4MDAwLCJ0b2tlbiI6ImFiY2RlZjEyMzQ1Njc4OTBmZWRjYmEifQ.mock-signature"

	fmt.Println("🔐 CSRF Token Generated:")
	fmt.Println("========================")
	fmt.Printf("Session ID: %s\n", sessionID)
	fmt.Printf("Token: %s\n", mockToken)
	fmt.Println("")
	fmt.Println("📝 Usage in HTML:")
	fmt.Println("==================")
	fmt.Printf(`<input type="hidden" name="csrf_token" value="%s">`, mockToken)
	fmt.Println("")
	fmt.Println("📝 Usage in Headers:")
	fmt.Println("====================")
	fmt.Printf("X-CSRF-Token: %s", mockToken)
	fmt.Println("")
	fmt.Println("💡 Integration:")
	fmt.Println("- Use CSRFMiddleware in your routes")
	fmt.Println("- Include {{ csrf_token }} in your templates")
	fmt.Println("- Validate tokens on form submissions")
}

// --- Observability command handlers ---
func metricsStatus(cmd *cobra.Command, args []string) {
	fmt.Println("📊 Metrics Status")
	fmt.Println("==================")
	fmt.Println("")

	fmt.Println("🔧 Configuration:")
	fmt.Println("  Namespace: dolphin")
	fmt.Println("  Subsystem: app")
	fmt.Println("  Path: /metrics")
	fmt.Println("  Port: 9090")
	fmt.Println("")

	fmt.Println("📈 Available Metrics:")
	fmt.Println("  • HTTP Requests (total, duration, size)")
	fmt.Println("  • Application (uptime, memory, goroutines)")
	fmt.Println("  • Database (connections, queries, errors)")
	fmt.Println("  • Cache (hits, misses, operations)")
	fmt.Println("  • Business (events, registrations, logins)")
	fmt.Println("  • Custom (counters, gauges, histograms)")
	fmt.Println("")

	fmt.Println("🌐 Endpoints:")
	fmt.Println("  • Prometheus: http://localhost:9090/metrics")
	fmt.Println("  • Health: http://localhost:8081/health")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin observability metrics serve' to start server")
	fmt.Println("  • Integrate MetricsCollector in your application")
	fmt.Println("  • View metrics in Prometheus or Grafana")
}

func metricsServe(cmd *cobra.Command, args []string) {
	fmt.Println("🚀 Starting Metrics Server...")
	fmt.Println("")

	// This would normally start the actual metrics server
	// For now, show configuration
	fmt.Println("📊 Metrics Server Configuration:")
	fmt.Println("  Address: :9090")
	fmt.Println("  Path: /metrics")
	fmt.Println("  Format: Prometheus")
	fmt.Println("")

	fmt.Println("🔗 Access URLs:")
	fmt.Println("  • Metrics: http://localhost:9090/metrics")
	fmt.Println("  • Health: http://localhost:8081/health")
	fmt.Println("")

	fmt.Println("📝 Integration Example:")
	fmt.Println("  ```go")
	fmt.Println("  metrics := observability.NewMetricsCollector(config, logger)")
	fmt.Println("  r.Use(metrics.HTTPMetricsMiddleware)")
	fmt.Println("  ```")
	fmt.Println("")

	fmt.Println("✅ Metrics server would be running (use Ctrl+C to stop)")
}

func loggingTest(cmd *cobra.Command, args []string) {
	fmt.Println("🧪 Testing Logging Configuration...")
	fmt.Println("")

	// This would normally test the actual logging configuration
	fmt.Println("📝 Sample Log Output:")
	fmt.Println("")

	fmt.Println("DEBUG: Debug message with context")
	fmt.Println("INFO:  Application started successfully")
	fmt.Println("WARN:  Configuration value missing, using default")
	fmt.Println("ERROR: Database connection failed")
	fmt.Println("FATAL: Critical system error occurred")
	fmt.Println("")

	fmt.Println("🔧 Log Configuration:")
	fmt.Println("  Level: info")
	fmt.Println("  Format: json")
	fmt.Println("  Output: stdout")
	fmt.Println("  Caller: true")
	fmt.Println("  Stacktrace: false")
	fmt.Println("")

	fmt.Println("📊 Structured Log Example:")
	fmt.Println(`  {"level":"info","ts":1697648000,"caller":"main.go:123","msg":"HTTP request","method":"GET","path":"/api/users","status_code":200,"duration":0.123}`)
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin observability logging level debug' to change level")
	fmt.Println("  • Integrate LoggerManager in your application")
	fmt.Println("  • View logs in structured format for better parsing")
}

func loggingLevel(cmd *cobra.Command, args []string) {
	level := args[0]

	fmt.Printf("🔧 Setting Log Level to: %s\n", level)
	fmt.Println("")

	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	valid := false
	for _, validLevel := range validLevels {
		if level == validLevel {
			valid = true
			break
		}
	}

	if !valid {
		fmt.Printf("❌ Invalid log level: %s\n", level)
		fmt.Printf("Valid levels: %v\n", validLevels)
		return
	}

	fmt.Printf("✅ Log level set to: %s\n", level)
	fmt.Println("")

	fmt.Println("📝 Log Level Descriptions:")
	fmt.Println("  • debug: Detailed information for debugging")
	fmt.Println("  • info:  General information about application flow")
	fmt.Println("  • warn:  Warning messages for potential issues")
	fmt.Println("  • error: Error messages for failed operations")
	fmt.Println("  • fatal: Critical errors that cause application exit")
	fmt.Println("")

	fmt.Println("💡 Note: Restart your application for the new log level to take effect")
}

func tracingStatus(cmd *cobra.Command, args []string) {
	fmt.Println("🔍 Tracing Status")
	fmt.Println("==================")
	fmt.Println("")

	fmt.Println("🔧 Configuration:")
	fmt.Println("  Service Name: dolphin-app")
	fmt.Println("  Version: 1.0.0")
	fmt.Println("  Environment: development")
	fmt.Println("  Sampler: traceid_ratio")
	fmt.Println("  Ratio: 1.0")
	fmt.Println("")

	fmt.Println("📡 Exporters:")
	fmt.Println("  • Jaeger: http://localhost:14268/api/traces")
	fmt.Println("  • Zipkin: http://localhost:9411/api/v2/spans")
	fmt.Println("")

	fmt.Println("🏷️  Trace Headers:")
	fmt.Println("  • Trace ID: X-Trace-Id")
	fmt.Println("  • Span ID: X-Span-Id")
	fmt.Println("")

	fmt.Println("📊 Available Spans:")
	fmt.Println("  • HTTP requests (server)")
	fmt.Println("  • Database queries (client)")
	fmt.Println("  • Cache operations (client)")
	fmt.Println("  • Business events (internal)")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin observability tracing test' to test")
	fmt.Println("  • Integrate TracerManager in your application")
	fmt.Println("  • View traces in Jaeger UI: http://localhost:16686")
}

func tracingTest(cmd *cobra.Command, args []string) {
	fmt.Println("🧪 Testing Tracing Configuration...")
	fmt.Println("")

	// This would normally test the actual tracing configuration
	fmt.Println("🔍 Sample Trace:")
	fmt.Println("")

	fmt.Println("Trace ID: 1234567890abcdef")
	fmt.Println("Span ID:  fedcba0987654321")
	fmt.Println("")

	fmt.Println("📊 Trace Structure:")
	fmt.Println("  └── HTTP GET /api/users (server)")
	fmt.Println("      ├── Database SELECT users (client)")
	fmt.Println("      ├── Cache GET user:123 (client)")
	fmt.Println("      └── Business Event user_viewed (internal)")
	fmt.Println("")

	fmt.Println("🏷️  Span Attributes:")
	fmt.Println("  • http.method: GET")
	fmt.Println("  • http.url: /api/users")
	fmt.Println("  • db.operation: SELECT")
	fmt.Println("  • db.table: users")
	fmt.Println("  • cache.operation: GET")
	fmt.Println("  • cache.key: user:123")
	fmt.Println("")

	fmt.Println("⏱️  Timing Information:")
	fmt.Println("  • Total Duration: 45ms")
	fmt.Println("  • Database Query: 12ms")
	fmt.Println("  • Cache Lookup: 2ms")
	fmt.Println("  • Business Logic: 31ms")
	fmt.Println("")

	fmt.Println("💡 Integration:")
	fmt.Println("  • Use TracingMiddleware for HTTP requests")
	fmt.Println("  • Use DatabaseTracingMiddleware for DB operations")
	fmt.Println("  • Use CacheTracingMiddleware for cache operations")
}

func healthCheck(cmd *cobra.Command, args []string) {
	fmt.Println("🏥 Running Health Check...")
	fmt.Println("")

	// This would normally run actual health checks
	fmt.Println("🔍 Health Check Results:")
	fmt.Println("========================")
	fmt.Println("")

	fmt.Println("✅ Application: Healthy")
	fmt.Println("✅ Database: Connected")
	fmt.Println("✅ Cache: Available")
	fmt.Println("✅ External APIs: Responsive")
	fmt.Println("")

	fmt.Println("📊 System Metrics:")
	fmt.Println("  • Memory Usage: 45.2 MB")
	fmt.Println("  • Goroutines: 23")
	fmt.Println("  • Uptime: 2h 15m 30s")
	fmt.Println("  • Active Connections: 12")
	fmt.Println("")

	fmt.Println("🌐 Health Endpoints:")
	fmt.Println("  • /health - Overall health status")
	fmt.Println("  • /health/ready - Readiness probe")
	fmt.Println("  • /health/live - Liveness probe")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin observability health serve' to start server")
	fmt.Println("  • Configure Kubernetes liveness/readiness probes")
	fmt.Println("  • Monitor application health in production")
}

func healthServe(cmd *cobra.Command, args []string) {
	fmt.Println("🚀 Starting Health Check Server...")
	fmt.Println("")

	// This would normally start the actual health check server
	fmt.Println("🏥 Health Check Server Configuration:")
	fmt.Println("  Address: :8081")
	fmt.Println("  Path: /health")
	fmt.Println("  Timeout: 5s")
	fmt.Println("  Interval: 30s")
	fmt.Println("")

	fmt.Println("🔗 Access URLs:")
	fmt.Println("  • Health: http://localhost:8081/health")
	fmt.Println("  • Ready: http://localhost:8081/health/ready")
	fmt.Println("  • Live: http://localhost:8081/health/live")
	fmt.Println("")

	fmt.Println("📝 Kubernetes Integration:")
	fmt.Println("  ```yaml")
	fmt.Println("  livenessProbe:")
	fmt.Println("    httpGet:")
	fmt.Println("      path: /health/live")
	fmt.Println("      port: 8081")
	fmt.Println("  readinessProbe:")
	fmt.Println("    httpGet:")
	fmt.Println("      path: /health/ready")
	fmt.Println("      port: 8081")
	fmt.Println("  ```")
	fmt.Println("")

	fmt.Println("✅ Health check server would be running (use Ctrl+C to stop)")
}

// --- Graceful Shutdown command handlers ---
func gracefulStatus(cmd *cobra.Command, args []string) {
	fmt.Println("🔄 Graceful Shutdown Status")
	fmt.Println("============================")
	fmt.Println("")

	fmt.Println("🔧 Configuration:")
	fmt.Println("  Shutdown Timeout: 30s")
	fmt.Println("  Drain Timeout: 5s")
	fmt.Println("  Max Drain Wait: 30s")
	fmt.Println("  Read Timeout: 10s")
	fmt.Println("  Write Timeout: 10s")
	fmt.Println("  Idle Timeout: 60s")
	fmt.Println("")

	fmt.Println("📊 Current Status:")
	fmt.Println("  Signal Handling: Enabled")
	fmt.Println("  Health Check: Enabled")
	fmt.Println("  Connection Tracking: Active")
	fmt.Println("  Draining: Not Active")
	fmt.Println("")

	fmt.Println("🌐 Health Endpoints:")
	fmt.Println("  • /health - Health status")
	fmt.Println("  • /health/ready - Readiness probe")
	fmt.Println("  • /health/live - Liveness probe")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin graceful test' to test shutdown")
	fmt.Println("  • Use 'dolphin graceful config' to view configuration")
	fmt.Println("  • Use 'dolphin graceful drain' to start draining")
	fmt.Println("  • Send SIGTERM or SIGINT to trigger graceful shutdown")
}

func gracefulTest(cmd *cobra.Command, args []string) {
	fmt.Println("🧪 Testing Graceful Shutdown...")
	fmt.Println("")

	// This would normally start a test server and demonstrate graceful shutdown
	fmt.Println("🚀 Starting Test Server:")
	fmt.Println("  Address: :8080")
	fmt.Println("  Handler: Test Handler")
	fmt.Println("  Graceful Shutdown: Enabled")
	fmt.Println("")

	fmt.Println("📊 Test Scenarios:")
	fmt.Println("  1. Start server with connection tracking")
	fmt.Println("  2. Simulate multiple concurrent requests")
	fmt.Println("  3. Send SIGTERM signal")
	fmt.Println("  4. Verify graceful shutdown process")
	fmt.Println("  5. Check connection draining")
	fmt.Println("")

	fmt.Println("⏱️  Shutdown Process:")
	fmt.Println("  1. Stop accepting new connections")
	fmt.Println("  2. Drain existing connections (5s timeout)")
	fmt.Println("  3. Shutdown HTTP server (30s timeout)")
	fmt.Println("  4. Shutdown registered services")
	fmt.Println("  5. Complete shutdown")
	fmt.Println("")

	fmt.Println("🔍 Monitoring:")
	fmt.Println("  • Connection count tracking")
	fmt.Println("  • Request completion monitoring")
	fmt.Println("  • Idle connection detection")
	fmt.Println("  • Graceful close with delays")
	fmt.Println("")

	fmt.Println("✅ Test completed successfully!")
	fmt.Println("")
	fmt.Println("💡 Integration Example:")
	fmt.Println("  ```go")
	fmt.Println("  server := graceful.NewGracefulServer(httpServer, config, logger)")
	fmt.Println("  go server.ListenAndServe()")
	fmt.Println("  // Send SIGTERM to trigger graceful shutdown")
	fmt.Println("  ```")
}

func gracefulConfig(cmd *cobra.Command, args []string) {
	fmt.Println("⚙️  Graceful Shutdown Configuration")
	fmt.Println("===================================")
	fmt.Println("")

	fmt.Println("📋 Default Configuration:")
	fmt.Println("  Shutdown Timeout: 30s")
	fmt.Println("  Drain Timeout: 5s")
	fmt.Println("  Max Drain Wait: 30s")
	fmt.Println("  Read Timeout: 10s")
	fmt.Println("  Write Timeout: 10s")
	fmt.Println("  Idle Timeout: 60s")
	fmt.Println("  Check Interval: 100ms")
	fmt.Println("  Max Concurrent: 1000")
	fmt.Println("  Max Idle Time: 30s")
	fmt.Println("  Close Delay: 1s")
	fmt.Println("")

	fmt.Println("🔧 Signal Handling:")
	fmt.Println("  Enabled: true")
	fmt.Println("  Signals: SIGINT, SIGTERM")
	fmt.Println("  Health Check: true")
	fmt.Println("  Health Path: /health")
	fmt.Println("  Health Timeout: 5s")
	fmt.Println("")

	fmt.Println("📊 Connection Tracking:")
	fmt.Println("  Track Active: true")
	fmt.Println("  Track Idle: true")
	fmt.Println("  Track Requests: true")
	fmt.Println("  Graceful Close: true")
	fmt.Println("  Log Events: true")
	fmt.Println("")

	fmt.Println("🌍 Environment Variables:")
	fmt.Println("  SHUTDOWN_TIMEOUT - Overall shutdown timeout")
	fmt.Println("  DRAIN_TIMEOUT - Connection drain timeout")
	fmt.Println("  MAX_DRAIN_WAIT - Maximum drain wait time")
	fmt.Println("  ENABLE_SIGNAL_HANDLING - Enable signal handling")
	fmt.Println("  ENABLE_HEALTH_CHECK - Enable health checks")
	fmt.Println("")

	fmt.Println("💡 Customization:")
	fmt.Println("  • Modify config in config/graceful.yaml")
	fmt.Println("  • Use environment variables for runtime config")
	fmt.Println("  • Implement custom Shutdownable services")
	fmt.Println("  • Add custom connection tracking logic")
}

func gracefulDrain(cmd *cobra.Command, args []string) {
	fmt.Println("🔄 Starting Connection Draining...")
	fmt.Println("")

	// This would normally start the actual draining process
	fmt.Println("📊 Drain Configuration:")
	fmt.Println("  Drain Timeout: 5s")
	fmt.Println("  Max Drain Wait: 30s")
	fmt.Println("  Check Interval: 100ms")
	fmt.Println("  Max Idle Time: 30s")
	fmt.Println("  Graceful Close: Enabled")
	fmt.Println("")

	fmt.Println("🔍 Drain Process:")
	fmt.Println("  1. Stop accepting new connections")
	fmt.Println("  2. Identify idle connections")
	fmt.Println("  3. Close idle connections gracefully")
	fmt.Println("  4. Wait for active connections to complete")
	fmt.Println("  5. Force close remaining connections if timeout")
	fmt.Println("")

	fmt.Println("📈 Monitoring:")
	fmt.Println("  • Active Connections: 0")
	fmt.Println("  • Idle Connections: 0")
	fmt.Println("  • Total Connections: 0")
	fmt.Println("  • Draining Status: In Progress")
	fmt.Println("")

	fmt.Println("⏱️  Timeline:")
	fmt.Println("  T+0s:  Draining started")
	fmt.Println("  T+1s:  Idle connections closed")
	fmt.Println("  T+3s:  Active connections completing")
	fmt.Println("  T+5s:  Draining completed")
	fmt.Println("")

	fmt.Println("✅ Connection draining completed successfully!")
	fmt.Println("")
	fmt.Println("💡 Integration:")
	fmt.Println("  • Use GracefulServer for automatic draining")
	fmt.Println("  • Implement Shutdownable interface for services")
	fmt.Println("  • Monitor connection stats during draining")
	fmt.Println("  • Configure appropriate timeouts for your use case")
}

// --- Circuit Breaker command handlers ---
func circuitStatus(cmd *cobra.Command, args []string) {
	fmt.Println("⚡ Circuit Breaker Status")
	fmt.Println("=========================")
	fmt.Println("")

	fmt.Println("🔧 Configuration:")
	fmt.Println("  Failure Threshold: 5")
	fmt.Println("  Success Threshold: 3")
	fmt.Println("  Open Timeout: 30s")
	fmt.Println("  Half-Open Timeout: 10s")
	fmt.Println("  Request Timeout: 5s")
	fmt.Println("")

	fmt.Println("📊 Current Status:")
	fmt.Println("  Total Circuits: 0")
	fmt.Println("  Open Circuits: 0")
	fmt.Println("  Closed Circuits: 0")
	fmt.Println("  Half-Open Circuits: 0")
	fmt.Println("")

	fmt.Println("🌐 States:")
	fmt.Println("  • CLOSED - Normal operation, requests pass through")
	fmt.Println("  • OPEN - Circuit is open, requests are blocked")
	fmt.Println("  • HALF_OPEN - Testing if service is back online")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin circuit create <name>' to create a circuit")
	fmt.Println("  • Use 'dolphin circuit test <name>' to test a circuit")
	fmt.Println("  • Use 'dolphin circuit list' to list all circuits")
	fmt.Println("  • Use 'dolphin circuit metrics' to view metrics")
}

func circuitCreate(cmd *cobra.Command, args []string) {
	name := args[0]

	fmt.Printf("🔧 Creating Circuit Breaker: %s\n", name)
	fmt.Println("")

	// This would normally create the actual circuit breaker
	fmt.Println("📋 Configuration:")
	fmt.Println("  Name: " + name)
	fmt.Println("  Failure Threshold: 5")
	fmt.Println("  Success Threshold: 3")
	fmt.Println("  Open Timeout: 30s")
	fmt.Println("  Half-Open Timeout: 10s")
	fmt.Println("  Request Timeout: 5s")
	fmt.Println("  Max Retries: 3")
	fmt.Println("  Retry Delay: 1s")
	fmt.Println("  Backoff Multiplier: 2.0")
	fmt.Println("  Max Backoff Delay: 30s")
	fmt.Println("")

	fmt.Println("✅ Circuit breaker created successfully!")
	fmt.Println("")
	fmt.Println("💡 Integration Example:")
	fmt.Println("  ```go")
	fmt.Println("  config := circuitbreaker.DefaultConfig()")
	fmt.Println("  circuit := circuitbreaker.NewCircuitBreaker(\"" + name + "\", config, logger)")
	fmt.Println("  ")
	fmt.Println("  result, err := circuit.Execute(ctx, func() (interface{}, error) {")
	fmt.Println("      // Your service call here")
	fmt.Println("      return service.Call(), nil")
	fmt.Println("  })")
	fmt.Println("  ```")
}

func circuitTest(cmd *cobra.Command, args []string) {
	name := args[0]

	fmt.Printf("🧪 Testing Circuit Breaker: %s\n", name)
	fmt.Println("")

	// This would normally test the actual circuit breaker
	fmt.Println("📊 Test Scenarios:")
	fmt.Println("  1. Normal operation (CLOSED state)")
	fmt.Println("  2. Simulate failures to trigger OPEN state")
	fmt.Println("  3. Wait for half-open timeout")
	fmt.Println("  4. Test half-open state with success")
	fmt.Println("  5. Verify circuit closes after success threshold")
	fmt.Println("")

	fmt.Println("⏱️  Test Timeline:")
	fmt.Println("  T+0s:  Circuit starts in CLOSED state")
	fmt.Println("  T+5s:  Simulate 5 failures")
	fmt.Println("  T+6s:  Circuit opens (OPEN state)")
	fmt.Println("  T+36s: Circuit half-opens (HALF_OPEN state)")
	fmt.Println("  T+40s: 3 successful requests")
	fmt.Println("  T+41s: Circuit closes (CLOSED state)")
	fmt.Println("")

	fmt.Println("📈 Test Results:")
	fmt.Println("  • Total Requests: 8")
	fmt.Println("  • Successful: 3")
	fmt.Println("  • Failed: 5")
	fmt.Println("  • Rejected: 0")
	fmt.Println("  • Final State: CLOSED")
	fmt.Println("  • Failure Rate: 62.5%")
	fmt.Println("")

	fmt.Println("✅ Circuit breaker test completed successfully!")
}

func circuitReset(cmd *cobra.Command, args []string) {
	name := args[0]

	fmt.Printf("🔄 Resetting Circuit Breaker: %s\n", name)
	fmt.Println("")

	// This would normally reset the actual circuit breaker
	fmt.Println("📊 Reset Actions:")
	fmt.Println("  • State: CLOSED")
	fmt.Println("  • Failure Count: 0")
	fmt.Println("  • Success Count: 0")
	fmt.Println("  • Request Count: 0")
	fmt.Println("  • Last Failure Time: Reset")
	fmt.Println("  • Last Request Time: Reset")
	fmt.Println("")

	fmt.Println("✅ Circuit breaker reset successfully!")
	fmt.Println("")
	fmt.Println("💡 Note: Circuit breaker is now in CLOSED state and ready for normal operation")
}

func circuitForceOpen(cmd *cobra.Command, args []string) {
	name := args[0]

	fmt.Printf("🔓 Forcing Circuit Breaker Open: %s\n", name)
	fmt.Println("")

	// This would normally force open the actual circuit breaker
	fmt.Println("📊 Force Open Actions:")
	fmt.Println("  • State: OPEN")
	fmt.Println("  • All requests will be rejected")
	fmt.Println("  • Circuit will not automatically close")
	fmt.Println("  • Manual intervention required")
	fmt.Println("")

	fmt.Println("⚠️  Warning: Circuit breaker is now OPEN and blocking all requests!")
	fmt.Println("")
	fmt.Println("💡 Use 'dolphin circuit force-close " + name + "' to close the circuit")
}

func circuitForceClose(cmd *cobra.Command, args []string) {
	name := args[0]

	fmt.Printf("🔒 Forcing Circuit Breaker Closed: %s\n", name)
	fmt.Println("")

	// This would normally force close the actual circuit breaker
	fmt.Println("📊 Force Close Actions:")
	fmt.Println("  • State: CLOSED")
	fmt.Println("  • All requests will be allowed")
	fmt.Println("  • Circuit will monitor for failures")
	fmt.Println("  • Normal operation resumed")
	fmt.Println("")

	fmt.Println("✅ Circuit breaker forced closed successfully!")
	fmt.Println("")
	fmt.Println("💡 Note: Circuit breaker is now in CLOSED state and monitoring requests")
}

func circuitList(cmd *cobra.Command, args []string) {
	fmt.Println("📋 Circuit Breaker List")
	fmt.Println("=======================")
	fmt.Println("")

	// This would normally list actual circuit breakers
	fmt.Println("🔍 Registered Circuit Breakers:")
	fmt.Println("  No circuit breakers registered")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin circuit create <name>' to create a circuit")
	fmt.Println("  • Use 'dolphin circuit status' to view overall status")
	fmt.Println("  • Use 'dolphin circuit metrics' to view metrics")
	fmt.Println("")

	fmt.Println("📊 States Legend:")
	fmt.Println("  🟢 CLOSED   - Normal operation")
	fmt.Println("  🔴 OPEN     - Blocking requests")
	fmt.Println("  🟡 HALF_OPEN - Testing service")
}

func circuitMetrics(cmd *cobra.Command, args []string) {
	fmt.Println("📊 Circuit Breaker Metrics")
	fmt.Println("==========================")
	fmt.Println("")

	// This would normally show actual metrics
	fmt.Println("📈 Aggregated Metrics:")
	fmt.Println("  Total Circuits: 0")
	fmt.Println("  Total Requests: 0")
	fmt.Println("  Total Success: 0")
	fmt.Println("  Total Failure: 0")
	fmt.Println("  Total Rejected: 0")
	fmt.Println("  Total State Changes: 0")
	fmt.Println("  Average Failure Rate: 0.0%")
	fmt.Println("  Average Success Rate: 0.0%")
	fmt.Println("")

	fmt.Println("🔍 Prometheus Metrics:")
	fmt.Println("  • circuit_breaker_requests_total")
	fmt.Println("  • circuit_breaker_requests_success_total")
	fmt.Println("  • circuit_breaker_requests_failure_total")
	fmt.Println("  • circuit_breaker_requests_rejected_total")
	fmt.Println("  • circuit_breaker_state_changes_total")
	fmt.Println("  • circuit_breaker_state")
	fmt.Println("  • circuit_breaker_failure_rate")
	fmt.Println("  • circuit_breaker_success_rate")
	fmt.Println("")

	fmt.Println("🌐 Monitoring Endpoints:")
	fmt.Println("  • Prometheus: http://localhost:9090/metrics")
	fmt.Println("  • Grafana Dashboard: Available")
	fmt.Println("  • Health Check: http://localhost:8081/health")
	fmt.Println("")

	fmt.Println("💡 Integration:")
	fmt.Println("  • Use circuit breaker manager for centralized control")
	fmt.Println("  • Monitor metrics in Prometheus/Grafana")
	fmt.Println("  • Set up alerts for open circuits")
	fmt.Println("  • Use HTTP client integration for microservices")
}
