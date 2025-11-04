package middleware

import (
	"dolphin/internal/auth"
	"dolphin/internal/middleware"

	"go.uber.org/zap"
)

// AuthMiddleware represents the authentication middleware
type AuthMiddleware = middleware.AuthMiddleware

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authManager *auth.AuthManager, logger *zap.Logger) *AuthMiddleware {
	return middleware.NewAuthMiddleware(authManager, logger)
}
