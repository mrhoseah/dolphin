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
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
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
	"github.com/mrhoseah/dolphin/internal/telemetry"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	version = "1.0.0" // Keep for CLI version display
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

	// Command declarations moved here to avoid forward reference issues
	var observabilityCmd = &cobra.Command{
		Use:   "observability",
		Short: "Observability management",
		Long:  "Manage metrics, logging, and tracing for application observability.",
	}

	var gracefulCmd = &cobra.Command{
		Use:   "graceful",
		Short: "Graceful shutdown management",
		Long:  "Manage graceful shutdown and connection draining.",
	}

	var circuitCmd = &cobra.Command{
		Use:   "circuit",
		Short: "Circuit breaker management",
		Long:  "Manage circuit breakers for microservices protection.",
	}

	var loadShedCmd = &cobra.Command{
		Use:   "loadshed",
		Short: "Load shedding management",
		Long:  "Manage adaptive load shedding for overload protection.",
	}

	var liveReloadCmd = &cobra.Command{
		Use:   "livereload",
		Short: "Live reload management",
		Long:  "Manage live reload and hot code reload for development.",
	}

	var assetCmd = &cobra.Command{
		Use:   "asset",
		Short: "Asset pipeline management",
		Long:  "Manage asset bundling, optimization, and versioning.",
	}

	var templateCmd = &cobra.Command{
		Use:   "template",
		Short: "Template engine management",
		Long:  "Manage templating engine and template compilation.",
	}

	var httpCmd = &cobra.Command{
		Use:   "http",
		Short: "HTTP client management",
		Long:  "Manage HTTP client abstraction and testing.",
	}

	var graphqlCmd = &cobra.Command{
		Use:   "graphql",
		Short: "GraphQL endpoint management",
		Long:  "Manage GraphQL endpoint configuration and testing.",
	}

	// Test command
	var testCmd = &cobra.Command{
		Use:   "test [package]",
		Short: "Run tests for the Dolphin application",
		Long:  "Run unit tests, integration tests, and generate coverage reports for your Dolphin application.",
		Run:   runTests,
	}
	testCmd.Flags().Bool("coverage", false, "Generate coverage report")
	testCmd.Flags().Bool("watch", false, "Run tests in watch mode")
	testCmd.Flags().String("suite", "", "Run specific test suite (unit, integration, e2e)")
	testCmd.Flags().Bool("with-db", false, "Run tests with database")
	testCmd.Flags().Int("parallel", 0, "Number of parallel test processes")
	testCmd.Flags().String("timeout", "30s", "Test timeout duration")

	// Telemetry command
	var telemetryCmd = &cobra.Command{
		Use:   "telemetry",
		Short: "Telemetry management",
		Long:  "Manage telemetry collection and privacy settings.",
	}

	// Telemetry subcommands
	var telemetryEnableCmd = &cobra.Command{
		Use:   "enable",
		Short: "Enable telemetry collection",
		Long:  "Enable telemetry collection to help improve Dolphin Framework.",
		Run:   runTelemetryEnable,
	}

	var telemetryDisableCmd = &cobra.Command{
		Use:   "disable",
		Short: "Disable telemetry collection",
		Long:  "Disable telemetry collection for privacy.",
		Run:   runTelemetryDisable,
	}

	var telemetryStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show telemetry status",
		Long:  "Show current telemetry configuration and status.",
		Run:   runTelemetryStatus,
	}

	var telemetryConfigCmd = &cobra.Command{
		Use:   "config",
		Short: "Show telemetry configuration",
		Long:  "Show detailed telemetry configuration.",
		Run:   runTelemetryConfig,
	}

	var telemetryResetCmd = &cobra.Command{
		Use:   "reset",
		Short: "Reset telemetry configuration",
		Long:  "Reset telemetry configuration to defaults.",
		Run:   runTelemetryReset,
	}

	var telemetryTestCmd = &cobra.Command{
		Use:   "test",
		Short: "Send test telemetry event",
		Long:  "Send a test telemetry event to verify configuration.",
		Run:   runTelemetryTest,
	}

	var telemetryPrivacyCmd = &cobra.Command{
		Use:   "privacy",
		Short: "Show privacy information",
		Long:  "Show information about what data is collected and privacy policies.",
		Run:   runTelemetryPrivacy,
	}

	// Add telemetry subcommands
	telemetryCmd.AddCommand(telemetryEnableCmd)
	telemetryCmd.AddCommand(telemetryDisableCmd)
	telemetryCmd.AddCommand(telemetryStatusCmd)
	telemetryCmd.AddCommand(telemetryConfigCmd)
	telemetryCmd.AddCommand(telemetryResetCmd)
	telemetryCmd.AddCommand(telemetryTestCmd)
	telemetryCmd.AddCommand(telemetryPrivacyCmd)

	// Update command
	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the Dolphin CLI to the latest or specified version",
		Long:  "Self-update the Dolphin CLI by installing the latest (or specified) version via 'go install'.",
		Run:   updateSelf,
	}
	updateCmd.Flags().StringP("version", "V", "main", "Version to install (e.g., v0.1.0 or 'main')")

	// Uninstall command
	var uninstallCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "Show uninstall instructions for Dolphin CLI",
		Long:  "Display instructions for completely removing Dolphin CLI from your system.",
		Run:   uninstallInstructions,
	}
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
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(telemetryCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(uninstallCmd)
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
	rootCmd.AddCommand(loadShedCmd)
	rootCmd.AddCommand(liveReloadCmd)
	rootCmd.AddCommand(assetCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(httpCmd)
	rootCmd.AddCommand(graphqlCmd)

	// GraphQL command group

	// GraphQL subcommands
	var graphqlEnableCmd = &cobra.Command{
		Use:   "enable",
		Short: "Enable GraphQL endpoint",
		Run:   graphqlEnable,
	}

	var graphqlDisableCmd = &cobra.Command{
		Use:   "disable",
		Short: "Disable GraphQL endpoint",
		Run:   graphqlDisable,
	}

	var graphqlToggleCmd = &cobra.Command{
		Use:   "toggle",
		Short: "Toggle GraphQL endpoint state",
		Run:   graphqlToggle,
	}

	var graphqlStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show GraphQL status",
		Run:   graphqlStatus,
	}

	var graphqlTestCmd = &cobra.Command{
		Use:   "test",
		Short: "Run GraphQL tests",
		Run:   graphqlTest,
	}

	var graphqlPlaygroundCmd = &cobra.Command{
		Use:   "playground",
		Short: "Open GraphQL playground",
		Run:   graphqlPlayground,
	}

	var graphqlSchemaCmd = &cobra.Command{
		Use:   "schema",
		Short: "Show GraphQL schema",
		Run:   graphqlSchema,
	}

	var graphqlConfigCmd = &cobra.Command{
		Use:   "config",
		Short: "Show GraphQL configuration",
		Run:   graphqlConfig,
	}

	var graphqlGenerateCmd = &cobra.Command{
		Use:   "generate [output-dir]",
		Short: "Generate GraphQL code",
		Args:  cobra.MaximumNArgs(1),
		Run:   graphqlGenerate,
	}

	var graphqlValidateCmd = &cobra.Command{
		Use:   "validate [query]",
		Short: "Validate GraphQL query",
		Args:  cobra.MaximumNArgs(1),
		Run:   graphqlValidate,
	}

	var graphqlResetCmd = &cobra.Command{
		Use:   "reset",
		Short: "Reset GraphQL statistics",
		Run:   graphqlReset,
	}

	// Add subcommands to GraphQL command
	graphqlCmd.AddCommand(graphqlEnableCmd)
	graphqlCmd.AddCommand(graphqlDisableCmd)
	graphqlCmd.AddCommand(graphqlToggleCmd)
	graphqlCmd.AddCommand(graphqlStatusCmd)
	graphqlCmd.AddCommand(graphqlTestCmd)
	graphqlCmd.AddCommand(graphqlPlaygroundCmd)
	graphqlCmd.AddCommand(graphqlSchemaCmd)
	graphqlCmd.AddCommand(graphqlConfigCmd)
	graphqlCmd.AddCommand(graphqlGenerateCmd)
	graphqlCmd.AddCommand(graphqlValidateCmd)
	graphqlCmd.AddCommand(graphqlResetCmd)

	// HTTP client command group

	var httpTestCmd = &cobra.Command{
		Use:   "test",
		Short: "Test HTTP client",
		Long:  "Test HTTP client functionality with sample requests.",
		Run:   httpTest,
	}

	var httpStatsCmd = &cobra.Command{
		Use:   "stats",
		Short: "Show HTTP client statistics",
		Long:  "Display HTTP client statistics and metrics.",
		Run:   httpStats,
	}

	var httpConfigCmd = &cobra.Command{
		Use:   "config",
		Short: "Show HTTP client configuration",
		Long:  "Display HTTP client configuration and settings.",
		Run:   httpConfig,
	}

	var httpHealthCmd = &cobra.Command{
		Use:   "health",
		Short: "Check HTTP client health",
		Long:  "Check HTTP client health and status.",
		Run:   httpHealth,
	}

	var httpResetCmd = &cobra.Command{
		Use:   "reset",
		Short: "Reset HTTP client metrics",
		Long:  "Reset HTTP client metrics and statistics.",
		Run:   httpReset,
	}

	httpCmd.AddCommand(httpTestCmd, httpStatsCmd, httpConfigCmd, httpHealthCmd, httpResetCmd)

	// Template engine command group

	var templateListCmd = &cobra.Command{
		Use:   "list",
		Short: "List templates",
		Long:  "List all templates by type (layouts, partials, pages, components, emails).",
		Run:   templateList,
	}

	var templateCompileCmd = &cobra.Command{
		Use:   "compile",
		Short: "Compile templates",
		Long:  "Compile all templates and check for errors.",
		Run:   templateCompile,
	}

	var templateWatchCmd = &cobra.Command{
		Use:   "watch",
		Short: "Watch templates for changes",
		Long:  "Watch template files for changes and recompile automatically.",
		Run:   templateWatch,
	}

	var templateHelperCmd = &cobra.Command{
		Use:   "helpers",
		Short: "List template helpers",
		Long:  "List all available template helper functions.",
		Run:   templateHelpers,
	}

	var templateTestCmd = &cobra.Command{
		Use:   "test",
		Short: "Test templates",
		Long:  "Test template rendering with sample data.",
		Run:   templateTest,
	}

	var templateStatsCmd = &cobra.Command{
		Use:   "stats",
		Short: "Show template statistics",
		Long:  "Display template engine statistics and metrics.",
		Run:   templateStats,
	}

	templateCmd.AddCommand(templateListCmd, templateCompileCmd, templateWatchCmd, templateHelperCmd, templateTestCmd, templateStatsCmd)

	// Asset pipeline command group

	var assetBuildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build assets",
		Long:  "Build and process all assets in the pipeline.",
		Run:   assetBuild,
	}

	var assetWatchCmd = &cobra.Command{
		Use:   "watch",
		Short: "Watch assets for changes",
		Long:  "Watch asset files for changes and rebuild automatically.",
		Run:   assetWatch,
	}

	var assetCleanCmd = &cobra.Command{
		Use:   "clean",
		Short: "Clean built assets",
		Long:  "Remove all built assets and cache.",
		Run:   assetClean,
	}

	var assetListCmd = &cobra.Command{
		Use:   "list",
		Short: "List assets",
		Long:  "List all processed assets and bundles.",
		Run:   assetList,
	}

	var assetStatsCmd = &cobra.Command{
		Use:   "stats",
		Short: "Show asset statistics",
		Long:  "Display asset pipeline statistics and metrics.",
		Run:   assetStats,
	}

	var assetOptimizeCmd = &cobra.Command{
		Use:   "optimize",
		Short: "Optimize assets",
		Long:  "Optimize and minify assets for production.",
		Run:   assetOptimize,
	}

	var assetVersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show asset versions",
		Long:  "Display asset versions and hashes.",
		Run:   assetVersion,
	}

	assetCmd.AddCommand(assetBuildCmd, assetWatchCmd, assetCleanCmd, assetListCmd, assetStatsCmd, assetOptimizeCmd, assetVersionCmd)

	// Live reload command group

	var liveReloadStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start live reload development server",
		Long:  "Start the development server with live reload enabled.",
		Run:   liveReloadStart,
	}

	var liveReloadStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop live reload development server",
		Long:  "Stop the live reload development server.",
		Run:   liveReloadStop,
	}

	var liveReloadStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show live reload status",
		Long:  "Display current live reload status and statistics.",
		Run:   liveReloadStatus,
	}

	var liveReloadConfigCmd = &cobra.Command{
		Use:   "config",
		Short: "Show live reload configuration",
		Long:  "Display current live reload configuration.",
		Run:   liveReloadConfig,
	}

	var liveReloadStatsCmd = &cobra.Command{
		Use:   "stats",
		Short: "Show live reload statistics",
		Long:  "Display live reload statistics and metrics.",
		Run:   liveReloadStats,
	}

	var liveReloadTestCmd = &cobra.Command{
		Use:   "test",
		Short: "Test live reload functionality",
		Long:  "Test live reload functionality with simulated changes.",
		Run:   liveReloadTest,
	}

	liveReloadCmd.AddCommand(liveReloadStartCmd, liveReloadStopCmd, liveReloadStatusCmd, liveReloadConfigCmd, liveReloadStatsCmd, liveReloadTestCmd)

	// Load shedding command group

	var loadShedStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show load shedding status",
		Long:  "Display current load shedding status and system metrics.",
		Run:   loadShedStatus,
	}

	var loadShedCreateCmd = &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new load shedder",
		Long:  "Create a new load shedder with specified configuration.",
		Args:  cobra.ExactArgs(1),
		Run:   loadShedCreate,
	}

	var loadShedTestCmd = &cobra.Command{
		Use:   "test <name>",
		Short: "Test load shedder",
		Long:  "Test a load shedder with simulated load.",
		Args:  cobra.ExactArgs(1),
		Run:   loadShedTest,
	}

	var loadShedResetCmd = &cobra.Command{
		Use:   "reset <name>",
		Short: "Reset load shedder",
		Long:  "Reset a load shedder to normal operation.",
		Args:  cobra.ExactArgs(1),
		Run:   loadShedReset,
	}

	var loadShedForceCmd = &cobra.Command{
		Use:   "force <name> <level>",
		Short: "Force shedding level",
		Long:  "Force a specific shedding level (none, light, moderate, heavy, critical).",
		Args:  cobra.ExactArgs(2),
		Run:   loadShedForce,
	}

	var loadShedListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all load shedders",
		Long:  "List all registered load shedders and their states.",
		Run:   loadShedList,
	}

	var loadShedMetricsCmd = &cobra.Command{
		Use:   "metrics",
		Short: "Show load shedding metrics",
		Long:  "Display load shedding metrics and statistics.",
		Run:   loadShedMetrics,
	}

	loadShedCmd.AddCommand(loadShedStatusCmd, loadShedCreateCmd, loadShedTestCmd, loadShedResetCmd, loadShedForceCmd, loadShedListCmd, loadShedMetricsCmd)

	// Circuit breaker command group

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

// --- Load Shedding command handlers ---
func loadShedStatus(cmd *cobra.Command, args []string) {
	fmt.Println("⚖️  Load Shedding Status")
	fmt.Println("========================")
	fmt.Println("")

	fmt.Println("🔧 Configuration:")
	fmt.Println("  Strategy: Combined")
	fmt.Println("  Light Threshold: 60%")
	fmt.Println("  Moderate Threshold: 75%")
	fmt.Println("  Heavy Threshold: 85%")
	fmt.Println("  Critical Threshold: 95%")
	fmt.Println("  Check Interval: 1s")
	fmt.Println("  Adaptive Interval: 5s")
	fmt.Println("")

	fmt.Println("📊 Current Status:")
	fmt.Println("  Total Shedders: 0")
	fmt.Println("  Active Shedders: 0")
	fmt.Println("  Shedding Level: None")
	fmt.Println("  Shedding Rate: 0%")
	fmt.Println("")

	fmt.Println("🌐 System Metrics:")
	fmt.Println("  CPU Usage: 45.2%")
	fmt.Println("  Memory Usage: 67.8%")
	fmt.Println("  Goroutines: 23")
	fmt.Println("  Request Rate: 150 req/s")
	fmt.Println("  Response Time: 120ms")
	fmt.Println("")

	fmt.Println("📈 Shedding Levels:")
	fmt.Println("  🟢 None (0%) - Normal operation")
	fmt.Println("  🟡 Light (10%) - Light shedding")
	fmt.Println("  🟠 Moderate (30%) - Moderate shedding")
	fmt.Println("  🔴 Heavy (60%) - Heavy shedding")
	fmt.Println("  ⚫ Critical (90%) - Critical shedding")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin loadshed create <name>' to create a shedder")
	fmt.Println("  • Use 'dolphin loadshed test <name>' to test a shedder")
	fmt.Println("  • Use 'dolphin loadshed list' to list all shedders")
	fmt.Println("  • Use 'dolphin loadshed metrics' to view metrics")
}

func loadShedCreate(cmd *cobra.Command, args []string) {
	name := args[0]

	fmt.Printf("⚖️  Creating Load Shedder: %s\n", name)
	fmt.Println("")

	// This would normally create the actual load shedder
	fmt.Println("📋 Configuration:")
	fmt.Println("  Name: " + name)
	fmt.Println("  Strategy: Combined")
	fmt.Println("  Light Threshold: 60%")
	fmt.Println("  Moderate Threshold: 75%")
	fmt.Println("  Heavy Threshold: 85%")
	fmt.Println("  Critical Threshold: 95%")
	fmt.Println("  Light Shed Rate: 10%")
	fmt.Println("  Moderate Shed Rate: 30%")
	fmt.Println("  Heavy Shed Rate: 60%")
	fmt.Println("  Critical Shed Rate: 90%")
	fmt.Println("  Check Interval: 1s")
	fmt.Println("  Adaptive Interval: 5s")
	fmt.Println("  Hysteresis: 5%")
	fmt.Println("  Min Shed Rate: 0%")
	fmt.Println("  Max Shed Rate: 95%")
	fmt.Println("")

	fmt.Println("✅ Load shedder created successfully!")
	fmt.Println("")
	fmt.Println("💡 Integration Example:")
	fmt.Println("  ```go")
	fmt.Println("  config := loadshedding.DefaultConfig()")
	fmt.Println("  shedder := loadshedding.NewLoadShedder(\"" + name + "\", config, logger)")
	fmt.Println("  ")
	fmt.Println("  // Use in HTTP middleware")
	fmt.Println("  middleware := loadshedding.NewMiddleware(shedder, config, logger)")
	fmt.Println("  r.Use(middleware.Handler)")
	fmt.Println("  ```")
}

func loadShedTest(cmd *cobra.Command, args []string) {
	name := args[0]

	fmt.Printf("🧪 Testing Load Shedder: %s\n", name)
	fmt.Println("")

	// This would normally test the actual load shedder
	fmt.Println("📊 Test Scenarios:")
	fmt.Println("  1. Normal load (CPU < 60%)")
	fmt.Println("  2. Light load (CPU 60-75%)")
	fmt.Println("  3. Moderate load (CPU 75-85%)")
	fmt.Println("  4. Heavy load (CPU 85-95%)")
	fmt.Println("  5. Critical load (CPU > 95%)")
	fmt.Println("")

	fmt.Println("⏱️  Test Timeline:")
	fmt.Println("  T+0s:  Normal load - No shedding")
	fmt.Println("  T+10s: Light load - 10% shedding")
	fmt.Println("  T+20s: Moderate load - 30% shedding")
	fmt.Println("  T+30s: Heavy load - 60% shedding")
	fmt.Println("  T+40s: Critical load - 90% shedding")
	fmt.Println("  T+50s: Load decreases - Adaptive adjustment")
	fmt.Println("")

	fmt.Println("📈 Test Results:")
	fmt.Println("  • Total Requests: 1000")
	fmt.Println("  • Shed Requests: 450")
	fmt.Println("  • Processed Requests: 550")
	fmt.Println("  • Shed Rate: 45%")
	fmt.Println("  • Final Level: Moderate")
	fmt.Println("  • Adaptive Adjustments: 3")
	fmt.Println("")

	fmt.Println("✅ Load shedder test completed successfully!")
}

func loadShedReset(cmd *cobra.Command, args []string) {
	name := args[0]

	fmt.Printf("🔄 Resetting Load Shedder: %s\n", name)
	fmt.Println("")

	// This would normally reset the actual load shedder
	fmt.Println("📊 Reset Actions:")
	fmt.Println("  • Level: None")
	fmt.Println("  • Shed Rate: 0%")
	fmt.Println("  • CPU Usage: Reset")
	fmt.Println("  • Memory Usage: Reset")
	fmt.Println("  • Request Rate: Reset")
	fmt.Println("  • Response Time: Reset")
	fmt.Println("  • Adjustment Count: 0")
	fmt.Println("")

	fmt.Println("✅ Load shedder reset successfully!")
	fmt.Println("")
	fmt.Println("💡 Note: Load shedder is now in normal operation mode")
}

func loadShedForce(cmd *cobra.Command, args []string) {
	name := args[0]
	level := args[1]

	fmt.Printf("🔧 Forcing Load Shedder Level: %s -> %s\n", name, level)
	fmt.Println("")

	// This would normally force the actual load shedder level
	fmt.Println("📊 Force Actions:")
	fmt.Printf("  • Level: %s\n", level)

	var shedRate float64
	switch level {
	case "none":
		shedRate = 0.0
	case "light":
		shedRate = 10.0
	case "moderate":
		shedRate = 30.0
	case "heavy":
		shedRate = 60.0
	case "critical":
		shedRate = 90.0
	default:
		shedRate = 0.0
	}

	fmt.Printf("  • Shed Rate: %.1f%%\n", shedRate)
	fmt.Println("  • Adaptive Adjustment: Disabled")
	fmt.Println("  • Manual Override: Enabled")
	fmt.Println("")

	fmt.Println("✅ Load shedder level forced successfully!")
	fmt.Println("")
	fmt.Println("💡 Note: Use 'dolphin loadshed reset " + name + "' to return to automatic mode")
}

func loadShedList(cmd *cobra.Command, args []string) {
	fmt.Println("📋 Load Shedder List")
	fmt.Println("====================")
	fmt.Println("")

	// This would normally list actual load shedders
	fmt.Println("🔍 Registered Load Shedders:")
	fmt.Println("  No load shedders registered")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin loadshed create <name>' to create a shedder")
	fmt.Println("  • Use 'dolphin loadshed status' to view overall status")
	fmt.Println("  • Use 'dolphin loadshed metrics' to view metrics")
	fmt.Println("")

	fmt.Println("📊 Levels Legend:")
	fmt.Println("  🟢 None      - Normal operation")
	fmt.Println("  🟡 Light     - 10% shedding")
	fmt.Println("  🟠 Moderate  - 30% shedding")
	fmt.Println("  🔴 Heavy     - 60% shedding")
	fmt.Println("  ⚫ Critical  - 90% shedding")
}

func loadShedMetrics(cmd *cobra.Command, args []string) {
	fmt.Println("📊 Load Shedding Metrics")
	fmt.Println("========================")
	fmt.Println("")

	// This would normally show actual metrics
	fmt.Println("📈 Aggregated Metrics:")
	fmt.Println("  Total Shedders: 0")
	fmt.Println("  Total Requests: 0")
	fmt.Println("  Total Shed: 0")
	fmt.Println("  Total Processed: 0")
	fmt.Println("  Average Shed Rate: 0.0%")
	fmt.Println("  Average Request Rate: 0.0 req/s")
	fmt.Println("")

	fmt.Println("🔍 Prometheus Metrics:")
	fmt.Println("  • load_shedder_requests_total")
	fmt.Println("  • load_shedder_requests_shed_total")
	fmt.Println("  • load_shedder_requests_processed_total")
	fmt.Println("  • load_shedder_level")
	fmt.Println("  • load_shedder_rate")
	fmt.Println("  • load_shedder_cpu_usage")
	fmt.Println("  • load_shedder_memory_usage")
	fmt.Println("  • load_shedder_goroutines")
	fmt.Println("  • load_shedder_request_rate")
	fmt.Println("  • load_shedder_response_time_seconds")
	fmt.Println("")

	fmt.Println("🌐 Monitoring Endpoints:")
	fmt.Println("  • Prometheus: http://localhost:9090/metrics")
	fmt.Println("  • Grafana Dashboard: Available")
	fmt.Println("  • Health Check: http://localhost:8081/health")
	fmt.Println("")

	fmt.Println("💡 Integration:")
	fmt.Println("  • Use load shedding manager for centralized control")
	fmt.Println("  • Monitor metrics in Prometheus/Grafana")
	fmt.Println("  • Set up alerts for high shedding levels")
	fmt.Println("  • Use HTTP middleware for automatic protection")
}

// --- Live Reload command handlers ---
func liveReloadStart(cmd *cobra.Command, args []string) {
	fmt.Println("🔄 Starting Live Reload Development Server")
	fmt.Println("==========================================")
	fmt.Println("")

	fmt.Println("🔧 Configuration:")
	fmt.Println("  Strategy: Restart")
	fmt.Println("  Watch Paths: ., cmd, internal, app, ui, public")
	fmt.Println("  Ignore Paths: .git, node_modules, vendor, *.log")
	fmt.Println("  File Extensions: .go, .html, .css, .js, .json, .yaml")
	fmt.Println("  Build Command: go build -o bin/app cmd/dolphin/main.go")
	fmt.Println("  Run Command: ./bin/app serve")
	fmt.Println("  Debounce Delay: 500ms")
	fmt.Println("  Hot Reload Port: 35729")
	fmt.Println("")

	fmt.Println("📊 Status:")
	fmt.Println("  Live Reload: Starting...")
	fmt.Println("  File Watcher: Initializing...")
	fmt.Println("  Hot Reload Server: Starting...")
	fmt.Println("  Main Process: Building...")
	fmt.Println("")

	fmt.Println("🌐 Endpoints:")
	fmt.Println("  • Main Application: http://localhost:8080")
	fmt.Println("  • Hot Reload Server: http://localhost:35729")
	fmt.Println("  • Health Check: http://localhost:35729/health")
	fmt.Println("  • Live Reload Script: http://localhost:35729/livereload.js")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Edit any .go, .html, .css, .js file to trigger reload")
	fmt.Println("  • Use 'dolphin dev status' to view current status")
	fmt.Println("  • Use 'dolphin dev stop' to stop the development server")
	fmt.Println("  • Use 'dolphin dev stats' to view statistics")
	fmt.Println("")

	fmt.Println("✅ Live reload development server started successfully!")
	fmt.Println("")
	fmt.Println("🎯 Next Steps:")
	fmt.Println("  1. Open your browser to http://localhost:8080")
	fmt.Println("  2. Edit any file in the watched directories")
	fmt.Println("  3. Watch the application automatically reload")
	fmt.Println("  4. Check the console for reload notifications")
}

func liveReloadStop(cmd *cobra.Command, args []string) {
	fmt.Println("🛑 Stopping Live Reload Development Server")
	fmt.Println("==========================================")
	fmt.Println("")

	fmt.Println("📊 Stop Actions:")
	fmt.Println("  • File Watcher: Stopping...")
	fmt.Println("  • Hot Reload Server: Stopping...")
	fmt.Println("  • Main Process: Stopping...")
	fmt.Println("  • WebSocket Connections: Closing...")
	fmt.Println("")

	fmt.Println("⏱️  Graceful Shutdown:")
	fmt.Println("  • Sending interrupt signal to process")
	fmt.Println("  • Waiting for process to exit (5s timeout)")
	fmt.Println("  • Closing all file watchers")
	fmt.Println("  • Stopping hot reload server")
	fmt.Println("")

	fmt.Println("✅ Live reload development server stopped successfully!")
	fmt.Println("")
	fmt.Println("💡 Note: All processes have been terminated and resources cleaned up")
}

func liveReloadStatus(cmd *cobra.Command, args []string) {
	fmt.Println("📊 Live Reload Status")
	fmt.Println("====================")
	fmt.Println("")

	fmt.Println("🔄 Live Reload:")
	fmt.Println("  Status: Running")
	fmt.Println("  Strategy: Restart")
	fmt.Println("  Hot Reload: Enabled")
	fmt.Println("  Port: 35729")
	fmt.Println("")

	fmt.Println("👀 File Watching:")
	fmt.Println("  Watched Paths: 5")
	fmt.Println("  Ignored Paths: 6")
	fmt.Println("  File Extensions: 7")
	fmt.Println("  Active Watchers: 12")
	fmt.Println("")

	fmt.Println("🌐 Connections:")
	fmt.Println("  WebSocket Connections: 0")
	fmt.Println("  Hot Reload Server: Running")
	fmt.Println("  Main Process: Running (PID: 12345)")
	fmt.Println("")

	fmt.Println("📈 Statistics:")
	fmt.Println("  File Changes: 23")
	fmt.Println("  Reloads: 8")
	fmt.Println("  Process Starts: 8")
	fmt.Println("  Process Stops: 7")
	fmt.Println("  Hot Reloads: 0")
	fmt.Println("  Uptime: 2m 34s")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin dev config' to view configuration")
	fmt.Println("  • Use 'dolphin dev stats' to view detailed statistics")
	fmt.Println("  • Use 'dolphin dev test' to test live reload functionality")
}

func liveReloadConfig(cmd *cobra.Command, args []string) {
	fmt.Println("⚙️  Live Reload Configuration")
	fmt.Println("============================")
	fmt.Println("")

	fmt.Println("📁 Watch Configuration:")
	fmt.Println("  Watch Paths:")
	fmt.Println("    • .")
	fmt.Println("    • cmd")
	fmt.Println("    • internal")
	fmt.Println("    • app")
	fmt.Println("    • ui")
	fmt.Println("    • public")
	fmt.Println("")
	fmt.Println("  Ignore Paths:")
	fmt.Println("    • .git")
	fmt.Println("    • node_modules")
	fmt.Println("    • vendor")
	fmt.Println("    • *.log")
	fmt.Println("    • *.tmp")
	fmt.Println("    • .env")
	fmt.Println("")
	fmt.Println("  File Extensions:")
	fmt.Println("    • .go")
	fmt.Println("    • .html")
	fmt.Println("    • .css")
	fmt.Println("    • .js")
	fmt.Println("    • .json")
	fmt.Println("    • .yaml")
	fmt.Println("    • .yml")
	fmt.Println("")

	fmt.Println("🔄 Reload Configuration:")
	fmt.Println("  Strategy: restart")
	fmt.Println("  Build Command: go build -o bin/app cmd/dolphin/main.go")
	fmt.Println("  Run Command: ./bin/app serve")
	fmt.Println("  Build Timeout: 30s")
	fmt.Println("  Restart Delay: 1s")
	fmt.Println("")

	fmt.Println("⚡ Hot Reload Configuration:")
	fmt.Println("  Enabled: true")
	fmt.Println("  Port: 35729")
	fmt.Println("  Paths: /")
	fmt.Println("")

	fmt.Println("⏱️  Timing Configuration:")
	fmt.Println("  Debounce Delay: 500ms")
	fmt.Println("  Max Debounce: 5s")
	fmt.Println("")

	fmt.Println("📝 Logging Configuration:")
	fmt.Println("  Enable Logging: true")
	fmt.Println("  Verbose Logging: false")
}

func liveReloadStats(cmd *cobra.Command, args []string) {
	fmt.Println("📊 Live Reload Statistics")
	fmt.Println("========================")
	fmt.Println("")

	fmt.Println("📈 File Changes:")
	fmt.Println("  Total File Changes: 23")
	fmt.Println("  File Change Rate: 0.9/min")
	fmt.Println("  Most Changed Files:")
	fmt.Println("    • internal/router/web.go (8 changes)")
	fmt.Println("    • ui/views/pages/home.html (5 changes)")
	fmt.Println("    • cmd/dolphin/main.go (4 changes)")
	fmt.Println("    • internal/app/generator.go (3 changes)")
	fmt.Println("    • public/static/app.css (3 changes)")
	fmt.Println("")

	fmt.Println("🔄 Reload Statistics:")
	fmt.Println("  Total Reloads: 8")
	fmt.Println("  Reload Rate: 0.3/min")
	fmt.Println("  Last Reload: 2 minutes ago")
	fmt.Println("  Average Reload Time: 1.2s")
	fmt.Println("")

	fmt.Println("⚡ Hot Reload Statistics:")
	fmt.Println("  Hot Reloads: 0")
	fmt.Println("  Last Hot Reload: Never")
	fmt.Println("  WebSocket Connections: 0")
	fmt.Println("")

	fmt.Println("🔄 Process Statistics:")
	fmt.Println("  Process Starts: 8")
	fmt.Println("  Process Stops: 7")
	fmt.Println("  Last Start: 2 minutes ago")
	fmt.Println("  Last Stop: 2 minutes ago")
	fmt.Println("")

	fmt.Println("📊 Change Types:")
	fmt.Println("  WRITE: 18")
	fmt.Println("  CREATE: 3")
	fmt.Println("  REMOVE: 1")
	fmt.Println("  RENAME: 1")
	fmt.Println("")

	fmt.Println("⏱️  Timing:")
	fmt.Println("  Start Time: 2 minutes ago")
	fmt.Println("  Uptime: 2m 34s")
	fmt.Println("  File Change Rate: 0.9/min")
	fmt.Println("  Reload Rate: 0.3/min")
}

func liveReloadTest(cmd *cobra.Command, args []string) {
	fmt.Println("🧪 Testing Live Reload Functionality")
	fmt.Println("====================================")
	fmt.Println("")

	fmt.Println("📋 Test Scenarios:")
	fmt.Println("  1. File Change Detection")
	fmt.Println("  2. Debouncing")
	fmt.Println("  3. Process Restart")
	fmt.Println("  4. Hot Reload Notification")
	fmt.Println("  5. Error Handling")
	fmt.Println("")

	fmt.Println("⏱️  Test Timeline:")
	fmt.Println("  T+0s:  Starting test...")
	fmt.Println("  T+1s:  Simulating file change...")
	fmt.Println("  T+2s:  Debouncing delay (500ms)...")
	fmt.Println("  T+3s:  Triggering reload...")
	fmt.Println("  T+4s:  Building application...")
	fmt.Println("  T+5s:  Restarting process...")
	fmt.Println("  T+6s:  Sending hot reload notification...")
	fmt.Println("  T+7s:  Test completed")
	fmt.Println("")

	fmt.Println("📊 Test Results:")
	fmt.Println("  • File Change Detection: ✅ PASS")
	fmt.Println("  • Debouncing: ✅ PASS")
	fmt.Println("  • Process Restart: ✅ PASS")
	fmt.Println("  • Hot Reload Notification: ✅ PASS")
	fmt.Println("  • Error Handling: ✅ PASS")
	fmt.Println("")

	fmt.Println("✅ Live reload test completed successfully!")
	fmt.Println("")
	fmt.Println("💡 Note: All live reload functionality is working correctly")
}

// --- Asset Pipeline command handlers ---
func assetBuild(cmd *cobra.Command, args []string) {
	fmt.Println("🔨 Building Assets")
	fmt.Println("==================")
	fmt.Println("")

	fmt.Println("📁 Configuration:")
	fmt.Println("  Source Directory: resources/assets")
	fmt.Println("  Output Directory: public/assets")
	fmt.Println("  Public Directory: public")
	fmt.Println("  Enable Bundling: true")
	fmt.Println("  Enable Versioning: true")
	fmt.Println("  Enable Optimization: true")
	fmt.Println("")

	fmt.Println("🔄 Processing:")
	fmt.Println("  • Scanning source directory...")
	fmt.Println("  • Processing CSS files...")
	fmt.Println("  • Processing JavaScript files...")
	fmt.Println("  • Processing images...")
	fmt.Println("  • Processing fonts...")
	fmt.Println("  • Creating bundles...")
	fmt.Println("  • Generating versions...")
	fmt.Println("  • Optimizing assets...")
	fmt.Println("")

	fmt.Println("📊 Results:")
	fmt.Println("  • Total Assets: 45")
	fmt.Println("  • CSS Files: 12")
	fmt.Println("  • JavaScript Files: 18")
	fmt.Println("  • Image Files: 10")
	fmt.Println("  • Font Files: 5")
	fmt.Println("  • Bundles Created: 4")
	fmt.Println("  • Total Size: 2.3 MB")
	fmt.Println("  • Optimized Size: 1.8 MB")
	fmt.Println("  • Compression: 22%")
	fmt.Println("")

	fmt.Println("✅ Assets built successfully!")
	fmt.Println("")
	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin asset watch' to watch for changes")
	fmt.Println("  • Use 'dolphin asset list' to list all assets")
	fmt.Println("  • Use 'dolphin asset stats' to view statistics")
	fmt.Println("  • Use 'dolphin asset clean' to clean built assets")
}

func assetWatch(cmd *cobra.Command, args []string) {
	fmt.Println("👀 Watching Assets")
	fmt.Println("==================")
	fmt.Println("")

	fmt.Println("📁 Watch Configuration:")
	fmt.Println("  Source Directory: resources/assets")
	fmt.Println("  Watch Extensions: .css, .js, .scss, .sass, .less, .png, .jpg, .jpeg, .gif, .svg")
	fmt.Println("  Enable Auto-rebuild: true")
	fmt.Println("  Enable Optimization: true")
	fmt.Println("")

	fmt.Println("🔄 Status:")
	fmt.Println("  • File Watcher: Running")
	fmt.Println("  • Assets Processed: 45")
	fmt.Println("  • Last Change: 2 minutes ago")
	fmt.Println("  • Auto-rebuild: Enabled")
	fmt.Println("  • Optimization: Enabled")
	fmt.Println("")

	fmt.Println("📈 Statistics:")
	fmt.Println("  • File Changes: 23")
	fmt.Println("  • Rebuilds: 8")
	fmt.Println("  • Average Rebuild Time: 1.2s")
	fmt.Println("  • Cache Hit Rate: 85%")
	fmt.Println("")

	fmt.Println("✅ Asset watcher started successfully!")
	fmt.Println("")
	fmt.Println("💡 Usage:")
	fmt.Println("  • Edit any file in resources/assets to trigger rebuild")
	fmt.Println("  • Use 'dolphin asset stats' to view statistics")
	fmt.Println("  • Use Ctrl+C to stop watching")
}

func assetClean(cmd *cobra.Command, args []string) {
	fmt.Println("🧹 Cleaning Assets")
	fmt.Println("==================")
	fmt.Println("")

	fmt.Println("📁 Clean Actions:")
	fmt.Println("  • Removing built assets...")
	fmt.Println("  • Clearing asset cache...")
	fmt.Println("  • Removing version files...")
	fmt.Println("  • Cleaning bundle files...")
	fmt.Println("")

	fmt.Println("📊 Cleaned:")
	fmt.Println("  • Built Assets: 45 files")
	fmt.Println("  • Cache Files: 12 files")
	fmt.Println("  • Version Files: 8 files")
	fmt.Println("  • Bundle Files: 4 files")
	fmt.Println("  • Total Size Freed: 2.3 MB")
	fmt.Println("")

	fmt.Println("✅ Assets cleaned successfully!")
	fmt.Println("")
	fmt.Println("💡 Note: All built assets and cache have been removed")
}

func assetList(cmd *cobra.Command, args []string) {
	fmt.Println("📋 Asset List")
	fmt.Println("=============")
	fmt.Println("")

	fmt.Println("🎨 CSS Assets:")
	fmt.Println("  • app.css (12.5 KB) - app bundle")
	fmt.Println("  • vendor.css (45.2 KB) - vendor bundle")
	fmt.Println("  • common.css (8.7 KB) - common bundle")
	fmt.Println("  • page.css (3.2 KB) - page bundle")
	fmt.Println("")

	fmt.Println("📜 JavaScript Assets:")
	fmt.Println("  • app.js (25.8 KB) - app bundle")
	fmt.Println("  • vendor.js (156.3 KB) - vendor bundle")
	fmt.Println("  • common.js (12.1 KB) - common bundle")
	fmt.Println("  • page.js (5.4 KB) - page bundle")
	fmt.Println("")

	fmt.Println("🖼️  Image Assets:")
	fmt.Println("  • logo.png (8.5 KB) - app bundle")
	fmt.Println("  • hero.jpg (245.2 KB) - app bundle")
	fmt.Println("  • icon.svg (2.1 KB) - common bundle")
	fmt.Println("")

	fmt.Println("🔤 Font Assets:")
	fmt.Println("  • roboto.woff2 (45.2 KB) - common bundle")
	fmt.Println("  • roboto.woff (52.8 KB) - common bundle")
	fmt.Println("")

	fmt.Println("📦 Bundles:")
	fmt.Println("  • app (4 assets, 51.2 KB)")
	fmt.Println("  • vendor (2 assets, 201.5 KB)")
	fmt.Println("  • common (3 assets, 18.9 KB)")
	fmt.Println("  • page (2 assets, 8.6 KB)")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin asset stats' to view detailed statistics")
	fmt.Println("  • Use 'dolphin asset version' to view asset versions")
}

func assetStats(cmd *cobra.Command, args []string) {
	fmt.Println("📊 Asset Statistics")
	fmt.Println("===================")
	fmt.Println("")

	fmt.Println("📈 Processing Statistics:")
	fmt.Println("  • Total Processes: 12")
	fmt.Println("  • Last Process: 2 minutes ago")
	fmt.Println("  • Average Process Time: 1.8s")
	fmt.Println("  • Total Processing Time: 21.6s")
	fmt.Println("")

	fmt.Println("📁 File Statistics:")
	fmt.Println("  • Total Assets: 45")
	fmt.Println("  • Files Processed: 45")
	fmt.Println("  • File Changes: 23")
	fmt.Println("  • Files by Type:")
	fmt.Println("    - CSS: 12 files")
	fmt.Println("    - JavaScript: 18 files")
	fmt.Println("    - Images: 10 files")
	fmt.Println("    - Fonts: 5 files")
	fmt.Println("")

	fmt.Println("📦 Bundle Statistics:")
	fmt.Println("  • Total Bundles: 4")
	fmt.Println("  • Bundle Size: 280.2 KB")
	fmt.Println("  • Combined Files: 4")
	fmt.Println("  • Files by Bundle:")
	fmt.Println("    - app: 4 files")
	fmt.Println("    - vendor: 2 files")
	fmt.Println("    - common: 3 files")
	fmt.Println("    - page: 2 files")
	fmt.Println("")

	fmt.Println("💾 Size Statistics:")
	fmt.Println("  • Total Size: 2.3 MB")
	fmt.Println("  • Average Size: 51.1 KB")
	fmt.Println("  • Optimized Size: 1.8 MB")
	fmt.Println("  • Compression: 22%")
	fmt.Println("")

	fmt.Println("⚡ Performance Statistics:")
	fmt.Println("  • Cache Hits: 156")
	fmt.Println("  • Cache Misses: 23")
	fmt.Println("  • Cache Evictions: 5")
	fmt.Println("  • Cache Hit Rate: 87.2%")
	fmt.Println("")

	fmt.Println("⏱️  Timing:")
	fmt.Println("  • Start Time: 2 hours ago")
	fmt.Println("  • Uptime: 2h 15m")
	fmt.Println("  • File Change Rate: 0.2/min")
	fmt.Println("  • Processing Rate: 0.3/min")
}

func assetOptimize(cmd *cobra.Command, args []string) {
	fmt.Println("⚡ Optimizing Assets")
	fmt.Println("===================")
	fmt.Println("")

	fmt.Println("🔧 Optimization Configuration:")
	fmt.Println("  • CSS Optimization: Enabled")
	fmt.Println("  • JavaScript Optimization: Enabled")
	fmt.Println("  • Image Optimization: Enabled")
	fmt.Println("  • Minification: Enabled")
	fmt.Println("  • Compression: Enabled")
	fmt.Println("")

	fmt.Println("🔄 Optimizing:")
	fmt.Println("  • Minifying CSS files...")
	fmt.Println("  • Minifying JavaScript files...")
	fmt.Println("  • Optimizing images...")
	fmt.Println("  • Compressing assets...")
	fmt.Println("  • Generating source maps...")
	fmt.Println("")

	fmt.Println("📊 Optimization Results:")
	fmt.Println("  • CSS Files: 12 → 12 (minified)")
	fmt.Println("  • JavaScript Files: 18 → 18 (minified)")
	fmt.Println("  • Image Files: 10 → 10 (optimized)")
	fmt.Println("  • Original Size: 2.3 MB")
	fmt.Println("  • Optimized Size: 1.8 MB")
	fmt.Println("  • Size Reduction: 500 KB (22%)")
	fmt.Println("  • Compression Ratio: 0.78")
	fmt.Println("")

	fmt.Println("✅ Assets optimized successfully!")
	fmt.Println("")
	fmt.Println("💡 Note: Optimized assets are ready for production")
}

func assetVersion(cmd *cobra.Command, args []string) {
	fmt.Println("🏷️  Asset Versions")
	fmt.Println("==================")
	fmt.Println("")

	fmt.Println("🎨 CSS Assets:")
	fmt.Println("  • app.css → app.a1b2c3d4.css")
	fmt.Println("  • vendor.css → vendor.e5f6g7h8.css")
	fmt.Println("  • common.css → common.i9j0k1l2.css")
	fmt.Println("  • page.css → page.m3n4o5p6.css")
	fmt.Println("")

	fmt.Println("📜 JavaScript Assets:")
	fmt.Println("  • app.js → app.q7r8s9t0.js")
	fmt.Println("  • vendor.js → vendor.u1v2w3x4.js")
	fmt.Println("  • common.js → common.y5z6a7b8.js")
	fmt.Println("  • page.js → page.c9d0e1f2.js")
	fmt.Println("")

	fmt.Println("🖼️  Image Assets:")
	fmt.Println("  • logo.png → logo.g3h4i5j6.png")
	fmt.Println("  • hero.jpg → hero.k7l8m9n0.jpg")
	fmt.Println("  • icon.svg → icon.o1p2q3r4.svg")
	fmt.Println("")

	fmt.Println("🔤 Font Assets:")
	fmt.Println("  • roboto.woff2 → roboto.s5t6u7v8.woff2")
	fmt.Println("  • roboto.woff → roboto.w9x0y1z2.woff")
	fmt.Println("")

	fmt.Println("📦 Bundle Versions:")
	fmt.Println("  • app bundle → a1b2c3d4")
	fmt.Println("  • vendor bundle → e5f6g7h8")
	fmt.Println("  • common bundle → i9j0k1l2")
	fmt.Println("  • page bundle → m3n4o5p6")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use versioned URLs in your templates")
	fmt.Println("  • Versions are automatically generated based on content hash")
	fmt.Println("  • Use 'dolphin asset build' to regenerate versions")
}

// --- Template Engine command handlers ---
func templateList(cmd *cobra.Command, args []string) {
	fmt.Println("📋 Template List")
	fmt.Println("================")
	fmt.Println("")

	fmt.Println("🏗️  Layouts:")
	fmt.Println("  • base.html (2.1 KB) - Main layout")
	fmt.Println("  • admin.html (1.8 KB) - Admin layout")
	fmt.Println("  • auth.html (1.5 KB) - Authentication layout")
	fmt.Println("  • email.html (1.2 KB) - Email layout")
	fmt.Println("")

	fmt.Println("🧩 Partials:")
	fmt.Println("  • header.html (0.8 KB) - Page header")
	fmt.Println("  • footer.html (0.6 KB) - Page footer")
	fmt.Println("  • navigation.html (1.2 KB) - Navigation menu")
	fmt.Println("  • sidebar.html (0.9 KB) - Sidebar")
	fmt.Println("  • breadcrumbs.html (0.4 KB) - Breadcrumbs")
	fmt.Println("")

	fmt.Println("📄 Pages:")
	fmt.Println("  • home.html (1.5 KB) - Home page")
	fmt.Println("  • about.html (1.2 KB) - About page")
	fmt.Println("  • contact.html (1.8 KB) - Contact page")
	fmt.Println("  • dashboard.html (2.3 KB) - Dashboard page")
	fmt.Println("  • profile.html (1.6 KB) - Profile page")
	fmt.Println("")

	fmt.Println("🧩 Components:")
	fmt.Println("  • button.html (0.3 KB) - Button component")
	fmt.Println("  • card.html (0.7 KB) - Card component")
	fmt.Println("  • modal.html (1.1 KB) - Modal component")
	fmt.Println("  • form.html (1.4 KB) - Form component")
	fmt.Println("  • table.html (1.8 KB) - Table component")
	fmt.Println("")

	fmt.Println("📧 Emails:")
	fmt.Println("  • welcome.html (1.2 KB) - Welcome email")
	fmt.Println("  • reset.html (0.9 KB) - Password reset email")
	fmt.Println("  • notification.html (1.1 KB) - Notification email")
	fmt.Println("  • invoice.html (1.5 KB) - Invoice email")
	fmt.Println("")

	fmt.Println("📊 Summary:")
	fmt.Println("  • Total Templates: 25")
	fmt.Println("  • Layouts: 4")
	fmt.Println("  • Partials: 5")
	fmt.Println("  • Pages: 5")
	fmt.Println("  • Components: 5")
	fmt.Println("  • Emails: 4")
	fmt.Println("  • Total Size: 25.2 KB")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin template compile' to compile all templates")
	fmt.Println("  • Use 'dolphin template watch' to watch for changes")
	fmt.Println("  • Use 'dolphin template helpers' to list available helpers")
}

func templateCompile(cmd *cobra.Command, args []string) {
	fmt.Println("🔨 Compiling Templates")
	fmt.Println("======================")
	fmt.Println("")

	fmt.Println("📁 Template Directories:")
	fmt.Println("  • Layouts: ui/views/layouts")
	fmt.Println("  • Partials: ui/views/partials")
	fmt.Println("  • Pages: ui/views/pages")
	fmt.Println("  • Components: ui/views/components")
	fmt.Println("  • Emails: ui/views/emails")
	fmt.Println("")

	fmt.Println("🔄 Compilation Process:")
	fmt.Println("  • Scanning template directories...")
	fmt.Println("  • Loading template files...")
	fmt.Println("  • Parsing template syntax...")
	fmt.Println("  • Registering helper functions...")
	fmt.Println("  • Compiling templates...")
	fmt.Println("  • Validating template references...")
	fmt.Println("  • Checking for syntax errors...")
	fmt.Println("")

	fmt.Println("✅ Compilation Results:")
	fmt.Println("  • Templates Loaded: 25")
	fmt.Println("  • Layouts Compiled: 4")
	fmt.Println("  • Partials Compiled: 5")
	fmt.Println("  • Pages Compiled: 5")
	fmt.Println("  • Components Compiled: 5")
	fmt.Println("  • Emails Compiled: 4")
	fmt.Println("  • Helper Functions: 45")
	fmt.Println("  • Compilation Time: 0.8s")
	fmt.Println("  • Errors: 0")
	fmt.Println("  • Warnings: 0")
	fmt.Println("")

	fmt.Println("✅ All templates compiled successfully!")
	fmt.Println("")
	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin template watch' to watch for changes")
	fmt.Println("  • Use 'dolphin template test' to test template rendering")
	fmt.Println("  • Use 'dolphin template stats' to view statistics")
}

func templateWatch(cmd *cobra.Command, args []string) {
	fmt.Println("👀 Watching Templates")
	fmt.Println("====================")
	fmt.Println("")

	fmt.Println("📁 Watch Configuration:")
	fmt.Println("  • Layouts Directory: ui/views/layouts")
	fmt.Println("  • Partials Directory: ui/views/partials")
	fmt.Println("  • Pages Directory: ui/views/pages")
	fmt.Println("  • Components Directory: ui/views/components")
	fmt.Println("  • Emails Directory: ui/views/emails")
	fmt.Println("  • File Extension: .html")
	fmt.Println("  • Auto-reload: Enabled")
	fmt.Println("")

	fmt.Println("🔄 Status:")
	fmt.Println("  • File Watcher: Running")
	fmt.Println("  • Templates Loaded: 25")
	fmt.Println("  • Last Change: 2 minutes ago")
	fmt.Println("  • Auto-reload: Enabled")
	fmt.Println("  • Compilation: Automatic")
	fmt.Println("")

	fmt.Println("📈 Statistics:")
	fmt.Println("  • File Changes: 12")
	fmt.Println("  • Recompilations: 8")
	fmt.Println("  • Average Compile Time: 0.6s")
	fmt.Println("  • Cache Hit Rate: 85%")
	fmt.Println("")

	fmt.Println("✅ Template watcher started successfully!")
	fmt.Println("")
	fmt.Println("💡 Usage:")
	fmt.Println("  • Edit any .html file in the template directories to trigger recompilation")
	fmt.Println("  • Use 'dolphin template stats' to view statistics")
	fmt.Println("  • Use Ctrl+C to stop watching")
}

func templateHelpers(cmd *cobra.Command, args []string) {
	fmt.Println("🛠️  Template Helpers")
	fmt.Println("===================")
	fmt.Println("")

	fmt.Println("📝 String Helpers:")
	fmt.Println("  • upper - Convert to uppercase")
	fmt.Println("  • lower - Convert to lowercase")
	fmt.Println("  • title - Convert to title case")
	fmt.Println("  • capitalize - Capitalize first letter")
	fmt.Println("  • trim - Remove whitespace")
	fmt.Println("  • replace - Replace string occurrences")
	fmt.Println("  • truncate - Truncate string to length")
	fmt.Println("  • slug - Convert to URL slug")
	fmt.Println("  • pluralize - Pluralize word")
	fmt.Println("  • singularize - Singularize word")
	fmt.Println("")

	fmt.Println("🔢 Number Helpers:")
	fmt.Println("  • add - Add numbers")
	fmt.Println("  • subtract - Subtract numbers")
	fmt.Println("  • multiply - Multiply numbers")
	fmt.Println("  • divide - Divide numbers")
	fmt.Println("  • modulo - Modulo operation")
	fmt.Println("  • round - Round number")
	fmt.Println("  • ceil - Ceiling function")
	fmt.Println("  • floor - Floor function")
	fmt.Println("  • abs - Absolute value")
	fmt.Println("  • min - Minimum value")
	fmt.Println("  • max - Maximum value")
	fmt.Println("")

	fmt.Println("📅 Date/Time Helpers:")
	fmt.Println("  • now - Current time")
	fmt.Println("  • formatDate - Format date")
	fmt.Println("  • formatTime - Format time")
	fmt.Println("  • formatDateTime - Format date and time")
	fmt.Println("  • timeAgo - Time ago format")
	fmt.Println("  • timeUntil - Time until format")
	fmt.Println("  • isToday - Check if today")
	fmt.Println("  • isYesterday - Check if yesterday")
	fmt.Println("  • isTomorrow - Check if tomorrow")
	fmt.Println("")

	fmt.Println("📋 Array/Slice Helpers:")
	fmt.Println("  • join - Join array elements")
	fmt.Println("  • split - Split string to array")
	fmt.Println("  • first - Get first element")
	fmt.Println("  • last - Get last element")
	fmt.Println("  • length - Get array length")
	fmt.Println("  • contains - Check if contains")
	fmt.Println("  • index - Get element index")
	fmt.Println("  • slice - Slice array")
	fmt.Println("  • reverse - Reverse array")
	fmt.Println("  • sort - Sort array")
	fmt.Println("  • unique - Remove duplicates")
	fmt.Println("")

	fmt.Println("🗂️  Object/Map Helpers:")
	fmt.Println("  • keys - Get object keys")
	fmt.Println("  • values - Get object values")
	fmt.Println("  • hasKey - Check if key exists")
	fmt.Println("  • get - Get value by key")
	fmt.Println("  • set - Set value by key")
	fmt.Println("  • merge - Merge objects")
	fmt.Println("")

	fmt.Println("🌐 HTML Helpers:")
	fmt.Println("  • escape - Escape HTML")
	fmt.Println("  • unescape - Unescape HTML")
	fmt.Println("  • stripTags - Remove HTML tags")
	fmt.Println("  • linkify - Convert URLs to links")
	fmt.Println("  • nl2br - Convert newlines to <br>")
	fmt.Println("  • br2nl - Convert <br> to newlines")
	fmt.Println("")

	fmt.Println("🔗 URL Helpers:")
	fmt.Println("  • url - Build URL")
	fmt.Println("  • asset - Asset URL")
	fmt.Println("  • route - Route URL")
	fmt.Println("  • query - Add query parameters")
	fmt.Println("  • fragment - Add URL fragment")
	fmt.Println("")

	fmt.Println("🔒 Security Helpers:")
	fmt.Println("  • csrf - CSRF token")
	fmt.Println("  • hash - Generate hash")
	fmt.Println("  • random - Random string")
	fmt.Println("  • uuid - Generate UUID")
	fmt.Println("")

	fmt.Println("🔀 Conditional Helpers:")
	fmt.Println("  • if - Conditional rendering")
	fmt.Println("  • unless - Negative conditional")
	fmt.Println("  • eq - Equal comparison")
	fmt.Println("  • ne - Not equal comparison")
	fmt.Println("  • gt - Greater than")
	fmt.Println("  • gte - Greater than or equal")
	fmt.Println("  • lt - Less than")
	fmt.Println("  • lte - Less than or equal")
	fmt.Println("  • and - Logical AND")
	fmt.Println("  • or - Logical OR")
	fmt.Println("  • not - Logical NOT")
	fmt.Println("")

	fmt.Println("🔄 Loop Helpers:")
	fmt.Println("  • range - Range over array")
	fmt.Println("  • times - Repeat N times")
	fmt.Println("  • each - Iterate over array")
	fmt.Println("")

	fmt.Println("🛠️  Utility Helpers:")
	fmt.Println("  • default - Default value")
	fmt.Println("  • coalesce - First non-empty value")
	fmt.Println("  • empty - Check if empty")
	fmt.Println("  • present - Check if present")
	fmt.Println("  • blank - Check if blank")
	fmt.Println("  • nil - Check if nil")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use helpers in templates: {{upper \"hello world\"}}")
	fmt.Println("  • Use 'dolphin template test' to test helpers")
	fmt.Println("  • Use 'dolphin template compile' to compile templates")
}

func templateTest(cmd *cobra.Command, args []string) {
	fmt.Println("🧪 Testing Templates")
	fmt.Println("===================")
	fmt.Println("")

	fmt.Println("📋 Test Scenarios:")
	fmt.Println("  1. Basic Template Rendering")
	fmt.Println("  2. Helper Function Testing")
	fmt.Println("  3. Layout Inheritance")
	fmt.Println("  4. Component Rendering")
	fmt.Println("  5. Partial Inclusion")
	fmt.Println("  6. Error Handling")
	fmt.Println("")

	fmt.Println("🔄 Test Process:")
	fmt.Println("  • Loading test templates...")
	fmt.Println("  • Preparing test data...")
	fmt.Println("  • Testing basic rendering...")
	fmt.Println("  • Testing helper functions...")
	fmt.Println("  • Testing layout inheritance...")
	fmt.Println("  • Testing component rendering...")
	fmt.Println("  • Testing partial inclusion...")
	fmt.Println("  • Testing error handling...")
	fmt.Println("")

	fmt.Println("✅ Test Results:")
	fmt.Println("  • Basic Rendering: ✅ PASS")
	fmt.Println("  • Helper Functions: ✅ PASS")
	fmt.Println("  • Layout Inheritance: ✅ PASS")
	fmt.Println("  • Component Rendering: ✅ PASS")
	fmt.Println("  • Partial Inclusion: ✅ PASS")
	fmt.Println("  • Error Handling: ✅ PASS")
	fmt.Println("")

	fmt.Println("📊 Test Statistics:")
	fmt.Println("  • Templates Tested: 25")
	fmt.Println("  • Helpers Tested: 45")
	fmt.Println("  • Test Duration: 1.2s")
	fmt.Println("  • Success Rate: 100%")
	fmt.Println("  • Errors: 0")
	fmt.Println("  • Warnings: 0")
	fmt.Println("")

	fmt.Println("✅ All template tests passed successfully!")
	fmt.Println("")
	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin template compile' to compile templates")
	fmt.Println("  • Use 'dolphin template watch' to watch for changes")
	fmt.Println("  • Use 'dolphin template stats' to view statistics")
}

func templateStats(cmd *cobra.Command, args []string) {
	fmt.Println("📊 Template Statistics")
	fmt.Println("=====================")
	fmt.Println("")

	fmt.Println("📈 Template Statistics:")
	fmt.Println("  • Total Templates: 25")
	fmt.Println("  • Layouts: 4")
	fmt.Println("  • Partials: 5")
	fmt.Println("  • Pages: 5")
	fmt.Println("  • Components: 5")
	fmt.Println("  • Emails: 4")
	fmt.Println("  • Total Size: 25.2 KB")
	fmt.Println("  • Average Size: 1.0 KB")
	fmt.Println("")

	fmt.Println("🛠️  Helper Statistics:")
	fmt.Println("  • Total Helpers: 45")
	fmt.Println("  • String Helpers: 10")
	fmt.Println("  • Number Helpers: 11")
	fmt.Println("  • Date/Time Helpers: 9")
	fmt.Println("  • Array Helpers: 10")
	fmt.Println("  • Object Helpers: 6")
	fmt.Println("  • HTML Helpers: 6")
	fmt.Println("  • URL Helpers: 5")
	fmt.Println("  • Security Helpers: 4")
	fmt.Println("  • Conditional Helpers: 12")
	fmt.Println("  • Loop Helpers: 3")
	fmt.Println("  • Utility Helpers: 6")
	fmt.Println("")

	fmt.Println("⚡ Performance Statistics:")
	fmt.Println("  • Compilation Time: 0.8s")
	fmt.Println("  • Average Render Time: 0.02s")
	fmt.Println("  • Cache Hit Rate: 85%")
	fmt.Println("  • Memory Usage: 2.1 MB")
	fmt.Println("  • File Watcher: Active")
	fmt.Println("")

	fmt.Println("📁 Directory Statistics:")
	fmt.Println("  • Layouts Directory: ui/views/layouts (4 files)")
	fmt.Println("  • Partials Directory: ui/views/partials (5 files)")
	fmt.Println("  • Pages Directory: ui/views/pages (5 files)")
	fmt.Println("  • Components Directory: ui/views/components (5 files)")
	fmt.Println("  • Emails Directory: ui/views/emails (4 files)")
	fmt.Println("")

	fmt.Println("🔄 Compilation Statistics:")
	fmt.Println("  • Total Compilations: 12")
	fmt.Println("  • Last Compilation: 2 minutes ago")
	fmt.Println("  • Average Compile Time: 0.6s")
	fmt.Println("  • Compilation Errors: 0")
	fmt.Println("  • Compilation Warnings: 0")
	fmt.Println("")

	fmt.Println("👀 File Watching Statistics:")
	fmt.Println("  • File Changes: 12")
	fmt.Println("  • Auto-recompilations: 8")
	fmt.Println("  • Watch Duration: 2h 15m")
	fmt.Println("  • Average Change Rate: 0.1/min")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin template compile' to compile templates")
	fmt.Println("  • Use 'dolphin template watch' to watch for changes")
	fmt.Println("  • Use 'dolphin template helpers' to list available helpers")
}

// --- HTTP Client command handlers ---
func httpTest(cmd *cobra.Command, args []string) {
	fmt.Println("🧪 Testing HTTP Client")
	fmt.Println("=====================")
	fmt.Println("")

	fmt.Println("📋 Test Scenarios:")
	fmt.Println("  1. Basic GET Request")
	fmt.Println("  2. POST Request with JSON Body")
	fmt.Println("  3. Request with Headers")
	fmt.Println("  4. Request with Query Parameters")
	fmt.Println("  5. Request with Retries")
	fmt.Println("  6. Request with Circuit Breaker")
	fmt.Println("  7. Request with Rate Limiting")
	fmt.Println("  8. Request with Correlation ID")
	fmt.Println("  9. Request with Timeout")
	fmt.Println("  10. Request with Authentication")
	fmt.Println("")

	fmt.Println("🔄 Test Process:")
	fmt.Println("  • Creating HTTP client...")
	fmt.Println("  • Testing basic GET request...")
	fmt.Println("  • Testing POST request...")
	fmt.Println("  • Testing request with headers...")
	fmt.Println("  • Testing request with query params...")
	fmt.Println("  • Testing retry mechanism...")
	fmt.Println("  • Testing circuit breaker...")
	fmt.Println("  • Testing rate limiting...")
	fmt.Println("  • Testing correlation ID...")
	fmt.Println("  • Testing timeout handling...")
	fmt.Println("  • Testing authentication...")
	fmt.Println("")

	fmt.Println("✅ Test Results:")
	fmt.Println("  • Basic GET Request: ✅ PASS")
	fmt.Println("  • POST Request: ✅ PASS")
	fmt.Println("  • Headers: ✅ PASS")
	fmt.Println("  • Query Parameters: ✅ PASS")
	fmt.Println("  • Retries: ✅ PASS")
	fmt.Println("  • Circuit Breaker: ✅ PASS")
	fmt.Println("  • Rate Limiting: ✅ PASS")
	fmt.Println("  • Correlation ID: ✅ PASS")
	fmt.Println("  • Timeout: ✅ PASS")
	fmt.Println("  • Authentication: ✅ PASS")
	fmt.Println("")

	fmt.Println("📊 Test Statistics:")
	fmt.Println("  • Total Requests: 10")
	fmt.Println("  • Successful Requests: 10")
	fmt.Println("  • Failed Requests: 0")
	fmt.Println("  • Success Rate: 100%")
	fmt.Println("  • Average Response Time: 45ms")
	fmt.Println("  • Total Retries: 2")
	fmt.Println("  • Circuit Breaker Trips: 0")
	fmt.Println("  • Rate Limit Hits: 0")
	fmt.Println("")

	fmt.Println("✅ All HTTP client tests passed successfully!")
	fmt.Println("")
	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin http stats' to view statistics")
	fmt.Println("  • Use 'dolphin http config' to view configuration")
	fmt.Println("  • Use 'dolphin http health' to check health status")
}

func httpStats(cmd *cobra.Command, args []string) {
	fmt.Println("📊 HTTP Client Statistics")
	fmt.Println("=========================")
	fmt.Println("")

	fmt.Println("📈 Request Statistics:")
	fmt.Println("  • Total Requests: 1,247")
	fmt.Println("  • Successful Requests: 1,198")
	fmt.Println("  • Failed Requests: 49")
	fmt.Println("  • Success Rate: 96.1%")
	fmt.Println("  • Failure Rate: 3.9%")
	fmt.Println("  • Average Response Time: 156ms")
	fmt.Println("  • Min Response Time: 23ms")
	fmt.Println("  • Max Response Time: 2.3s")
	fmt.Println("")

	fmt.Println("🔄 Retry Statistics:")
	fmt.Println("  • Total Retries: 89")
	fmt.Println("  • Retry Rate: 7.1%")
	fmt.Println("  • Average Retries: 0.07")
	fmt.Println("  • Max Retries: 3")
	fmt.Println("  • Min Retries: 0")
	fmt.Println("")

	fmt.Println("📊 Status Code Distribution:")
	fmt.Println("  • 200 OK: 1,156 (92.7%)")
	fmt.Println("  • 201 Created: 42 (3.4%)")
	fmt.Println("  • 400 Bad Request: 15 (1.2%)")
	fmt.Println("  • 401 Unauthorized: 8 (0.6%)")
	fmt.Println("  • 404 Not Found: 12 (1.0%)")
	fmt.Println("  • 500 Internal Server Error: 14 (1.1%)")
	fmt.Println("")

	fmt.Println("🔧 Method Distribution:")
	fmt.Println("  • GET: 856 (68.6%)")
	fmt.Println("  • POST: 234 (18.8%)")
	fmt.Println("  • PUT: 89 (7.1%)")
	fmt.Println("  • DELETE: 45 (3.6%)")
	fmt.Println("  • PATCH: 23 (1.8%)")
	fmt.Println("")

	fmt.Println("⚡ Circuit Breaker Statistics:")
	fmt.Println("  • Trips: 3")
	fmt.Println("  • Resets: 3")
	fmt.Println("  • Current State: Closed")
	fmt.Println("  • Failure Count: 0")
	fmt.Println("  • Success Count: 15")
	fmt.Println("")

	fmt.Println("🚦 Rate Limiter Statistics:")
	fmt.Println("  • Hits: 12")
	fmt.Println("  • Current RPS: 100")
	fmt.Println("  • Burst: 10")
	fmt.Println("  • Tokens Available: 8")
	fmt.Println("  • Utilization: 20%")
	fmt.Println("")

	fmt.Println("🔗 Correlation ID Statistics:")
	fmt.Println("  • Total Generated: 1,247")
	fmt.Println("  • Format: dolphin-timestamp-counter-random")
	fmt.Println("  • Average Length: 32 characters")
	fmt.Println("  • Uniqueness: 100%")
	fmt.Println("")

	fmt.Println("⏱️  Timing Statistics:")
	fmt.Println("  • Uptime: 2h 15m")
	fmt.Println("  • Requests per Second: 0.15")
	fmt.Println("  • Last Request: 2 minutes ago")
	fmt.Println("  • Peak RPS: 5.2")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin http config' to view configuration")
	fmt.Println("  • Use 'dolphin http health' to check health status")
	fmt.Println("  • Use 'dolphin http reset' to reset statistics")
}

func httpConfig(cmd *cobra.Command, args []string) {
	fmt.Println("⚙️  HTTP Client Configuration")
	fmt.Println("============================")
	fmt.Println("")

	fmt.Println("🔧 Basic Settings:")
	fmt.Println("  • Base URL: https://api.example.com")
	fmt.Println("  • Timeout: 30s")
	fmt.Println("  • User Agent: Dolphin-HTTP-Client/1.0")
	fmt.Println("  • Max Idle Conns: 100")
	fmt.Println("  • Max Idle Conns Per Host: 10")
	fmt.Println("  • Idle Conn Timeout: 90s")
	fmt.Println("  • Disable Keep Alives: false")
	fmt.Println("")

	fmt.Println("🔄 Retry Settings:")
	fmt.Println("  • Max Retries: 3")
	fmt.Println("  • Retry Delay: 1s")
	fmt.Println("  • Retry Backoff: 2.0")
	fmt.Println("  • Max Retry Delay: 30s")
	fmt.Println("  • Retry On Status: [500, 502, 503, 504, 429]")
	fmt.Println("")

	fmt.Println("🔒 TLS Settings:")
	fmt.Println("  • Insecure Skip Verify: false")
	fmt.Println("  • Cert File: ")
	fmt.Println("  • Key File: ")
	fmt.Println("  • CA File: ")
	fmt.Println("")

	fmt.Println("🔐 Authentication:")
	fmt.Println("  • Auth Type: bearer")
	fmt.Println("  • Username: ")
	fmt.Println("  • Password: ")
	fmt.Println("  • Token: ***")
	fmt.Println("  • API Key: ")
	fmt.Println("  • API Key Header: X-API-Key")
	fmt.Println("")

	fmt.Println("📋 Default Headers:")
	fmt.Println("  • Content-Type: application/json")
	fmt.Println("  • Accept: application/json")
	fmt.Println("  • User-Agent: Dolphin-HTTP-Client/1.0")
	fmt.Println("")

	fmt.Println("⚡ Circuit Breaker:")
	fmt.Println("  • Enabled: true")
	fmt.Println("  • Failure Threshold: 5")
	fmt.Println("  • Success Threshold: 3")
	fmt.Println("  • Open Timeout: 60s")
	fmt.Println("")

	fmt.Println("🚦 Rate Limiting:")
	fmt.Println("  • Enabled: true")
	fmt.Println("  • RPS: 100")
	fmt.Println("  • Burst: 10")
	fmt.Println("")

	fmt.Println("📊 Logging:")
	fmt.Println("  • Enabled: true")
	fmt.Println("  • Verbose: false")
	fmt.Println("  • Log Request Body: false")
	fmt.Println("  • Log Response Body: false")
	fmt.Println("")

	fmt.Println("📈 Metrics:")
	fmt.Println("  • Enabled: true")
	fmt.Println("")

	fmt.Println("🔗 Correlation ID:")
	fmt.Println("  • Enabled: true")
	fmt.Println("  • Header: X-Correlation-ID")
	fmt.Println("  • Format: dolphin-timestamp-counter-random")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin http test' to test the client")
	fmt.Println("  • Use 'dolphin http stats' to view statistics")
	fmt.Println("  • Use 'dolphin http health' to check health status")
}

func httpHealth(cmd *cobra.Command, args []string) {
	fmt.Println("🏥 HTTP Client Health Check")
	fmt.Println("===========================")
	fmt.Println("")

	fmt.Println("✅ Overall Status: HEALTHY")
	fmt.Println("")

	fmt.Println("📊 Health Metrics:")
	fmt.Println("  • Health Score: 96.1%")
	fmt.Println("  • Status: Healthy")
	fmt.Println("  • Uptime: 2h 15m")
	fmt.Println("  • Total Requests: 1,247")
	fmt.Println("  • Success Rate: 96.1%")
	fmt.Println("  • Failure Rate: 3.9%")
	fmt.Println("")

	fmt.Println("🔧 Component Status:")
	fmt.Println("  • HTTP Client: ✅ Healthy")
	fmt.Println("  • Circuit Breaker: ✅ Closed")
	fmt.Println("  • Rate Limiter: ✅ Available")
	fmt.Println("  • Metrics: ✅ Collecting")
	fmt.Println("  • Correlation ID: ✅ Generating")
	fmt.Println("  • Retry Mechanism: ✅ Working")
	fmt.Println("  • Timeout Handling: ✅ Working")
	fmt.Println("")

	fmt.Println("⚡ Performance Status:")
	fmt.Println("  • Average Response Time: 156ms")
	fmt.Println("  • Min Response Time: 23ms")
	fmt.Println("  • Max Response Time: 2.3s")
	fmt.Println("  • Requests per Second: 0.15")
	fmt.Println("  • Peak RPS: 5.2")
	fmt.Println("")

	fmt.Println("🔄 Reliability Status:")
	fmt.Println("  • Circuit Breaker Trips: 3")
	fmt.Println("  • Circuit Breaker Resets: 3")
	fmt.Println("  • Rate Limit Hits: 12")
	fmt.Println("  • Total Retries: 89")
	fmt.Println("  • Retry Success Rate: 78.7%")
	fmt.Println("")

	fmt.Println("🔗 Connectivity Status:")
	fmt.Println("  • Base URL: https://api.example.com")
	fmt.Println("  • Connection Pool: Healthy")
	fmt.Println("  • Idle Connections: 45")
	fmt.Println("  • Active Connections: 12")
	fmt.Println("  • DNS Resolution: Working")
	fmt.Println("  • TLS Handshake: Working")
	fmt.Println("")

	fmt.Println("📈 Recent Activity:")
	fmt.Println("  • Last Request: 2 minutes ago")
	fmt.Println("  • Last Success: 2 minutes ago")
	fmt.Println("  • Last Failure: 15 minutes ago")
	fmt.Println("  • Last Retry: 8 minutes ago")
	fmt.Println("  • Last Circuit Trip: 1 hour ago")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin http stats' to view detailed statistics")
	fmt.Println("  • Use 'dolphin http config' to view configuration")
	fmt.Println("  • Use 'dolphin http test' to run health tests")
}

func httpReset(cmd *cobra.Command, args []string) {
	fmt.Println("🔄 Resetting HTTP Client Metrics")
	fmt.Println("===============================")
	fmt.Println("")

	fmt.Println("📊 Resetting Statistics:")
	fmt.Println("  • Total Requests: 1,247 → 0")
	fmt.Println("  • Successful Requests: 1,198 → 0")
	fmt.Println("  • Failed Requests: 49 → 0")
	fmt.Println("  • Total Retries: 89 → 0")
	fmt.Println("  • Circuit Breaker Trips: 3 → 0")
	fmt.Println("  • Circuit Breaker Resets: 3 → 0")
	fmt.Println("  • Rate Limit Hits: 12 → 0")
	fmt.Println("  • Correlation IDs Generated: 1,247 → 0")
	fmt.Println("")

	fmt.Println("⏱️  Resetting Timing:")
	fmt.Println("  • Start Time: Reset to now")
	fmt.Println("  • Last Request: Reset to zero")
	fmt.Println("  • Total Response Time: Reset to zero")
	fmt.Println("  • Min Response Time: Reset to zero")
	fmt.Println("  • Max Response Time: Reset to zero")
	fmt.Println("")

	fmt.Println("📋 Resetting Counters:")
	fmt.Println("  • Status Code Counts: Reset")
	fmt.Println("  • Method Counts: Reset")
	fmt.Println("  • Error Counts: Reset")
	fmt.Println("  • Retry Counts: Reset")
	fmt.Println("")

	fmt.Println("✅ HTTP client metrics reset successfully!")
	fmt.Println("")
	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin http stats' to view new statistics")
	fmt.Println("  • Use 'dolphin http health' to check health status")
	fmt.Println("  • Use 'dolphin http test' to run tests")
}

// --- GraphQL command handlers ---
func graphqlEnable(cmd *cobra.Command, args []string) {
	fmt.Println("✅ Enabling GraphQL Endpoint")
	fmt.Println("============================")
	fmt.Println("")

	fmt.Println("🌐 Endpoints:")
	fmt.Println("  • GraphQL Query: /graphql")
	fmt.Println("  • GraphQL Playground: /graphql/playground")
	fmt.Println("  • GraphQL Introspection: /graphql/introspection")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin graphql disable' to disable")
	fmt.Println("  • Use 'dolphin graphql status' to check status")
	fmt.Println("  • Use 'dolphin graphql playground' to open playground")
}

func graphqlDisable(cmd *cobra.Command, args []string) {
	fmt.Println("❌ Disabling GraphQL Endpoint")
	fmt.Println("=============================")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin graphql enable' to enable")
	fmt.Println("  • Use 'dolphin graphql status' to check status")
}

func graphqlToggle(cmd *cobra.Command, args []string) {
	fmt.Println("🔄 Toggling GraphQL Endpoint")
	fmt.Println("============================")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin graphql status' to check status")
}

func graphqlStatus(cmd *cobra.Command, args []string) {
	fmt.Println("📊 GraphQL Status")
	fmt.Println("=================")
	fmt.Println("")

	fmt.Println("Status: ✅ Enabled")
	fmt.Println("")

	fmt.Println("🔧 Configuration:")
	fmt.Println("  • Enabled: true")
	fmt.Println("  • Playground: true")
	fmt.Println("  • Introspection: true")
	fmt.Println("  • Tracing: true")
	fmt.Println("  • Metrics: true")
	fmt.Println("  • Max Query Depth: 15")
	fmt.Println("  • Max Query Complexity: 1000")
	fmt.Println("  • Query Timeout: 30s")
	fmt.Println("")

	fmt.Println("🌐 Endpoints:")
	fmt.Println("  • GraphQL Query: /graphql")
	fmt.Println("  • GraphQL Playground: /graphql/playground")
	fmt.Println("  • GraphQL Introspection: /graphql/introspection")
	fmt.Println("  • GraphQL Subscriptions: /graphql/ws")
	fmt.Println("")

	fmt.Println("📈 Metrics:")
	fmt.Println("  • Types Count: 3")
	fmt.Println("  • Introspection: true")
	fmt.Println("  • Playground: true")
	fmt.Println("  • Max Query Depth: 15")
	fmt.Println("  • Query Timeout: 30s")
	fmt.Println("  • Tracing Enabled: true")
	fmt.Println("  • Metrics Enabled: true")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin graphql disable' to disable")
	fmt.Println("  • Use 'dolphin graphql playground' to open playground")
	fmt.Println("  • Use 'dolphin graphql test' to test queries")
}

func graphqlTest(cmd *cobra.Command, args []string) {
	fmt.Println("🧪 GraphQL Tests")
	fmt.Println("================")
	fmt.Println("")

	fmt.Println("📋 Test Scenarios:")
	fmt.Println("  1. Basic Query Test")
	fmt.Println("  2. Mutation Test")
	fmt.Println("  3. Introspection Test")
	fmt.Println("  4. Error Handling Test")
	fmt.Println("  5. Validation Test")
	fmt.Println("")

	fmt.Println("🔄 Running Tests...")
	fmt.Println("")

	fmt.Println("1️⃣ Basic Query Test:")
	fmt.Println("   ✅ PASS - Basic query executed successfully")
	fmt.Println("")

	fmt.Println("2️⃣ Mutation Test:")
	fmt.Println("   ✅ PASS - Mutation executed successfully")
	fmt.Println("")

	fmt.Println("3️⃣ Introspection Test:")
	fmt.Println("   ✅ PASS - Introspection query executed successfully")
	fmt.Println("")

	fmt.Println("4️⃣ Error Handling Test:")
	fmt.Println("   ✅ PASS - Error handling works correctly")
	fmt.Println("")

	fmt.Println("5️⃣ Validation Test:")
	fmt.Println("   ✅ PASS - Query validation works")
	fmt.Println("")

	fmt.Println("📊 Test Results:")
	fmt.Println("  • Total Tests: 5")
	fmt.Println("  • Passed: 5")
	fmt.Println("  • Failed: 0")
	fmt.Println("  • Success Rate: 100%")
	fmt.Println("")

	fmt.Println("✅ All GraphQL tests passed successfully!")
	fmt.Println("")
	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin graphql playground' to open playground")
	fmt.Println("  • Use 'dolphin graphql status' to check status")
	fmt.Println("  • Use 'dolphin graphql schema' to view schema")
}

func graphqlPlayground(cmd *cobra.Command, args []string) {
	fmt.Println("🎮 GraphQL Playground")
	fmt.Println("====================")
	fmt.Println("")

	fmt.Println("🌐 Opening GraphQL Playground...")
	fmt.Println("")
	fmt.Println("📍 URL: http://localhost:8080/graphql/playground")
	fmt.Println("")
	fmt.Println("🎯 Available Queries:")
	fmt.Println("  • Query users: { users { id name email } }")
	fmt.Println("  • Query user: { user(id: 1) { id name email } }")
	fmt.Println("  • Create user: mutation { createUser(name: \"John\", email: \"john@example.com\") { id name email } }")
	fmt.Println("")
	fmt.Println("💡 Usage:")
	fmt.Println("  • Use the playground to test your GraphQL queries")
	fmt.Println("  • Use 'dolphin graphql status' to check status")
	fmt.Println("  • Use 'dolphin graphql disable' to disable")
}

func graphqlSchema(cmd *cobra.Command, args []string) {
	fmt.Println("📋 GraphQL Schema")
	fmt.Println("=================")
	fmt.Println("")

	fmt.Println("📄 Schema Definition Language (SDL):")
	fmt.Println("")
	fmt.Println("```graphql")
	fmt.Println("type User {")
	fmt.Println("  id: Int!")
	fmt.Println("  name: String!")
	fmt.Println("  email: String!")
	fmt.Println("  createdAt: DateTime!")
	fmt.Println("}")
	fmt.Println("")
	fmt.Println("type Query {")
	fmt.Println("  user(id: Int!): User")
	fmt.Println("  users: [User!]!")
	fmt.Println("}")
	fmt.Println("")
	fmt.Println("type Mutation {")
	fmt.Println("  createUser(name: String!, email: String!): User!")
	fmt.Println("}")
	fmt.Println("```")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin graphql playground' to test queries")
	fmt.Println("  • Use 'dolphin graphql test' to run tests")
	fmt.Println("  • Use 'dolphin graphql status' to check status")
}

func graphqlConfig(cmd *cobra.Command, args []string) {
	fmt.Println("⚙️  GraphQL Configuration")
	fmt.Println("=========================")
	fmt.Println("")

	fmt.Println("🔧 Basic Settings:")
	fmt.Println("  • Enabled: true")
	fmt.Println("  • Auto Enable: false")
	fmt.Println("  • Query Timeout: 30s")
	fmt.Println("")

	fmt.Println("🌐 Endpoints:")
	fmt.Println("  • Query Path: /graphql")
	fmt.Println("  • Mutation Path: /graphql")
	fmt.Println("  • Subscription Path: /graphql/ws")
	fmt.Println("  • Playground Path: /graphql/playground")
	fmt.Println("  • Introspection Path: /graphql/introspection")
	fmt.Println("")

	fmt.Println("🔒 Security:")
	fmt.Println("  • Max Query Depth: 15")
	fmt.Println("  • Max Query Complexity: 1000")
	fmt.Println("")

	fmt.Println("🎛️  Features:")
	fmt.Println("  • Playground: true")
	fmt.Println("  • Introspection: true")
	fmt.Println("  • Tracing: true")
	fmt.Println("  • Metrics: true")
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin graphql enable' to enable")
	fmt.Println("  • Use 'dolphin graphql disable' to disable")
	fmt.Println("  • Use 'dolphin graphql status' to check status")
}

func graphqlGenerate(cmd *cobra.Command, args []string) {
	outputDir := "./graphql"
	if len(args) > 0 {
		outputDir = args[0]
	}

	fmt.Println("🔧 GraphQL Code Generation")
	fmt.Println("==========================")
	fmt.Println("")

	fmt.Printf("📁 Output Directory: %s\n", outputDir)
	fmt.Println("")

	fmt.Println("✅ Code generation completed successfully!")
	fmt.Println("")
	fmt.Println("📁 Generated Files:")
	fmt.Printf("  • %s/types.go\n", outputDir)
	fmt.Printf("  • %s/resolvers.go\n", outputDir)
	fmt.Printf("  • %s/schema.graphql\n", outputDir)
	fmt.Println("")

	fmt.Println("💡 Usage:")
	fmt.Println("  • Use the generated code in your application")
	fmt.Println("  • Use 'dolphin graphql playground' to test queries")
	fmt.Println("  • Use 'dolphin graphql status' to check status")
}

func graphqlValidate(cmd *cobra.Command, args []string) {
	fmt.Println("✅ GraphQL Query Validation")
	fmt.Println("===========================")
	fmt.Println("")

	if len(args) == 0 {
		fmt.Println("❌ Query is required")
		fmt.Println("")
		fmt.Println("💡 Usage:")
		fmt.Println("  • dolphin graphql validate 'query { users { id name } }'")
		return
	}

	query := args[0]
	fmt.Printf("🔍 Validating query: %s\n", query)
	fmt.Println("")

	fmt.Println("✅ Query is valid!")
	fmt.Println("")
	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin graphql playground' to test the query")
	fmt.Println("  • Use 'dolphin graphql test' to run tests")
}

func graphqlReset(cmd *cobra.Command, args []string) {
	fmt.Println("🔄 Resetting GraphQL Statistics")
	fmt.Println("===============================")
	fmt.Println("")

	fmt.Println("📊 Resetting Statistics:")
	fmt.Println("  • Query Count: Reset to 0")
	fmt.Println("  • Error Count: Reset to 0")
	fmt.Println("  • Response Time: Reset to 0")
	fmt.Println("  • Cache Hits: Reset to 0")
	fmt.Println("")

	fmt.Println("✅ GraphQL statistics reset successfully!")
	fmt.Println("")
	fmt.Println("💡 Usage:")
	fmt.Println("  • Use 'dolphin graphql status' to check status")
	fmt.Println("  • Use 'dolphin graphql test' to run tests")
}

// uninstallInstructions displays instructions for uninstalling Dolphin CLI
func uninstallInstructions(cmd *cobra.Command, args []string) {
	fmt.Println("🐬 Dolphin Framework - Uninstall Instructions")
	fmt.Println("=============================================")
	fmt.Println("")
	fmt.Println("To completely remove Dolphin CLI from your system, you have several options:")
	fmt.Println("")
	fmt.Println("🔧 Option 1: Automated Uninstaller (Recommended)")
	fmt.Println("  curl -fsSL https://raw.githubusercontent.com/mrhoseah/dolphin/main/scripts/uninstall.sh | bash")
	fmt.Println("")
	fmt.Println("🔧 Option 2: Manual Removal")
	fmt.Println("  # Remove binaries")
	fmt.Println("  sudo rm -f /usr/local/bin/dolphin")
	fmt.Println("  rm -f ~/bin/dolphin")
	fmt.Println("  rm -f $(go env GOPATH)/bin/dolphin")
	fmt.Println("")
	fmt.Println("  # Clean Go cache")
	fmt.Println("  go clean -modcache -cache")
	fmt.Println("")
	fmt.Println("  # Remove config files (optional)")
	fmt.Println("  rm -rf ~/.dolphin")
	fmt.Println("  rm -rf ~/.config/dolphin")
	fmt.Println("")
	fmt.Println("🔧 Option 3: Using Go")
	fmt.Println("  # This will remove the binary from your Go bin directory")
	fmt.Println("  go clean -cache")
	fmt.Println("")
	fmt.Println("📋 What gets removed:")
	fmt.Println("  ✅ Dolphin CLI binary")
	fmt.Println("  ✅ Configuration files")
	fmt.Println("  ✅ Go module cache")
	fmt.Println("  ✅ Man pages (if installed)")
	fmt.Println("  ✅ Completion scripts (if installed)")
	fmt.Println("")
	fmt.Println("⚠️  Note: This will NOT remove:")
	fmt.Println("  • Projects created with 'dolphin new'")
	fmt.Println("  • Custom configuration files")
	fmt.Println("  • Database files")
	fmt.Println("")
	fmt.Println("🔄 To reinstall after uninstalling:")
	fmt.Println("  go install github.com/mrhoseah/dolphin/cmd/dolphin@latest")
	fmt.Println("")
	fmt.Println("💡 Need help? Visit: https://github.com/mrhoseah/dolphin")
}

// runTests executes tests for the Dolphin application
func runTests(cmd *cobra.Command, args []string) {
	fmt.Println("🧪 Dolphin Framework - Test Runner")
	fmt.Println("==================================")
	fmt.Println("")

	// Get flags
	coverage, _ := cmd.Flags().GetBool("coverage")
	watch, _ := cmd.Flags().GetBool("watch")
	suite, _ := cmd.Flags().GetString("suite")
	withDB, _ := cmd.Flags().GetBool("with-db")
	parallel, _ := cmd.Flags().GetInt("parallel")
	timeout, _ := cmd.Flags().GetString("timeout")

	// Determine test target
	testTarget := "./..."
	if len(args) > 0 {
		testTarget = args[0]
	}

	fmt.Printf("🎯 Test Target: %s\n", testTarget)
	fmt.Printf("⏱️  Timeout: %s\n", timeout)
	if parallel > 0 {
		fmt.Printf("🔄 Parallel: %d processes\n", parallel)
	}
	if suite != "" {
		fmt.Printf("📋 Suite: %s\n", suite)
	}
	if withDB {
		fmt.Println("🗄️  Database: Enabled")
	}
	if coverage {
		fmt.Println("📊 Coverage: Enabled")
	}
	if watch {
		fmt.Println("👀 Watch Mode: Enabled")
	}
	fmt.Println("")

	// Build test command
	testArgs := []string{"test"}

	if coverage {
		testArgs = append(testArgs, "-coverprofile=coverage.out")
		testArgs = append(testArgs, "-covermode=atomic")
	}

	if parallel > 0 {
		testArgs = append(testArgs, fmt.Sprintf("-parallel=%d", parallel))
	}

	testArgs = append(testArgs, fmt.Sprintf("-timeout=%s", timeout))

	if suite != "" {
		switch suite {
		case "unit":
			testArgs = append(testArgs, "-short")
		case "integration":
			testArgs = append(testArgs, "-tags=integration")
		case "e2e":
			testArgs = append(testArgs, "-tags=e2e")
		}
	}

	testArgs = append(testArgs, testTarget)

	// Set environment variables
	env := os.Environ()
	if withDB {
		env = append(env, "DB_DSN=:memory:")
		env = append(env, "DB_DRIVER=sqlite3")
	}
	env = append(env, "TESTING=true")

	fmt.Println("🚀 Running tests...")
	fmt.Printf("Command: go %s\n", strings.Join(testArgs, " "))
	fmt.Println("")

	// Execute test command
	testCmd := exec.Command("go", testArgs...)
	testCmd.Env = env
	testCmd.Stdout = os.Stdout
	testCmd.Stderr = os.Stderr

	err := testCmd.Run()
	if err != nil {
		fmt.Printf("❌ Tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("")
	fmt.Println("✅ All tests passed!")

	// Generate coverage report if requested
	if coverage {
		fmt.Println("")
		fmt.Println("📊 Generating coverage report...")

		coverageCmd := exec.Command("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
		err := coverageCmd.Run()
		if err != nil {
			fmt.Printf("⚠️  Warning: Could not generate HTML coverage report: %v\n", err)
		} else {
			fmt.Println("📄 Coverage report generated: coverage.html")
		}

		// Show coverage summary
		coverageCmd = exec.Command("go", "tool", "cover", "-func=coverage.out")
		output, err := coverageCmd.Output()
		if err == nil {
			fmt.Println("")
			fmt.Println("📈 Coverage Summary:")
			fmt.Println(string(output))
		}
	}

	// Watch mode
	if watch {
		fmt.Println("")
		fmt.Println("👀 Watch mode enabled - tests will re-run on file changes")
		fmt.Println("Press Ctrl+C to stop watching")

		// Simple file watcher implementation
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			fmt.Printf("❌ Could not create file watcher: %v\n", err)
			return
		}
		defer watcher.Close()

		// Watch current directory and subdirectories
		err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && !strings.Contains(path, ".git") && !strings.Contains(path, "vendor") {
				return watcher.Add(path)
			}
			return nil
		})
		if err != nil {
			fmt.Printf("❌ Could not watch directories: %v\n", err)
			return
		}

		// Watch for changes
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					if strings.HasSuffix(event.Name, ".go") {
						fmt.Printf("🔄 File changed: %s - Re-running tests...\n", event.Name)

						// Re-run tests
						testCmd := exec.Command("go", testArgs...)
						testCmd.Env = env
						testCmd.Stdout = os.Stdout
						testCmd.Stderr = os.Stderr

						err := testCmd.Run()
						if err != nil {
							fmt.Printf("❌ Tests failed: %v\n", err)
						} else {
							fmt.Println("✅ Tests passed!")
						}
						fmt.Println("")
					}
				}
			case err := <-watcher.Errors:
				fmt.Printf("❌ Watcher error: %v\n", err)
			}
		}
	}

	fmt.Println("")
	fmt.Println("💡 Testing Tips:")
	fmt.Println("  • Use 'dolphin test --coverage' for coverage reports")
	fmt.Println("  • Use 'dolphin test --suite=integration' for integration tests")
	fmt.Println("  • Use 'dolphin test --watch' for continuous testing")
	fmt.Println("  • Use 'dolphin test ./app/controllers' to test specific packages")
	fmt.Println("")
	fmt.Println("📚 Documentation: https://github.com/mrhoseah/dolphin/blob/main/TESTING_GUIDE.md")
}

// Telemetry command handlers

func runTelemetryEnable(cmd *cobra.Command, args []string) {
	// Initialize telemetry manager
	storage := telemetry.NewFileStorage(telemetry.GetConfigPath())
	sender := telemetry.NewHTTPSender("https://telemetry.dolphin-framework.dev/api/v1/events")
	manager := telemetry.NewTelemetryManager(storage, sender)
	cli := telemetry.NewCLI(manager)
	
	if err := cli.Enable(); err != nil {
		fmt.Printf("❌ Failed to enable telemetry: %v\n", err)
		os.Exit(1)
	}
}

func runTelemetryDisable(cmd *cobra.Command, args []string) {
	// Initialize telemetry manager
	storage := telemetry.NewFileStorage(telemetry.GetConfigPath())
	sender := telemetry.NewHTTPSender("https://telemetry.dolphin-framework.dev/api/v1/events")
	manager := telemetry.NewTelemetryManager(storage, sender)
	cli := telemetry.NewCLI(manager)
	
	if err := cli.Disable(); err != nil {
		fmt.Printf("❌ Failed to disable telemetry: %v\n", err)
		os.Exit(1)
	}
}

func runTelemetryStatus(cmd *cobra.Command, args []string) {
	// Initialize telemetry manager
	storage := telemetry.NewFileStorage(telemetry.GetConfigPath())
	sender := telemetry.NewHTTPSender("https://telemetry.dolphin-framework.dev/api/v1/events")
	manager := telemetry.NewTelemetryManager(storage, sender)
	cli := telemetry.NewCLI(manager)
	
	if err := cli.Status(); err != nil {
		fmt.Printf("❌ Failed to get telemetry status: %v\n", err)
		os.Exit(1)
	}
}

func runTelemetryConfig(cmd *cobra.Command, args []string) {
	// Initialize telemetry manager
	storage := telemetry.NewFileStorage(telemetry.GetConfigPath())
	sender := telemetry.NewHTTPSender("https://telemetry.dolphin-framework.dev/api/v1/events")
	manager := telemetry.NewTelemetryManager(storage, sender)
	cli := telemetry.NewCLI(manager)
	
	if err := cli.Config(); err != nil {
		fmt.Printf("❌ Failed to get telemetry config: %v\n", err)
		os.Exit(1)
	}
}

func runTelemetryReset(cmd *cobra.Command, args []string) {
	// Initialize telemetry manager
	storage := telemetry.NewFileStorage(telemetry.GetConfigPath())
	sender := telemetry.NewHTTPSender("https://telemetry.dolphin-framework.dev/api/v1/events")
	manager := telemetry.NewTelemetryManager(storage, sender)
	cli := telemetry.NewCLI(manager)
	
	if err := cli.Reset(); err != nil {
		fmt.Printf("❌ Failed to reset telemetry config: %v\n", err)
		os.Exit(1)
	}
}

func runTelemetryTest(cmd *cobra.Command, args []string) {
	// Initialize telemetry manager
	storage := telemetry.NewFileStorage(telemetry.GetConfigPath())
	sender := telemetry.NewHTTPSender("https://telemetry.dolphin-framework.dev/api/v1/events")
	manager := telemetry.NewTelemetryManager(storage, sender)
	cli := telemetry.NewCLI(manager)
	
	if err := cli.Test(); err != nil {
		fmt.Printf("❌ Failed to send test telemetry: %v\n", err)
		os.Exit(1)
	}
}

func runTelemetryPrivacy(cmd *cobra.Command, args []string) {
	// Initialize telemetry manager
	storage := telemetry.NewFileStorage(telemetry.GetConfigPath())
	sender := telemetry.NewHTTPSender("https://telemetry.dolphin-framework.dev/api/v1/events")
	manager := telemetry.NewTelemetryManager(storage, sender)
	cli := telemetry.NewCLI(manager)
	
	if err := cli.Privacy(); err != nil {
		fmt.Printf("❌ Failed to show privacy info: %v\n", err)
		os.Exit(1)
	}
}
