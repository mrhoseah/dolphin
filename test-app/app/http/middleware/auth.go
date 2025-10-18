package middleware

import (
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}
		
		// Check for Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}
		
		token := tokenParts[1]
		
		// TODO: Implement JWT token validation
		// For now, just check if token exists
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		
		// TODO: Parse token and set user context
		// c.Set("user", user)
		
		c.Next()
	}
}

// GuestMiddleware redirects authenticated users
func GuestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Check if user is authenticated
		// If authenticated, redirect to dashboard
		c.Next()
	}
}
