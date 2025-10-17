package graphql

import (
	"context"
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"
	"go.uber.org/zap"
)

// Directive represents a custom GraphQL directive
type Directive struct {
	Name        string
	Description string
	Locations   []string // FIELD, OBJECT, ARGUMENT, etc.
	Handler     DirectiveHandler
}

// DirectiveHandler handles directive execution
type DirectiveHandler func(ctx context.Context, params DirectiveParams) (interface{}, error)

// DirectiveParams contains parameters for directive execution
type DirectiveParams struct {
	Source  interface{}
	Args    map[string]interface{}
	Field   *graphql.Field
	Object  *graphql.Object
	Context context.Context
	Logger  *zap.Logger
}

// DirectiveRegistry manages custom directives
type DirectiveRegistry struct {
	directives map[string]*Directive
	logger     *zap.Logger
}

// NewDirectiveRegistry creates a new directive registry
func NewDirectiveRegistry(logger *zap.Logger) *DirectiveRegistry {
	return &DirectiveRegistry{
		directives: make(map[string]*Directive),
		logger:     logger,
	}
}

// RegisterDirective registers a custom directive
func (dr *DirectiveRegistry) RegisterDirective(directive *Directive) {
	dr.directives[directive.Name] = directive
	dr.logger.Info("Registered directive", zap.String("name", directive.Name))
}

// GetDirective gets a directive by name
func (dr *DirectiveRegistry) GetDirective(name string) (*Directive, bool) {
	directive, exists := dr.directives[name]
	return directive, exists
}

// ListDirectives returns all registered directives
func (dr *DirectiveRegistry) ListDirectives() []*Directive {
	directives := make([]*Directive, 0, len(dr.directives))
	for _, directive := range dr.directives {
		directives = append(directives, directive)
	}
	return directives
}

// CreateAuthDirective creates an authorization directive
func CreateAuthDirective() *Directive {
	return &Directive{
		Name:        "auth",
		Description: "Requires authentication with specified role",
		Locations:   []string{"FIELD", "OBJECT"},
		Handler: func(ctx context.Context, params DirectiveParams) (interface{}, error) {
			// Get role from directive arguments
			role, ok := params.Args["role"].(string)
			if !ok {
				return nil, fmt.Errorf("role argument is required for @auth directive")
			}

			// Check if user is authenticated and has the required role
			userRole, exists := ctx.Value("user_role").(string)
			if !exists {
				return nil, fmt.Errorf("authentication required")
			}

			if userRole != role && userRole != "admin" {
				return nil, fmt.Errorf("insufficient permissions: required role %s", role)
			}

			params.Logger.Info("Authorization successful",
				zap.String("required_role", role),
				zap.String("user_role", userRole),
			)

			return params.Source, nil
		},
	}
}

// CreateCacheDirective creates a caching directive
func CreateCacheDirective() *Directive {
	return &Directive{
		Name:        "cache",
		Description: "Sets cache control for the field",
		Locations:   []string{"FIELD"},
		Handler: func(ctx context.Context, params DirectiveParams) (interface{}, error) {
			// Get cache settings from directive arguments
			maxAge, _ := params.Args["maxAge"].(int)
			if maxAge == 0 {
				maxAge = 300 // Default 5 minutes
			}

			// Set cache headers in context
			ctx = context.WithValue(ctx, "cache_control", fmt.Sprintf("max-age=%d", maxAge))
			params.Logger.Info("Cache directive applied",
				zap.Int("max_age", maxAge),
			)

			return params.Source, nil
		},
	}
}

// CreateTransformDirective creates a transformation directive
func CreateTransformDirective() *Directive {
	return &Directive{
		Name:        "transform",
		Description: "Transforms the field value",
		Locations:   []string{"FIELD"},
		Handler: func(ctx context.Context, params DirectiveParams) (interface{}, error) {
			// Get transformation type from directive arguments
			transform, ok := params.Args["type"].(string)
			if !ok {
				return params.Source, nil
			}

			// Apply transformation based on type
			switch transform {
			case "uppercase":
				if str, ok := params.Source.(string); ok {
					return strings.ToUpper(str), nil
				}
			case "lowercase":
				if str, ok := params.Source.(string); ok {
					return strings.ToLower(str), nil
				}
			case "trim":
				if str, ok := params.Source.(string); ok {
					return strings.TrimSpace(str), nil
				}
			default:
				params.Logger.Warn("Unknown transformation type",
					zap.String("type", transform),
				)
			}

			return params.Source, nil
		},
	}
}

// CreateDeprecatedDirective creates a deprecation directive
func CreateDeprecatedDirective() *Directive {
	return &Directive{
		Name:        "deprecated",
		Description: "Marks a field as deprecated",
		Locations:   []string{"FIELD"},
		Handler: func(ctx context.Context, params DirectiveParams) (interface{}, error) {
			reason, _ := params.Args["reason"].(string)
			if reason == "" {
				reason = "This field is deprecated"
			}

			params.Logger.Warn("Deprecated field accessed",
				zap.String("reason", reason),
			)

			return params.Source, nil
		},
	}
}

// CreateRateLimitDirective creates a rate limiting directive
func CreateRateLimitDirective() *Directive {
	return &Directive{
		Name:        "rateLimit",
		Description: "Applies rate limiting to the field",
		Locations:   []string{"FIELD"},
		Handler: func(ctx context.Context, params DirectiveParams) (interface{}, error) {
			// Get rate limit settings from directive arguments
			requests, _ := params.Args["requests"].(int)
			window, _ := params.Args["window"].(int)
			if requests == 0 {
				requests = 100 // Default 100 requests
			}
			if window == 0 {
				window = 60 // Default 60 seconds
			}

			// Check rate limit (simplified implementation)
			userID, exists := ctx.Value("user_id").(string)
			if !exists {
				userID = "anonymous"
			}

			// In a real implementation, you would check against a rate limiter
			params.Logger.Info("Rate limit check",
				zap.String("user_id", userID),
				zap.Int("requests", requests),
				zap.Int("window", window),
			)

			return params.Source, nil
		},
	}
}

// CreateValidationDirective creates a validation directive
func CreateValidationDirective() *Directive {
	return &Directive{
		Name:        "validate",
		Description: "Validates field values",
		Locations:   []string{"FIELD"},
		Handler: func(ctx context.Context, params DirectiveParams) (interface{}, error) {
			// Get validation rules from directive arguments
			rules, ok := params.Args["rules"].(map[string]interface{})
			if !ok {
				return params.Source, nil
			}

			// Apply validation rules
			for rule, value := range rules {
				switch rule {
				case "minLength":
					if str, ok := params.Source.(string); ok {
						if minLen, ok := value.(int); ok && len(str) < minLen {
							return nil, fmt.Errorf("string too short: minimum length is %d", minLen)
						}
					}
				case "maxLength":
					if str, ok := params.Source.(string); ok {
						if maxLen, ok := value.(int); ok && len(str) > maxLen {
							return nil, fmt.Errorf("string too long: maximum length is %d", maxLen)
						}
					}
				case "pattern":
					if str, ok := params.Source.(string); ok {
						if pattern, ok := value.(string); ok {
							// In a real implementation, you would use regex
							params.Logger.Info("Pattern validation",
								zap.String("value", str),
								zap.String("pattern", pattern),
							)
						}
					}
				}
			}

			return params.Source, nil
		},
	}
}

// ApplyDirectives applies directives to a field
func (dr *DirectiveRegistry) ApplyDirectives(
	ctx context.Context,
	source interface{},
	field *graphql.Field,
	directives []string,
) (interface{}, error) {
	result := source

	for _, directiveName := range directives {
		directive, exists := dr.GetDirective(directiveName)
		if !exists {
			dr.logger.Warn("Unknown directive",
				zap.String("name", directiveName),
			)
			continue
		}

		// Create directive parameters
		params := DirectiveParams{
			Source:  result,
			Args:    make(map[string]interface{}), // In a real implementation, parse from field definition
			Field:   field,
			Context: ctx,
			Logger:  dr.logger,
		}

		// Execute directive
		newResult, err := directive.Handler(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("directive %s failed: %w", directiveName, err)
		}

		result = newResult
	}

	return result, nil
}

// CreateDirectiveField creates a field with directive support
func (dr *DirectiveRegistry) CreateDirectiveField(
	name string,
	fieldType graphql.Output,
	directives []string,
	resolver graphql.FieldResolveFn,
) *graphql.Field {
	return &graphql.Field{
		Type: fieldType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// First resolve the field
			result, err := resolver(p)
			if err != nil {
				return nil, err
			}

			// Then apply directives
			return dr.ApplyDirectives(p.Context, result, &graphql.Field{Type: fieldType}, directives)
		},
	}
}

// InitializeDefaultDirectives initializes default directives
func (dr *DirectiveRegistry) InitializeDefaultDirectives() {
	dr.RegisterDirective(CreateAuthDirective())
	dr.RegisterDirective(CreateCacheDirective())
	dr.RegisterDirective(CreateTransformDirective())
	dr.RegisterDirective(CreateDeprecatedDirective())
	dr.RegisterDirective(CreateRateLimitDirective())
	dr.RegisterDirective(CreateValidationDirective())
}
