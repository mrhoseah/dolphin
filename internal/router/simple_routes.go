package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// SimpleRoutes provides a CodeIgniter-like simple route API
// Usage:
//   routes := router.SimpleRoutes()
//   routes.Get("/", homeHandler)
//   routes.Post("/users", createUserHandler)
//   routes.Auth("/dashboard", dashboardHandler)
//   routes.View("/about", "pages/about")
type SimpleRoutes struct {
	router *Router
	routes *Routes
}

// NewSimpleRoutes creates a new SimpleRoutes instance
func NewSimpleRoutes(router *Router) *SimpleRoutes {
	return &SimpleRoutes{
		router: router,
		routes: NewRoutes(router.router),
	}
}

// Get registers a GET route
func (sr *SimpleRoutes) Get(path string, handler http.HandlerFunc) {
	sr.routes.Get(path, handler)
}

// Post registers a POST route
func (sr *SimpleRoutes) Post(path string, handler http.HandlerFunc) {
	sr.routes.Post(path, handler)
}

// Put registers a PUT route
func (sr *SimpleRoutes) Put(path string, handler http.HandlerFunc) {
	sr.routes.Put(path, handler)
}

// Patch registers a PATCH route
func (sr *SimpleRoutes) Patch(path string, handler http.HandlerFunc) {
	sr.routes.Patch(path, handler)
}

// Delete registers a DELETE route
func (sr *SimpleRoutes) Delete(path string, handler http.HandlerFunc) {
	sr.routes.Delete(path, handler)
}

// Any registers a route for all HTTP methods
func (sr *SimpleRoutes) Any(path string, handler http.HandlerFunc) {
	sr.routes.Any(path, handler)
}

// Match registers a route for specific HTTP methods
func (sr *SimpleRoutes) Match(methods []string, path string, handler http.HandlerFunc) {
	sr.routes.Match(methods, path, handler)
}

// Auth registers a route with authentication middleware
func (sr *SimpleRoutes) Auth(path string, handler http.HandlerFunc) {
	authMiddleware := sr.router.GetAuthMiddleware()
	sr.router.router.With(authMiddleware.Authenticate).Get(path, handler)
}

// AuthPost registers a POST route with authentication middleware
func (sr *SimpleRoutes) AuthPost(path string, handler http.HandlerFunc) {
	authMiddleware := sr.router.GetAuthMiddleware()
	sr.router.router.With(authMiddleware.Authenticate).Post(path, handler)
}

// Guest registers a route for guest-only access
func (sr *SimpleRoutes) Guest(path string, handler http.HandlerFunc) {
	authMiddleware := sr.router.GetAuthMiddleware()
	sr.router.router.With(authMiddleware.Guest).Get(path, handler)
}

// GuestPost registers a POST route for guest-only access
func (sr *SimpleRoutes) GuestPost(path string, handler http.HandlerFunc) {
	authMiddleware := sr.router.GetAuthMiddleware()
	sr.router.router.With(authMiddleware.Guest).Post(path, handler)
}

// Role registers a route with role-based access control
func (sr *SimpleRoutes) Role(path string, role string, handler http.HandlerFunc) {
	authMiddleware := sr.router.GetAuthMiddleware()
	sr.router.router.With(
		authMiddleware.Authenticate,
		authMiddleware.RoleMiddleware(role),
	).Get(path, handler)
}

// Permission registers a route with permission-based access control
func (sr *SimpleRoutes) Permission(path string, permission string, handler http.HandlerFunc) {
	authMiddleware := sr.router.GetAuthMiddleware()
	sr.router.router.With(
		authMiddleware.Authenticate,
		authMiddleware.PermissionMiddleware(permission),
	).Get(path, handler)
}

// View registers a route that renders a Fin template
func (sr *SimpleRoutes) View(path string, templateName string) {
	sr.Get(path, func(w http.ResponseWriter, req *http.Request) {
		data := make(map[string]interface{})
		if err := sr.router.renderFin(w, req, templateName, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

// ViewWithAuth registers a route that renders a Fin template with authentication
// The template will automatically check @auth directive and redirect if not authenticated
func (sr *SimpleRoutes) ViewWithAuth(path string, templateName string) {
	authMiddleware := sr.router.GetAuthMiddleware()
	sr.router.router.With(authMiddleware.Authenticate).Get(path, func(w http.ResponseWriter, req *http.Request) {
		data := make(map[string]interface{})
		if err := sr.router.renderFin(w, req, templateName, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

// Group creates a route group
func (sr *SimpleRoutes) Group(prefix string, middleware []func(http.Handler) http.Handler, fn func(*RouteGroup)) {
	sr.routes.Group(prefix, middleware, fn)
}

// AuthGroup creates an authenticated route group
func (sr *SimpleRoutes) AuthGroup(prefix string, fn func(*RouteGroup)) {
	authMiddleware := sr.router.GetAuthMiddleware()
	sr.routes.Group(prefix, []func(http.Handler) http.Handler{
		authMiddleware.Authenticate,
	}, fn)
}

// GuestGroup creates a guest-only route group
func (sr *SimpleRoutes) GuestGroup(prefix string, fn func(*RouteGroup)) {
	authMiddleware := sr.router.GetAuthMiddleware()
	sr.routes.Group(prefix, []func(http.Handler) http.Handler{
		authMiddleware.Guest,
	}, fn)
}

// RoleGroup creates a role-based route group
func (sr *SimpleRoutes) RoleGroup(prefix string, role string, fn func(*RouteGroup)) {
	authMiddleware := sr.router.GetAuthMiddleware()
	sr.routes.Group(prefix, []func(http.Handler) http.Handler{
		authMiddleware.Authenticate,
		authMiddleware.RoleMiddleware(role),
	}, fn)
}

// PermissionGroup creates a permission-based route group
func (sr *SimpleRoutes) PermissionGroup(prefix string, permission string, fn func(*RouteGroup)) {
	authMiddleware := sr.router.GetAuthMiddleware()
	sr.routes.Group(prefix, []func(http.Handler) http.Handler{
		authMiddleware.Authenticate,
		authMiddleware.PermissionMiddleware(permission),
	}, fn)
}

// Resource registers RESTful resource routes
func (sr *SimpleRoutes) Resource(path string, controller ResourceController) {
	sr.router.router.Route(path, func(r chi.Router) {
		if controller.Index != nil {
			r.Get("/", controller.Index)
		}
		if controller.Create != nil {
			r.Get("/create", controller.Create)
		}
		if controller.Store != nil {
			r.Post("/", controller.Store)
		}
		if controller.Show != nil {
			r.Get("/{id}", controller.Show)
		}
		if controller.Edit != nil {
			r.Get("/{id}/edit", controller.Edit)
		}
		if controller.Update != nil {
			r.Put("/{id}", controller.Update)
			r.Patch("/{id}", controller.Update)
		}
		if controller.Destroy != nil {
			r.Delete("/{id}", controller.Destroy)
		}
	})
}

// ResourceWithAuth registers RESTful resource routes with authentication
func (sr *SimpleRoutes) ResourceWithAuth(path string, controller ResourceController) {
	authMiddleware := sr.router.GetAuthMiddleware()
	sr.router.router.Route(path, func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)
		if controller.Index != nil {
			r.Get("/", controller.Index)
		}
		if controller.Create != nil {
			r.Get("/create", controller.Create)
		}
		if controller.Store != nil {
			r.Post("/", controller.Store)
		}
		if controller.Show != nil {
			r.Get("/{id}", controller.Show)
		}
		if controller.Edit != nil {
			r.Get("/{id}/edit", controller.Edit)
		}
		if controller.Update != nil {
			r.Put("/{id}", controller.Update)
			r.Patch("/{id}", controller.Update)
		}
		if controller.Destroy != nil {
			r.Delete("/{id}", controller.Destroy)
		}
	})
}

