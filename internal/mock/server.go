package mock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MockResponse represents a mock API response
type MockResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Body       interface{}            `json:"body"`
	Delay      int                    `json:"delay_ms,omitempty"` // Delay in milliseconds
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// MockEndpoint represents a mock API endpoint
type MockEndpoint struct {
	Method     string                 `json:"method"`
	Path       string                 `json:"path"`
	Response   MockResponse           `json:"response"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Conditions []Condition            `json:"conditions,omitempty"` // Conditions for dynamic responses
}

// Condition represents a condition for dynamic mock responses
type Condition struct {
	Field    string      `json:"field"`    // Query param, header, or body field
	Operator string      `json:"operator"` // eq, ne, contains, etc.
	Value    interface{} `json:"value"`
	Response MockResponse `json:"response"`
}

// Server manages mock API endpoints
type Server struct {
	endpoints map[string]*MockEndpoint // key: method+path
	mu        sync.RWMutex
	logger    *zap.Logger
	basePath  string
}

// NewServer creates a new mock server
func NewServer(logger *zap.Logger, basePath string) *Server {
	s := &Server{
		endpoints: make(map[string]*MockEndpoint),
		logger:    logger,
		basePath:  basePath,
	}
	s.loadFromDirectory()
	return s
}

// Register registers a mock endpoint
func (s *Server) Register(endpoint *MockEndpoint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(endpoint.Method, endpoint.Path)
	s.endpoints[key] = endpoint
	s.logger.Info("Mock endpoint registered",
		zap.String("method", endpoint.Method),
		zap.String("path", endpoint.Path),
	)
}

// Handle handles HTTP requests with mock responses
func (s *Server) Handle(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	key := s.makeKey(r.Method, r.URL.Path)
	endpoint, exists := s.endpoints[key]
	s.mu.RUnlock()

	if !exists {
		// Try to find a wildcard match
		endpoint = s.findWildcardMatch(r.Method, r.URL.Path)
		if endpoint == nil {
			http.Error(w, "Mock endpoint not found", http.StatusNotFound)
			return
		}
	}

	// Apply delay if specified
	if endpoint.Response.Delay > 0 {
		time.Sleep(time.Duration(endpoint.Response.Delay) * time.Millisecond)
	}

	// Check conditions for dynamic responses
	response := s.selectResponse(endpoint, r)
	if response == nil {
		response = &endpoint.Response
	}

	// Set headers
	for key, value := range response.Headers {
		w.Header().Set(key, value)
	}

	// Set status code
	w.WriteHeader(response.StatusCode)

	// Write body
	if response.Body != nil {
		bodyBytes, err := json.Marshal(response.Body)
		if err != nil {
			s.logger.Error("Failed to marshal mock response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Write(bodyBytes)
	}
}

// selectResponse selects a response based on conditions
func (s *Server) selectResponse(endpoint *MockEndpoint, r *http.Request) *MockResponse {
	for _, condition := range endpoint.Conditions {
		if s.evaluateCondition(condition, r) {
			return &condition.Response
		}
	}
	return nil
}

// evaluateCondition evaluates a condition against the request
func (s *Server) evaluateCondition(condition Condition, r *http.Request) bool {
	// Check query parameters
	if value := r.URL.Query().Get(condition.Field); value != "" {
		return s.compareValues(value, condition.Operator, condition.Value)
	}

	// Check headers
	if value := r.Header.Get(condition.Field); value != "" {
		return s.compareValues(value, condition.Operator, condition.Value)
	}

	// Check body (for POST/PUT requests)
	if r.Body != nil && (r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH") {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			if value, exists := body[condition.Field]; exists {
				return s.compareValues(fmt.Sprintf("%v", value), condition.Operator, condition.Value)
			}
		}
	}

	return false
}

// compareValues compares two values based on operator
func (s *Server) compareValues(value1 string, operator string, value2 interface{}) bool {
	value2Str := fmt.Sprintf("%v", value2)

	switch operator {
	case "eq", "==", "equals":
		return value1 == value2Str
	case "ne", "!=", "not_equals":
		return value1 != value2Str
	case "contains":
		return strings.Contains(value1, value2Str)
	case "starts_with":
		return strings.HasPrefix(value1, value2Str)
	case "ends_with":
		return strings.HasSuffix(value1, value2Str)
	case "gt", ">":
		return value1 > value2Str
	case "lt", "<":
		return value1 < value2Str
	default:
		return false
	}
}

// findWildcardMatch finds a matching endpoint using wildcards
func (s *Server) findWildcardMatch(method, path string) *MockEndpoint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Simple wildcard matching (you can enhance this)
	for key, endpoint := range s.endpoints {
		if strings.HasPrefix(key, method+"|") {
			// Check if path matches with wildcards
			if s.pathMatches(endpoint.Path, path) {
				return endpoint
			}
		}
	}
	return nil
}

// pathMatches checks if a path pattern matches a given path
func (s *Server) pathMatches(pattern, path string) bool {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, patternPart := range patternParts {
		if patternPart == "*" || patternPart == "{*}" {
			continue
		}
		if patternPart != pathParts[i] {
			return false
		}
	}
	return true
}

// makeKey creates a key for endpoint lookup
func (s *Server) makeKey(method, path string) string {
	return method + "|" + path
}

// loadFromDirectory loads mock endpoints from JSON files
func (s *Server) loadFromDirectory() {
	if s.basePath == "" {
		return
	}

	files, err := filepath.Glob(filepath.Join(s.basePath, "*.json"))
	if err != nil {
		s.logger.Error("Failed to read mock directory", zap.Error(err))
		return
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			s.logger.Warn("Failed to read mock file", zap.String("file", file), zap.Error(err))
			continue
		}

		var endpoints []MockEndpoint
		if err := json.Unmarshal(data, &endpoints); err != nil {
			// Try single endpoint
			var endpoint MockEndpoint
			if err := json.Unmarshal(data, &endpoint); err != nil {
				s.logger.Warn("Failed to parse mock file", zap.String("file", file), zap.Error(err))
				continue
			}
			endpoints = []MockEndpoint{endpoint}
		}

		for _, endpoint := range endpoints {
			s.Register(&endpoint)
		}
	}
}

// LoadFromFile loads mock endpoints from a JSON file
func (s *Server) LoadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var endpoints []MockEndpoint
	if err := json.Unmarshal(data, &endpoints); err != nil {
		// Try single endpoint
		var endpoint MockEndpoint
		if err := json.Unmarshal(data, &endpoint); err != nil {
			return err
		}
		endpoints = []MockEndpoint{endpoint}
	}

	for _, endpoint := range endpoints {
		s.Register(&endpoint)
	}

	return nil
}

// GetAll returns all registered endpoints
func (s *Server) GetAll() map[string]*MockEndpoint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*MockEndpoint)
	for key, endpoint := range s.endpoints {
		endpointCopy := *endpoint
		result[key] = &endpointCopy
	}
	return result
}

