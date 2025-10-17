package security

import (
	"fmt"
	"net/http"
	"strings"
)

// SecurityPreset defines a security header configuration preset
type SecurityPreset struct {
	Name        string
	Description string
	Headers     map[string]string
}

// SecurityPresets contains predefined security configurations
var SecurityPresets = map[string]SecurityPreset{
	"strict": {
		Name:        "Strict Security",
		Description: "Maximum security with strict CSP and all security headers",
		Headers: map[string]string{
			"Strict-Transport-Security":           "max-age=31536000; includeSubDomains; preload",
			"X-Content-Type-Options":              "nosniff",
			"X-Frame-Options":                     "DENY",
			"X-XSS-Protection":                    "1; mode=block",
			"Referrer-Policy":                     "strict-origin-when-cross-origin",
			"Permissions-Policy":                  "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()",
			"Content-Security-Policy":             "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self';",
			"Cross-Origin-Embedder-Policy":        "require-corp",
			"Cross-Origin-Opener-Policy":          "same-origin",
			"Cross-Origin-Resource-Policy":        "same-origin",
			"X-Permitted-Cross-Domain-Policies":   "none",
		},
	},
	"balanced": {
		Name:        "Balanced Security",
		Description: "Good security with reasonable CSP for development",
		Headers: map[string]string{
			"Strict-Transport-Security":           "max-age=31536000; includeSubDomains",
			"X-Content-Type-Options":              "nosniff",
			"X-Frame-Options":                     "SAMEORIGIN",
			"X-XSS-Protection":                    "1; mode=block",
			"Referrer-Policy":                     "strict-origin-when-cross-origin",
			"Permissions-Policy":                  "geolocation=(), microphone=(), camera=()",
			"Content-Security-Policy":             "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self';",
		},
	},
	"development": {
		Name:        "Development Security",
		Description: "Minimal security for development with relaxed CSP",
		Headers: map[string]string{
			"X-Content-Type-Options":              "nosniff",
			"X-Frame-Options":                     "SAMEORIGIN",
			"X-XSS-Protection":                    "1; mode=block",
			"Content-Security-Policy":             "default-src 'self' 'unsafe-inline' 'unsafe-eval' data: blob:; img-src 'self' data: https:; font-src 'self' data:;",
		},
	},
	"api": {
		Name:        "API Security",
		Description: "Security headers optimized for API endpoints",
		Headers: map[string]string{
			"Strict-Transport-Security":           "max-age=31536000; includeSubDomains",
			"X-Content-Type-Options":              "nosniff",
			"X-Frame-Options":                     "DENY",
			"X-XSS-Protection":                    "1; mode=block",
			"Referrer-Policy":                     "strict-origin-when-cross-origin",
			"Permissions-Policy":                  "geolocation=(), microphone=(), camera=()",
			"Content-Security-Policy":             "default-src 'none'; frame-ancestors 'none';",
			"Cross-Origin-Embedder-Policy":        "require-corp",
			"Cross-Origin-Opener-Policy":          "same-origin",
			"Cross-Origin-Resource-Policy":        "same-origin",
		},
	},
	"minimal": {
		Name:        "Minimal Security",
		Description: "Basic security headers only",
		Headers: map[string]string{
			"X-Content-Type-Options": "nosniff",
			"X-Frame-Options":        "SAMEORIGIN",
			"X-XSS-Protection":       "1; mode=block",
		},
	},
}

// SecurityHeaderManager manages security header configurations
type SecurityHeaderManager struct {
	preset     string
	customHeaders map[string]string
	overrides  map[string]string
}

// NewSecurityHeaderManager creates a new security header manager
func NewSecurityHeaderManager(preset string) *SecurityHeaderManager {
	return &SecurityHeaderManager{
		preset:        preset,
		customHeaders: make(map[string]string),
		overrides:     make(map[string]string),
	}
}

// SetPreset changes the security preset
func (shm *SecurityHeaderManager) SetPreset(preset string) error {
	if _, exists := SecurityPresets[preset]; !exists {
		return fmt.Errorf("unknown security preset: %s", preset)
	}
	shm.preset = preset
	return nil
}

// AddCustomHeader adds a custom security header
func (shm *SecurityHeaderManager) AddCustomHeader(name, value string) {
	shm.customHeaders[name] = value
}

// OverrideHeader overrides a header from the preset
func (shm *SecurityHeaderManager) OverrideHeader(name, value string) {
	shm.overrides[name] = value
}

// GetHeaders returns all security headers for the current configuration
func (shm *SecurityHeaderManager) GetHeaders() map[string]string {
	headers := make(map[string]string)

	// Start with preset headers
	if preset, exists := SecurityPresets[shm.preset]; exists {
		for name, value := range preset.Headers {
			headers[name] = value
		}
	}

	// Apply overrides
	for name, value := range shm.overrides {
		headers[name] = value
	}

	// Add custom headers
	for name, value := range shm.customHeaders {
		headers[name] = value
	}

	return headers
}

// SecurityHeadersMiddleware creates a middleware that applies security headers
func SecurityHeadersMiddleware(manager *SecurityHeaderManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headers := manager.GetHeaders()
			
			// Apply all security headers
			for name, value := range headers {
				w.Header().Set(name, value)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddlewareWithPreset creates a middleware with a specific preset
func SecurityHeadersMiddlewareWithPreset(preset string) func(http.Handler) http.Handler {
	manager := NewSecurityHeaderManager(preset)
	return SecurityHeadersMiddleware(manager)
}

// SecurityHeadersMiddlewareWithCustom creates a middleware with custom configuration
func SecurityHeadersMiddlewareWithCustom(preset string, overrides, custom map[string]string) func(http.Handler) http.Handler {
	manager := NewSecurityHeaderManager(preset)
	
	// Apply overrides
	for name, value := range overrides {
		manager.OverrideHeader(name, value)
	}
	
	// Apply custom headers
	for name, value := range custom {
		manager.AddCustomHeader(name, value)
	}
	
	return SecurityHeadersMiddleware(manager)
}

// CSPBuilder helps build Content Security Policy headers
type CSPBuilder struct {
	directives map[string][]string
}

// NewCSPBuilder creates a new CSP builder
func NewCSPBuilder() *CSPBuilder {
	return &CSPBuilder{
		directives: make(map[string][]string),
	}
}

// AddDirective adds a CSP directive
func (csp *CSPBuilder) AddDirective(name string, values ...string) *CSPBuilder {
	csp.directives[name] = append(csp.directives[name], values...)
	return csp
}

// SetDefaultSrc sets the default-src directive
func (csp *CSPBuilder) SetDefaultSrc(values ...string) *CSPBuilder {
	return csp.AddDirective("default-src", values...)
}

// SetScriptSrc sets the script-src directive
func (csp *CSPBuilder) SetScriptSrc(values ...string) *CSPBuilder {
	return csp.AddDirective("script-src", values...)
}

// SetStyleSrc sets the style-src directive
func (csp *CSPBuilder) SetStyleSrc(values ...string) *CSPBuilder {
	return csp.AddDirective("style-src", values...)
}

// SetImgSrc sets the img-src directive
func (csp *CSPBuilder) SetImgSrc(values ...string) *CSPBuilder {
	return csp.AddDirective("img-src", values...)
}

// SetFontSrc sets the font-src directive
func (csp *CSPBuilder) SetFontSrc(values ...string) *CSPBuilder {
	return csp.AddDirective("font-src", values...)
}

// SetConnectSrc sets the connect-src directive
func (csp *CSPBuilder) SetConnectSrc(values ...string) *CSPBuilder {
	return csp.AddDirective("connect-src", values...)
}

// SetFrameAncestors sets the frame-ancestors directive
func (csp *CSPBuilder) SetFrameAncestors(values ...string) *CSPBuilder {
	return csp.AddDirective("frame-ancestors", values...)
}

// SetBaseURI sets the base-uri directive
func (csp *CSPBuilder) SetBaseURI(values ...string) *CSPBuilder {
	return csp.AddDirective("base-uri", values...)
}

// SetFormAction sets the form-action directive
func (csp *CSPBuilder) SetFormAction(values ...string) *CSPBuilder {
	return csp.AddDirective("form-action", values...)
}

// Build returns the CSP header value
func (csp *CSPBuilder) Build() string {
	var parts []string
	
	for name, values := range csp.directives {
		if len(values) > 0 {
			parts = append(parts, fmt.Sprintf("%s %s", name, strings.Join(values, " ")))
		}
	}
	
	return strings.Join(parts, "; ")
}

// SecurityConfig represents the overall security configuration
type SecurityConfig struct {
	Preset        string            `yaml:"preset" json:"preset"`
	Overrides     map[string]string `yaml:"overrides" json:"overrides"`
	CustomHeaders map[string]string `yaml:"custom_headers" json:"custom_headers"`
	EnableHSTS    bool              `yaml:"enable_hsts" json:"enable_hsts"`
	EnableCSP     bool              `yaml:"enable_csp" json:"enable_csp"`
	EnableCORS    bool              `yaml:"enable_cors" json:"enable_cors"`
}

// DefaultSecurityConfig returns a default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		Preset:        "balanced",
		Overrides:     make(map[string]string),
		CustomHeaders: make(map[string]string),
		EnableHSTS:    true,
		EnableCSP:     true,
		EnableCORS:    false,
	}
}

// GetManager creates a SecurityHeaderManager from the config
func (sc *SecurityConfig) GetManager() *SecurityHeaderManager {
	manager := NewSecurityHeaderManager(sc.Preset)
	
	// Apply overrides
	for name, value := range sc.Overrides {
		manager.OverrideHeader(name, value)
	}
	
	// Apply custom headers
	for name, value := range sc.CustomHeaders {
		manager.AddCustomHeader(name, value)
	}
	
	return manager
}

// GetAvailablePresets returns all available security presets
func GetAvailablePresets() []string {
	var presets []string
	for name := range SecurityPresets {
		presets = append(presets, name)
	}
	return presets
}

// GetPresetInfo returns information about a specific preset
func GetPresetInfo(preset string) (SecurityPreset, error) {
	if p, exists := SecurityPresets[preset]; exists {
		return p, nil
	}
	return SecurityPreset{}, fmt.Errorf("preset not found: %s", preset)
}
