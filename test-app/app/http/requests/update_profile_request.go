package requests

type UpdateProfileRequest struct {
	Name  string `form:"name" binding:"required,min=2,max=255"`
	Email string `form:"email" binding:"required,email,max=255"`
}
