package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// Category handles category related requests
type Category struct{}

// NewCategory creates a new Category controller
func NewCategory() *Category {
	return &Category{}
}

// Index handles GET /category
func (c *Category) Index(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]interface{}{
		"message": "List of category",
		"data":    []interface{}{},
	})
}

// Show handles GET /category/{id}
func (c *Category) Show(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	render.JSON(w, r, map[string]interface{}{
		"message": "Show category",
		"id":      id,
		"data":    map[string]interface{}{},
	})
}

// Store handles POST /category
func (c *Category) Store(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]interface{}{
		"message": "category created successfully",
		"data":    map[string]interface{}{},
	})
}

// Update handles PUT /category/{id}
func (c *Category) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	render.JSON(w, r, map[string]interface{}{
		"message": "Category updated successfully",
		"id":      id,
		"data":    map[string]interface{}{},
	})
}

// Destroy handles DELETE /category/{id}
func (c *Category) Destroy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	render.JSON(w, r, map[string]interface{}{
		"message": "category deleted successfully",
		"id":      id,
	})
}