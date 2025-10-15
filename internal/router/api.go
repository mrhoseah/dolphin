package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mrhoseah/dolphin/app/http/controllers"
	"github.com/mrhoseah/dolphin/internal/auth"
	dolphinMiddleware "github.com/mrhoseah/dolphin/internal/middleware"
)

// setupAPIRoutes configures API routes
func (r *Router) setupAPIRoutes(router chi.Router) {
	// Setup Dolphin-style authentication
	sessionStore := auth.NewMemorySessionStore()
	authManager := auth.SetupAuth(r.app.DB().GetDB(), sessionStore)

	// Initialize Dolphin-style auth middleware
	dolphinAuthMiddleware := dolphinMiddleware.NewAuthMiddleware(authManager, r.app.Logger())

	// Initialize controllers
	dolphinAuthController := controllers.NewAuthController(authManager, r.app.Logger())

	// Authentication routes (Dolphin style)
	router.Route("/auth", func(auth chi.Router) {
		// Public routes (guests only)
		auth.Group(func(guest chi.Router) {
			guest.Use(dolphinAuthMiddleware.Guest)
			guest.Post("/login", dolphinAuthController.Login)
			guest.Post("/register", dolphinAuthController.Register)
		})

		// Protected routes (authenticated users only)
		auth.Group(func(protected chi.Router) {
			protected.Use(dolphinAuthMiddleware.Authenticate)
			protected.Post("/logout", dolphinAuthController.Logout)
			protected.Get("/me", dolphinAuthController.Me)
		})

		// Public status routes
		auth.Get("/check", dolphinAuthController.Check)
		auth.Get("/guest", dolphinAuthController.Guest)
	})

	// Protected API routes
	router.Route("/api", func(api chi.Router) {
		api.Use(dolphinAuthMiddleware.Authenticate)

		// User routes
		api.Route("/users", func(users chi.Router) {
			users.Get("/", r.placeholderHandler)
			users.Post("/", r.placeholderHandler)
			users.Get("/{id}", r.placeholderHandler)
			users.Put("/{id}", r.placeholderHandler)
			users.Delete("/{id}", r.placeholderHandler)
		})

		// Admin routes (role-based)
		api.Route("/admin", func(admin chi.Router) {
			admin.Use(dolphinAuthMiddleware.RoleMiddleware("admin"))
			admin.Get("/dashboard", r.placeholderHandler)
			admin.Get("/users", r.placeholderHandler)
		})

		// Posts routes with permissions
		api.Route("/posts", func(posts chi.Router) {
			posts.Get("/", r.placeholderHandler) // read permission
			posts.Post("/", dolphinAuthMiddleware.PermissionMiddleware("write")(http.HandlerFunc(r.placeholderHandler)).ServeHTTP)
			posts.Get("/{id}", r.placeholderHandler)
			posts.Put("/{id}", dolphinAuthMiddleware.PermissionMiddleware("write")(http.HandlerFunc(r.placeholderHandler)).ServeHTTP)
			posts.Delete("/{id}", dolphinAuthMiddleware.PermissionMiddleware("delete")(http.HandlerFunc(r.placeholderHandler)).ServeHTTP)
		})
	})
}
