package controllers

import (
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
)

// AuthController handles authentication requests
type AuthController struct{}

// NewAuthController creates a new AuthController
func NewAuthController() *AuthController {
	return &AuthController{}
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com" binding:"required,email"`
	Password string `json:"password" example:"password123" binding:"required"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Name     string `json:"name" example:"John Doe" binding:"required"`
	Email    string `json:"email" example:"user@example.com" binding:"required,email"`
	Password string `json:"password" example:"password123" binding:"required,min=6"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType string    `json:"token_type" example:"Bearer"`
	ExpiresAt time.Time `json:"expires_at" example:"2024-01-01T12:00:00Z"`
	User      User      `json:"user"`
}

// Login handles POST /auth/login
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} SuccessResponse{data=AuthResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.JSON(w, r, ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request data",
		})
		return
	}

	// Simulate authentication logic
	if req.Email != "user@example.com" || req.Password != "password123" {
		w.WriteHeader(http.StatusUnauthorized)
		render.JSON(w, r, ErrorResponse{
			Error:   "Unauthorized",
			Message: "Invalid credentials",
		})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 1,
		"email":   req.Email,
		"role":    "user",
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to generate token",
		})
		return
	}

	user := User{ID: 1, Name: "John Doe", Email: req.Email}

	render.JSON(w, r, SuccessResponse{
		Message: "Login successful",
		Data: AuthResponse{
			Token:     tokenString,
			TokenType: "Bearer",
			ExpiresAt: time.Now().Add(time.Hour * 24),
			User:      user,
		},
	})
}

// Register handles POST /auth/register
// @Summary User registration
// @Description Register a new user account
// @Tags authentication
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration data"
// @Success 201 {object} SuccessResponse{data=AuthResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.JSON(w, r, ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request data",
		})
		return
	}

	// Simulate user creation logic
	user := User{ID: 2, Name: req.Name, Email: req.Email}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    "user",
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to generate token",
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, SuccessResponse{
		Message: "User registered successfully",
		Data: AuthResponse{
			Token:     tokenString,
			TokenType: "Bearer",
			ExpiresAt: time.Now().Add(time.Hour * 24),
			User:      user,
		},
	})
}

// Logout handles POST /auth/logout
// @Summary User logout
// @Description Logout user and invalidate token
// @Tags authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/logout [post]
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, SuccessResponse{
		Message: "Logout successful",
		Data:    nil,
	})
}

// RefreshToken handles POST /auth/refresh
// @Summary Refresh JWT token
// @Description Refresh an expired JWT token
// @Tags authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse{data=AuthResponse}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/refresh [post]
func (c *AuthController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Generate new JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 1,
		"email":   "user@example.com",
		"role":    "user",
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to generate token",
		})
		return
	}

	user := User{ID: 1, Name: "John Doe", Email: "user@example.com"}

	render.JSON(w, r, SuccessResponse{
		Message: "Token refreshed successfully",
		Data: AuthResponse{
			Token:     tokenString,
			TokenType: "Bearer",
			ExpiresAt: time.Now().Add(time.Hour * 24),
			User:      user,
		},
	})
}
