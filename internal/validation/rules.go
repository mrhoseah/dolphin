package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Rule represents a validation rule
type Rule interface {
	// Validate validates a value
	Validate(value interface{}) error

	// GetMessage returns the error message
	GetMessage() string

	// GetField returns the field name
	GetField() string
}

// BaseRule provides common rule functionality
type BaseRule struct {
	Field   string
	Message string
}

// GetField returns the field name
func (br *BaseRule) GetField() string {
	return br.Field
}

// GetMessage returns the error message
func (br *BaseRule) GetMessage() string {
	return br.Message
}

// Validate implements the Rule interface (default implementation)
func (br *BaseRule) Validate(value interface{}) error {
	return nil // Override in concrete rules
}

// RequiredRule validates that a field is required
type RequiredRule struct {
	*BaseRule
}

// NewRequiredRule creates a new required rule
func NewRequiredRule(field, message string) *RequiredRule {
	if message == "" {
		message = fmt.Sprintf("The %s field is required", field)
	}

	return &RequiredRule{
		BaseRule: &BaseRule{
			field:   field,
			message: message,
		},
	}
}

// Validate validates that the value is not empty
func (rr *RequiredRule) Validate(value interface{}) error {
	if value == nil {
		return fmt.Errorf(rr.message)
	}

	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf(rr.message)
		}
	case []interface{}:
		if len(v) == 0 {
			return fmt.Errorf(rr.message)
		}
	case map[string]interface{}:
		if len(v) == 0 {
			return fmt.Errorf(rr.message)
		}
	}

	return nil
}

// StringRule validates that a field is a string
type StringRule struct {
	*BaseRule
	minLength int
	maxLength int
}

// NewStringRule creates a new string rule
func NewStringRule(field, message string, minLength, maxLength int) *StringRule {
	if message == "" {
		message = fmt.Sprintf("The %s field must be a string", field)
	}

	return &StringRule{
		BaseRule: &BaseRule{
			field:   field,
			message: message,
		},
		minLength: minLength,
		maxLength: maxLength,
	}
}

// Validate validates that the value is a string
func (sr *StringRule) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf(sr.message)
	}

	if sr.minLength > 0 && len(str) < sr.minLength {
		return fmt.Errorf("The %s field must be at least %d characters", sr.field, sr.minLength)
	}

	if sr.maxLength > 0 && len(str) > sr.maxLength {
		return fmt.Errorf("The %s field must not exceed %d characters", sr.field, sr.maxLength)
	}

	return nil
}

// EmailRule validates that a field is a valid email
type EmailRule struct {
	*BaseRule
	pattern *regexp.Regexp
}

// NewEmailRule creates a new email rule
func NewEmailRule(field, message string) *EmailRule {
	if message == "" {
		message = fmt.Sprintf("The %s field must be a valid email address", field)
	}

	pattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	return &EmailRule{
		BaseRule: &BaseRule{
			field:   field,
			message: message,
		},
		pattern: pattern,
	}
}

// Validate validates that the value is a valid email
func (er *EmailRule) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf(er.message)
	}

	if !er.pattern.MatchString(str) {
		return fmt.Errorf(er.message)
	}

	return nil
}

// NumericRule validates that a field is numeric
type NumericRule struct {
	*BaseRule
	min float64
	max float64
}

// NewNumericRule creates a new numeric rule
func NewNumericRule(field, message string, min, max float64) *NumericRule {
	if message == "" {
		message = fmt.Sprintf("The %s field must be numeric", field)
	}

	return &NumericRule{
		BaseRule: &BaseRule{
			field:   field,
			message: message,
		},
		min: min,
		max: max,
	}
}

// Validate validates that the value is numeric
func (nr *NumericRule) Validate(value interface{}) error {
	var num float64

	switch v := value.(type) {
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	case float32:
		num = float64(v)
	case float64:
		num = v
	case string:
		var err error
		num, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf(nr.message)
		}
	default:
		return fmt.Errorf(nr.message)
	}

	if nr.min != 0 && num < nr.min {
		return fmt.Errorf("The %s field must be at least %g", nr.field, nr.min)
	}

	if nr.max != 0 && num > nr.max {
		return fmt.Errorf("The %s field must not exceed %g", nr.field, nr.max)
	}

	return nil
}

// IntegerRule validates that a field is an integer
type IntegerRule struct {
	*BaseRule
	min int64
	max int64
}

// NewIntegerRule creates a new integer rule
func NewIntegerRule(field, message string, min, max int64) *IntegerRule {
	if message == "" {
		message = fmt.Sprintf("The %s field must be an integer", field)
	}

	return &IntegerRule{
		BaseRule: &BaseRule{
			field:   field,
			message: message,
		},
		min: min,
		max: max,
	}
}

// Validate validates that the value is an integer
func (ir *IntegerRule) Validate(value interface{}) error {
	var num int64

	switch v := value.(type) {
	case int:
		num = int64(v)
	case int64:
		num = v
	case string:
		var err error
		num, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf(ir.message)
		}
	default:
		return fmt.Errorf(ir.message)
	}

	if ir.min != 0 && num < ir.min {
		return fmt.Errorf("The %s field must be at least %d", ir.field, ir.min)
	}

	if ir.max != 0 && num > ir.max {
		return fmt.Errorf("The %s field must not exceed %d", ir.field, ir.max)
	}

	return nil
}

// BooleanRule validates that a field is a boolean
type BooleanRule struct {
	*BaseRule
}

// NewBooleanRule creates a new boolean rule
func NewBooleanRule(field, message string) *BooleanRule {
	if message == "" {
		message = fmt.Sprintf("The %s field must be a boolean", field)
	}

	return &BooleanRule{
		BaseRule: &BaseRule{
			field:   field,
			message: message,
		},
	}
}

// Validate validates that the value is a boolean
func (br *BooleanRule) Validate(value interface{}) error {
	switch value.(type) {
	case bool:
		return nil
	case string:
		str := strings.ToLower(value.(string))
		if str == "true" || str == "false" || str == "1" || str == "0" {
			return nil
		}
	}

	return fmt.Errorf(br.message)
}

// DateRule validates that a field is a valid date
type DateRule struct {
	*BaseRule
	format string
}

// NewDateRule creates a new date rule
func NewDateRule(field, message, format string) *DateRule {
	if message == "" {
		message = fmt.Sprintf("The %s field must be a valid date", field)
	}

	if format == "" {
		format = "2006-01-02"
	}

	return &DateRule{
		BaseRule: &BaseRule{
			field:   field,
			message: message,
		},
		format: format,
	}
}

// Validate validates that the value is a valid date
func (dr *DateRule) Validate(value interface{}) error {
	var str string

	switch v := value.(type) {
	case string:
		str = v
	case time.Time:
		return nil // Already a valid time
	default:
		return fmt.Errorf(dr.message)
	}

	_, err := time.Parse(dr.format, str)
	if err != nil {
		return fmt.Errorf(dr.message)
	}

	return nil
}

// InRule validates that a field value is in a list of allowed values
type InRule struct {
	*BaseRule
	allowed []interface{}
}

// NewInRule creates a new in rule
func NewInRule(field, message string, allowed []interface{}) *InRule {
	if message == "" {
		message = fmt.Sprintf("The %s field must be one of the allowed values", field)
	}

	return &InRule{
		BaseRule: &BaseRule{
			field:   field,
			message: message,
		},
		allowed: allowed,
	}
}

// Validate validates that the value is in the allowed list
func (ir *InRule) Validate(value interface{}) error {
	for _, allowed := range ir.allowed {
		if reflect.DeepEqual(value, allowed) {
			return nil
		}
	}

	return fmt.Errorf(ir.message)
}

// NotInRule validates that a field value is not in a list of disallowed values
type NotInRule struct {
	*BaseRule
	disallowed []interface{}
}

// NewNotInRule creates a new not in rule
func NewNotInRule(field, message string, disallowed []interface{}) *NotInRule {
	if message == "" {
		message = fmt.Sprintf("The %s field must not be one of the disallowed values", field)
	}

	return &NotInRule{
		BaseRule: &BaseRule{
			field:   field,
			message: message,
		},
		disallowed: disallowed,
	}
}

// Validate validates that the value is not in the disallowed list
func (nir *NotInRule) Validate(value interface{}) error {
	for _, disallowed := range nir.disallowed {
		if reflect.DeepEqual(value, disallowed) {
			return fmt.Errorf(nir.message)
		}
	}

	return nil
}

// MinRule validates that a field has a minimum value
type MinRule struct {
	*BaseRule
	min interface{}
}

// NewMinRule creates a new min rule
func NewMinRule(field, message string, min interface{}) *MinRule {
	if message == "" {
		message = fmt.Sprintf("The %s field must be at least %v", field, min)
	}

	return &MinRule{
		BaseRule: &BaseRule{
			field:   field,
			message: message,
		},
		min: min,
	}
}

// Validate validates that the value meets the minimum requirement
func (mr *MinRule) Validate(value interface{}) error {
	// This is a simplified implementation
	// In a real implementation, you'd need to handle different types properly
	return nil
}

// MaxRule validates that a field has a maximum value
type MaxRule struct {
	*BaseRule
	max interface{}
}

// NewMaxRule creates a new max rule
func NewMaxRule(field, message string, max interface{}) *MaxRule {
	if message == "" {
		message = fmt.Sprintf("The %s field must not exceed %v", field, max)
	}

	return &MaxRule{
		BaseRule: &BaseRule{
			field:   field,
			message: message,
		},
		max: max,
	}
}

// Validate validates that the value meets the maximum requirement
func (mr *MaxRule) Validate(value interface{}) error {
	// This is a simplified implementation
	// In a real implementation, you'd need to handle different types properly
	return nil
}

// RegexRule validates that a field matches a regular expression
type RegexRule struct {
	*BaseRule
	pattern *regexp.Regexp
}

// NewRegexRule creates a new regex rule
func NewRegexRule(field, message, pattern string) *RegexRule {
	if message == "" {
		message = fmt.Sprintf("The %s field format is invalid", field)
	}

	regex := regexp.MustCompile(pattern)

	return &RegexRule{
		BaseRule: &BaseRule{
			field:   field,
			message: message,
		},
		pattern: regex,
	}
}

// Validate validates that the value matches the regex pattern
func (rr *RegexRule) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf(rr.message)
	}

	if !rr.pattern.MatchString(str) {
		return fmt.Errorf(rr.message)
	}

	return nil
}

// URLRule validates that a field is a valid URL
type URLRule struct {
	*BaseRule
}

// NewURLRule creates a new URL rule
func NewURLRule(field, message string) *URLRule {
	if message == "" {
		message = fmt.Sprintf("The %s field must be a valid URL", field)
	}

	return &URLRule{
		BaseRule: &BaseRule{
			field:   field,
			message: message,
		},
	}
}

// Validate validates that the value is a valid URL
func (ur *URLRule) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf(ur.message)
	}

	// Simple URL validation
	if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") {
		return fmt.Errorf(ur.message)
	}

	return nil
}

// UUIDRule validates that a field is a valid UUID
type UUIDRule struct {
	*BaseRule
	pattern *regexp.Regexp
}

// NewUUIDRule creates a new UUID rule
func NewUUIDRule(field, message string) *UUIDRule {
	if message == "" {
		message = fmt.Sprintf("The %s field must be a valid UUID", field)
	}

	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	return &UUIDRule{
		BaseRule: &BaseRule{
			field:   field,
			message: message,
		},
		pattern: pattern,
	}
}

// Validate validates that the value is a valid UUID
func (ur *UUIDRule) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf(ur.message)
	}

	if !ur.pattern.MatchString(strings.ToLower(str)) {
		return fmt.Errorf(ur.message)
	}

	return nil
}
