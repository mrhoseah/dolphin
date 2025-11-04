package performance

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Middleware records performance metrics for requests
func Middleware(monitor *Monitor, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			// Record response time
			duration := time.Since(start).Seconds() * 1000 // Convert to milliseconds
			monitor.Record("response_time", duration, r.URL.Path)

			// Record status code
			monitor.Record("status_code", float64(rw.statusCode), r.URL.Path)

			logger.Debug("Request processed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Float64("duration_ms", duration),
				zap.Int("status", rw.statusCode),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

