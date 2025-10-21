package requests

type LoginRequest struct {
	Email    string `form:"email" binding:"required,email"`
	Password string `form:"password" binding:"required"`
	Remember bool   `form:"remember"`
}
