package controllers

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/mrhoseah/dolphin/internal/auth"
	"go.uber.org/zap"
)

// AuthController handles authentication requests
type AuthController struct {
	authService *auth.AuthManager
	logger      *zap.Logger
}

// NewAuthController creates a new authentication controller
func NewAuthController(authService *auth.AuthManager, logger *zap.Logger) *AuthController {
	return &AuthController{
		authService: authService,
		logger:      logger,
	}
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body auth.LoginRequest true "Login credentials"
// @Success 200 {object} auth.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequestDto
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		c.logger.Error("Failed to decode login request", zap.Error(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Email and password are required"})
		return
	}

	// Authenticate user
	err := c.authService.LoginWithCredentials(map[string]string{
		"email":    req.Email,
		"password": req.Password,
	})
	if err != nil {
		c.logger.Warn("Login failed", zap.String("email", req.Email), zap.Error(err))
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, map[string]string{"error": "Invalid credentials"})
		return
	}

	// Get authenticated user
	user := c.authService.User()
	if user == nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, map[string]string{"error": "Authentication failed"})
		return
	}

	c.logger.Info("User logged in successfully", zap.String("email", req.Email))
	render.JSON(w, r, map[string]interface{}{
		"message": "Login successful",
		"user":    user,
	})
}

// Register handles user registration
// @Summary Register new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param user body auth.RegisterRequest true "User registration data"
// @Success 201 {object} auth.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /auth/register [post]
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req auth.RegisterRequestDto
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		c.logger.Error("Failed to decode registration request", zap.Error(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "All fields are required"})
		return
	}

	// Register user (placeholder - implement user creation logic)
	// For now, just return success
	c.logger.Info("User registered successfully", zap.String("email", req.Email))
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]string{
			"email":     req.Email,
			"firstName": req.FirstName,
			"lastName":  req.LastName,
		},
	})
}

// Logout handles user logout
// @Summary Logout user
// @Description Logout the authenticated user
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/logout [post]
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	user := c.authService.User()
	if user == nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, map[string]string{"error": "Authentication required"})
		return
	}

	// Logout user
	c.authService.Logout()

	c.logger.Info("User logged out successfully", zap.Uint("user_id", user.GetID()))
	render.JSON(w, r, map[string]string{"message": "Logged out successfully"})
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Generate new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token body map[string]string true "Refresh token"
// @Success 200 {object} auth.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (c *AuthController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		c.logger.Error("Failed to decode refresh request", zap.Error(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.RefreshToken == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Refresh token is required"})
		return
	}

	// Refresh token (placeholder - implement token refresh logic)
	c.logger.Info("Token refresh requested", zap.String("refresh_token", req.RefreshToken))
	render.JSON(w, r, map[string]string{"message": "Token refresh not implemented yet"})
}

// Me returns current user information
// @Summary Get current user
// @Description Get information about the currently authenticated user
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} auth.User
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
func (c *AuthController) Me(w http.ResponseWriter, r *http.Request) {
	user := c.authService.User()
	if user == nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, map[string]string{"error": "Authentication required"})
		return
	}

	render.JSON(w, r, user)
}

// Check returns authentication status
// @Summary Check authentication status
// @Description Check if user is authenticated
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/check [get]
func (c *AuthController) Check(w http.ResponseWriter, r *http.Request) {
	user := c.authService.User()
	if user != nil {
		render.JSON(w, r, map[string]interface{}{
			"authenticated": true,
			"user":          user,
		})
	} else {
		render.JSON(w, r, map[string]interface{}{
			"authenticated": false,
		})
	}
}

// Guest returns guest status
// @Summary Check guest status
// @Description Check if user is a guest (not authenticated)
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/guest [get]
func (c *AuthController) Guest(w http.ResponseWriter, r *http.Request) {
	user := c.authService.User()
	isGuest := user == nil
	render.JSON(w, r, map[string]interface{}{
		"guest": isGuest,
	})
}
