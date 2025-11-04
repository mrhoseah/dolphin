package middleware

import (
	"dolphin/internal/apikey"
	"dolphin/internal/middleware"
	"net/http"
)

// APIKeyAuth middleware for API key authentication
func APIKeyAuth(manager *apikey.APIKeyManager) func(http.Handler) http.Handler {
	return middleware.APIKeyAuth(manager)
}

// RequireScope middleware checks if API key has required scope
func RequireScope(scope string) func(http.Handler) http.Handler {
	return middleware.RequireScope(scope)
}

