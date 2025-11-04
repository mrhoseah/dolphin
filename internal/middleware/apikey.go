package middleware

import (
	"context"
	"net/http"
	"strings"

	"dolphin/internal/apikey"
)

// APIKeyAuth middleware for API key authentication
func APIKeyAuth(manager *apikey.APIKeyManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get API key from header or query parameter
			key := r.Header.Get("X-API-Key")
			if key == "" {
				key = r.URL.Query().Get("api_key")
			}
			if key == "" {
				// Try Bearer token format
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					key = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if key == "" {
				http.Error(w, "API key required", http.StatusUnauthorized)
				return
			}

			// Validate API key
			apiKey, err := manager.Validate(key)
			if err != nil {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			// Add API key to context
			ctx := context.WithValue(r.Context(), "api_key", apiKey)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireScope middleware checks if API key has required scope
func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey, ok := r.Context().Value("api_key").(*apikey.APIKey)
			if !ok {
				http.Error(w, "API key required", http.StatusUnauthorized)
				return
			}

			manager, ok := r.Context().Value("api_key_manager").(*apikey.APIKeyManager)
			if !ok {
				http.Error(w, "API key manager not found", http.StatusInternalServerError)
				return
			}

			if !manager.HasScope(apiKey, scope) {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetAPIKeyFromContext gets API key from request context
func GetAPIKeyFromContext(ctx context.Context) (*apikey.APIKey, bool) {
	apiKey, ok := ctx.Value("api_key").(*apikey.APIKey)
	return apiKey, ok
}

