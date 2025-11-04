package router

import (
	"net/http"

	"dolphin/internal/app"
	"dolphin/internal/router"

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

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Router.ServeHTTP(w, req)
}
