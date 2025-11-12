package router

import (
	"html"
	"net/http"
	"strings"
)

// SanitizeString sanitizes a string by escaping HTML and trimming whitespace
func SanitizeString(s string) string {
	s = strings.TrimSpace(s)
	s = html.EscapeString(s)
	return s
}

// SanitizeRequest sanitizes request form values
func SanitizeRequest(r *http.Request) {
	if err := r.ParseForm(); err != nil {
		return
	}

	// Sanitize all form values
	for key, values := range r.PostForm {
		for i, value := range values {
			r.PostForm[key][i] = SanitizeString(value)
		}
	}

	for key, values := range r.Form {
		for i, value := range values {
			r.Form[key][i] = SanitizeString(value)
		}
	}
}

// SanitizeMiddleware creates middleware for request sanitization
func SanitizeMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only sanitize POST, PUT, PATCH requests
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				SanitizeRequest(r)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SanitizeHTML sanitizes HTML content (basic implementation)
func SanitizeHTML(html string) string {
	// Remove script tags
	html = strings.ReplaceAll(html, "<script", "&lt;script")
	html = strings.ReplaceAll(html, "</script>", "&lt;/script&gt;")
	
	// Remove event handlers
	html = strings.ReplaceAll(html, "onerror=", "")
	html = strings.ReplaceAll(html, "onclick=", "")
	html = strings.ReplaceAll(html, "onload=", "")
	
	return html
}

