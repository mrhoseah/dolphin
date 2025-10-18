package validation

import (
	"reflect"
	"strconv"
	"strings"
)

// Validator manages validation rules and errors
type Validator struct {
	rules  map[string][]Rule
	errors map[string][]string
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		rules:  make(map[string][]Rule),
		errors: make(map[string][]string),
	}
}

// AddRule adds a validation rule for a field
func (v *Validator) AddRule(field string, rule Rule) {
	if v.rules[field] == nil {
		v.rules[field] = make([]Rule, 0)
	}
	v.rules[field] = append(v.rules[field], rule)
}

// AddRules adds multiple validation rules for a field
func (v *Validator) AddRules(field string, rules ...Rule) {
	for _, rule := range rules {
		v.AddRule(field, rule)
	}
}

// Validate validates data against all rules
func (v *Validator) Validate(data map[string]interface{}) bool {
	v.errors = make(map[string][]string)
	isValid := true

	for field, rules := range v.rules {
		value, exists := data[field]

		for _, rule := range rules {
			// Skip validation if field doesn't exist and rule is not required
			if !exists && !isRequiredRule(rule) {
				continue
			}

			if err := rule.Validate(value); err != nil {
				if v.errors[field] == nil {
					v.errors[field] = make([]string, 0)
				}
				v.errors[field] = append(v.errors[field], err.Error())
				isValid = false
			}
		}
	}

	return isValid
}

// HasErrors checks if there are any validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// GetErrors returns all validation errors
func (v *Validator) GetErrors() map[string][]string {
	return v.errors
}

// GetError returns the first error for a field
func (v *Validator) GetError(field string) string {
	if errors, exists := v.errors[field]; exists && len(errors) > 0 {
		return errors[0]
	}
	return ""
}

// GetErrorsForField returns all errors for a specific field
func (v *Validator) GetErrorsForField(field string) []string {
	if errors, exists := v.errors[field]; exists {
		return errors
	}
	return []string{}
}

// GetFirstError returns the first error from any field
func (v *Validator) GetFirstError() string {
	for _, errors := range v.errors {
		if len(errors) > 0 {
			return errors[0]
		}
	}
	return ""
}

// GetAllErrors returns all errors as a flat slice
func (v *Validator) GetAllErrors() []string {
	var allErrors []string
	for _, errors := range v.errors {
		allErrors = append(allErrors, errors...)
	}
	return allErrors
}

// ClearErrors clears all validation errors
func (v *Validator) ClearErrors() {
	v.errors = make(map[string][]string)
}

// isRequiredRule checks if a rule is a required rule
func isRequiredRule(rule Rule) bool {
	_, ok := rule.(*RequiredRule)
	return ok
}

// Request represents a validated request
type Request struct {
	data      map[string]interface{}
	validator *Validator
	errors    map[string][]string
}

// NewRequest creates a new request with validation
func NewRequest(data map[string]interface{}) *Request {
	return &Request{
		data:      data,
		validator: NewValidator(),
		errors:    make(map[string][]string),
	}
}

// Validate validates the request data
func (r *Request) Validate() bool {
	isValid := r.validator.Validate(r.data)
	r.errors = r.validator.GetErrors()
	return isValid
}

// HasErrors checks if there are any validation errors
func (r *Request) HasErrors() bool {
	return len(r.errors) > 0
}

// GetErrors returns all validation errors
func (r *Request) GetErrors() map[string][]string {
	return r.errors
}

// GetError returns the first error for a field
func (r *Request) GetError(field string) string {
	return r.validator.GetError(field)
}

// GetErrorsForField returns all errors for a specific field
func (r *Request) GetErrorsForField(field string) []string {
	return r.validator.GetErrorsForField(field)
}

// GetFirstError returns the first error from any field
func (r *Request) GetFirstError() string {
	return r.validator.GetFirstError()
}

// GetAllErrors returns all errors as a flat slice
func (r *Request) GetAllErrors() []string {
	return r.validator.GetAllErrors()
}

// Get returns a value from the request data
func (r *Request) Get(field string) interface{} {
	return r.data[field]
}

// GetString returns a string value from the request data
func (r *Request) GetString(field string) string {
	if value, exists := r.data[field]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt returns an integer value from the request data
func (r *Request) GetInt(field string) int {
	if value, exists := r.data[field]; exists {
		switch v := value.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			// Try to parse string to int
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}

// GetFloat returns a float value from the request data
func (r *Request) GetFloat(field string) float64 {
	if value, exists := r.data[field]; exists {
		switch v := value.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case string:
			// Try to parse string to float
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return 0.0
}

// GetBool returns a boolean value from the request data
func (r *Request) GetBool(field string) bool {
	if value, exists := r.data[field]; exists {
		switch v := value.(type) {
		case bool:
			return v
		case string:
			return strings.ToLower(v) == "true" || v == "1"
		case int:
			return v != 0
		}
	}
	return false
}

// GetArray returns an array value from the request data
func (r *Request) GetArray(field string) []interface{} {
	if value, exists := r.data[field]; exists {
		if arr, ok := value.([]interface{}); ok {
			return arr
		}
	}
	return []interface{}{}
}

// GetMap returns a map value from the request data
func (r *Request) GetMap(field string) map[string]interface{} {
	if value, exists := r.data[field]; exists {
		if m, ok := value.(map[string]interface{}); ok {
			return m
		}
	}
	return map[string]interface{}{}
}

// Set sets a value in the request data
func (r *Request) Set(field string, value interface{}) {
	r.data[field] = value
}

// Has checks if a field exists in the request data
func (r *Request) Has(field string) bool {
	_, exists := r.data[field]
	return exists
}

// All returns all request data
func (r *Request) All() map[string]interface{} {
	return r.data
}

// Only returns only the specified fields from request data
func (r *Request) Only(fields ...string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, field := range fields {
		if value, exists := r.data[field]; exists {
			result[field] = value
		}
	}
	return result
}

// Except returns all fields except the specified ones
func (r *Request) Except(fields ...string) map[string]interface{} {
	result := make(map[string]interface{})
	excluded := make(map[string]bool)

	for _, field := range fields {
		excluded[field] = true
	}

	for field, value := range r.data {
		if !excluded[field] {
			result[field] = value
		}
	}

	return result
}

// AddRule adds a validation rule
func (r *Request) AddRule(field string, rule Rule) {
	r.validator.AddRule(field, rule)
}

// AddRules adds multiple validation rules
func (r *Request) AddRules(field string, rules ...Rule) {
	r.validator.AddRules(field, rules...)
}

// Fluent validation methods

// Required marks a field as required
func (r *Request) Required(field string, message ...string) *Request {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	r.AddRule(field, NewRequiredRule(field, msg))
	return r
}

// String validates that a field is a string
func (r *Request) String(field string, message ...string) *Request {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	r.AddRule(field, NewStringRule(field, msg, 0, 0))
	return r
}

// StringWithLength validates that a field is a string with specific length constraints
func (r *Request) StringWithLength(field string, minLength, maxLength int, message ...string) *Request {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	r.AddRule(field, NewStringRule(field, msg, minLength, maxLength))
	return r
}

// Email validates that a field is a valid email
func (r *Request) Email(field string, message ...string) *Request {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	r.AddRule(field, NewEmailRule(field, msg))
	return r
}

// Numeric validates that a field is numeric
func (r *Request) Numeric(field string, message ...string) *Request {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	r.AddRule(field, NewNumericRule(field, msg, 0, 0))
	return r
}

// NumericRange validates that a field is numeric within a range
func (r *Request) NumericRange(field string, min, max float64, message ...string) *Request {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	r.AddRule(field, NewNumericRule(field, msg, min, max))
	return r
}

// Integer validates that a field is an integer
func (r *Request) Integer(field string, message ...string) *Request {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	r.AddRule(field, NewIntegerRule(field, msg, 0, 0))
	return r
}

// IntegerRange validates that a field is an integer within a range
func (r *Request) IntegerRange(field string, min, max int64, message ...string) *Request {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	r.AddRule(field, NewIntegerRule(field, msg, min, max))
	return r
}

// Boolean validates that a field is a boolean
func (r *Request) Boolean(field string, message ...string) *Request {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	r.AddRule(field, NewBooleanRule(field, msg))
	return r
}

// Date validates that a field is a valid date
func (r *Request) Date(field string, format string, message ...string) *Request {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	r.AddRule(field, NewDateRule(field, msg, format))
	return r
}

// In validates that a field value is in a list of allowed values
func (r *Request) In(field string, allowed []interface{}, message ...string) *Request {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	r.AddRule(field, NewInRule(field, msg, allowed))
	return r
}

// NotIn validates that a field value is not in a list of disallowed values
func (r *Request) NotIn(field string, disallowed []interface{}, message ...string) *Request {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	r.AddRule(field, NewNotInRule(field, msg, disallowed))
	return r
}

// Regex validates that a field matches a regular expression
func (r *Request) Regex(field string, pattern string, message ...string) *Request {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	r.AddRule(field, NewRegexRule(field, msg, pattern))
	return r
}

// URL validates that a field is a valid URL
func (r *Request) URL(field string, message ...string) *Request {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	r.AddRule(field, NewURLRule(field, msg))
	return r
}

// UUID validates that a field is a valid UUID
func (r *Request) UUID(field string, message ...string) *Request {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	r.AddRule(field, NewUUIDRule(field, msg))
	return r
}

// Custom adds a custom validation rule
func (r *Request) Custom(field string, rule Rule) *Request {
	r.AddRule(field, rule)
	return r
}

// Helper function to create validation rules from struct tags
func ParseRulesFromStruct(s interface{}) map[string][]Rule {
	rules := make(map[string][]Rule)

	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldName := strings.ToLower(field.Name)

		// Parse validation tags
		if tag := field.Tag.Get("validate"); tag != "" {
			ruleStrings := strings.Split(tag, "|")

			for _, ruleString := range ruleStrings {
				ruleString = strings.TrimSpace(ruleString)
				if ruleString == "" {
					continue
				}

				rule := parseRuleFromString(fieldName, ruleString)
				if rule != nil {
					if rules[fieldName] == nil {
						rules[fieldName] = make([]Rule, 0)
					}
					rules[fieldName] = append(rules[fieldName], rule)
				}
			}
		}
	}

	return rules
}

// parseRuleFromString parses a validation rule from a string
func parseRuleFromString(field, ruleString string) Rule {
	parts := strings.SplitN(ruleString, ":", 2)
	ruleName := parts[0]

	switch ruleName {
	case "required":
		return NewRequiredRule(field, "")
	case "string":
		return NewStringRule(field, "", 0, 0)
	case "email":
		return NewEmailRule(field, "")
	case "numeric":
		return NewNumericRule(field, "", 0, 0)
	case "integer":
		return NewIntegerRule(field, "", 0, 0)
	case "boolean":
		return NewBooleanRule(field, "")
	case "date":
		format := "2006-01-02"
		if len(parts) > 1 {
			format = parts[1]
		}
		return NewDateRule(field, "", format)
	case "url":
		return NewURLRule(field, "")
	case "uuid":
		return NewUUIDRule(field, "")
	default:
		// Handle parameterized rules
		if len(parts) > 1 {
			params := strings.Split(parts[1], ",")

			switch ruleName {
			case "min":
				if len(params) > 0 {
					if min, err := strconv.ParseFloat(params[0], 64); err == nil {
						return NewNumericRule(field, "", min, 0)
					}
				}
			case "max":
				if len(params) > 0 {
					if max, err := strconv.ParseFloat(params[0], 64); err == nil {
						return NewNumericRule(field, "", 0, max)
					}
				}
			case "min_length":
				if len(params) > 0 {
					if min, err := strconv.Atoi(params[0]); err == nil {
						return NewStringRule(field, "", min, 0)
					}
				}
			case "max_length":
				if len(params) > 0 {
					if max, err := strconv.Atoi(params[0]); err == nil {
						return NewStringRule(field, "", 0, max)
					}
				}
			case "regex":
				if len(params) > 0 {
					return NewRegexRule(field, "", params[0])
				}
			}
		}
	}

	return nil
}
