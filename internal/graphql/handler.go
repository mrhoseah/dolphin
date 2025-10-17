package graphql

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/graphql-go/graphql/gqlerrors"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for GraphQL
type Handler struct {
	schemaManager *SchemaManager
	logger        *zap.Logger
}

// NewHandler creates a new GraphQL handler
func NewHandler(schemaManager *SchemaManager, logger *zap.Logger) *Handler {
	return &Handler{
		schemaManager: schemaManager,
		logger:        logger,
	}
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   interface{}                `json:"data,omitempty"`
	Errors []gqlerrors.FormattedError `json:"errors,omitempty"`
}

// ServeHTTP handles HTTP requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if GraphQL is enabled
	if !h.schemaManager.IsEnabled() {
		h.handleDisabled(w, r)
		return
	}

	// Set CORS headers
	h.setCORSHeaders(w, r)

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow POST requests for GraphQL
	if r.Method != "POST" {
		h.handleMethodNotAllowed(w, r)
		return
	}

	// Parse request
	var req GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleBadRequest(w, r, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	// Validate query
	if strings.TrimSpace(req.Query) == "" {
		h.handleBadRequest(w, r, "Query is required")
		return
	}

	// Execute GraphQL query
	ctx := r.Context()
	result := h.schemaManager.Execute(ctx, req.Query, req.Variables)

	// Prepare response
	response := GraphQLResponse{
		Data:   result.Data,
		Errors: result.Errors,
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode GraphQL response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log request
	h.logger.Info("GraphQL request processed",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Bool("has_errors", len(result.Errors) > 0),
		zap.Int("error_count", len(result.Errors)),
	)
}

// handleDisabled handles requests when GraphQL is disabled
func (h *Handler) handleDisabled(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)

	response := map[string]interface{}{
		"error":     "GraphQL endpoint is disabled",
		"message":   "GraphQL functionality is currently disabled. Please contact the administrator.",
		"status":    "disabled",
		"code":      503,
		"timestamp": time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(response)
}

// handleMethodNotAllowed handles unsupported HTTP methods
func (h *Handler) handleMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)

	response := map[string]interface{}{
		"error":     "Method Not Allowed",
		"message":   fmt.Sprintf("Method %s is not allowed for GraphQL endpoint", r.Method),
		"status":    "method_not_allowed",
		"code":      405,
		"timestamp": time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(response)
}

// handleBadRequest handles bad requests
func (h *Handler) handleBadRequest(w http.ResponseWriter, r *http.Request, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := map[string]interface{}{
		"error":     "Bad Request",
		"message":   message,
		"status":    "bad_request",
		"code":      400,
		"timestamp": time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(response)
}

// setCORSHeaders sets CORS headers
func (h *Handler) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

// PlaygroundHandler handles GraphQL playground requests
func (h *Handler) PlaygroundHandler(w http.ResponseWriter, r *http.Request) {
	// Check if GraphQL is enabled
	if !h.schemaManager.IsEnabled() {
		h.handleDisabled(w, r)
		return
	}

	// Check if playground is enabled
	if !h.schemaManager.GetConfig().EnablePlayground {
		http.Error(w, "GraphQL Playground is disabled", http.StatusForbidden)
		return
	}

	// Only allow GET requests for playground
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Serve playground HTML
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(h.schemaManager.GetPlaygroundHTML()))
}

// IntrospectionHandler handles GraphQL introspection requests
func (h *Handler) IntrospectionHandler(w http.ResponseWriter, r *http.Request) {
	// Check if GraphQL is enabled
	if !h.schemaManager.IsEnabled() {
		h.handleDisabled(w, r)
		return
	}

	// Check if introspection is enabled
	if !h.schemaManager.GetConfig().EnableIntrospection {
		http.Error(w, "GraphQL Introspection is disabled", http.StatusForbidden)
		return
	}

	// Only allow POST requests for introspection
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Execute introspection query
	ctx := r.Context()
	introspectionQuery := h.schemaManager.GetIntrospectionQuery()
	result := h.schemaManager.Execute(ctx, introspectionQuery, nil)

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Write response
	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.Error("Failed to encode introspection response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HealthHandler handles GraphQL health check requests
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := "disabled"
	code := 503
	message := "GraphQL endpoint is disabled"

	if h.schemaManager.IsEnabled() {
		status = "enabled"
		code = 200
		message = "GraphQL endpoint is enabled"
	}

	response := map[string]interface{}{
		"status":                status,
		"message":               message,
		"code":                  code,
		"enabled":               h.schemaManager.IsEnabled(),
		"playground_enabled":    h.schemaManager.GetConfig().EnablePlayground,
		"introspection_enabled": h.schemaManager.GetConfig().EnableIntrospection,
		"timestamp":             time.Now().Unix(),
	}

	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// StatusHandler handles GraphQL status requests
func (h *Handler) StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := h.schemaManager.GetMetrics()
	metrics["enabled"] = h.schemaManager.IsEnabled()
	metrics["timestamp"] = time.Now().Unix()

	json.NewEncoder(w).Encode(metrics)
}
