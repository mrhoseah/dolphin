package validation

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Validator defines the interface for validation
type Validator interface {
	Validate(data interface{}) error
	ValidateField(field interface{}, rules []string) error
}

// Rule defines a validation rule
type Rule struct {
	Name    string
	Value   interface{}
	Message string
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError
}

func (e ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "no validation errors"
	}

	var messages []string
	for _, err := range e.Errors {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// HasErrors returns true if there are validation errors
func (e ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// AddError adds a validation error
func (e *ValidationErrors) AddError(field, message string, value interface{}) {
	e.Errors = append(e.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// GetErrors returns all validation errors
func (e ValidationErrors) GetErrors() []ValidationError {
	return e.Errors
}

// FieldValidator validates individual fields
type FieldValidator struct {
	rules map[string]func(interface{}, string) error
}

// NewFieldValidator creates a new field validator
func NewFieldValidator() *FieldValidator {
	v := &FieldValidator{
		rules: make(map[string]func(interface{}, string) error),
	}

	// Register default rules
	v.registerDefaultRules()

	return v
}

// registerDefaultRules registers default validation rules
func (v *FieldValidator) registerDefaultRules() {
	v.rules["required"] = v.validateRequired
	v.rules["email"] = v.validateEmail
	v.rules["min"] = v.validateMin
	v.rules["max"] = v.validateMax
	v.rules["min_length"] = v.validateMinLength
	v.rules["max_length"] = v.validateMaxLength
	v.rules["numeric"] = v.validateNumeric
	v.rules["alpha"] = v.validateAlpha
	v.rules["alpha_numeric"] = v.validateAlphaNumeric
	v.rules["url"] = v.validateURL
	v.rules["date"] = v.validateDate
	v.rules["regex"] = v.validateRegex
	v.rules["in"] = v.validateIn
	v.rules["not_in"] = v.validateNotIn
	v.rules["confirmed"] = v.validateConfirmed
	v.rules["different"] = v.validateDifferent
	v.rules["same"] = v.validateSame
}

// RegisterRule registers a custom validation rule
func (v *FieldValidator) RegisterRule(name string, rule func(interface{}, string) error) {
	v.rules[name] = rule
}

// ValidateField validates a single field with rules
func (v *FieldValidator) ValidateField(field interface{}, rules []string) error {
	var errors ValidationErrors

	for _, rule := range rules {
		ruleParts := strings.Split(rule, ":")
		ruleName := ruleParts[0]
		ruleValue := ""
		if len(ruleParts) > 1 {
			ruleValue = strings.Join(ruleParts[1:], ":")
		}

		if validator, exists := v.rules[ruleName]; exists {
			if err := validator(field, ruleValue); err != nil {
				errors.AddError("field", err.Error(), field)
			}
		} else {
			errors.AddError("field", fmt.Sprintf("unknown validation rule: %s", ruleName), field)
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// Validate validates a struct with validation tags
func (v *FieldValidator) Validate(data interface{}) error {
	var errors ValidationErrors

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("validation target must be a struct")
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get validation rules from struct tag
		rules := fieldType.Tag.Get("validate")
		if rules == "" {
			continue
		}

		ruleList := strings.Split(rules, "|")
		fieldName := fieldType.Name

		// Use json tag name if available
		if jsonTag := fieldType.Tag.Get("json"); jsonTag != "" {
			fieldName = strings.Split(jsonTag, ",")[0]
		}

		for _, rule := range ruleList {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			ruleParts := strings.Split(rule, ":")
			ruleName := ruleParts[0]
			ruleValue := ""
			if len(ruleParts) > 1 {
				ruleValue = strings.Join(ruleParts[1:], ":")
			}

			if validator, exists := v.rules[ruleName]; exists {
				if err := validator(field.Interface(), ruleValue); err != nil {
					errors.AddError(fieldName, err.Error(), field.Interface())
				}
			}
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// Validation rule implementations

func (v *FieldValidator) validateRequired(value interface{}, ruleValue string) error {
	if value == nil {
		return fmt.Errorf("field is required")
	}

	switch val := value.(type) {
	case string:
		if strings.TrimSpace(val) == "" {
			return fmt.Errorf("field is required")
		}
	case int, int8, int16, int32, int64:
		if val == 0 {
			return fmt.Errorf("field is required")
		}
	case uint, uint8, uint16, uint32, uint64:
		if val == 0 {
			return fmt.Errorf("field is required")
		}
	case float32, float64:
		if val == 0.0 {
			return fmt.Errorf("field is required")
		}
	case bool:
		// Bool is always valid, even if false
		return nil
	}

	return nil
}

func (v *FieldValidator) validateEmail(value interface{}, ruleValue string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	if str == "" {
		return nil // Empty string is valid (use required rule for that)
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(str) {
		return fmt.Errorf("field must be a valid email address")
	}

	return nil
}

func (v *FieldValidator) validateMin(value interface{}, ruleValue string) error {
	min, err := strconv.ParseFloat(ruleValue, 64)
	if err != nil {
		return fmt.Errorf("invalid min rule value: %s", ruleValue)
	}

	switch val := value.(type) {
	case int:
		if float64(val) < min {
			return fmt.Errorf("field must be at least %v", min)
		}
	case int8:
		if float64(val) < min {
			return fmt.Errorf("field must be at least %v", min)
		}
	case int16:
		if float64(val) < min {
			return fmt.Errorf("field must be at least %v", min)
		}
	case int32:
		if float64(val) < min {
			return fmt.Errorf("field must be at least %v", min)
		}
	case int64:
		if float64(val) < min {
			return fmt.Errorf("field must be at least %v", min)
		}
	case uint:
		if float64(val) < min {
			return fmt.Errorf("field must be at least %v", min)
		}
	case uint8:
		if float64(val) < min {
			return fmt.Errorf("field must be at least %v", min)
		}
	case uint16:
		if float64(val) < min {
			return fmt.Errorf("field must be at least %v", min)
		}
	case uint32:
		if float64(val) < min {
			return fmt.Errorf("field must be at least %v", min)
		}
	case uint64:
		if float64(val) < min {
			return fmt.Errorf("field must be at least %v", min)
		}
	case float32:
		if float64(val) < min {
			return fmt.Errorf("field must be at least %v", min)
		}
	case float64:
		if val < min {
			return fmt.Errorf("field must be at least %v", min)
		}
	default:
		return fmt.Errorf("field must be a number")
	}

	return nil
}

func (v *FieldValidator) validateMax(value interface{}, ruleValue string) error {
	max, err := strconv.ParseFloat(ruleValue, 64)
	if err != nil {
		return fmt.Errorf("invalid max rule value: %s", ruleValue)
	}

	switch val := value.(type) {
	case int:
		if float64(val) > max {
			return fmt.Errorf("field must be at most %v", max)
		}
	case int8:
		if float64(val) > max {
			return fmt.Errorf("field must be at most %v", max)
		}
	case int16:
		if float64(val) > max {
			return fmt.Errorf("field must be at most %v", max)
		}
	case int32:
		if float64(val) > max {
			return fmt.Errorf("field must be at most %v", max)
		}
	case int64:
		if float64(val) > max {
			return fmt.Errorf("field must be at most %v", max)
		}
	case uint:
		if float64(val) > max {
			return fmt.Errorf("field must be at most %v", max)
		}
	case uint8:
		if float64(val) > max {
			return fmt.Errorf("field must be at most %v", max)
		}
	case uint16:
		if float64(val) > max {
			return fmt.Errorf("field must be at most %v", max)
		}
	case uint32:
		if float64(val) > max {
			return fmt.Errorf("field must be at most %v", max)
		}
	case uint64:
		if float64(val) > max {
			return fmt.Errorf("field must be at most %v", max)
		}
	case float32:
		if float64(val) > max {
			return fmt.Errorf("field must be at most %v", max)
		}
	case float64:
		if val > max {
			return fmt.Errorf("field must be at most %v", max)
		}
	default:
		return fmt.Errorf("field must be a number")
	}

	return nil
}

func (v *FieldValidator) validateMinLength(value interface{}, ruleValue string) error {
	minLen, err := strconv.Atoi(ruleValue)
	if err != nil {
		return fmt.Errorf("invalid min_length rule value: %s", ruleValue)
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	if len(str) < minLen {
		return fmt.Errorf("field must be at least %d characters long", minLen)
	}

	return nil
}

func (v *FieldValidator) validateMaxLength(value interface{}, ruleValue string) error {
	maxLen, err := strconv.Atoi(ruleValue)
	if err != nil {
		return fmt.Errorf("invalid max_length rule value: %s", ruleValue)
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	if len(str) > maxLen {
		return fmt.Errorf("field must be at most %d characters long", maxLen)
	}

	return nil
}

func (v *FieldValidator) validateNumeric(value interface{}, ruleValue string) error {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return nil
	case string:
		str := value.(string)
		if str == "" {
			return nil // Empty string is valid (use required rule for that)
		}
		_, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fmt.Errorf("field must be numeric")
		}
		return nil
	default:
		return fmt.Errorf("field must be numeric")
	}
}

func (v *FieldValidator) validateAlpha(value interface{}, ruleValue string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	if str == "" {
		return nil // Empty string is valid (use required rule for that)
	}

	for _, char := range str {
		if !unicode.IsLetter(char) {
			return fmt.Errorf("field must contain only letters")
		}
	}

	return nil
}

func (v *FieldValidator) validateAlphaNumeric(value interface{}, ruleValue string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	if str == "" {
		return nil // Empty string is valid (use required rule for that)
	}

	for _, char := range str {
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
			return fmt.Errorf("field must contain only letters and numbers")
		}
	}

	return nil
}

func (v *FieldValidator) validateURL(value interface{}, ruleValue string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	if str == "" {
		return nil // Empty string is valid (use required rule for that)
	}

	_, err := url.ParseRequestURI(str)
	if err != nil {
		return fmt.Errorf("field must be a valid URL")
	}

	return nil
}

func (v *FieldValidator) validateDate(value interface{}, ruleValue string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	if str == "" {
		return nil // Empty string is valid (use required rule for that)
	}

	format := "2006-01-02"
	if ruleValue != "" {
		format = ruleValue
	}

	_, err := time.Parse(format, str)
	if err != nil {
		return fmt.Errorf("field must be a valid date in format %s", format)
	}

	return nil
}

func (v *FieldValidator) validateRegex(value interface{}, ruleValue string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	if str == "" {
		return nil // Empty string is valid (use required rule for that)
	}

	regex, err := regexp.Compile(ruleValue)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %s", ruleValue)
	}

	if !regex.MatchString(str) {
		return fmt.Errorf("field does not match required pattern")
	}

	return nil
}

func (v *FieldValidator) validateIn(value interface{}, ruleValue string) error {
	if ruleValue == "" {
		return fmt.Errorf("in rule requires a list of values")
	}

	allowedValues := strings.Split(ruleValue, ",")
	for i, val := range allowedValues {
		allowedValues[i] = strings.TrimSpace(val)
	}

	str := fmt.Sprintf("%v", value)
	for _, allowed := range allowedValues {
		if str == allowed {
			return nil
		}
	}

	return fmt.Errorf("field must be one of: %s", strings.Join(allowedValues, ", "))
}

func (v *FieldValidator) validateNotIn(value interface{}, ruleValue string) error {
	if ruleValue == "" {
		return fmt.Errorf("not_in rule requires a list of values")
	}

	forbiddenValues := strings.Split(ruleValue, ",")
	for i, val := range forbiddenValues {
		forbiddenValues[i] = strings.TrimSpace(val)
	}

	str := fmt.Sprintf("%v", value)
	for _, forbidden := range forbiddenValues {
		if str == forbidden {
			return fmt.Errorf("field must not be one of: %s", strings.Join(forbiddenValues, ", "))
		}
	}

	return nil
}

func (v *FieldValidator) validateConfirmed(value interface{}, ruleValue string) error {
	// This is a placeholder - confirmation validation typically requires
	// comparing two fields (e.g., password and password_confirmation)
	// This would need to be implemented at the struct level
	return nil
}

func (v *FieldValidator) validateDifferent(value interface{}, ruleValue string) error {
	// This is a placeholder - different validation typically requires
	// comparing two fields
	// This would need to be implemented at the struct level
	return nil
}

func (v *FieldValidator) validateSame(value interface{}, ruleValue string) error {
	// This is a placeholder - same validation typically requires
	// comparing two fields
	// This would need to be implemented at the struct level
	return nil
}
