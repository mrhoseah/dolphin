package router

import (
	"net/http"

	"github.com/mrhoseah/dolphin/internal/app"
	"github.com/mrhoseah/dolphin/internal/router"
	"github.com/mrhoseah/dolphin/internal/template"

	"github.com/go-chi/chi/v5"
)

// Router represents the HTTP router wrapper
type Router struct {
	*router.Router
}

// New creates a new router instance
func New(app *app.App) *Router {
	return &Router{
		Router: router.New(app),
	}
}

// GetRouter returns the underlying chi router for custom route setup
func (r *Router) GetRouter() chi.Router {
	return r.Router.GetChiRouter()
}

// GetFinEngine returns the Fin template engine
func (r *Router) GetFinEngine() template.FinTemplateEngine {
	return r.Router.GetFinEngine()
}

// SetupRoutes configures Dolphin's default routes
// Call this after adding your custom routes to allow them to override defaults
func (r *Router) SetupRoutes() {
	r.Router.SetupRoutes()
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Router.ServeHTTP(w, req)
}
