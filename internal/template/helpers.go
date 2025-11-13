package template

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

// TemplateHelpers provides helper functions for Fin templates
type TemplateHelpers struct {
	request *http.Request
}

// NewTemplateHelpers creates a new template helpers instance
func NewTemplateHelpers(r *http.Request) *TemplateHelpers {
	return &TemplateHelpers{request: r}
}

// GetTemplateHelpers returns a map of helper functions for Fin templates
func GetTemplateHelpers(r *http.Request) template.FuncMap {
	helpers := NewTemplateHelpers(r)
	return template.FuncMap{
		// URL helpers
		"url":      helpers.URL,
		"route":    helpers.Route,
		"asset":    helpers.Asset,
		"secure_asset": helpers.SecureAsset,
		
		// String helpers
		"str_limit":    helpers.StrLimit,
		"str_contains": helpers.StrContains,
		"str_starts_with": helpers.StrStartsWith,
		"str_ends_with": helpers.StrEndsWith,
		"str_replace": helpers.StrReplace,
		"str_slug": helpers.StrSlug,
		"str_title": helpers.StrTitle,
		"str_upper": helpers.StrUpper,
		"str_lower": helpers.StrLower,
		"str_plural": helpers.StrPlural,
		"str_singular": helpers.StrSingular,
		
		// Array/Collection helpers
		"count": helpers.Count,
		"empty": helpers.Empty,
		"first": helpers.First,
		"last": helpers.Last,
		"join": helpers.Join,
		"in_array": helpers.InArray,
		
		// Date/Time helpers
		"now": helpers.Now,
		"date": helpers.Date,
		"date_format": helpers.DateFormat,
		"time_ago": helpers.TimeAgo,
		"time_diff": helpers.TimeDiff,
		
		// Number helpers
		"number_format": helpers.NumberFormat,
		"currency": helpers.Currency,
		
		// HTML helpers
		"e": helpers.Escape,
		"raw": helpers.Raw,
		"old": helpers.Old,
		"checked": helpers.Checked,
		"selected": helpers.Selected,
		
		// Auth helpers
		"auth": helpers.Auth,
		"guest": helpers.Guest,
		"user": helpers.User,
		
		// CSRF helpers
		"csrf_token": helpers.CSRFToken,
		"csrf_field": helpers.CSRFField,
		
		// Form helpers
		"method_field": helpers.MethodField,
		
		// Conditional helpers
		"if_helper": helpers.If,
		"unless":    helpers.Unless,
		"isset":     helpers.Isset,
		
		// JSON helpers
		"json": helpers.JSON,
		"json_encode": helpers.JSONEncode,
		"json_decode": helpers.JSONDecode,
	},
}

// URL generates a URL
func (th *TemplateHelpers) URL(path string) string {
	if th.request == nil {
		return path
	}
	scheme := "http"
	if th.request.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s", scheme, th.request.Host, path)
}

// Route generates a URL for a named route (placeholder)
func (th *TemplateHelpers) Route(name string, params ...interface{}) string {
	// In a real implementation, this would look up the route by name
	return fmt.Sprintf("/route/%s", name)
}

// Asset generates an asset URL
func (th *TemplateHelpers) Asset(path string) string {
	return fmt.Sprintf("/assets/%s", strings.TrimPrefix(path, "/"))
}

// SecureAsset generates a secure asset URL
func (th *TemplateHelpers) SecureAsset(path string) string {
	if th.request == nil {
		return th.Asset(path)
	}
	scheme := "https"
	return fmt.Sprintf("%s://%s/assets/%s", scheme, th.request.Host, strings.TrimPrefix(path, "/"))
}

// StrLimit limits a string to a certain length
func (th *TemplateHelpers) StrLimit(str string, limit int) string {
	if len(str) <= limit {
		return str
	}
	return str[:limit] + "..."
}

// StrContains checks if a string contains a substring
func (th *TemplateHelpers) StrContains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

// StrStartsWith checks if a string starts with a prefix
func (th *TemplateHelpers) StrStartsWith(str, prefix string) bool {
	return strings.HasPrefix(str, prefix)
}

// StrEndsWith checks if a string ends with a suffix
func (th *TemplateHelpers) StrEndsWith(str, suffix string) bool {
	return strings.HasSuffix(str, suffix)
}

// StrReplace replaces occurrences in a string
func (th *TemplateHelpers) StrReplace(search, replace, subject string) string {
	return strings.ReplaceAll(subject, search, replace)
}

// StrSlug converts a string to a URL-friendly slug
func (th *TemplateHelpers) StrSlug(str string) string {
	// Simple slug implementation
	str = strings.ToLower(str)
	str = strings.ReplaceAll(str, " ", "-")
	return str
}

// StrTitle converts a string to title case
func (th *TemplateHelpers) StrTitle(str string) string {
	return strings.Title(str)
}

// StrUpper converts a string to uppercase
func (th *TemplateHelpers) StrUpper(str string) string {
	return strings.ToUpper(str)
}

// StrLower converts a string to lowercase
func (th *TemplateHelpers) StrLower(str string) string {
	return strings.ToLower(str)
}

// StrPlural pluralizes a word (simple implementation)
func (th *TemplateHelpers) StrPlural(word string) string {
	if strings.HasSuffix(word, "y") {
		return strings.TrimSuffix(word, "y") + "ies"
	}
	if strings.HasSuffix(word, "s") || strings.HasSuffix(word, "x") || strings.HasSuffix(word, "z") {
		return word + "es"
	}
	return word + "s"
}

// StrSingular singularizes a word (simple implementation)
func (th *TemplateHelpers) StrSingular(word string) string {
	if strings.HasSuffix(word, "ies") {
		return strings.TrimSuffix(word, "ies") + "y"
	}
	if strings.HasSuffix(word, "es") {
		return strings.TrimSuffix(word, "es")
	}
	if strings.HasSuffix(word, "s") {
		return strings.TrimSuffix(word, "s")
	}
	return word
}

// Count returns the count of items
func (th *TemplateHelpers) Count(items interface{}) int {
	switch v := items.(type) {
	case []interface{}:
		return len(v)
	case map[string]interface{}:
		return len(v)
	case string:
		return len(v)
	default:
		return 0
	}
}

// Empty checks if a value is empty
func (th *TemplateHelpers) Empty(value interface{}) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

// First returns the first element
func (th *TemplateHelpers) First(items interface{}) interface{} {
	switch v := items.(type) {
	case []interface{}:
		if len(v) > 0 {
			return v[0]
		}
	}
	return nil
}

// Last returns the last element
func (th *TemplateHelpers) Last(items interface{}) interface{} {
	switch v := items.(type) {
	case []interface{}:
		if len(v) > 0 {
			return v[len(v)-1]
		}
	}
	return nil
}

// Join joins array elements with a string
func (th *TemplateHelpers) Join(items interface{}, separator string) string {
	switch v := items.(type) {
	case []string:
		return strings.Join(v, separator)
	case []interface{}:
		strs := make([]string, len(v))
		for i, item := range v {
			strs[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(strs, separator)
	}
	return ""
}

// InArray checks if a value exists in an array
func (th *TemplateHelpers) InArray(needle interface{}, haystack interface{}) bool {
	switch h := haystack.(type) {
	case []interface{}:
		for _, item := range h {
			if item == needle {
				return true
			}
		}
	case []string:
		needleStr := fmt.Sprintf("%v", needle)
		for _, item := range h {
			if item == needleStr {
				return true
			}
		}
	}
	return false
}

// Now returns the current time
func (th *TemplateHelpers) Now() time.Time {
	return time.Now()
}

// Date formats a date
func (th *TemplateHelpers) Date(format string, t time.Time) string {
	// Simple date formatting
	return t.Format(format)
}

// DateFormat formats a date with a format string
func (th *TemplateHelpers) DateFormat(format string, t time.Time) string {
	return t.Format(format)
}

// TimeAgo returns a human-readable time ago string
func (th *TemplateHelpers) TimeAgo(t time.Time) string {
	duration := time.Since(t)
	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := int(duration.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}

// TimeDiff returns the difference between two times
func (th *TemplateHelpers) TimeDiff(t1, t2 time.Time) time.Duration {
	return t2.Sub(t1)
}

// NumberFormat formats a number
func (th *TemplateHelpers) NumberFormat(number float64, decimals int) string {
	return fmt.Sprintf("%.*f", decimals, number)
}

// Currency formats a number as currency
func (th *TemplateHelpers) Currency(amount float64, currency string) string {
	switch currency {
	case "USD":
		return fmt.Sprintf("$%.2f", amount)
	case "EUR":
		return fmt.Sprintf("€%.2f", amount)
	case "GBP":
		return fmt.Sprintf("£%.2f", amount)
	default:
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
}

// Escape escapes HTML
func (th *TemplateHelpers) Escape(str string) template.HTML {
	return template.HTML(template.HTMLEscapeString(str))
}

// Raw returns raw HTML
func (th *TemplateHelpers) Raw(html string) template.HTML {
	return template.HTML(html)
}

// Old retrieves old input value (from form validation)
func (th *TemplateHelpers) Old(key string, defaultValue ...string) string {
	if th.request == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	if err := th.request.ParseForm(); err == nil {
		if values, ok := th.request.Form[key]; ok && len(values) > 0 {
			return values[0]
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// Checked checks if a checkbox should be checked
func (th *TemplateHelpers) Checked(condition bool) string {
	if condition {
		return "checked"
	}
	return ""
}

// Selected checks if an option should be selected
func (th *TemplateHelpers) Selected(condition bool) string {
	if condition {
		return "selected"
	}
	return ""
}

// Auth checks if user is authenticated
func (th *TemplateHelpers) Auth() bool {
	// This would check the actual auth state
	// For now, return false as placeholder
	return false
}

// Guest checks if user is a guest
func (th *TemplateHelpers) Guest() bool {
	return !th.Auth()
}

// User returns the current user
func (th *TemplateHelpers) User() interface{} {
	// This would return the actual user
	return nil
}

// CSRFToken returns the CSRF token
func (th *TemplateHelpers) CSRFToken() string {
	// This would return the actual CSRF token
	return ""
}

// CSRFField returns a CSRF hidden field
func (th *TemplateHelpers) CSRFField() template.HTML {
	token := th.CSRFToken()
	return template.HTML(fmt.Sprintf(`<input type="hidden" name="_token" value="%s">`, token))
}

// MethodField returns a method spoofing field
func (th *TemplateHelpers) MethodField(method string) template.HTML {
	return template.HTML(fmt.Sprintf(`<input type="hidden" name="_method" value="%s">`, method))
}

// If returns value1 if condition is true, otherwise value2
func (th *TemplateHelpers) If(condition bool, value1, value2 interface{}) interface{} {
	if condition {
		return value1
	}
	return value2
}

// Unless returns value1 if condition is false, otherwise value2
func (th *TemplateHelpers) Unless(condition bool, value1, value2 interface{}) interface{} {
	if !condition {
		return value1
	}
	return value2
}

// Isset checks if a value is set (not nil)
func (th *TemplateHelpers) Isset(value interface{}) bool {
	return value != nil
}

// JSON encodes data as JSON
func (th *TemplateHelpers) JSON(data interface{}) string {
	// This would use json.Marshal
	return fmt.Sprintf("%v", data)
}

// JSONEncode encodes data as JSON
func (th *TemplateHelpers) JSONEncode(data interface{}) string {
	return th.JSON(data)
}

// JSONDecode decodes JSON (placeholder)
func (th *TemplateHelpers) JSONDecode(json string) interface{} {
	// This would use json.Unmarshal
	return nil
}

