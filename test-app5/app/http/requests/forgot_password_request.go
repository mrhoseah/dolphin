package requests

type ForgotPasswordRequest struct {
	Email string `form:"email" binding:"required,email"`
}
