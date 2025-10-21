package controllers

import (
	"net/http"
	"time"
	"crypto/rand"
	"encoding/hex"
	
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	
	"test-app2/models"
	"test-app2/services"
	"test-app2/http/requests"
)

type AuthController struct {
	authService *services.AuthService
	mailService *services.MailService
	db          *gorm.DB
}

func NewAuthController(authService *services.AuthService, mailService *services.MailService, db *gorm.DB) *AuthController {
	return &AuthController{
		authService: authService,
		mailService: mailService,
		db:          db,
	}
}

// ShowLoginForm displays the login form
func (ac *AuthController) ShowLoginForm(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/login", gin.H{
		"title": "Login",
	})
}

// Login handles user authentication
func (ac *AuthController) Login(c *gin.Context) {
	var req requests.LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "auth/login", gin.H{
			"title": "Login",
			"errors": err.Error(),
		})
		return
	}

	user, err := ac.authService.Attempt(req.Email, req.Password)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "auth/login", gin.H{
			"title": "Login",
			"error": "Invalid credentials",
		})
		return
	}

	// Set session
	session := ac.authService.CreateSession(user)
	c.SetCookie("session_token", session.Token, 3600*24*7, "/", "", false, true)

	// Redirect to intended page or dashboard
	intended := c.Query("intended")
	if intended != "" {
		c.Redirect(http.StatusFound, intended)
		return
	}
	c.Redirect(http.StatusFound, "/dashboard")
}

// ShowRegisterForm displays the registration form
func (ac *AuthController) ShowRegisterForm(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/register", gin.H{
		"title": "Register",
	})
}

// Register handles user registration
func (ac *AuthController) Register(c *gin.Context) {
	var req requests.RegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "auth/register", gin.H{
			"title": "Register",
			"errors": err.Error(),
		})
		return
	}

	user, err := ac.authService.Register(req.Name, req.Email, req.Password)
	if err != nil {
		c.HTML(http.StatusBadRequest, "auth/register", gin.H{
			"title": "Register",
			"error": err.Error(),
		})
		return
	}

	// Send email verification
	ac.mailService.SendEmailVerification(user)

	c.HTML(http.StatusOK, "auth/verify-email", gin.H{
		"title": "Verify Email",
		"user":  user,
	})
}

// ShowForgotPasswordForm displays the forgot password form
func (ac *AuthController) ShowForgotPasswordForm(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/forgot-password", gin.H{
		"title": "Forgot Password",
	})
}

// ForgotPassword handles forgot password request
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var req requests.ForgotPasswordRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "auth/forgot-password", gin.H{
			"title": "Forgot Password",
			"errors": err.Error(),
		})
		return
	}

	user := &models.User{}
	if err := ac.db.Where("email = ?", req.Email).First(user).Error; err != nil {
		// Don't reveal if email exists or not
		c.HTML(http.StatusOK, "auth/forgot-password", gin.H{
			"title": "Forgot Password",
			"message": "If your email exists, you will receive a password reset link.",
		})
		return
	}

	// Generate reset token
	token := generateRandomToken()
	reset := &models.PasswordReset{
		Email:     user.Email,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour), // Token expires in 1 hour
	}

	ac.db.Create(reset)

	// Send reset email
	ac.mailService.SendPasswordReset(user, token)

	c.HTML(http.StatusOK, "auth/forgot-password", gin.H{
		"title": "Forgot Password",
		"message": "If your email exists, you will receive a password reset link.",
	})
}

// ShowResetPasswordForm displays the reset password form
func (ac *AuthController) ShowResetPasswordForm(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.HTML(http.StatusBadRequest, "auth/reset-password", gin.H{
			"title": "Reset Password",
			"error": "Invalid reset token",
		})
		return
	}

	reset := &models.PasswordReset{}
	if err := ac.db.Where("token = ?", token).First(reset).Error; err != nil {
		c.HTML(http.StatusBadRequest, "auth/reset-password", gin.H{
			"title": "Reset Password",
			"error": "Invalid reset token",
		})
		return
	}

	if reset.IsExpired() {
		c.HTML(http.StatusBadRequest, "auth/reset-password", gin.H{
			"title": "Reset Password",
			"error": "Reset token has expired",
		})
		return
	}

	c.HTML(http.StatusOK, "auth/reset-password", gin.H{
		"title": "Reset Password",
		"token": token,
	})
}

// ResetPassword handles password reset
func (ac *AuthController) ResetPassword(c *gin.Context) {
	var req requests.ResetPasswordRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "auth/reset-password", gin.H{
			"title": "Reset Password",
			"errors": err.Error(),
			"token": req.Token,
		})
		return
	}

	reset := &models.PasswordReset{}
	if err := ac.db.Where("token = ?", req.Token).First(reset).Error; err != nil {
		c.HTML(http.StatusBadRequest, "auth/reset-password", gin.H{
			"title": "Reset Password",
			"error": "Invalid reset token",
			"token": req.Token,
		})
		return
	}

	if reset.IsExpired() {
		c.HTML(http.StatusBadRequest, "auth/reset-password", gin.H{
			"title": "Reset Password",
			"error": "Reset token has expired",
			"token": req.Token,
		})
		return
	}

	user := &models.User{}
	if err := ac.db.Where("email = ?", reset.Email).First(user).Error; err != nil {
		c.HTML(http.StatusBadRequest, "auth/reset-password", gin.H{
			"title": "Reset Password",
			"error": "User not found",
			"token": req.Token,
		})
		return
	}

	// Update password
	if err := user.SetPassword(req.Password); err != nil {
		c.HTML(http.StatusInternalServerError, "auth/reset-password", gin.H{
			"title": "Reset Password",
			"error": "Failed to update password",
			"token": req.Token,
		})
		return
	}

	ac.db.Save(user)

	// Delete reset token
	ac.db.Delete(reset)

	c.Redirect(http.StatusFound, "/login?message=Password reset successfully")
}

// ShowVerifyEmailForm displays the email verification form
func (ac *AuthController) ShowVerifyEmailForm(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/verify-email", gin.H{
		"title": "Verify Email",
	})
}

// VerifyEmail handles email verification
func (ac *AuthController) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.HTML(http.StatusBadRequest, "auth/verify-email", gin.H{
			"title": "Verify Email",
			"error": "Invalid verification token",
		})
		return
	}

	// In a real implementation, you would verify the token
	// For now, we'll just show a success message
	c.HTML(http.StatusOK, "auth/verify-email", gin.H{
		"title": "Verify Email",
		"message": "Email verified successfully!",
	})
}

// ResendVerification resends email verification
func (ac *AuthController) ResendVerification(c *gin.Context) {
	email := c.PostForm("email")
	if email == "" {
		c.HTML(http.StatusBadRequest, "auth/verify-email", gin.H{
			"title": "Verify Email",
			"error": "Email is required",
		})
		return
	}

	user := &models.User{}
	if err := ac.db.Where("email = ?", email).First(user).Error; err != nil {
		c.HTML(http.StatusBadRequest, "auth/verify-email", gin.H{
			"title": "Verify Email",
			"error": "User not found",
		})
		return
	}

	ac.mailService.SendEmailVerification(user)

	c.HTML(http.StatusOK, "auth/verify-email", gin.H{
		"title": "Verify Email",
		"message": "Verification email sent!",
	})
}

// Dashboard displays the user dashboard
func (ac *AuthController) Dashboard(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "dashboard", gin.H{
		"title": "Dashboard",
		"user":  user,
	})
}

// ShowProfileForm displays the profile form
func (ac *AuthController) ShowProfileForm(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "auth/profile", gin.H{
		"title": "Profile",
		"user":  user,
	})
}

// UpdateProfile handles profile updates
func (ac *AuthController) UpdateProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var req requests.UpdateProfileRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "auth/profile", gin.H{
			"title": "Profile",
			"user":  user,
			"errors": err.Error(),
		})
		return
	}

	// Update user profile
	userModel := user.(*models.User)
	userModel.Name = req.Name
	userModel.Email = req.Email

	ac.db.Save(userModel)

	c.HTML(http.StatusOK, "auth/profile", gin.H{
		"title": "Profile",
		"user":  userModel,
		"message": "Profile updated successfully!",
	})
}

// Logout handles user logout
func (ac *AuthController) Logout(c *gin.Context) {
	// Clear session cookie
	c.SetCookie("session_token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/login")
}

// Helper function to generate random token
func generateRandomToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
