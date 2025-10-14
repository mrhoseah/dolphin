package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Rule represents a validation rule
type Rule interface {
	Validate(value interface{}) error
	Message() string
}

// RequiredRule validates that a field is not empty
type RequiredRule struct{}

func (r RequiredRule) Validate(value interface{}) error {
	if value == nil {
		return fmt.Errorf("field is required")
	}
	
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("field is required")
		}
	case []interface{}:
		if len(v) == 0 {
			return fmt.Errorf("field is required")
		}
	case map[string]interface{}:
		if len(v) == 0 {
			return fmt.Errorf("field is required")
		}
	}
	
	return nil
}

func (r RequiredRule) Message() string {
	return "This field is required"
}

// EmailRule validates email format
type EmailRule struct{}

func (r EmailRule) Validate(value interface{}) error {
	if value == nil {
		return nil // Let RequiredRule handle nil values
	}
	
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("email must be a string")
	}
	
	if str == "" {
		return nil // Let RequiredRule handle empty strings
	}
	
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(str) {
		return fmt.Errorf("invalid email format")
	}
	
	return nil
}

func (r EmailRule) Message() string {
	return "Must be a valid email address"
}

// MinLengthRule validates minimum string length
type MinLengthRule struct {
	Min int
}

func (r MinLengthRule) Validate(value interface{}) error {
	if value == nil {
		return nil
	}
	
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}
	
	if len(str) < r.Min {
		return fmt.Errorf("minimum length is %d characters", r.Min)
	}
	
	return nil
}

func (r MinLengthRule) Message() string {
	return fmt.Sprintf("Minimum length is %d characters", r.Min)
}

// MaxLengthRule validates maximum string length
type MaxLengthRule struct {
	Max int
}

func (r MaxLengthRule) Validate(value interface{}) error {
	if value == nil {
		return nil
	}
	
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}
	
	if len(str) > r.Max {
		return fmt.Errorf("maximum length is %d characters", r.Max)
	}
	
	return nil
}

func (r MaxLengthRule) Message() string {
	return fmt.Sprintf("Maximum length is %d characters", r.Max)
}

// MinRule validates minimum numeric value
type MinRule struct {
	Min float64
}

func (r MinRule) Validate(value interface{}) error {
	if value == nil {
		return nil
	}
	
	var num float64
	switch v := value.(type) {
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	case float64:
		num = v
	case float32:
		num = float64(v)
	default:
		return fmt.Errorf("value must be numeric")
	}
	
	if num < r.Min {
		return fmt.Errorf("minimum value is %f", r.Min)
	}
	
	return nil
}

func (r MinRule) Message() string {
	return fmt.Sprintf("Minimum value is %f", r.Min)
}

// MaxRule validates maximum numeric value
type MaxRule struct {
	Max float64
}

func (r MaxRule) Validate(value interface{}) error {
	if value == nil {
		return nil
	}
	
	var num float64
	switch v := value.(type) {
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	case float64:
		num = v
	case float32:
		num = float64(v)
	default:
		return fmt.Errorf("value must be numeric")
	}
	
	if num > r.Max {
		return fmt.Errorf("maximum value is %f", r.Max)
	}
	
	return nil
}

func (r MaxRule) Message() string {
	return fmt.Sprintf("Maximum value is %f", r.Max)
}

// InRule validates that value is in a list of allowed values
type InRule struct {
	Values []interface{}
}

func (r InRule) Validate(value interface{}) error {
	if value == nil {
		return nil
	}
	
	for _, allowed := range r.Values {
		if reflect.DeepEqual(value, allowed) {
			return nil
		}
	}
	
	return fmt.Errorf("value must be one of: %v", r.Values)
}

func (r InRule) Message() string {
	return fmt.Sprintf("Value must be one of: %v", r.Values)
}

// RegexRule validates against a regular expression
type RegexRule struct {
	Pattern *regexp.Regexp
}

func (r RegexRule) Validate(value interface{}) error {
	if value == nil {
		return nil
	}
	
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}
	
	if !r.Pattern.MatchString(str) {
		return fmt.Errorf("value does not match required pattern")
	}
	
	return nil
}

func (r RegexRule) Message() string {
	return "Value does not match required pattern"
}

// Validator handles validation of data
type Validator struct {
	rules map[string][]Rule
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string][]Rule),
	}
}

// AddRule adds a validation rule for a field
func (v *Validator) AddRule(field string, rule Rule) {
	v.rules[field] = append(v.rules[field], rule)
}

// Validate validates data against rules
func (v *Validator) Validate(data map[string]interface{}) map[string][]string {
	errors := make(map[string][]string)
	
	for field, rules := range v.rules {
		value := data[field]
		
		for _, rule := range rules {
			if err := rule.Validate(value); err != nil {
				if errors[field] == nil {
					errors[field] = []string{}
				}
				errors[field] = append(errors[field], err.Error())
			}
		}
	}
	
	return errors
}

// HasErrors checks if there are validation errors
func (v *Validator) HasErrors(errors map[string][]string) bool {
	return len(errors) > 0
}

// RequestValidator provides validation for HTTP requests
type RequestValidator struct {
	validator *Validator
}

// NewRequestValidator creates a new request validator
func NewRequestValidator() *RequestValidator {
	return &RequestValidator{
		validator: NewValidator(),
	}
}

// ValidateRequest validates request data
func (rv *RequestValidator) ValidateRequest(data map[string]interface{}, rules map[string][]Rule) map[string][]string {
	// Clear existing rules
	rv.validator.rules = make(map[string][]Rule)
	
	// Add new rules
	for field, fieldRules := range rules {
		for _, rule := range fieldRules {
			rv.validator.AddRule(field, rule)
		}
	}
	
	return rv.validator.Validate(data)
}

// Common validation rules
var (
	Required = RequiredRule{}
	Email    = EmailRule{}
)

// MinLength creates a minimum length rule
func MinLength(min int) Rule {
	return MinLengthRule{Min: min}
}

// MaxLength creates a maximum length rule
func MaxLength(max int) Rule {
	return MaxLengthRule{Max: max}
}

// Min creates a minimum value rule
func Min(min float64) Rule {
	return MinRule{Min: min}
}

// Max creates a maximum value rule
func Max(max float64) Rule {
	return MaxRule{Max: max}
}

// In creates an "in" rule
func In(values ...interface{}) Rule {
	return InRule{Values: values}
}

// Regex creates a regex rule
func Regex(pattern string) Rule {
	return RegexRule{Pattern: regexp.MustCompile(pattern)}
}
