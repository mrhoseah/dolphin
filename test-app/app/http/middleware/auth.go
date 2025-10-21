package middleware

import (
	"net/http"
	"strings"
	"time"
	
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	
	"app/models"
	"app/services"
)

type AuthMiddleware struct {
	db          *gorm.DB
	authService *services.AuthService
}

func NewAuthMiddleware(db *gorm.DB, authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		db:          db,
		authService: authService,
	}
}

// AuthMiddleware validates session tokens
func (am *AuthMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionToken, err := c.Cookie("session_token")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		user, err := am.authService.ValidateSession(sessionToken)
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

// GuestMiddleware redirects authenticated users
func (am *AuthMiddleware) GuestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionToken, err := c.Cookie("session_token")
		if err != nil {
			c.Next()
			return
		}

		_, err = am.authService.ValidateSession(sessionToken)
		if err == nil {
			c.Redirect(http.StatusFound, "/dashboard")
			c.Abort()
			return
		}

		c.Next()
	}
}

// EnsureEmailVerified middleware ensures user's email is verified
func (am *AuthMiddleware) EnsureEmailVerified() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		userModel := user.(*models.User)
		if !userModel.IsEmailVerified() {
			c.Redirect(http.StatusFound, "/verify-email")
			c.Abort()
			return
		}

		c.Next()
	}
}
