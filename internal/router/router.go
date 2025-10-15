package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/mrhoseah/dolphin/internal/app"
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

	// Serve uploaded files
	r.router.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./storage/uploads/"))))
}

// Handler methods

func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"dolphin-framework"}`))
}
