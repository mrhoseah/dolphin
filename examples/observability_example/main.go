package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/mrhoseah/dolphin/internal/observability"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()

	// Example 1: Basic Observability Setup
	fmt.Println("=== Example 1: Basic Observability Setup ===")

	// Create observability configuration
	config := observability.DefaultObservabilityConfig()
	config.Metrics.Enabled = true
	config.Logging.Level = observability.InfoLevel
	config.Tracing.Enabled = true

	// Create observability manager
	om, err := observability.NewObservabilityManager(config, logger)
	if err != nil {
		log.Fatalf("Failed to create observability manager: %v", err)
	}

	// Start observability services
	if err := om.Start(); err != nil {
		log.Fatalf("Failed to start observability services: %v", err)
	}
	defer om.Stop(context.Background())

	// Example 2: Structured Logging
	fmt.Println("\n=== Example 2: Structured Logging ===")

	// Get logger
	appLogger := om.GetLogger()

	// Log different types of events
	appLogger.Info("Application started",
		zap.String("version", "1.0.0"),
		zap.String("environment", "development"))

	appLogger.Warn("Configuration value missing, using default",
		zap.String("config_key", "database.timeout"),
		zap.String("default_value", "30s"))

	appLogger.Error("Database connection failed",
		zap.Error(fmt.Errorf("connection timeout")),
		zap.String("database", "postgres"),
		zap.Duration("timeout", 5*time.Second))

	// Log business events
	om.LogBusinessEvent("user_registration", "success", map[string]interface{}{
		"user_id":             "12345",
		"email":               "user@example.com",
		"registration_method": "email",
	})

	// Log security events
	om.LogSecurityEvent("login_attempt", "medium", map[string]interface{}{
		"user_id":    "12345",
		"ip_address": "192.168.1.100",
		"user_agent": "Mozilla/5.0...",
		"success":    true,
	})

	// Log audit events
	om.LogAudit("user_update", "profile", "12345", true, map[string]interface{}{
		"fields_changed": []string{"email", "name"},
		"ip_address":     "192.168.1.100",
	})

	// Example 3: Metrics Collection
	fmt.Println("\n=== Example 3: Metrics Collection ===")

	// Record HTTP metrics (normally done by middleware)
	om.LogHTTPRequest(&http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/api/users"},
		Header: make(http.Header),
	}, 200, 150*time.Millisecond, 1024)

	// Record database metrics
	om.LogDatabaseQuery("SELECT", "users", 25*time.Millisecond, nil)
	om.LogDatabaseQuery("INSERT", "users", 45*time.Millisecond, fmt.Errorf("constraint violation"))

	// Record cache metrics
	om.LogCacheOperation("GET", "redis", "user:123", true, nil)
	om.LogCacheOperation("SET", "redis", "user:123", false, nil)
	om.LogCacheOperation("DELETE", "redis", "user:123", false, fmt.Errorf("key not found"))

	// Record business metrics
	om.RecordUserRegistration()
	om.RecordUserLogin()
	om.RecordAPICall("/api/users", "GET", "200")

	// Example 4: Distributed Tracing
	fmt.Println("\n=== Example 4: Distributed Tracing ===")

	// Start a root span
	ctx, span := om.StartSpan(context.Background(), "user_operation")
	defer om.FinishSpan(ctx)

	// Add span attributes
	om.SetSpanAttributes(ctx, map[string]interface{}{
		"user_id":    "12345",
		"operation":  "profile_update",
		"ip_address": "192.168.1.100",
	})

	// Simulate database operation with tracing
	ctx, dbSpan := om.StartSpan(ctx, "database_query")
	om.SetSpanAttributes(ctx, map[string]interface{}{
		"db.operation": "UPDATE",
		"db.table":     "users",
		"db.query":     "UPDATE users SET name = ? WHERE id = ?",
	})

	// Simulate work
	time.Sleep(10 * time.Millisecond)

	// Add span event
	om.AddSpanEvent(ctx, "query_executed", map[string]interface{}{
		"rows_affected":  1,
		"execution_time": "10ms",
	})

	om.FinishSpan(ctx)

	// Simulate cache operation with tracing
	ctx, cacheSpan := om.StartSpan(ctx, "cache_operation")
	om.SetSpanAttributes(ctx, map[string]interface{}{
		"cache.operation": "SET",
		"cache.key":       "user:123",
		"cache.ttl":       "3600s",
	})

	// Simulate work
	time.Sleep(2 * time.Millisecond)

	om.FinishSpan(ctx)

	// Add final span event
	om.AddSpanEvent(ctx, "operation_completed", map[string]interface{}{
		"total_duration": "12ms",
		"success":        true,
	})

	// Example 5: Performance Monitoring
	fmt.Println("\n=== Example 5: Performance Monitoring ===")

	// Log performance metrics
	om.LogPerformance("user_profile_update", 12*time.Millisecond, map[string]interface{}{
		"database_queries":   2,
		"cache_operations":   1,
		"external_api_calls": 0,
		"memory_usage":       "45.2MB",
	})

	// Example 6: Custom Metrics
	fmt.Println("\n=== Example 6: Custom Metrics ===")

	// Create custom metrics
	if om.GetMetrics() != nil {
		// Custom counter
		userActions := om.GetMetrics().CreateCustomCounter(
			"user_actions_total",
			"Total number of user actions",
			[]string{"action_type", "user_id"},
		)
		userActions.WithLabelValues("profile_update", "12345").Inc()

		// Custom gauge
		activeUsers := om.GetMetrics().CreateCustomGauge(
			"active_users",
			"Number of currently active users",
			[]string{"status"},
		)
		activeUsers.WithLabelValues("online").Set(42)

		// Custom histogram
		requestDuration := om.GetMetrics().CreateCustomHistogram(
			"custom_request_duration_seconds",
			"Custom request duration in seconds",
			nil, // Use default buckets
			[]string{"endpoint", "method"},
		)
		requestDuration.WithLabelValues("/api/users", "GET").Observe(0.150)
	}

	// Example 7: Health Checks
	fmt.Println("\n=== Example 7: Health Checks ===")

	// Get health status
	summary := om.GetSummary()
	fmt.Printf("Observability Summary: %+v\n", summary)

	// Example 8: HTTP Middleware Integration
	fmt.Println("\n=== Example 8: HTTP Middleware Integration ===")

	// Create HTTP server with observability middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get trace IDs for correlation
		traceID := om.GetTraceID(r.Context())
		spanID := om.GetSpanID(r.Context())

		w.Header().Set("X-Trace-Id", traceID)
		w.Header().Set("X-Span-Id", spanID)

		fmt.Fprintf(w, "Hello, World! Trace: %s, Span: %s", traceID, spanID)
	})

	// Apply all observability middlewares
	middlewares := om.GetHTTPMiddlewares()
	handler := mux
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	// Start server in goroutine
	go func() {
		fmt.Println("Starting server on :8080 with observability...")
		if err := http.ListenAndServe(":8080", handler); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	// Test the server
	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		log.Printf("Failed to test server: %v", err)
	} else {
		fmt.Println("Server Response Headers:")
		for name, values := range resp.Header {
			for _, value := range values {
				fmt.Printf("  %s: %s\n", name, value)
			}
		}
		resp.Body.Close()
	}

	// Example 9: Error Handling with Observability
	fmt.Println("\n=== Example 9: Error Handling with Observability ===")

	// Log errors with context
	err = fmt.Errorf("database connection failed: timeout after 5s")
	om.LogError(err, "Failed to connect to database",
		zap.String("database", "postgres"),
		zap.String("host", "localhost"),
		zap.Int("port", 5432),
		zap.Duration("timeout", 5*time.Second))

	// Example 10: Context-Aware Logging
	fmt.Println("\n=== Example 10: Context-Aware Logging ===")

	// Create context with request ID
	ctx = context.WithValue(context.Background(), "request_id", "req-12345")
	ctx = context.WithValue(ctx, "user_id", "user-67890")
	ctx = context.WithValue(ctx, "trace_id", "trace-abcdef")

	// Get context-aware logger
	contextLogger := om.GetLogging().WithContext(ctx)
	contextLogger.Info("Processing request with context",
		zap.String("action", "user_profile_update"),
		zap.String("ip_address", "192.168.1.100"))

	fmt.Println("\n🎉 All observability examples completed successfully!")
	fmt.Println("\n📚 Next Steps:")
	fmt.Println("1. Use 'dolphin observability metrics status' to view metrics")
	fmt.Println("2. Use 'dolphin observability logging test' to test logging")
	fmt.Println("3. Use 'dolphin observability tracing test' to test tracing")
	fmt.Println("4. Use 'dolphin observability health check' to run health checks")
	fmt.Println("5. Integrate ObservabilityManager in your application")
	fmt.Println("6. View metrics in Prometheus: http://localhost:9090/metrics")
	fmt.Println("7. View traces in Jaeger: http://localhost:16686")
}
