package loadshedding

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Middleware represents HTTP middleware for load shedding
type Middleware struct {
	shedder *LoadShedder
	logger  *zap.Logger

	// Response customization
	errorResponse    []byte
	errorStatusCode  int
	errorContentType string

	// Metrics
	metrics *Metrics
}

// MiddlewareConfig represents middleware configuration
type MiddlewareConfig struct {
	ErrorResponse    []byte `yaml:"error_response" json:"error_response"`
	ErrorStatusCode  int    `yaml:"error_status_code" json:"error_status_code"`
	ErrorContentType string `yaml:"error_content_type" json:"error_content_type"`
}

// DefaultMiddlewareConfig returns default middleware configuration
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		ErrorResponse:    []byte(`{"error":"Service temporarily unavailable","code":"LOAD_SHEDDING"}`),
		ErrorStatusCode:  http.StatusServiceUnavailable,
		ErrorContentType: "application/json",
	}
}

// NewMiddleware creates a new load shedding middleware
func NewMiddleware(shedder *LoadShedder, config *MiddlewareConfig, logger *zap.Logger) *Middleware {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}

	return &Middleware{
		shedder:          shedder,
		logger:           logger,
		errorResponse:    config.ErrorResponse,
		errorStatusCode:  config.ErrorStatusCode,
		errorContentType: config.ErrorContentType,
		metrics:          NewMetrics("middleware"),
	}
}

// Handler returns the HTTP middleware handler
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Check if request should be shed
		if m.shedder.ShouldShed(r.Context()) {
			m.handleShedRequest(w, r)
			return
		}

		// Process request
		m.processRequest(w, r, next, start)
	})
}

// handleShedRequest handles a request that should be shed
func (m *Middleware) handleShedRequest(w http.ResponseWriter, r *http.Request) {
	// Record shed request
	m.metrics.RecordRequest(true)

	// Set response headers
	w.Header().Set("Content-Type", m.errorContentType)
	w.Header().Set("X-Load-Shedding", "true")
	w.Header().Set("X-Shedding-Level", m.shedder.GetCurrentLevel().String())
	w.Header().Set("X-Shedding-Rate", fmt.Sprintf("%.2f", m.shedder.GetCurrentShedRate()))

	// Write error response
	w.WriteHeader(m.errorStatusCode)
	w.Write(m.errorResponse)

	// Log shed request
	if m.logger != nil {
		m.logger.Warn("Request shed due to load",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("level", m.shedder.GetCurrentLevel().String()),
			zap.Float64("rate", m.shedder.GetCurrentShedRate()))
	}
}

// processRequest processes a request that should not be shed
func (m *Middleware) processRequest(w http.ResponseWriter, r *http.Request, next http.Handler, start time.Time) {
	// Record processed request
	m.metrics.RecordRequest(false)

	// Wrap response writer to capture metrics
	wrapped := &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	// Process request
	next.ServeHTTP(wrapped, r)

	// Record response time
	duration := time.Since(start)
	m.metrics.RecordResponseTime(duration)

	// Add shedding headers
	wrapped.Header().Set("X-Load-Shedding", "false")
	wrapped.Header().Set("X-Shedding-Level", m.shedder.GetCurrentLevel().String())
	wrapped.Header().Set("X-Shedding-Rate", fmt.Sprintf("%.2f", m.shedder.GetCurrentShedRate()))
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetStats returns middleware statistics
func (m *Middleware) GetStats() MetricsStats {
	return m.metrics.GetStats()
}

// LoadSheddingManager manages multiple load shedding middlewares
type LoadSheddingManager struct {
	shedders    map[string]*LoadShedder
	middlewares map[string]*Middleware
	metrics     *MetricsCollector
	logger      *zap.Logger
	mu          sync.RWMutex
}

// NewLoadSheddingManager creates a new load shedding manager
func NewLoadSheddingManager(logger *zap.Logger) *LoadSheddingManager {
	return &LoadSheddingManager{
		shedders:    make(map[string]*LoadShedder),
		middlewares: make(map[string]*Middleware),
		metrics:     NewMetricsCollector(logger),
		logger:      logger,
	}
}

// CreateShedder creates a new load shedder
func (lsm *LoadSheddingManager) CreateShedder(name string, config *Config) (*LoadShedder, error) {
	if name == "" {
		return nil, fmt.Errorf("shedder name cannot be empty")
	}

	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	// Check if shedder already exists
	if _, exists := lsm.shedders[name]; exists {
		return nil, fmt.Errorf("load shedder %s already exists", name)
	}

	// Create shedder
	shedder := NewLoadShedder(config, lsm.logger)
	lsm.shedders[name] = shedder

	// Register with metrics collector
	lsm.metrics.RegisterShedder(name, shedder.GetMetrics())

	if lsm.logger != nil {
		lsm.logger.Info("Load shedder created",
			zap.String("shedder", name),
			zap.String("strategy", config.Strategy.String()))
	}

	return shedder, nil
}

// CreateMiddleware creates a new middleware for a shedder
func (lsm *LoadSheddingManager) CreateMiddleware(name string, shedderName string, config *MiddlewareConfig) (*Middleware, error) {
	if name == "" {
		return nil, fmt.Errorf("middleware name cannot be empty")
	}

	lsm.mu.RLock()
	shedder, exists := lsm.shedders[shedderName]
	lsm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("load shedder %s not found", shedderName)
	}

	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	// Check if middleware already exists
	if _, exists := lsm.middlewares[name]; exists {
		return nil, fmt.Errorf("middleware %s already exists", name)
	}

	// Create middleware
	middleware := NewMiddleware(shedder, config, lsm.logger)
	lsm.middlewares[name] = middleware

	if lsm.logger != nil {
		lsm.logger.Info("Load shedding middleware created",
			zap.String("middleware", name),
			zap.String("shedder", shedderName))
	}

	return middleware, nil
}

// GetShedder returns a shedder by name
func (lsm *LoadSheddingManager) GetShedder(name string) (*LoadShedder, bool) {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	shedder, exists := lsm.shedders[name]
	return shedder, exists
}

// GetMiddleware returns a middleware by name
func (lsm *LoadSheddingManager) GetMiddleware(name string) (*Middleware, bool) {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	middleware, exists := lsm.middlewares[name]
	return middleware, exists
}

// RemoveShedder removes a shedder
func (lsm *LoadSheddingManager) RemoveShedder(name string) error {
	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	shedder, exists := lsm.shedders[name]
	if !exists {
		return fmt.Errorf("load shedder %s not found", name)
	}

	// Stop the shedder
	shedder.Stop()

	// Unregister from metrics
	lsm.metrics.UnregisterShedder(name)

	// Remove shedder
	delete(lsm.shedders, name)

	// Remove associated middlewares
	for middlewareName, middleware := range lsm.middlewares {
		if middleware.shedder == shedder {
			delete(lsm.middlewares, middlewareName)
		}
	}

	if lsm.logger != nil {
		lsm.logger.Info("Load shedder removed",
			zap.String("shedder", name))
	}

	return nil
}

// GetShedderNames returns all shedder names
func (lsm *LoadSheddingManager) GetShedderNames() []string {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	names := make([]string, 0, len(lsm.shedders))
	for name := range lsm.shedders {
		names = append(names, name)
	}
	return names
}

// GetMiddlewareNames returns all middleware names
func (lsm *LoadSheddingManager) GetMiddlewareNames() []string {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	names := make([]string, 0, len(lsm.middlewares))
	for name := range lsm.middlewares {
		names = append(names, name)
	}
	return names
}

// GetAllStats returns statistics for all shedders
func (lsm *LoadSheddingManager) GetAllStats() map[string]Stats {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	stats := make(map[string]Stats)
	for name, shedder := range lsm.shedders {
		stats[name] = shedder.GetStats()
	}
	return stats
}

// GetAggregatedStats returns aggregated statistics
func (lsm *LoadSheddingManager) GetAggregatedStats() AggregatedStats {
	return lsm.metrics.GetAggregatedStats()
}

// ResetAll resets all shedders
func (lsm *LoadSheddingManager) ResetAll() {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	for name, shedder := range lsm.shedders {
		shedder.Reset()

		if lsm.logger != nil {
			lsm.logger.Info("Load shedder reset",
				zap.String("shedder", name))
		}
	}

	// Reset metrics
	lsm.metrics.ResetAll()
}

// Stop stops all shedders
func (lsm *LoadSheddingManager) Stop() {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	for name, shedder := range lsm.shedders {
		shedder.Stop()

		if lsm.logger != nil {
			lsm.logger.Info("Load shedder stopped",
				zap.String("shedder", name))
		}
	}
}

// GetManagerStats returns manager statistics
func (lsm *LoadSheddingManager) GetManagerStats() ManagerStats {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	shedderCount := len(lsm.shedders)
	middlewareCount := len(lsm.middlewares)

	// Count shedders by level
	levelCounts := make(map[SheddingLevel]int)
	for _, shedder := range lsm.shedders {
		level := shedder.GetCurrentLevel()
		levelCounts[level]++
	}

	return ManagerStats{
		ShedderCount:    shedderCount,
		MiddlewareCount: middlewareCount,
		LevelCounts:     levelCounts,
	}
}

// ManagerStats represents manager statistics
type ManagerStats struct {
	ShedderCount    int                   `json:"shedder_count"`
	MiddlewareCount int                   `json:"middleware_count"`
	LevelCounts     map[SheddingLevel]int `json:"level_counts"`
}
