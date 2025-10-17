package validation

import (
	"fmt"
	"html"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Sanitizer defines the interface for data sanitization
type Sanitizer interface {
	Sanitize(data interface{}) error
	SanitizeField(field interface{}, rules []string) (interface{}, error)
}

// FieldSanitizer sanitizes individual fields
type FieldSanitizer struct {
	rules map[string]func(interface{}, string) (interface{}, error)
}

// NewFieldSanitizer creates a new field sanitizer
func NewFieldSanitizer() *FieldSanitizer {
	s := &FieldSanitizer{
		rules: make(map[string]func(interface{}, string) (interface{}, error)),
	}

	// Register default sanitization rules
	s.registerDefaultRules()

	return s
}

// registerDefaultRules registers default sanitization rules
func (s *FieldSanitizer) registerDefaultRules() {
	s.rules["trim"] = s.sanitizeTrim
	s.rules["lowercase"] = s.sanitizeLowercase
	s.rules["uppercase"] = s.sanitizeUppercase
	s.rules["escape_html"] = s.sanitizeEscapeHTML
	s.rules["unescape_html"] = s.sanitizeUnescapeHTML
	s.rules["strip_html"] = s.sanitizeStripHTML
	s.rules["strip_whitespace"] = s.sanitizeStripWhitespace
	s.rules["normalize_whitespace"] = s.sanitizeNormalizeWhitespace
	s.rules["remove_special_chars"] = s.sanitizeRemoveSpecialChars
	s.rules["keep_alphanumeric"] = s.sanitizeKeepAlphanumeric
	s.rules["normalize_email"] = s.sanitizeNormalizeEmail
	s.rules["normalize_phone"] = s.sanitizeNormalizePhone
	s.rules["slug"] = s.sanitizeSlug
	s.rules["limit_length"] = s.sanitizeLimitLength
	s.rules["remove_emojis"] = s.sanitizeRemoveEmojis
	s.rules["normalize_unicode"] = s.sanitizeNormalizeUnicode
}

// RegisterRule registers a custom sanitization rule
func (s *FieldSanitizer) RegisterRule(name string, rule func(interface{}, string) (interface{}, error)) {
	s.rules[name] = rule
}

// SanitizeField sanitizes a single field with rules
func (s *FieldSanitizer) SanitizeField(field interface{}, rules []string) (interface{}, error) {
	result := field

	for _, rule := range rules {
		ruleParts := strings.Split(rule, ":")
		ruleName := ruleParts[0]
		ruleValue := ""
		if len(ruleParts) > 1 {
			ruleValue = strings.Join(ruleParts[1:], ":")
		}

		if sanitizer, exists := s.rules[ruleName]; exists {
			var err error
			result, err = sanitizer(result, ruleValue)
			if err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

// Sanitize sanitizes a struct with sanitization tags
func (s *FieldSanitizer) Sanitize(data interface{}) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("sanitization target must be a struct")
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanInterface() || !field.CanSet() {
			continue
		}

		// Get sanitization rules from struct tag
		rules := fieldType.Tag.Get("sanitize")
		if rules == "" {
			continue
		}

		ruleList := strings.Split(rules, "|")

		result := field.Interface()
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

			if sanitizer, exists := s.rules[ruleName]; exists {
				var err error
				result, err = sanitizer(result, ruleValue)
				if err != nil {
					return err
				}
			}
		}

		// Set the sanitized value back to the field
		field.Set(reflect.ValueOf(result))
	}

	return nil
}

// Sanitization rule implementations

func (s *FieldSanitizer) sanitizeTrim(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	return strings.TrimSpace(str), nil
}

func (s *FieldSanitizer) sanitizeLowercase(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	return strings.ToLower(str), nil
}

func (s *FieldSanitizer) sanitizeUppercase(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	return strings.ToUpper(str), nil
}

func (s *FieldSanitizer) sanitizeEscapeHTML(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	return html.EscapeString(str), nil
}

func (s *FieldSanitizer) sanitizeUnescapeHTML(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	return html.UnescapeString(str), nil
}

func (s *FieldSanitizer) sanitizeStripHTML(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	// Simple HTML tag removal
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	return htmlRegex.ReplaceAllString(str, ""), nil
}

func (s *FieldSanitizer) sanitizeStripWhitespace(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	// Remove all whitespace characters
	whitespaceRegex := regexp.MustCompile(`\s+`)
	return whitespaceRegex.ReplaceAllString(str, ""), nil
}

func (s *FieldSanitizer) sanitizeNormalizeWhitespace(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	// Replace multiple whitespace characters with single space
	whitespaceRegex := regexp.MustCompile(`\s+`)
	return whitespaceRegex.ReplaceAllString(strings.TrimSpace(str), " "), nil
}

func (s *FieldSanitizer) sanitizeRemoveSpecialChars(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	// Keep only alphanumeric characters and spaces
	specialCharsRegex := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	return specialCharsRegex.ReplaceAllString(str, ""), nil
}

func (s *FieldSanitizer) sanitizeKeepAlphanumeric(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	// Keep only alphanumeric characters
	alphanumericRegex := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return alphanumericRegex.ReplaceAllString(str, ""), nil
}

func (s *FieldSanitizer) sanitizeNormalizeEmail(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	// Convert to lowercase and trim whitespace
	return strings.TrimSpace(strings.ToLower(str)), nil
}

func (s *FieldSanitizer) sanitizeNormalizePhone(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	// Remove all non-digit characters
	phoneRegex := regexp.MustCompile(`\D`)
	return phoneRegex.ReplaceAllString(str, ""), nil
}

func (s *FieldSanitizer) sanitizeSlug(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	// Convert to lowercase
	str = strings.ToLower(str)

	// Replace spaces and special characters with hyphens
	str = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(str, "")
	str = regexp.MustCompile(`[\s_-]+`).ReplaceAllString(str, "-")
	str = strings.Trim(str, "-")

	return str, nil
}

func (s *FieldSanitizer) sanitizeLimitLength(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	if ruleValue == "" {
		return str, nil
	}

	maxLen, err := strconv.Atoi(ruleValue)
	if err != nil {
		return str, fmt.Errorf("invalid limit_length value: %s", ruleValue)
	}

	if len(str) > maxLen {
		return str[:maxLen], nil
	}

	return str, nil
}

func (s *FieldSanitizer) sanitizeRemoveEmojis(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	// Remove emoji and other Unicode symbols
	var result strings.Builder
	for _, r := range str {
		if !unicode.IsSymbol(r) && !isEmoji(r) {
			result.WriteRune(r)
		}
	}

	return result.String(), nil
}

func (s *FieldSanitizer) sanitizeNormalizeUnicode(value interface{}, ruleValue string) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	// Normalize Unicode characters
	// This is a simplified implementation
	// In a real implementation, you'd use golang.org/x/text/unicode/norm
	return strings.ToLower(str), nil
}

// isEmoji checks if a rune is an emoji
func isEmoji(r rune) bool {
	return r >= 0x1F600 && r <= 0x1F64F || // Emoticons
		r >= 0x1F300 && r <= 0x1F5FF || // Misc Symbols and Pictographs
		r >= 0x1F680 && r <= 0x1F6FF || // Transport and Map
		r >= 0x1F1E0 && r <= 0x1F1FF || // Regional indicator symbols
		r >= 0x2600 && r <= 0x26FF || // Misc symbols
		r >= 0x2700 && r <= 0x27BF || // Dingbats
		r >= 0xFE00 && r <= 0xFE0F || // Variation Selectors
		r >= 0x1F900 && r <= 0x1F9FF || // Supplemental Symbols and Pictographs
		r >= 0x1F018 && r <= 0x1F270 || // Various other emoji ranges
		r == 0x200D || // Zero width joiner
		r == 0x200C || // Zero width non-joiner
		r == 0xFE0F || // Variation selector-16
		r == 0x1F1E6 && r <= 0x1F1FF // Regional indicator symbols
}

// RequestSanitizer provides high-level request sanitization
type RequestSanitizer struct {
	fieldSanitizer *FieldSanitizer
}

// NewRequestSanitizer creates a new request sanitizer
func NewRequestSanitizer() *RequestSanitizer {
	return &RequestSanitizer{
		fieldSanitizer: NewFieldSanitizer(),
	}
}

// SanitizeRequest sanitizes common request fields
func (rs *RequestSanitizer) SanitizeRequest(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for key, value := range data {
		switch v := value.(type) {
		case string:
			// Apply common string sanitizations
			sanitized[key] = rs.sanitizeString(v)
		case map[string]interface{}:
			// Recursively sanitize nested objects
			sanitized[key] = rs.SanitizeRequest(v)
		case []interface{}:
			// Sanitize array elements
			sanitizedArray := make([]interface{}, len(v))
			for i, item := range v {
				if str, ok := item.(string); ok {
					sanitizedArray[i] = rs.sanitizeString(str)
				} else {
					sanitizedArray[i] = item
				}
			}
			sanitized[key] = sanitizedArray
		default:
			sanitized[key] = value
		}
	}

	return sanitized
}

// sanitizeString applies common string sanitizations
func (rs *RequestSanitizer) sanitizeString(str string) string {
	// Trim whitespace
	str = strings.TrimSpace(str)

	// Normalize whitespace
	str = regexp.MustCompile(`\s+`).ReplaceAllString(str, " ")

	// Escape HTML
	str = html.EscapeString(str)

	return str
}

// SanitizeFormData sanitizes form data
func (rs *RequestSanitizer) SanitizeFormData(data map[string][]string) map[string][]string {
	sanitized := make(map[string][]string)

	for key, values := range data {
		sanitizedValues := make([]string, len(values))
		for i, value := range values {
			sanitizedValues[i] = rs.sanitizeString(value)
		}
		sanitized[key] = sanitizedValues
	}

	return sanitized
}
