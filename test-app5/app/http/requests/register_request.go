package requests

type RegisterRequest struct {
	Name                 string `form:"name" binding:"required,min=2,max=255"`
	Email                string `form:"email" binding:"required,email,max=255"`
	Password             string `form:"password" binding:"required,min=8"`
	PasswordConfirmation string `form:"password_confirmation" binding:"required,eqfield=Password"`
	Terms                bool   `form:"terms" binding:"required"`
}
