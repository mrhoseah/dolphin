package observability

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// MetricsCollector manages application metrics
type MetricsCollector struct {
	// HTTP metrics
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpRequestSize     *prometheus.HistogramVec
	httpResponseSize    *prometheus.HistogramVec
	httpActiveRequests  prometheus.Gauge

	// Application metrics
	appUptime          prometheus.Counter
	appMemoryUsage     prometheus.Gauge
	appGoroutineCount  prometheus.Gauge
	appGCPauseDuration *prometheus.HistogramVec

	// Database metrics
	dbConnectionsActive prometheus.Gauge
	dbConnectionsIdle   prometheus.Gauge
	dbQueryDuration     *prometheus.HistogramVec
	dbQueryErrors       *prometheus.CounterVec

	// Cache metrics
	cacheHits       *prometheus.CounterVec
	cacheMisses     *prometheus.CounterVec
	cacheOperations *prometheus.CounterVec
	cacheSize       *prometheus.GaugeVec

	// Business metrics
	businessEvents    *prometheus.CounterVec
	userRegistrations prometheus.Counter
	userLogins        prometheus.Counter
	apiCalls          *prometheus.CounterVec

	// Custom metrics
	customCounters   map[string]*prometheus.CounterVec
	customGauges     map[string]*prometheus.GaugeVec
	customHistograms map[string]*prometheus.HistogramVec

	// Internal state
	startTime time.Time
	mu        sync.RWMutex
	logger    *zap.Logger
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled              bool              `yaml:"enabled" json:"enabled"`
	Namespace            string            `yaml:"namespace" json:"namespace"`
	Subsystem            string            `yaml:"subsystem" json:"subsystem"`
	Path                 string            `yaml:"path" json:"path"`
	Port                 int               `yaml:"port" json:"port"`
	Labels               map[string]string `yaml:"labels" json:"labels"`
	Buckets              []float64         `yaml:"buckets" json:"buckets"`
	EnableGoMetrics      bool              `yaml:"enable_go_metrics" json:"enable_go_metrics"`
	EnableProcessMetrics bool              `yaml:"enable_process_metrics" json:"enable_process_metrics"`
}

// DefaultMetricsConfig returns default metrics configuration
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Enabled:              true,
		Namespace:            "dolphin",
		Subsystem:            "app",
		Path:                 "/metrics",
		Port:                 9090,
		Labels:               make(map[string]string),
		Buckets:              prometheus.DefBuckets,
		EnableGoMetrics:      true,
		EnableProcessMetrics: true,
	}
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(config *MetricsConfig, logger *zap.Logger) *MetricsCollector {
	if config == nil {
		config = DefaultMetricsConfig()
	}

	mc := &MetricsCollector{
		startTime:        time.Now(),
		logger:           logger,
		customCounters:   make(map[string]*prometheus.CounterVec),
		customGauges:     make(map[string]*prometheus.GaugeVec),
		customHistograms: make(map[string]*prometheus.HistogramVec),
	}

	// Initialize HTTP metrics
	mc.httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code", "handler"},
	)

	mc.httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   config.Buckets,
		},
		[]string{"method", "path", "status_code", "handler"},
	)

	mc.httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "http_request_size_bytes",
			Help:      "HTTP request size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path", "handler"},
	)

	mc.httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "http_response_size_bytes",
			Help:      "HTTP response size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path", "status_code", "handler"},
	)

	mc.httpActiveRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "http_active_requests",
			Help:      "Number of active HTTP requests",
		},
	)

	// Initialize application metrics
	mc.appUptime = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "uptime_seconds_total",
			Help:      "Application uptime in seconds",
		},
	)

	mc.appMemoryUsage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "memory_usage_bytes",
			Help:      "Current memory usage in bytes",
		},
	)

	mc.appGoroutineCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "goroutines_total",
			Help:      "Current number of goroutines",
		},
	)

	mc.appGCPauseDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "gc_pause_duration_seconds",
			Help:      "GC pause duration in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 15),
		},
		[]string{"gc_type"},
	)

	// Initialize database metrics
	mc.dbConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: "database",
			Name:      "connections_active",
			Help:      "Number of active database connections",
		},
	)

	mc.dbConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: "database",
			Name:      "connections_idle",
			Help:      "Number of idle database connections",
		},
	)

	mc.dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: "database",
			Name:      "query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   config.Buckets,
		},
		[]string{"operation", "table"},
	)

	mc.dbQueryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: "database",
			Name:      "query_errors_total",
			Help:      "Total number of database query errors",
		},
		[]string{"operation", "table", "error_type"},
	)

	// Initialize cache metrics
	mc.cacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: "cache",
			Name:      "hits_total",
			Help:      "Total number of cache hits",
		},
		[]string{"cache_name", "key_pattern"},
	)

	mc.cacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: "cache",
			Name:      "misses_total",
			Help:      "Total number of cache misses",
		},
		[]string{"cache_name", "key_pattern"},
	)

	mc.cacheOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: "cache",
			Name:      "operations_total",
			Help:      "Total number of cache operations",
		},
		[]string{"cache_name", "operation", "status"},
	)

	mc.cacheSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: "cache",
			Name:      "size_bytes",
			Help:      "Cache size in bytes",
		},
		[]string{"cache_name"},
	)

	// Initialize business metrics
	mc.businessEvents = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: "business",
			Name:      "events_total",
			Help:      "Total number of business events",
		},
		[]string{"event_type", "status"},
	)

	mc.userRegistrations = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: "business",
			Name:      "user_registrations_total",
			Help:      "Total number of user registrations",
		},
	)

	mc.userLogins = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: "business",
			Name:      "user_logins_total",
			Help:      "Total number of user logins",
		},
	)

	mc.apiCalls = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: "api",
			Name:      "calls_total",
			Help:      "Total number of API calls",
		},
		[]string{"endpoint", "method", "status"},
	)

	// Start background metrics collection
	go mc.collectSystemMetrics()

	return mc
}

// HTTPMetricsMiddleware creates middleware for HTTP metrics
func (mc *MetricsCollector) HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Increment active requests
		mc.httpActiveRequests.Inc()
		defer mc.httpActiveRequests.Dec()

		// Wrap response writer to capture status code and size
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Record request size
		requestSize := r.ContentLength
		if requestSize < 0 {
			requestSize = 0
		}

		// Get handler name
		handler := "unknown"
		if h := r.Context().Value("handler"); h != nil {
			if handlerName, ok := h.(string); ok {
				handler = handlerName
			}
		}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := fmt.Sprintf("%d", wrapped.statusCode)

		mc.httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, statusCode, handler).Inc()
		mc.httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, statusCode, handler).Observe(duration)
		mc.httpRequestSize.WithLabelValues(r.Method, r.URL.Path, handler).Observe(float64(requestSize))
		mc.httpResponseSize.WithLabelValues(r.Method, r.URL.Path, statusCode, handler).Observe(float64(wrapped.size))
	})
}

// RecordDatabaseQuery records database query metrics
func (mc *MetricsCollector) RecordDatabaseQuery(operation, table string, duration time.Duration, err error) {
	mc.dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())

	if err != nil {
		errorType := "unknown"
		if err != nil {
			errorType = "query_error"
		}
		mc.dbQueryErrors.WithLabelValues(operation, table, errorType).Inc()
	}
}

// RecordCacheOperation records cache operation metrics
func (mc *MetricsCollector) RecordCacheOperation(cacheName, operation, keyPattern, status string) {
	mc.cacheOperations.WithLabelValues(cacheName, operation, status).Inc()
}

// RecordCacheHit records a cache hit
func (mc *MetricsCollector) RecordCacheHit(cacheName, keyPattern string) {
	mc.cacheHits.WithLabelValues(cacheName, keyPattern).Inc()
}

// RecordCacheMiss records a cache miss
func (mc *MetricsCollector) RecordCacheMiss(cacheName, keyPattern string) {
	mc.cacheMisses.WithLabelValues(cacheName, keyPattern).Inc()
}

// RecordBusinessEvent records a business event
func (mc *MetricsCollector) RecordBusinessEvent(eventType, status string) {
	mc.businessEvents.WithLabelValues(eventType, status).Inc()
}

// RecordUserRegistration records a user registration
func (mc *MetricsCollector) RecordUserRegistration() {
	mc.userRegistrations.Inc()
}

// RecordUserLogin records a user login
func (mc *MetricsCollector) RecordUserLogin() {
	mc.userLogins.Inc()
}

// RecordAPICall records an API call
func (mc *MetricsCollector) RecordAPICall(endpoint, method, status string) {
	mc.apiCalls.WithLabelValues(endpoint, method, status).Inc()
}

// SetDatabaseConnections sets database connection metrics
func (mc *MetricsCollector) SetDatabaseConnections(active, idle int) {
	mc.dbConnectionsActive.Set(float64(active))
	mc.dbConnectionsIdle.Set(float64(idle))
}

// SetCacheSize sets cache size metric
func (mc *MetricsCollector) SetCacheSize(cacheName string, size int64) {
	mc.cacheSize.WithLabelValues(cacheName).Set(float64(size))
}

// CreateCustomCounter creates a custom counter metric
func (mc *MetricsCollector) CreateCustomCounter(name, help string, labels []string) *prometheus.CounterVec {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if counter, exists := mc.customCounters[name]; exists {
		return counter
	}

	counter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "dolphin",
			Subsystem: "custom",
			Name:      name,
			Help:      help,
		},
		labels,
	)

	mc.customCounters[name] = counter
	return counter
}

// CreateCustomGauge creates a custom gauge metric
func (mc *MetricsCollector) CreateCustomGauge(name, help string, labels []string) *prometheus.GaugeVec {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if gauge, exists := mc.customGauges[name]; exists {
		return gauge
	}

	gauge := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "dolphin",
			Subsystem: "custom",
			Name:      name,
			Help:      help,
		},
		labels,
	)

	mc.customGauges[name] = gauge
	return gauge
}

// CreateCustomHistogram creates a custom histogram metric
func (mc *MetricsCollector) CreateCustomHistogram(name, help string, buckets []float64, labels []string) *prometheus.HistogramVec {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if histogram, exists := mc.customHistograms[name]; exists {
		return histogram
	}

	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	histogram := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "dolphin",
			Subsystem: "custom",
			Name:      name,
			Help:      help,
			Buckets:   buckets,
		},
		labels,
	)

	mc.customHistograms[name] = histogram
	return histogram
}

// GetMetricsHandler returns the Prometheus metrics handler
func (mc *MetricsCollector) GetMetricsHandler() http.Handler {
	return promhttp.Handler()
}

// StartMetricsServer starts the metrics server
func (mc *MetricsCollector) StartMetricsServer(config *MetricsConfig) error {
	if !config.Enabled {
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle(config.Path, mc.GetMetricsHandler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: mux,
	}

	go func() {
		mc.logger.Info("Starting metrics server",
			zap.String("addr", server.Addr),
			zap.String("path", config.Path))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mc.logger.Error("Metrics server error", zap.Error(err))
		}
	}()

	return nil
}

// MetricsConfigFromEnv creates metrics config from environment variables
func MetricsConfigFromEnv() *MetricsConfig {
	config := DefaultMetricsConfig()

	if enabled := os.Getenv("METRICS_ENABLED"); enabled == "false" {
		config.Enabled = false
	}
	if namespace := os.Getenv("METRICS_NAMESPACE"); namespace != "" {
		config.Namespace = namespace
	}
	if subsystem := os.Getenv("METRICS_SUBSYSTEM"); subsystem != "" {
		config.Subsystem = subsystem
	}
	if path := os.Getenv("METRICS_PATH"); path != "" {
		config.Path = path
	}
	if port := os.Getenv("METRICS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Port = p
		}
	}

	return config
}

// collectSystemMetrics collects system metrics in the background
func (mc *MetricsCollector) collectSystemMetrics() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Update uptime
		mc.appUptime.Add(time.Since(mc.startTime).Seconds())
		mc.startTime = time.Now()

		// Update memory usage
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		mc.appMemoryUsage.Set(float64(m.Alloc))

		// Update goroutine count
		mc.appGoroutineCount.Set(float64(runtime.NumGoroutine()))

		// Update GC metrics
		mc.appGCPauseDuration.WithLabelValues("GC").Observe(float64(m.PauseNs[(m.NumGC+255)%256]) / 1e9)
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += int64(size)
	return size, err
}

// MetricsSummary represents a summary of metrics
type MetricsSummary struct {
	HTTPRequests    int64   `json:"http_requests"`
	HTTPDuration    float64 `json:"http_duration_avg"`
	MemoryUsage     int64   `json:"memory_usage_bytes"`
	GoroutineCount  int     `json:"goroutine_count"`
	DatabaseQueries int64   `json:"database_queries"`
	CacheHits       int64   `json:"cache_hits"`
	CacheMisses     int64   `json:"cache_misses"`
	Uptime          float64 `json:"uptime_seconds"`
}

// GetSummary returns a summary of current metrics
func (mc *MetricsCollector) GetSummary() *MetricsSummary {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MetricsSummary{
		HTTPRequests:    0, // Would need to track this separately
		HTTPDuration:    0, // Would need to track this separately
		MemoryUsage:     int64(m.Alloc),
		GoroutineCount:  runtime.NumGoroutine(),
		DatabaseQueries: 0, // Would need to track this separately
		CacheHits:       0, // Would need to track this separately
		CacheMisses:     0, // Would need to track this separately
		Uptime:          time.Since(mc.startTime).Seconds(),
	}
}
