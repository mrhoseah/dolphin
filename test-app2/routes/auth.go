package routes

import (
	"github.com/gin-gonic/gin"
	"app/http/controllers"
	"app/http/middleware"
)

func SetupAuthRoutes(r *gin.Engine, authController *controllers.AuthController, authMiddleware *middleware.AuthMiddleware) {
	// Guest routes (only accessible when not authenticated)
	guest := r.Group("/")
	guest.Use(authMiddleware.GuestMiddleware())
	{
		guest.GET("/login", authController.ShowLoginForm)
		guest.POST("/login", authController.Login)
		guest.GET("/register", authController.ShowRegisterForm)
		guest.POST("/register", authController.Register)
		guest.GET("/forgot-password", authController.ShowForgotPasswordForm)
		guest.POST("/forgot-password", authController.ForgotPassword)
		guest.GET("/reset-password", authController.ShowResetPasswordForm)
		guest.POST("/reset-password", authController.ResetPassword)
		guest.GET("/verify-email", authController.ShowVerifyEmailForm)
		guest.POST("/verify-email", authController.ResendVerification)
	}

	// Protected routes (require authentication)
	protected := r.Group("/")
	protected.Use(authMiddleware.AuthMiddleware())
	{
		protected.GET("/dashboard", authController.Dashboard)
		protected.GET("/profile", authController.ShowProfileForm)
		protected.POST("/profile", authController.UpdateProfile)
		protected.POST("/logout", authController.Logout)
	}

	// Email verification route (accessible with token)
	r.GET("/verify-email", authController.VerifyEmail)
}
