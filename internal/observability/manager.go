package observability

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ObservabilityManager manages all observability features
type ObservabilityManager struct {
	metrics *MetricsCollector
	logging *LoggerManager
	tracing *TracerManager
	config  *ObservabilityConfig
	logger  *zap.Logger
}

// ObservabilityConfig represents the overall observability configuration
type ObservabilityConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Metrics configuration
	Metrics *MetricsConfig `yaml:"metrics" json:"metrics"`

	// Logging configuration
	Logging *LogConfig `yaml:"logging" json:"logging"`

	// Tracing configuration
	Tracing *TraceConfig `yaml:"tracing" json:"tracing"`

	// Health check configuration
	HealthCheck *HealthCheckConfig `yaml:"health_check" json:"health_check"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Enabled  bool          `yaml:"enabled" json:"enabled"`
	Path     string        `yaml:"path" json:"path"`
	Port     int           `yaml:"port" json:"port"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
	Interval time.Duration `yaml:"interval" json:"interval"`
	Checks   []string      `yaml:"checks" json:"checks"`
}

// DefaultObservabilityConfig returns default observability configuration
func DefaultObservabilityConfig() *ObservabilityConfig {
	return &ObservabilityConfig{
		Enabled: true,
		Metrics: DefaultMetricsConfig(),
		Logging: DefaultLogConfig(),
		Tracing: DefaultTraceConfig(),
		HealthCheck: &HealthCheckConfig{
			Enabled:  true,
			Path:     "/health",
			Port:     8081,
			Timeout:  5 * time.Second,
			Interval: 30 * time.Second,
			Checks:   []string{"database", "cache", "external_apis"},
		},
	}
}

// NewObservabilityManager creates a new observability manager
func NewObservabilityManager(config *ObservabilityConfig, logger *zap.Logger) (*ObservabilityManager, error) {
	if config == nil {
		config = DefaultObservabilityConfig()
	}

	if !config.Enabled {
		return &ObservabilityManager{
			config: config,
			logger: logger,
		}, nil
	}

	om := &ObservabilityManager{
		config: config,
		logger: logger,
	}

	// Initialize metrics
	if config.Metrics != nil && config.Metrics.Enabled {
		metrics := NewMetricsCollector(config.Metrics, logger)
		om.metrics = metrics
	}

	// Initialize logging
	if config.Logging != nil {
		logging, err := NewLoggerManager(config.Logging)
		if err != nil {
			return nil, fmt.Errorf("failed to create logger manager: %w", err)
		}
		om.logging = logging
	}

	// Initialize tracing
	if config.Tracing != nil && config.Tracing.Enabled {
		tracing, err := NewTracerManager(config.Tracing, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create tracer manager: %w", err)
		}
		om.tracing = tracing
	}

	return om, nil
}

// GetMetrics returns the metrics collector
func (om *ObservabilityManager) GetMetrics() *MetricsCollector {
	return om.metrics
}

// GetLogging returns the logger manager
func (om *ObservabilityManager) GetLogging() *LoggerManager {
	return om.logging
}

// GetTracing returns the tracer manager
func (om *ObservabilityManager) GetTracing() *TracerManager {
	return om.tracing
}

// GetLogger returns the main logger
func (om *ObservabilityManager) GetLogger() *zap.Logger {
	if om.logging != nil {
		return om.logging.GetLogger()
	}
	return om.logger
}

// GetSugarLogger returns the sugar logger
func (om *ObservabilityManager) GetSugarLogger() *zap.SugaredLogger {
	if om.logging != nil {
		return om.logging.GetSugarLogger()
	}
	return om.logger.Sugar()
}

// Start starts all observability services
func (om *ObservabilityManager) Start() error {
	if !om.config.Enabled {
		return nil
	}

	// Start metrics server
	if om.metrics != nil {
		if err := om.metrics.StartMetricsServer(om.config.Metrics); err != nil {
			return fmt.Errorf("failed to start metrics server: %w", err)
		}
	}

	// Start health check server
	if om.config.HealthCheck != nil && om.config.HealthCheck.Enabled {
		if err := om.startHealthCheckServer(); err != nil {
			return fmt.Errorf("failed to start health check server: %w", err)
		}
	}

	return nil
}

// Stop stops all observability services
func (om *ObservabilityManager) Stop(ctx context.Context) error {
	if !om.config.Enabled {
		return nil
	}

	// Stop tracing provider
	if om.tracing != nil && om.tracing.provider != nil {
		if err := om.tracing.provider.Shutdown(ctx); err != nil {
			om.logger.Error("Failed to shutdown tracer provider", zap.Error(err))
		}
	}

	return nil
}

// GetHTTPMiddlewares returns all HTTP middlewares
func (om *ObservabilityManager) GetHTTPMiddlewares() []func(http.Handler) http.Handler {
	var middlewares []func(http.Handler) http.Handler

	// Add tracing middleware first
	if om.tracing != nil {
		middlewares = append(middlewares, TracingMiddleware(om.tracing))
	}

	// Add logging middleware
	if om.logging != nil {
		middlewares = append(middlewares, LoggingMiddleware(om.logging))
	}

	// Add metrics middleware
	if om.metrics != nil {
		middlewares = append(middlewares, om.metrics.HTTPMetricsMiddleware)
	}

	return middlewares
}

// LogHTTPRequest logs an HTTP request with all observability features
func (om *ObservabilityManager) LogHTTPRequest(r *http.Request, statusCode int, duration time.Duration, size int64) {
	// Log with structured logging
	if om.logging != nil {
		om.logging.LogHTTPRequest(r, statusCode, duration, size)
	}

	// Record metrics
	if om.metrics != nil {
		// This would be handled by the metrics middleware
		// but we can add additional business metrics here
	}
}

// LogDatabaseQuery logs a database query with all observability features
func (om *ObservabilityManager) LogDatabaseQuery(operation, table string, duration time.Duration, err error) {
	// Log with structured logging
	if om.logging != nil {
		om.logging.LogDatabaseQuery(operation, table, duration, err)
	}

	// Record metrics
	if om.metrics != nil {
		om.metrics.RecordDatabaseQuery(operation, table, duration, err)
	}
}

// LogCacheOperation logs a cache operation with all observability features
func (om *ObservabilityManager) LogCacheOperation(operation, cacheName, key string, hit bool, err error) {
	// Log with structured logging
	if om.logging != nil {
		om.logging.LogCacheOperation(operation, cacheName, key, hit, err)
	}

	// Record metrics
	if om.metrics != nil {
		status := "success"
		if err != nil {
			status = "error"
		}
		om.metrics.RecordCacheOperation(cacheName, operation, key, status)

		if hit {
			om.metrics.RecordCacheHit(cacheName, key)
		} else {
			om.metrics.RecordCacheMiss(cacheName, key)
		}
	}
}

// LogBusinessEvent logs a business event with all observability features
func (om *ObservabilityManager) LogBusinessEvent(eventType, status string, data map[string]interface{}) {
	// Log with structured logging
	if om.logging != nil {
		om.logging.LogBusinessEvent(eventType, status, data)
	}

	// Record metrics
	if om.metrics != nil {
		om.metrics.RecordBusinessEvent(eventType, status)
	}
}

// LogError logs an error with all observability features
func (om *ObservabilityManager) LogError(err error, message string, fields ...zap.Field) {
	// Log with structured logging
	if om.logging != nil {
		om.logging.LogError(err, message, fields...)
	}

	// Add error metrics if needed
	if om.metrics != nil {
		// Could add error counter metrics here
	}
}

// LogSecurityEvent logs a security event with all observability features
func (om *ObservabilityManager) LogSecurityEvent(eventType, severity string, data map[string]interface{}) {
	// Log with structured logging
	if om.logging != nil {
		om.logging.LogSecurityEvent(eventType, severity, data)
	}

	// Record security metrics
	if om.metrics != nil {
		om.metrics.RecordBusinessEvent("security_"+eventType, severity)
	}
}

// LogAudit logs an audit event with all observability features
func (om *ObservabilityManager) LogAudit(action, resource, userID string, success bool, details map[string]interface{}) {
	// Log with structured logging
	if om.logging != nil {
		om.logging.LogAudit(action, resource, userID, success, details)
	}

	// Record audit metrics
	if om.metrics != nil {
		status := "success"
		if !success {
			status = "failure"
		}
		om.metrics.RecordBusinessEvent("audit_"+action, status)
	}
}

// LogPerformance logs performance metrics with all observability features
func (om *ObservabilityManager) LogPerformance(operation string, duration time.Duration, metrics map[string]interface{}) {
	// Log with structured logging
	if om.logging != nil {
		om.logging.LogPerformance(operation, duration, metrics)
	}

	// Record performance metrics
	if om.metrics != nil {
		// Could add custom performance metrics here
	}
}

// StartSpan starts a new span with tracing
func (om *ObservabilityManager) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if om.tracing != nil {
		return om.tracing.StartSpan(ctx, name, opts...)
	}
	return ctx, trace.SpanFromContext(ctx)
}

// StartSpanWithAttributes starts a span with attributes
func (om *ObservabilityManager) StartSpanWithAttributes(ctx context.Context, name string, attrs map[string]interface{}, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if om.tracing != nil {
		return om.tracing.StartSpanWithAttributes(ctx, name, attrs, opts...)
	}
	return ctx, trace.SpanFromContext(ctx)
}

// AddSpanEvent adds an event to the current span
func (om *ObservabilityManager) AddSpanEvent(ctx context.Context, name string, attrs map[string]interface{}) {
	if om.tracing != nil {
		om.tracing.AddSpanEvent(ctx, name, attrs)
	}
}

// SetSpanAttributes sets attributes on the current span
func (om *ObservabilityManager) SetSpanAttributes(ctx context.Context, attrs map[string]interface{}) {
	if om.tracing != nil {
		om.tracing.SetSpanAttributes(ctx, attrs)
	}
}

// SetSpanStatus sets the status of the current span
func (om *ObservabilityManager) SetSpanStatus(ctx context.Context, code codes.Code, description string) {
	if om.tracing != nil {
		om.tracing.SetSpanStatus(ctx, code, description)
	}
}

// FinishSpan finishes the current span
func (om *ObservabilityManager) FinishSpan(ctx context.Context) {
	if om.tracing != nil {
		om.tracing.FinishSpan(ctx)
	}
}

// GetTraceID returns the trace ID from context
func (om *ObservabilityManager) GetTraceID(ctx context.Context) string {
	if om.tracing != nil {
		return om.tracing.GetTraceID(ctx)
	}
	return ""
}

// GetSpanID returns the span ID from context
func (om *ObservabilityManager) GetSpanID(ctx context.Context) string {
	if om.tracing != nil {
		return om.tracing.GetSpanID(ctx)
	}
	return ""
}

// startHealthCheckServer starts the health check server
func (om *ObservabilityManager) startHealthCheckServer() error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc(om.config.HealthCheck.Path, om.healthCheckHandler)

	// Readiness check endpoint
	mux.HandleFunc(om.config.HealthCheck.Path+"/ready", om.readinessCheckHandler)

	// Liveness check endpoint
	mux.HandleFunc(om.config.HealthCheck.Path+"/live", om.livenessCheckHandler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", om.config.HealthCheck.Port),
		Handler: mux,
	}

	go func() {
		om.logger.Info("Starting health check server",
			zap.String("addr", server.Addr),
			zap.String("path", om.config.HealthCheck.Path))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			om.logger.Error("Health check server error", zap.Error(err))
		}
	}()

	return nil
}

// healthCheckHandler handles health check requests
func (om *ObservabilityManager) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	status := "healthy"
	statusCode := http.StatusOK

	// Check various components
	checks := make(map[string]string)

	// Check metrics
	if om.metrics != nil {
		checks["metrics"] = "ok"
	} else {
		checks["metrics"] = "disabled"
	}

	// Check logging
	if om.logging != nil {
		checks["logging"] = "ok"
	} else {
		checks["logging"] = "disabled"
	}

	// Check tracing
	if om.tracing != nil {
		checks["tracing"] = "ok"
	} else {
		checks["tracing"] = "disabled"
	}

	// Return health status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"status":"%s","checks":%v,"timestamp":"%s"}`,
		status, checks, time.Now().Format(time.RFC3339))
}

// readinessCheckHandler handles readiness check requests
func (om *ObservabilityManager) readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the application is ready to serve traffic
	ready := true
	checks := make(map[string]string)

	// Add readiness checks here
	// For example: database connectivity, external service availability, etc.

	if ready {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ready","checks":%v}`, checks)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"not_ready","checks":%v}`, checks)
	}
}

// livenessCheckHandler handles liveness check requests
func (om *ObservabilityManager) livenessCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the application is alive
	alive := true

	if alive {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"alive"}`)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"status":"dead"}`)
	}
}

// GetSummary returns a summary of observability status
func (om *ObservabilityManager) GetSummary() map[string]interface{} {
	summary := map[string]interface{}{
		"enabled": om.config.Enabled,
		"metrics": om.metrics != nil,
		"logging": om.logging != nil,
		"tracing": om.tracing != nil,
	}

	if om.metrics != nil {
		summary["metrics_summary"] = om.metrics.GetSummary()
	}

	return summary
}

// ObservabilityConfigFromEnv creates observability config from environment variables
func ObservabilityConfigFromEnv() *ObservabilityConfig {
	config := DefaultObservabilityConfig()

	// Override with environment variables
	if enabled := os.Getenv("OBSERVABILITY_ENABLED"); enabled == "false" {
		config.Enabled = false
	}

	// Override metrics config
	config.Metrics = MetricsConfigFromEnv()

	// Override logging config
	config.Logging = LogConfigFromEnv()

	// Override tracing config
	config.Tracing = TraceConfigFromEnv()

	return config
}
