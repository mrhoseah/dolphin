package controllers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// UserController handles user-related requests
type UserController struct{}

// NewUserController creates a new UserController
func NewUserController() *UserController {
	return &UserController{}
}

// User represents a user model
type User struct {
	ID    int    `json:"id" example:"1"`
	Name  string `json:"name" example:"John Doe"`
	Email string `json:"email" example:"john@example.com"`
}

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Name  string `json:"name" example:"John Doe" binding:"required"`
	Email string `json:"email" example:"john@example.com" binding:"required,email"`
}

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	Name  string `json:"name" example:"John Doe"`
	Email string `json:"email" example:"john@example.com"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Bad Request"`
	Message string `json:"message" example:"Invalid request data"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message" example:"Success"`
	Data    interface{} `json:"data"`
}

// Index handles GET /users
// @Summary Get all users
// @Description Get a list of all users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse{data=[]User}
// @Failure 500 {object} ErrorResponse
// @Router /users [get]
func (c *UserController) Index(w http.ResponseWriter, r *http.Request) {
	users := []User{
		{ID: 1, Name: "John Doe", Email: "john@example.com"},
		{ID: 2, Name: "Jane Smith", Email: "jane@example.com"},
	}

	render.JSON(w, r, SuccessResponse{
		Message: "Users retrieved successfully",
		Data:    users,
	})
}

// Show handles GET /users/{id}
// @Summary Get user by ID
// @Description Get a specific user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} SuccessResponse{data=User}
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [get]
func (c *UserController) Show(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		render.JSON(w, r, ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid user ID",
		})
		return
	}

	user := User{ID: id, Name: "John Doe", Email: "john@example.com"}

	render.JSON(w, r, SuccessResponse{
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// Store handles POST /users
// @Summary Create a new user
// @Description Create a new user with the provided data
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User data"
// @Success 201 {object} SuccessResponse{data=User}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [post]
func (c *UserController) Store(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.JSON(w, r, ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request data",
		})
		return
	}

	user := User{ID: 1, Name: req.Name, Email: req.Email}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, SuccessResponse{
		Message: "User created successfully",
		Data:    user,
	})
}

// Update handles PUT /users/{id}
// @Summary Update user
// @Description Update an existing user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body UpdateUserRequest true "User data"
// @Success 200 {object} SuccessResponse{data=User}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [put]
func (c *UserController) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		render.JSON(w, r, ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid user ID",
		})
		return
	}

	var req UpdateUserRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.JSON(w, r, ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request data",
		})
		return
	}

	user := User{ID: id, Name: req.Name, Email: req.Email}

	render.JSON(w, r, SuccessResponse{
		Message: "User updated successfully",
		Data:    user,
	})
}

// Destroy handles DELETE /users/{id}
// @Summary Delete user
// @Description Delete a user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [delete]
func (c *UserController) Destroy(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		render.JSON(w, r, ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid user ID",
		})
		return
	}

	render.JSON(w, r, SuccessResponse{
		Message: "User deleted successfully",
		Data:    map[string]int{"id": id},
	})
}
