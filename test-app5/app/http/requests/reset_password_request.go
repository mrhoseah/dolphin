package requests

type ResetPasswordRequest struct {
	Token                string `form:"token" binding:"required"`
	Email                string `form:"email" binding:"required,email"`
	Password             string `form:"password" binding:"required,min=8"`
	PasswordConfirmation string `form:"password_confirmation" binding:"required,eqfield=Password"`
}
