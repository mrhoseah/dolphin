package router

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// MetricsCollector collects API metrics
type MetricsCollector struct {
	logger        *zap.Logger
	requestCount  int64
	errorCount    int64
	responseTimes []time.Duration
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		logger:        logger,
		responseTimes: make([]time.Duration, 0),
	}
}

// MetricsMiddleware creates middleware for collecting API metrics
func MetricsMiddleware(collector *MetricsCollector) func(next http.Handler) http.Handler {
	if collector == nil {
		collector = NewMetricsCollector(nil)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Wrap response writer to capture status code
			mw := &metricsWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(mw, r)

			// Calculate metrics
			duration := time.Since(start)
			collector.requestCount++
			
			if mw.statusCode >= 400 {
				collector.errorCount++
			}

			// Log metrics
			if collector.logger != nil {
				collector.logger.Info("Request metrics",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int("status", mw.statusCode),
					zap.Duration("duration", duration),
					zap.String("ip", GetIPAddress(r)),
				)
			}

			// Store response time (keep last 100)
			collector.responseTimes = append(collector.responseTimes, duration)
			if len(collector.responseTimes) > 100 {
				collector.responseTimes = collector.responseTimes[1:]
			}
		})
	}
}

// metricsWriter captures status code for metrics
type metricsWriter struct {
	http.ResponseWriter
	statusCode int
}

func (mw *metricsWriter) WriteHeader(code int) {
	mw.statusCode = code
	mw.ResponseWriter.WriteHeader(code)
}

// GetMetrics returns current metrics
func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	var avgResponseTime time.Duration
	if len(mc.responseTimes) > 0 {
		var total time.Duration
		for _, rt := range mc.responseTimes {
			total += rt
		}
		avgResponseTime = total / time.Duration(len(mc.responseTimes))
	}

	return map[string]interface{}{
		"request_count":     mc.requestCount,
		"error_count":       mc.errorCount,
		"success_rate":      float64(mc.requestCount-mc.errorCount) / float64(mc.requestCount) * 100,
		"avg_response_time": avgResponseTime.Milliseconds(),
	}
}

