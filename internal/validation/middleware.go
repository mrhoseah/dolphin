package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"go.uber.org/zap"
)

// ValidationMiddleware creates a middleware for request validation
func ValidationMiddleware(validator *Validator, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Parse request body
			var data interface{}
			if r.Body != nil {
				if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
					logger.Warn("Failed to parse JSON body", zap.Error(err))
					render.Status(r, http.StatusBadRequest)
					render.JSON(w, r, map[string]string{
						"error": "Invalid JSON format",
					})
					return
				}
			}

			// Validate the data
			if !validator.Validate(data.(map[string]interface{})) {
				logger.Warn("Validation failed")
				render.Status(r, http.StatusUnprocessableEntity)
				render.JSON(w, r, map[string]interface{}{
					"error":   "Validation failed",
					"details": "Validation failed",
				})
				return
			}

			// Store validated data in context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "validated_data", data)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SanitizationMiddleware creates a middleware for request sanitization
func SanitizationMiddleware(sanitizer *RequestSanitizer, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Parse request body
			var data interface{}
			if r.Body != nil {
				if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
					logger.Warn("Failed to parse JSON body", zap.Error(err))
					render.Status(r, http.StatusBadRequest)
					render.JSON(w, r, map[string]string{
						"error": "Invalid JSON format",
					})
					return
				}
			}

			// Sanitize the data
			if dataMap, ok := data.(map[string]interface{}); ok {
				sanitizedData := sanitizer.SanitizeRequest(dataMap)
				data = sanitizedData
			}

			// Store sanitized data in context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "sanitized_data", data)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FormValidationMiddleware creates a middleware for form validation
func FormValidationMiddleware(validator *Validator, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Parse form data
			if err := r.ParseForm(); err != nil {
				logger.Warn("Failed to parse form data", zap.Error(err))
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, map[string]string{
					"error": "Invalid form data",
				})
				return
			}

			// Convert form data to map
			formData := make(map[string]interface{})
			for key, values := range r.Form {
				if len(values) == 1 {
					formData[key] = values[0]
				} else {
					formData[key] = values
				}
			}

			// Validate the form data
			if !validator.Validate(formData) {
				logger.Warn("Form validation failed")
				render.Status(r, http.StatusUnprocessableEntity)
				render.JSON(w, r, map[string]interface{}{
					"error":   "Form validation failed",
					"details": "Validation failed",
				})
				return
			}

			// Store validated form data in context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "validated_form", formData)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FormSanitizationMiddleware creates a middleware for form sanitization
func FormSanitizationMiddleware(sanitizer *RequestSanitizer, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Parse form data
			if err := r.ParseForm(); err != nil {
				logger.Warn("Failed to parse form data", zap.Error(err))
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, map[string]string{
					"error": "Invalid form data",
				})
				return
			}

			// Sanitize form data
			sanitizedForm := sanitizer.SanitizeFormData(r.Form)

			// Convert sanitized form data to map
			formData := make(map[string]interface{})
			for key, values := range sanitizedForm {
				if len(values) == 1 {
					formData[key] = values[0]
				} else {
					formData[key] = values
				}
			}

			// Store sanitized form data in context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "sanitized_form", formData)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetValidatedData extracts validated data from context
func GetValidatedData(ctx context.Context) interface{} {
	return ctx.Value("validated_data")
}

// GetSanitizedData extracts sanitized data from context
func GetSanitizedData(ctx context.Context) interface{} {
	return ctx.Value("sanitized_data")
}

// GetValidatedForm extracts validated form data from context
func GetValidatedForm(ctx context.Context) map[string]interface{} {
	if data, ok := ctx.Value("validated_form").(map[string]interface{}); ok {
		return data
	}
	return nil
}

// GetSanitizedForm extracts sanitized form data from context
func GetSanitizedForm(ctx context.Context) map[string]interface{} {
	if data, ok := ctx.Value("sanitized_form").(map[string]interface{}); ok {
		return data
	}
	return nil
}

// ValidationManager manages validation and sanitization
type ValidationManager struct {
	validator    *Validator
	sanitizer    *RequestSanitizer
	reqSanitizer *RequestSanitizer
	logger       *zap.Logger
}

// NewValidationManager creates a new validation manager
func NewValidationManager(logger *zap.Logger) *ValidationManager {
	return &ValidationManager{
		validator:    NewValidator(),
		sanitizer:    NewRequestSanitizer(),
		reqSanitizer: NewRequestSanitizer(),
		logger:       logger,
	}
}

// GetValidator returns the field validator
func (vm *ValidationManager) GetValidator() *Validator {
	return vm.validator
}

// GetSanitizer returns the field sanitizer
func (vm *ValidationManager) GetSanitizer() *RequestSanitizer {
	return vm.sanitizer
}

// GetRequestSanitizer returns the request sanitizer
func (vm *ValidationManager) GetRequestSanitizer() *RequestSanitizer {
	return vm.reqSanitizer
}

// ValidateAndSanitize validates and sanitizes data
func (vm *ValidationManager) ValidateAndSanitize(data interface{}) error {
	// First sanitize
	if dataMap, ok := data.(map[string]interface{}); ok {
		sanitizedData := vm.sanitizer.SanitizeRequest(dataMap)
		data = sanitizedData
	}

	// Then validate
	if !vm.validator.Validate(data.(map[string]interface{})) {
		return fmt.Errorf("validation failed")
	}

	return nil
}

// ValidateStruct validates a struct with validation tags
func (vm *ValidationManager) ValidateStruct(data interface{}) error {
	if dataMap, ok := data.(map[string]interface{}); ok {
		if !vm.validator.Validate(dataMap) {
			return fmt.Errorf("validation failed")
		}
	}
	return nil
}

// SanitizeStruct sanitizes a struct with sanitization tags
func (vm *ValidationManager) SanitizeStruct(data interface{}) error {
	if dataMap, ok := data.(map[string]interface{}); ok {
		sanitizedData := vm.sanitizer.SanitizeRequest(dataMap)
		// Update the original data with sanitized values
		for k, v := range sanitizedData {
			dataMap[k] = v
		}
	}
	return nil
}

// ValidateField validates a single field (placeholder - not implemented)
func (vm *ValidationManager) ValidateField(field interface{}, rules []string) error {
	return fmt.Errorf("ValidateField not implemented")
}

// SanitizeField sanitizes a single field
func (vm *ValidationManager) SanitizeField(field interface{}, rules []string) (interface{}, error) {
	return nil, fmt.Errorf("SanitizeField not implemented")
}

// CreateValidationMiddleware creates a validation middleware
func (vm *ValidationManager) CreateValidationMiddleware() func(next http.Handler) http.Handler {
	return ValidationMiddleware(vm.validator, vm.logger)
}

// CreateSanitizationMiddleware creates a sanitization middleware
func (vm *ValidationManager) CreateSanitizationMiddleware() func(next http.Handler) http.Handler {
	return SanitizationMiddleware(vm.sanitizer, vm.logger)
}

// CreateFormValidationMiddleware creates a form validation middleware
func (vm *ValidationManager) CreateFormValidationMiddleware() func(next http.Handler) http.Handler {
	return FormValidationMiddleware(vm.validator, vm.logger)
}

// CreateFormSanitizationMiddleware creates a form sanitization middleware
func (vm *ValidationManager) CreateFormSanitizationMiddleware() func(next http.Handler) http.Handler {
	return FormSanitizationMiddleware(vm.reqSanitizer, vm.logger)
}

// Example usage structs

// UserRegistrationRequest represents a user registration request
type UserRegistrationRequest struct {
	Username        string `json:"username" validate:"required,min_length:3,max_length:20,alpha_numeric" sanitize:"trim,lowercase"`
	Email           string `json:"email" validate:"required,email" sanitize:"trim,lowercase"`
	Password        string `json:"password" validate:"required,min_length:8" sanitize:"trim"`
	ConfirmPassword string `json:"confirm_password" validate:"required" sanitize:"trim"`
	FirstName       string `json:"first_name" validate:"required,alpha" sanitize:"trim,titlecase"`
	LastName        string `json:"last_name" validate:"required,alpha" sanitize:"trim,titlecase"`
	Age             int    `json:"age" validate:"required,min:18,max:120"`
	Bio             string `json:"bio" validate:"max_length:500" sanitize:"trim,strip_html,normalize_whitespace"`
	Website         string `json:"website" validate:"url" sanitize:"trim,lowercase"`
}

// PostCreateRequest represents a post creation request
type PostCreateRequest struct {
	Title       string   `json:"title" validate:"required,min_length:5,max_length:200" sanitize:"trim,strip_html"`
	Content     string   `json:"content" validate:"required,min_length:10" sanitize:"trim,strip_html,normalize_whitespace"`
	Tags        []string `json:"tags" validate:"max_length:10" sanitize:"trim,lowercase"`
	IsPublished bool     `json:"is_published"`
	Category    string   `json:"category" validate:"required,in:tech,business,lifestyle" sanitize:"trim,lowercase"`
}

// CommentCreateRequest represents a comment creation request
type CommentCreateRequest struct {
	PostID   uint   `json:"post_id" validate:"required"`
	Content  string `json:"content" validate:"required,min_length:1,max_length:1000" sanitize:"trim,strip_html,normalize_whitespace"`
	ParentID *uint  `json:"parent_id"`
}
