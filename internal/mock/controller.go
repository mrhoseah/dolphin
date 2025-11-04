package mock

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// Controller handles mock server API endpoints
type Controller struct {
	server *Server
	logger *zap.Logger
}

// NewController creates a new mock controller
func NewController(server *Server, logger *zap.Logger) *Controller {
	return &Controller{
		server: server,
		logger: logger,
	}
}

// List returns all mock endpoints
func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	endpoints := c.server.GetAll()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"endpoints": endpoints,
	})
}

// Register registers a new mock endpoint
func (c *Controller) Register(w http.ResponseWriter, r *http.Request) {
	var endpoint MockEndpoint

	if err := json.NewDecoder(r.Body).Decode(&endpoint); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c.server.Register(&endpoint)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Mock endpoint registered",
	})
}

// Handle handles mock requests
func (c *Controller) Handle(w http.ResponseWriter, r *http.Request) {
	c.server.Handle(w, r)
}

