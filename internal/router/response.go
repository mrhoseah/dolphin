package router

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
)

// Response represents a standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta contains pagination and metadata information
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    Meta        `json:"meta"`
}

// JSONResponse sends a JSON response with the given status code
func JSONResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	render.Status(r, status)
	render.JSON(w, r, data)
}

// Success sends a successful JSON response
func Success(w http.ResponseWriter, r *http.Request, data interface{}, message string) {
	response := Response{
		Success: true,
		Data:    data,
		Message: message,
	}
	JSONResponse(w, r, http.StatusOK, response)
}

// Created sends a 201 Created response
func Created(w http.ResponseWriter, r *http.Request, data interface{}, message string) {
	response := Response{
		Success: true,
		Data:    data,
		Message: message,
	}
	JSONResponse(w, r, http.StatusCreated, response)
}

// ErrorResponse sends an error JSON response
func ErrorResponse(w http.ResponseWriter, r *http.Request, status int, message string, err error) {
	response := Response{
		Success: false,
		Error:   message,
	}
	if err != nil {
		response.Error = err.Error()
	}
	JSONResponse(w, r, status, response)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(w http.ResponseWriter, r *http.Request, message string) {
	ErrorResponse(w, r, http.StatusBadRequest, message, nil)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(w http.ResponseWriter, r *http.Request, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	ErrorResponse(w, r, http.StatusUnauthorized, message, nil)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(w http.ResponseWriter, r *http.Request, message string) {
	if message == "" {
		message = "Forbidden"
	}
	ErrorResponse(w, r, http.StatusForbidden, message, nil)
}

// NotFound sends a 404 Not Found response
func NotFound(w http.ResponseWriter, r *http.Request, message string) {
	if message == "" {
		message = "Resource not found"
	}
	ErrorResponse(w, r, http.StatusNotFound, message, nil)
}

// ValidationError sends a 422 Unprocessable Entity response with validation errors
func ValidationError(w http.ResponseWriter, r *http.Request, errors map[string][]string) {
	response := Response{
		Success: false,
		Error:   "Validation failed",
		Data:    errors,
	}
	JSONResponse(w, r, http.StatusUnprocessableEntity, response)
}

// InternalError sends a 500 Internal Server Error response
func InternalError(w http.ResponseWriter, r *http.Request, message string, err error) {
	if message == "" {
		message = "Internal server error"
	}
	ErrorResponse(w, r, http.StatusInternalServerError, message, err)
}

// Paginated sends a paginated JSON response
func Paginated(w http.ResponseWriter, r *http.Request, data interface{}, page, perPage int, total int64) {
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	
	response := PaginatedResponse{
		Success: true,
		Data:    data,
		Meta: Meta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	}
	JSONResponse(w, r, http.StatusOK, response)
}

// NoContent sends a 204 No Content response
func NoContent(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusNoContent)
	w.WriteHeader(http.StatusNoContent)
}

// ParsePagination parses pagination parameters from query string
func ParsePagination(r *http.Request) (page, perPage int) {
	page = 1
	perPage = 15 // default

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	return page, perPage
}

// ParseQueryInt parses an integer from query parameters
func ParseQueryInt(r *http.Request, key string, defaultValue int) int {
	if val := r.URL.Query().Get(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// ParseQueryString parses a string from query parameters
func ParseQueryString(r *http.Request, key string, defaultValue string) string {
	if val := r.URL.Query().Get(key); val != "" {
		return val
	}
	return defaultValue
}

// ParseQueryBool parses a boolean from query parameters
func ParseQueryBool(r *http.Request, key string, defaultValue bool) bool {
	if val := r.URL.Query().Get(key); val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// BindJSON binds JSON request body to a struct
func BindJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

