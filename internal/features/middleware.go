package features

import (
	"net/http"
)

// Middleware checks if a feature flag is enabled before allowing request
func Middleware(manager *Manager, flagName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !manager.IsEnabled(flagName) {
				http.Error(w, "Feature not available", http.StatusNotFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// OptionalMiddleware allows request but adds header indicating feature status
func OptionalMiddleware(manager *Manager, flagName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if manager.IsEnabled(flagName) {
				w.Header().Set("X-Feature-"+flagName, "enabled")
			} else {
				w.Header().Set("X-Feature-"+flagName, "disabled")
			}
			next.ServeHTTP(w, r)
		})
	}
}

