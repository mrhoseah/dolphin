package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"dolphin/internal/app"
	"dolphin/internal/auth"
	"dolphin/internal/maintenance"
	loggingMiddleware "dolphin/internal/middleware/logging"
	recoveryMiddleware "dolphin/internal/middleware/recovery"
	"dolphin/internal/storage"
	"dolphin/internal/template"

	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// Router handles HTTP routing
type Router struct {
	app                *app.App
	router             *chi.Mux
	maintenanceManager *maintenance.Manager
	authManager        *auth.AuthManager
	finEngine          template.FinTemplateEngine
}

// New creates a new router instance
func New(app *app.App) *Router {
	r := &Router{
		app:                app,
		router:             chi.NewRouter(),
		maintenanceManager: maintenance.NewManager("storage/framework/maintenance.json"),
	}

	// Initialize web auth manager (session-based)
	r.authManager = auth.NewAuthManager()

	// Initialize Fin template engine
	finConfig := &template.Config{
		ViewsPath:    "views",
		CachePath:    "storage/cache/views",
		CacheEnabled: true,
		DebugMode:    app.Config().App.Debug,
		Extensions:   []string{".fin.html"}, // Only .fin.html is supported
	}
	r.finEngine = template.NewFinEngine(finConfig)

	r.setupMiddleware()
	// Don't setup routes here - allow apps to add custom routes first
	// r.setupRoutes() will be called after custom routes are added

	return r
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

// GetChiRouter returns the underlying chi router
func (r *Router) GetChiRouter() *chi.Mux {
	return r.router
}

// GetFinEngine returns the Fin template engine
func (r *Router) GetFinEngine() template.FinTemplateEngine {
	return r.finEngine
}

// Mount attaches a sub-router at a given pattern
func (r *Router) Mount(pattern string, sr chi.Router) {
	r.router.Mount(pattern, sr)
}

// Use adds a middleware to the router
func (r *Router) Use(mwf func(http.Handler) http.Handler) {
	r.router.Use(mwf)
}

// setupMiddleware configures global middleware
func (r *Router) setupMiddleware() {
	// Maintenance mode middleware (should be first)
	maintenanceMiddleware := maintenance.NewMiddleware(r.maintenanceManager)
	r.router.Use(maintenanceMiddleware.Handle)

	// Request ID middleware
	r.router.Use(middleware.RequestID)

	// Real IP middleware
	r.router.Use(middleware.RealIP)

	// Logger middleware
	r.router.Use(loggingMiddleware.New(r.app.Logger()))

	// Recovery middleware with debug mode
	r.router.Use(recoveryMiddleware.NewWithConfig(recoveryMiddleware.RecoveryConfig{
		Logger:      r.app.Logger(),
		DebugMode:   r.app.Config().App.Debug,
		Environment: r.app.Config().App.Environment,
	}))

	// Timeout middleware
	r.router.Use(middleware.Timeout(30))

	// CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure based on your needs
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.router.Use(corsMiddleware.Handler)

	// Compress middleware
	r.router.Use(middleware.Compress(5))
}

// SetupRoutes configures application routes
// This should be called after custom routes are added to allow apps to override defaults
func (r *Router) SetupRoutes() {
	r.setupRoutes()
}

// setupRoutes configures application routes
func (r *Router) setupRoutes() {
	// Health check endpoints (Kubernetes-ready)
	r.router.Get("/health", r.healthCheck)
	r.router.Get("/health/live", r.healthLiveness)
	r.router.Get("/health/ready", r.healthReadiness)

	// Maintenance status endpoint
	r.router.Get("/maintenance/status", r.maintenanceStatus)

	// Swagger documentation
	r.router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	// API routes
	r.router.Route("/api", func(api chi.Router) {
		// API v1 routes
		api.Route("/v1", func(v1 chi.Router) {
			r.setupAPIRoutes(v1)
		})
	})

	// Web routes - setup AFTER custom routes can be added
	// This allows apps to override default routes
	// Note: Don't wrap in Route("/") - add directly to allow root path matching
	r.setupWebRoutes(r.router)

	// Static file serving
	r.setupStaticRoutes()

	// Custom 404 handler - must be last
	r.router.NotFound(r.handleNotFound)
	r.router.MethodNotAllowed(r.handleMethodNotAllowed)
}

// placeholderHandler is a temporary handler for routes without controllers
func (r *Router) placeholderHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Controller not implemented yet"}`))
}

// setupStaticRoutes configures static file serving
func (r *Router) setupStaticRoutes() {
	// Serve static files from public directory
	fileServer := http.FileServer(http.Dir("./public/"))
	r.router.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Ensure storage symlink exists (public/storage → storage/app/public)
	// This allows public access to files stored in storage/app/public
	if !storage.SymlinkExists("public") {
		// Try to create symlink automatically
		if err := storage.CreateSymlink("public", "storage"); err != nil {
			r.app.Logger().Warn("Storage symlink not found. Run 'dolphin storage:link' to create it.",
				zap.Error(err),
			)
		} else {
			r.app.Logger().Info("Storage symlink created automatically")
		}
	}

	// Serve storage files from public/storage (symlinked to storage/app/public)
	r.router.Handle("/storage/*", http.StripPrefix("/storage/", http.FileServer(http.Dir("./public/storage/"))))

	// Serve uploaded files (backward compatibility)
	r.router.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./storage/uploads/"))))
}

// Handler methods

func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"dolphin-framework"}`))
}

func (r *Router) healthLiveness(w http.ResponseWriter, req *http.Request) {
	// Liveness probe - checks if the application is alive
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"alive","service":"dolphin-framework"}`))
}

func (r *Router) healthReadiness(w http.ResponseWriter, req *http.Request) {
	// Readiness probe - checks if the application is ready to serve traffic
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"dolphin-framework"}`))
}

func (r *Router) maintenanceStatus(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := r.maintenanceManager.Status()

	if enabled, ok := status["enabled"].(bool); ok && enabled {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	// Convert status to JSON
	jsonData, err := json.Marshal(status)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

// handleNotFound handles 404 Not Found errors
func (r *Router) handleNotFound(w http.ResponseWriter, req *http.Request) {
	r.renderErrorPage(w, req, http.StatusNotFound, "Page Not Found", "The page you're looking for doesn't exist or has been moved.")
}

// handleMethodNotAllowed handles 405 Method Not Allowed errors
func (r *Router) handleMethodNotAllowed(w http.ResponseWriter, req *http.Request) {
	r.renderErrorPage(w, req, http.StatusMethodNotAllowed, "Method Not Allowed", "The requested method is not allowed for this resource.")
}

// renderErrorPage renders an elegant error page
func (r *Router) renderErrorPage(w http.ResponseWriter, req *http.Request, statusCode int, title, message string) {
	// Check if this is an API request
	if strings.HasPrefix(req.URL.Path, "/api/") {
		r.renderAPIError(w, req, statusCode, title, message)
		return
	}

	// Try to render error page using Fin template
	errorData := map[string]interface{}{
		"StatusCode": statusCode,
		"Title":      title,
		"Message":    message,
		"Path":       req.URL.Path,
		"Method":     req.Method,
		"Debug":      r.app.Config().App.Debug,
		"Environment": r.app.Config().App.Environment,
	}

	// Try to render error page template
	html, err := r.finEngine.Render("errors/error", errorData)
	if err != nil {
		// Fallback to built-in error page
		r.renderBuiltInErrorPage(w, req, statusCode, title, message, errorData)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(html))
}

// renderAPIError renders JSON error response for API requests
func (r *Router) renderAPIError(w http.ResponseWriter, req *http.Request, statusCode int, title, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":   title,
		"message": message,
		"status":  statusCode,
	}

	// Add debug info in development
	if r.app.Config().App.Debug || r.app.Config().App.Environment == "development" {
		response["path"] = req.URL.Path
		response["method"] = req.Method
	}

	json.NewEncoder(w).Encode(response)
}

// renderBuiltInErrorPage renders a built-in elegant error page
func (r *Router) renderBuiltInErrorPage(w http.ResponseWriter, req *http.Request, statusCode int, title, message string, data map[string]interface{}) {
	debugMode := r.app.Config().App.Debug || r.app.Config().App.Environment == "development"
	env := r.app.Config().App.Environment

	html := r.generateErrorPageHTML(statusCode, title, message, req.URL.Path, req.Method, debugMode, env)
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(html))
}

// generateErrorPageHTML generates elegant error page HTML
func (r *Router) generateErrorPageHTML(statusCode int, title, message, path, method string, debugMode bool, env string) string {
	statusEmoji := "🐬"
	
	switch statusCode {
	case 404:
		statusEmoji = "🔍"
	case 500:
		statusEmoji = "⚠️"
	case 403:
		statusEmoji = "🔒"
	case 401:
		statusEmoji = "🔐"
	}

	debugSection := ""
	if debugMode {
		debugSection = fmt.Sprintf(`
		<div class="bg-gray-50 border border-gray-200 rounded-lg p-6 mt-6">
			<h3 class="text-lg font-semibold text-gray-900 mb-4">Debug Information</h3>
			<div class="space-y-2 text-sm">
				<p><strong class="text-gray-700">Path:</strong> <code class="bg-gray-200 px-2 py-1 rounded">%s</code></p>
				<p><strong class="text-gray-700">Method:</strong> <code class="bg-gray-200 px-2 py-1 rounded">%s</code></p>
				<p><strong class="text-gray-700">Status Code:</strong> <code class="bg-gray-200 px-2 py-1 rounded">%d</code></p>
				<p><strong class="text-gray-700">Environment:</strong> <code class="bg-gray-200 px-2 py-1 rounded">%s</code></p>
			</div>
		</div>`, path, method, statusCode, env)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s - Dolphin Framework</title>
	<script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gradient-to-br from-blue-50 to-indigo-100 min-h-screen flex items-center justify-center p-4">
	<div class="bg-white rounded-2xl shadow-2xl max-w-2xl w-full p-8 md:p-12">
		<div class="text-center mb-8">
			<div class="text-6xl md:text-8xl mb-4">%s</div>
			<h1 class="text-3xl md:text-4xl font-bold text-gray-900 mb-2">%s</h1>
			<p class="text-lg text-gray-600">%s</p>
		</div>
		
		<div class="border-t border-gray-200 pt-6">
			<div class="flex flex-col sm:flex-row gap-4 justify-center">
				<a href="/" class="inline-flex items-center justify-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors">
					<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"></path>
					</svg>
					Go Home
				</a>
				<button onclick="window.history.back()" class="inline-flex items-center justify-center px-6 py-3 border border-gray-300 text-base font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors">
					<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path>
					</svg>
					Go Back
				</button>
			</div>
		</div>
		
		%s
	</div>
</body>
	</html>`, title, statusEmoji, title, message, debugSection)
}
