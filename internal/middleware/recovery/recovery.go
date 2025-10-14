package middleware

import (
	"encoding/json"
	"net/http"
	"runtime"

	"go.uber.org/zap"
)

// Recovery middleware for panic recovery
func New(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Get stack trace
					stack := make([]byte, 4096)
					length := runtime.Stack(stack, false)

					// Log the panic
					logger.Error("Panic recovered",
						zap.Any("error", err),
						zap.String("stack", string(stack[:length])),
						zap.String("method", r.Method),
						zap.String("url", r.URL.String()),
					)

					// Return error response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					response := map[string]interface{}{
						"error":   "Internal server error",
						"message": "An unexpected error occurred",
					}

					json.NewEncoder(w).Encode(response)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
