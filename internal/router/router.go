package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/mrhoseah/dolphin/internal/app"
	"github.com/mrhoseah/dolphin/internal/auth"
	"github.com/mrhoseah/dolphin/internal/maintenance"
	authMiddleware "github.com/mrhoseah/dolphin/internal/middleware"
	loggingMiddleware "github.com/mrhoseah/dolphin/internal/middleware/logging"
	recoveryMiddleware "github.com/mrhoseah/dolphin/internal/middleware/recovery"
	"github.com/mrhoseah/dolphin/internal/session"
	"github.com/mrhoseah/dolphin/internal/storage"
	"github.com/mrhoseah/dolphin/internal/template"

	"github.com/gorilla/sessions"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// Router handles HTTP routing
type Router struct {
	app                *app.App
	router             *chi.Mux
	maintenanceManager *maintenance.Manager
	authManager        *auth.AuthManager
	sessionManager     *session.SessionManager
	sessionStore       *sessions.CookieStore
	finEngine          template.FinTemplateEngine
}

// New creates a new router instance
func New(app *app.App) *Router {
	r := &Router{
		app:                app,
		router:             chi.NewRouter(),
		maintenanceManager: maintenance.NewManager("storage/framework/maintenance.json"),
	}

	// Initialize session manager with secure defaults
	secretKey := app.Config().App.Key
	if secretKey == "" || secretKey == "your-application-key-here" {
		// Warn about insecure default key in production
		if app.Config().App.Environment == "production" {
			app.Logger().Warn("Using default session key in production is insecure. Please set a strong secret key in configuration.")
		}
		secretKey = "dolphin-secret-key-change-in-production"
	}
	
	// Ensure secret key is at least 32 bytes for security
	if len(secretKey) < 32 {
		app.Logger().Warn("Session secret key is too short. Recommended minimum length is 32 characters.")
	}
	
	r.sessionManager = session.NewSessionManager(secretKey)
	r.sessionStore = sessions.NewCookieStore([]byte(secretKey))
	
	// Configure secure session options
	isProduction := app.Config().App.Environment == "production"
	r.sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,      // Prevent XSS attacks
		Secure:   isProduction, // Only send over HTTPS in production
		SameSite: http.SameSiteLaxMode, // CSRF protection
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

// GetAuthManager returns the auth manager
func (r *Router) GetAuthManager() *auth.AuthManager {
	return r.authManager
}

// Mount attaches a sub-router at a given pattern
func (r *Router) Mount(pattern string, sr chi.Router) {
	r.router.Mount(pattern, sr)
}

// GetAuthMiddleware returns a new auth middleware instance
// Use this to protect Fin template routes:
//
//	authMiddleware := router.GetAuthMiddleware()
//	router.GetChiRouter().Group(func(r chi.Router) {
//		r.Use(authMiddleware.Authenticate)
//		r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
//			router.renderFin(w, r, "pages/dashboard", map[string]interface{}{})
//		})
//	})
func (r *Router) GetAuthMiddleware() *authMiddleware.AuthMiddleware {
	return authMiddleware.NewAuthMiddleware(r.authManager, r.app.Logger())
}

// GetMetricsCollector returns a new metrics collector instance
func (r *Router) GetMetricsCollector() *MetricsCollector {
	return NewMetricsCollector(r.app.Logger())
}

// GetSSEServer returns a new SSE server instance
func (r *Router) GetSSEServer() *SSEServer {
	return NewSSEServer(r.app.Logger())
}

// Routes returns a simple route registration helper
func (r *Router) Routes() *Routes {
	return NewRoutesWithRouter(r)
}

// SimpleRoutes returns a CodeIgniter-like simple route API
func (r *Router) SimpleRoutes() *SimpleRoutes {
	return NewSimpleRoutes(r)
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

	// Session middleware (must be before auth)
	r.router.Use(session.SessionMiddleware(r.sessionManager, "dolphin_session"))

	// Auth guard initialization middleware
	// This sets up the HTTP session guard with request/response for each request
	r.router.Use(r.authGuardMiddleware())

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

	// Request ID middleware (if not already set by chi)
	r.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Set request ID in context if not already present
			if reqID := req.Header.Get("X-Request-ID"); reqID != "" {
				req = SetRequestID(req, reqID)
			} else if reqID := middleware.GetReqID(req.Context()); reqID != "" {
				req = SetRequestID(req, reqID)
			}
			
			// Set IP address in context (chi middleware already sets this)
			ip := req.RemoteAddr
			if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
				// Take the first IP from X-Forwarded-For header
				ips := strings.Split(forwarded, ",")
				if len(ips) > 0 {
					ip = strings.TrimSpace(ips[0])
				}
			}
			req = SetIPAddress(req, ip)
			
			next.ServeHTTP(w, req)
		})
	})

	// CORS middleware with secure defaults
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure based on your needs in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.router.Use(corsMiddleware.Handler)

	// Security headers middleware
	r.router.Use(r.securityHeadersMiddleware())

	// Compress middleware
	r.router.Use(middleware.Compress(5))
}

// securityHeadersMiddleware adds security headers to all responses.
// This helps protect against common web vulnerabilities.
func (r *Router) securityHeadersMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Prevent clickjacking
			w.Header().Set("X-Frame-Options", "DENY")
			
			// Prevent MIME type sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")
			
			// Enable XSS protection
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			
			// Referrer policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			
			// Content Security Policy (basic - apps should customize)
			if r.app.Config().App.Environment == "production" {
				w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';")
			}
			
			// Permissions Policy (formerly Feature Policy)
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			
			next.ServeHTTP(w, req)
		})
	}
}

// authGuardMiddleware initializes the HTTP session guard for each request
func (r *Router) authGuardMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Get or create HTTP session guard
			guard := r.authManager.Guard("web")
			if guard == nil {
				// Create HTTP session guard if it doesn't exist
				// Note: Apps should register their own provider via auth manager
				// This is a fallback for basic functionality
				provider := auth.NewDatabaseProvider(r.app.DB().GetDB(), &auth.User{})
				httpGuard := auth.NewHTTPSessionGuard("web", provider, r.sessionStore, "dolphin_session")
				r.authManager.RegisterGuard("web", httpGuard)
				r.authManager.SetDefaultGuard("web")
				guard = httpGuard
				
				r.app.Logger().Debug("Created default HTTP session guard",
					zap.String("guard", "web"),
				)
			}

			// If it's an HTTPGuard, set the request/response for session access
			if httpGuard, ok := guard.(auth.HTTPGuard); ok {
				httpGuard.SetRequestResponse(req, w)
			}

			next.ServeHTTP(w, req)
		})
	}
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

// setupStaticRoutes configures static file serving with proper security headers.
// Static files are served from the public directory with appropriate caching headers.
func (r *Router) setupStaticRoutes() {
	// Serve static files from public directory
	fileServer := http.FileServer(http.Dir("./public/"))
	
	// Wrap file server with security headers for static content
	staticHandler := r.staticFileHandler(fileServer)
	r.router.Handle("/static/*", http.StripPrefix("/static/", staticHandler))

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
	r.router.Handle("/storage/*", http.StripPrefix("/storage/", r.staticFileHandler(http.FileServer(http.Dir("./public/storage/")))))

	// Serve uploaded files (backward compatibility)
	r.router.Handle("/uploads/*", http.StripPrefix("/uploads/", r.staticFileHandler(http.FileServer(http.Dir("./storage/uploads/")))))
}

// staticFileHandler wraps a file server with appropriate headers for static content.
func (r *Router) staticFileHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Set cache headers for static files (1 year cache)
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		
		// Security headers for static files
		w.Header().Set("X-Content-Type-Options", "nosniff")
		
		next.ServeHTTP(w, req)
	})
}

// Handler methods

// healthCheck handles the basic health check endpoint.
// Returns a JSON response indicating the service is operational.
func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"status":  "ok",
		"service": "dolphin-framework",
	}
	json.NewEncoder(w).Encode(response)
}

// healthLiveness handles the Kubernetes liveness probe endpoint.
// Returns a JSON response indicating the application is alive and should not be restarted.
func (r *Router) healthLiveness(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"status":  "alive",
		"service": "dolphin-framework",
	}
	json.NewEncoder(w).Encode(response)
}

// healthReadiness handles the Kubernetes readiness probe endpoint.
// Returns a JSON response indicating the application is ready to serve traffic.
// This can be extended to check database connectivity, external services, etc.
func (r *Router) healthReadiness(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Check if database is available (basic readiness check)
	ready := true
	if r.app.DB() != nil {
		sqlDB, err := r.app.DB().GetDB().DB()
		if err == nil {
			if err := sqlDB.Ping(); err != nil {
				ready = false
			}
		}
	}
	
	if ready {
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"status":  "ready",
			"service": "dolphin-framework",
		}
		json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		response := map[string]interface{}{
			"status":  "not_ready",
			"service": "dolphin-framework",
			"reason":  "database_unavailable",
		}
		json.NewEncoder(w).Encode(response)
	}
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

	// Try to render error page template (must include .fin.html extension)
	html, err := r.finEngine.Render("errors/error.fin.html", errorData)
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
