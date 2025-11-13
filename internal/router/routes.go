package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RouteGroup represents a group of routes with shared middleware
type RouteGroup struct {
	router     chi.Router
	middleware []func(http.Handler) http.Handler
	prefix     string
}

// Routes provides a simple, CodeIgniter-like route registration API
type Routes struct {
	routerInstance *Router
	router         *chi.Mux
	groups         map[string]*RouteGroup
}

// NewRoutes creates a new Routes instance
func NewRoutes(router *chi.Mux) *Routes {
	return &Routes{
		router: router,
		groups: make(map[string]*RouteGroup),
	}
}

// NewRoutesWithRouter creates a new Routes instance with Router reference
func NewRoutesWithRouter(routerInstance *Router) *Routes {
	return &Routes{
		routerInstance: routerInstance,
		router:         routerInstance.router,
		groups:         make(map[string]*RouteGroup),
	}
}

// Get registers a GET route
func (r *Routes) Get(path string, handler http.HandlerFunc) {
	r.router.Get(path, handler)
}

// Post registers a POST route
func (r *Routes) Post(path string, handler http.HandlerFunc) {
	r.router.Post(path, handler)
}

// Put registers a PUT route
func (r *Routes) Put(path string, handler http.HandlerFunc) {
	r.router.Put(path, handler)
}

// Patch registers a PATCH route
func (r *Routes) Patch(path string, handler http.HandlerFunc) {
	r.router.Patch(path, handler)
}

// Delete registers a DELETE route
func (r *Routes) Delete(path string, handler http.HandlerFunc) {
	r.router.Delete(path, handler)
}

// Options registers an OPTIONS route
func (r *Routes) Options(path string, handler http.HandlerFunc) {
	r.router.Options(path, handler)
}

// Head registers a HEAD route
func (r *Routes) Head(path string, handler http.HandlerFunc) {
	r.router.Head(path, handler)
}

// Any registers a route for all HTTP methods
func (r *Routes) Any(path string, handler http.HandlerFunc) {
	r.router.HandleFunc(path, handler)
}

// Match registers a route for specific HTTP methods
func (r *Routes) Match(methods []string, path string, handler http.HandlerFunc) {
	for _, method := range methods {
		switch method {
		case "GET":
			r.Get(path, handler)
		case "POST":
			r.Post(path, handler)
		case "PUT":
			r.Put(path, handler)
		case "PATCH":
			r.Patch(path, handler)
		case "DELETE":
			r.Delete(path, handler)
		case "OPTIONS":
			r.Options(path, handler)
		case "HEAD":
			r.Head(path, handler)
		}
	}
}

// Group creates a route group with shared middleware
func (r *Routes) Group(prefix string, middleware []func(http.Handler) http.Handler, fn func(*RouteGroup)) {
	group := &RouteGroup{
		router:     r.router,
		middleware: middleware,
		prefix:     prefix,
	}
	r.groups[prefix] = group
	fn(group)
}

// RouteGroup methods

// Get registers a GET route in the group
func (rg *RouteGroup) Get(path string, handler http.HandlerFunc) {
	fullPath := rg.prefix + path
	rg.router.Route(fullPath, func(r chi.Router) {
		for _, mw := range rg.middleware {
			r.Use(mw)
		}
		r.Get("/", handler)
	})
}

// Post registers a POST route in the group
func (rg *RouteGroup) Post(path string, handler http.HandlerFunc) {
	fullPath := rg.prefix + path
	rg.router.Route(fullPath, func(r chi.Router) {
		for _, mw := range rg.middleware {
			r.Use(mw)
		}
		r.Post("/", handler)
	})
}

// Put registers a PUT route in the group
func (rg *RouteGroup) Put(path string, handler http.HandlerFunc) {
	fullPath := rg.prefix + path
	rg.router.Route(fullPath, func(r chi.Router) {
		for _, mw := range rg.middleware {
			r.Use(mw)
		}
		r.Put("/", handler)
	})
}

// Patch registers a PATCH route in the group
func (rg *RouteGroup) Patch(path string, handler http.HandlerFunc) {
	fullPath := rg.prefix + path
	rg.router.Route(fullPath, func(r chi.Router) {
		for _, mw := range rg.middleware {
			r.Use(mw)
		}
		r.Patch("/", handler)
	})
}

// Delete registers a DELETE route in the group
func (rg *RouteGroup) Delete(path string, handler http.HandlerFunc) {
	fullPath := rg.prefix + path
	rg.router.Route(fullPath, func(r chi.Router) {
		for _, mw := range rg.middleware {
			r.Use(mw)
		}
		r.Delete("/", handler)
	})
}

// Resource registers RESTful resource routes
func (rg *RouteGroup) Resource(path string, controller ResourceController) {
	rg.router.Route(rg.prefix+path, func(r chi.Router) {
		for _, mw := range rg.middleware {
			r.Use(mw)
		}
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

// ResourceController defines methods for RESTful resource routes
type ResourceController struct {
	Index   http.HandlerFunc
	Create  http.HandlerFunc
	Store   http.HandlerFunc
	Show    http.HandlerFunc
	Edit    http.HandlerFunc
	Update  http.HandlerFunc
	Destroy http.HandlerFunc
}

// AuthGroup creates a route group with authentication middleware
func (r *Routes) AuthGroup(prefix string, fn func(*RouteGroup)) {
	authMiddleware := r.router.(*Router).GetAuthMiddleware()
	r.Group(prefix, []func(http.Handler) http.Handler{
		authMiddleware.Authenticate,
	}, fn)
}

// GuestGroup creates a route group for guest-only routes
func (r *Routes) GuestGroup(prefix string, fn func(*RouteGroup)) {
	authMiddleware := r.router.(*Router).GetAuthMiddleware()
	r.Group(prefix, []func(http.Handler) http.Handler{
		authMiddleware.Guest,
	}, fn)
}

// RoleGroup creates a route group with role-based middleware
func (r *Routes) RoleGroup(prefix string, role string, fn func(*RouteGroup)) {
	authMiddleware := r.router.(*Router).GetAuthMiddleware()
	r.Group(prefix, []func(http.Handler) http.Handler{
		authMiddleware.Authenticate,
		authMiddleware.RoleMiddleware(role),
	}, fn)
}

// PermissionGroup creates a route group with permission-based middleware
func (r *Routes) PermissionGroup(prefix string, permission string, fn func(*RouteGroup)) {
	authMiddleware := r.router.(*Router).GetAuthMiddleware()
	r.Group(prefix, []func(http.Handler) http.Handler{
		authMiddleware.Authenticate,
		authMiddleware.PermissionMiddleware(permission),
	}, fn)
}

// View registers a route that renders a Fin template
func (r *Routes) View(path string, templateName string) {
	r.Get(path, func(w http.ResponseWriter, req *http.Request) {
		router := r.router.(*Router)
		data := make(map[string]interface{})
		if err := router.renderFin(w, req, templateName, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

// ViewWithAuth registers a route that renders a Fin template with authentication
func (r *Routes) ViewWithAuth(path string, templateName string) {
	authMiddleware := r.router.(*Router).GetAuthMiddleware()
	r.router.With(authMiddleware.Authenticate).Get(path, func(w http.ResponseWriter, req *http.Request) {
		router := r.router.(*Router)
		data := make(map[string]interface{})
		if err := router.renderFin(w, req, templateName, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

