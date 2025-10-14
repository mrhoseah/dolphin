package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mrhoseah/dolphin/app/http/controllers"
	"github.com/mrhoseah/dolphin/internal/app"
	AppMiddleware "github.com/mrhoseah/dolphin/internal/middleware/auth"
	loggingMiddleware "github.com/mrhoseah/dolphin/internal/middleware/logging"
	recoveryMiddleware "github.com/mrhoseah/dolphin/internal/middleware/recovery"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Router handles HTTP routing
type Router struct {
	app    *app.App
	router *chi.Mux
}

// New creates a new router instance
func New(app *app.App) *Router {
	r := &Router{
		app:    app,
		router: chi.NewRouter(),
	}

	r.setupMiddleware()
	r.setupRoutes()

	return r
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

// setupMiddleware configures global middleware
func (r *Router) setupMiddleware() {
	// Request ID middleware
	r.router.Use(middleware.RequestID)

	// Real IP middleware
	r.router.Use(middleware.RealIP)

	// Logger middleware
	r.router.Use(loggingMiddleware.New(r.app.Logger()))

	// Recovery middleware
	r.router.Use(recoveryMiddleware.New(r.app.Logger()))

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

// setupRoutes configures application routes
func (r *Router) setupRoutes() {
	// Health check endpoint
	r.router.Get("/health", r.healthCheck)

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

	// Web routes
	r.router.Route("/", func(web chi.Router) {
		r.setupWebRoutes(web)
	})

	// Static file serving
	r.setupStaticRoutes()
}

// setupAPIRoutes configures API routes
func (r *Router) setupAPIRoutes(router chi.Router) {
	// Initialize controllers
	authController := controllers.NewAuthController()
	userController := controllers.NewUserController()

	// Authentication routes
	router.Route("/auth", func(auth chi.Router) {
		auth.Post("/login", authController.Login)
		auth.Post("/register", authController.Register)
		auth.Post("/logout", authController.Logout)
		auth.Post("/refresh", authController.RefreshToken)
	})

	// Protected routes
	router.Route("/protected", func(protected chi.Router) {
		protected.Use(AppMiddleware.New(r.app.Config().JWT.Secret))

		protected.Get("/user", r.handleGetUser)
		protected.Put("/user", userController.Update)
		protected.Delete("/user", userController.Destroy)
	})

	// Resource routes
	router.Route("/users", func(users chi.Router) {
		users.Get("/", userController.Index)
		users.Post("/", userController.Store)
		users.Get("/{id}", userController.Show)
		users.Put("/{id}", userController.Update)
		users.Delete("/{id}", userController.Destroy)
	})

	router.Route("/posts", func(posts chi.Router) {
		posts.Get("/", r.handleGetPosts)
		posts.Post("/", r.handleCreatePost)
		posts.Get("/{id}", r.handleGetPost)
		posts.Put("/{id}", r.handleUpdatePost)
		posts.Delete("/{id}", r.handleDeletePost)
	})
}

// setupWebRoutes configures web routes
func (r *Router) setupWebRoutes(router chi.Router) {
	// Home page
	router.Get("/", r.handleHome)

	// Dashboard (protected)
	router.Route("/dashboard", func(dashboard chi.Router) {
		dashboard.Use(AppMiddleware.New(r.app.Config().JWT.Secret))
		dashboard.Get("/", r.handleDashboard)
	})

	// Admin routes
	router.Route("/admin", func(admin chi.Router) {
		admin.Use(AppMiddleware.New(r.app.Config().JWT.Secret))
		admin.Use(r.adminMiddleware)

		admin.Get("/", r.handleAdminDashboard)
		admin.Get("/users", r.handleAdminUsers)
		admin.Get("/posts", r.handleAdminPosts)
	})
}

// setupStaticRoutes configures static file serving
func (r *Router) setupStaticRoutes() {
	// Serve static files from public directory
	fileServer := http.FileServer(http.Dir("./public/"))
	r.router.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Serve uploaded files
	r.router.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./storage/uploads/"))))
}

// Handler methods

func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"dolphin-framework"}`))
}

func (r *Router) handleHome(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dolphin Framework</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen flex items-center justify-center">
        <div class="max-w-md w-full bg-white rounded-lg shadow-md p-6">
            <div class="text-center">
                <h1 class="text-3xl font-bold text-gray-900 mb-4">🐬 Dolphin Framework</h1>
                <p class="text-gray-600 mb-6">Enterprise-grade Go web framework</p>
                <div class="space-y-2">
                    <a href="/api/v1/auth/login" class="block w-full bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-600 transition">API Login</a>
                    <a href="/dashboard" class="block w-full bg-green-500 text-white py-2 px-4 rounded hover:bg-green-600 transition">Dashboard</a>
                </div>
            </div>
        </div>
    </div>
</body>
</html>
	`))
}

func (r *Router) handleDashboard(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dashboard - Dolphin Framework</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen">
        <nav class="bg-white shadow">
            <div class="max-w-7xl mx-auto px-4">
                <div class="flex justify-between h-16">
                    <div class="flex items-center">
                        <h1 class="text-xl font-semibold">🐬 Dolphin Dashboard</h1>
                    </div>
                    <div class="flex items-center space-x-4">
                        <a href="/api/v1/auth/logout" class="text-gray-500 hover:text-gray-700">Logout</a>
                    </div>
                </div>
            </div>
        </nav>
        <div class="max-w-7xl mx-auto py-6 px-4">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div class="bg-white rounded-lg shadow p-6">
                    <h3 class="text-lg font-medium text-gray-900">Users</h3>
                    <p class="text-gray-600">Manage user accounts</p>
                </div>
                <div class="bg-white rounded-lg shadow p-6">
                    <h3 class="text-lg font-medium text-gray-900">Posts</h3>
                    <p class="text-gray-600">Manage blog posts</p>
                </div>
                <div class="bg-white rounded-lg shadow p-6">
                    <h3 class="text-lg font-medium text-gray-900">Settings</h3>
                    <p class="text-gray-600">Application settings</p>
                </div>
            </div>
        </div>
    </div>
</body>
</html>
	`))
}

// API handlers
func (r *Router) handleLogin(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Login endpoint - implement authentication logic"}`))
}

func (r *Router) handleRegister(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Register endpoint - implement registration logic"}`))
}

func (r *Router) handleLogout(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Logout endpoint - implement logout logic"}`))
}

func (r *Router) handleRefreshToken(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Refresh token endpoint - implement token refresh logic"}`))
}

func (r *Router) handleGetUser(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Get user endpoint - implement user retrieval logic"}`))
}

func (r *Router) handleUpdateUser(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Update user endpoint - implement user update logic"}`))
}

func (r *Router) handleDeleteUser(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Delete user endpoint - implement user deletion logic"}`))
}

func (r *Router) handleGetUsers(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Get users endpoint - implement users list logic"}`))
}

func (r *Router) handleCreateUser(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Create user endpoint - implement user creation logic"}`))
}

func (r *Router) handleGetPosts(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Get posts endpoint - implement posts list logic"}`))
}

func (r *Router) handleCreatePost(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Create post endpoint - implement post creation logic"}`))
}

func (r *Router) handleGetPost(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Get post endpoint - implement post retrieval logic"}`))
}

func (r *Router) handleUpdatePost(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Update post endpoint - implement post update logic"}`))
}

func (r *Router) handleDeletePost(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Delete post endpoint - implement post deletion logic"}`))
}

func (r *Router) handleAdminDashboard(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<h1>Admin Dashboard</h1><p>Admin panel - implement admin interface</p>`))
}

func (r *Router) handleAdminUsers(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<h1>Admin Users</h1><p>User management - implement admin user interface</p>`))
}

func (r *Router) handleAdminPosts(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<h1>Admin Posts</h1><p>Post management - implement admin post interface</p>`))
}

// Middleware
func (r *Router) adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Check if user has admin role
		// This is a placeholder - implement proper admin role checking
		next.ServeHTTP(w, req)
	})
}
