package middleware

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"dolphin/internal/auth"
	"go.uber.org/zap"
)

// AuthMiddleware handles Dolphin-style authentication
type AuthMiddleware struct {
	authManager *auth.AuthManager
	logger      *zap.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authManager *auth.AuthManager, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authManager: authManager,
		logger:      logger,
	}
}

// Authenticate middleware that requires authentication
// For web requests (HTML), redirects to login page
// For API requests (JSON), returns JSON error response
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is an API request
		isAPIRequest := strings.HasPrefix(r.URL.Path, "/api/") ||
			r.Header.Get("Accept") == "application/json" ||
			r.Header.Get("Content-Type") == "application/json"

		if !m.authManager.Check() {
			m.logger.Warn("Unauthenticated request",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))

			if isAPIRequest {
				// Return JSON response for API requests
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{
					"message": "Unauthenticated",
				})
			} else {
				// Redirect to login for web requests
				// Store the intended URL for redirect after login
				loginURL := "/auth/login"
				if r.URL.Path != "/" && r.URL.Path != "/auth/login" {
					loginURL = "/auth/login?redirect=" + r.URL.Path
				}
				http.Redirect(w, r, loginURL, http.StatusFound)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RedirectIfAuthenticated middleware that redirects authenticated users
func (m *AuthMiddleware) RedirectIfAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.authManager.Check() {
			// Redirect to dashboard or home
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// EnsureEmailIsVerified middleware that ensures user's email is verified
func (m *AuthMiddleware) EnsureEmailIsVerified(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := m.authManager.User()
		if user == nil {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, map[string]string{
				"message": "Unauthenticated",
			})
			return
		}

		// Check if user has verified email (you'll need to add this field to your User model)
		// For now, we'll assume all users are verified
		// if !user.IsEmailVerified() {
		//     render.Status(r, http.StatusForbidden)
		//     render.JSON(w, r, map[string]string{
		//         "message": "Email not verified",
		//     })
		//     return
		// }

		next.ServeHTTP(w, r)
	})
}

// ThrottleLoginAttempts middleware that throttles login attempts
func (m *AuthMiddleware) ThrottleLoginAttempts(maxAttempts int, decayMinutes int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This is a simplified implementation
			// In a real implementation, you'd use Redis or another cache to track attempts

			// For now, we'll just log the attempt
			m.logger.Info("Login attempt",
				zap.String("ip", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()))

			next.ServeHTTP(w, r)
		})
	}
}

// RoleMiddleware middleware that checks for specific roles
func (m *AuthMiddleware) RoleMiddleware(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := m.authManager.User()
			if user == nil {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{
					"message": "Unauthenticated",
				})
				return
			}

			// Simple role checking (you might want to implement a proper role system)
			hasRole := false
			for _, role := range roles {
				if m.hasRole(user, role) {
					hasRole = true
					break
				}
			}

			if !hasRole {
				m.logger.Warn("Insufficient permissions",
					zap.String("user_id", user.(string)), // Cast to string or appropriate type
					zap.Strings("required_roles", roles))
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, map[string]string{
					"message": "Insufficient permissions",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// PermissionMiddleware middleware that checks for specific permissions
func (m *AuthMiddleware) PermissionMiddleware(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := m.authManager.User()
			if user == nil {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{
					"message": "Unauthenticated",
				})
				return
			}

			// Simple permission checking (you might want to implement a proper permission system)
			hasPermission := false
			for _, permission := range permissions {
				if m.hasPermission(user, permission) {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				m.logger.Warn("Insufficient permissions",
					zap.String("user_id", user.(string)), // Cast to string or appropriate type
					zap.Strings("required_permissions", permissions))
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, map[string]string{
					"message": "Insufficient permissions",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// hasRole checks if user has the required role
func (m *AuthMiddleware) hasRole(user interface{}, role string) bool {
	// Simplified role checking
	// In a real implementation, you'd have a roles table or user roles
	switch role {
	case "admin":
		// For now, admin role is hardcoded based on email
		return user.(string) == "admin@example.com" // Cast to string
	case "user":
		return true // All authenticated users have user role
	default:
		return false
	}
}

// hasPermission checks if user has the required permission
func (m *AuthMiddleware) hasPermission(user interface{}, permission string) bool {
	// Simplified permission checking
	// In a real implementation, you'd have a permissions table
	switch permission {
	case "read":
		return true // All authenticated users can read
	case "write":
		return user.(string) == "admin@example.com" // Cast to string
	case "delete":
		return user.(string) == "admin@example.com" // Cast to string
	default:
		return false
	}
}

// Guest middleware that ensures the user is a guest
func (m *AuthMiddleware) Guest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.authManager.Check() {
			render.Status(r, http.StatusForbidden)
			render.JSON(w, r, map[string]string{
				"message": "Already authenticated",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// OptionalAuth middleware that adds user info if authenticated
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This middleware doesn't block requests, just adds user info to context
		// if available
		next.ServeHTTP(w, r)
	})
}
