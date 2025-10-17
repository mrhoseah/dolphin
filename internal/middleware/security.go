package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"
)

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// HSTS (HTTP Strict Transport Security)
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

			// X-Content-Type-Options
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// X-Frame-Options
			w.Header().Set("X-Frame-Options", "DENY")

			// X-XSS-Protection
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Referrer-Policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions-Policy
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			// Content-Security-Policy
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';")

			next.ServeHTTP(w, r)
		})
	}
}

// CSRFProtectionMiddleware provides CSRF protection
func CSRFProtectionMiddleware(secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CSRF for safe methods
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			// Get CSRF token from header or form
			token := r.Header.Get("X-CSRF-Token")
			if token == "" {
				token = r.FormValue("_token")
			}

			// Validate CSRF token
			if !isValidCSRFToken(token, secret) {
				http.Error(w, "CSRF token mismatch", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isValidCSRFToken validates a CSRF token
func isValidCSRFToken(token, secret string) bool {
	// This is a simplified implementation
	// In a real implementation, you'd use a proper CSRF token validation
	return token != "" && len(token) > 10
}

// RequestSizeLimitMiddleware limits the size of request bodies
func RequestSizeLimitMiddleware(maxSize int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)

			next.ServeHTTP(w, r)
		})
	}
}

// TimeoutMiddleware adds a timeout to requests
func TimeoutMiddleware(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a response writer that can handle timeouts
			timeoutWriter := &timeoutResponseWriter{
				ResponseWriter: w,
				timeout:        timeout,
			}

			// Handle the request with timeout
			done := make(chan bool)
			go func() {
				next.ServeHTTP(timeoutWriter, r.WithContext(ctx))
				done <- true
			}()

			select {
			case <-done:
				// Request completed successfully
			case <-ctx.Done():
				// Request timed out
				timeoutWriter.WriteHeader(http.StatusRequestTimeout)
				timeoutWriter.Write([]byte("Request timeout"))
			}
		})
	}
}

// timeoutResponseWriter wraps http.ResponseWriter to handle timeouts
type timeoutResponseWriter struct {
	http.ResponseWriter
	timeout time.Duration
}

func (w *timeoutResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *timeoutResponseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

// IPWhitelistMiddleware allows only whitelisted IPs
func IPWhitelistMiddleware(allowedIPs []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			// Check if IP is whitelisted
			allowed := false
			for _, ip := range allowedIPs {
				if clientIP == ip {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IPBlacklistMiddleware blocks blacklisted IPs
func IPBlacklistMiddleware(blockedIPs []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			// Check if IP is blacklisted
			for _, ip := range blockedIPs {
				if clientIP == ip {
					http.Error(w, "Access denied", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitByIPMiddleware limits requests per IP
func RateLimitByIPMiddleware(requestsPerMinute int) func(next http.Handler) http.Handler {
	// This is a simplified implementation
	// In a real implementation, you'd use a proper rate limiting library
	requestCounts := make(map[string][]time.Time)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			now := time.Now()

			// Clean old requests
			if requests, exists := requestCounts[clientIP]; exists {
				var validRequests []time.Time
				for _, reqTime := range requests {
					if now.Sub(reqTime) < time.Minute {
						validRequests = append(validRequests, reqTime)
					}
				}
				requestCounts[clientIP] = validRequests
			}

			// Check rate limit
			if len(requestCounts[clientIP]) >= requestsPerMinute {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// Add current request
			requestCounts[clientIP] = append(requestCounts[clientIP], now)

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(ip, ","); idx != -1 {
			ip = ip[:idx]
		}
		return strings.TrimSpace(ip)
	}

	// Check X-Real-IP header
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// SecurityConfig defines security configuration
type SecurityConfig struct {
	EnableHSTS          bool
	EnableCSRF          bool
	EnableCSP           bool
	EnableXSSProtection bool
	EnableFrameOptions  bool
	MaxRequestSize      int64
	RequestTimeout      time.Duration
	AllowedIPs          []string
	BlockedIPs          []string
	RateLimitPerMinute  int
	CSRFSecret          string
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		EnableHSTS:          true,
		EnableCSRF:          true,
		EnableCSP:           true,
		EnableXSSProtection: true,
		EnableFrameOptions:  true,
		MaxRequestSize:      10 * 1024 * 1024, // 10MB
		RequestTimeout:      30 * time.Second,
		RateLimitPerMinute:  100,
		CSRFSecret:          "your-csrf-secret-key",
	}
}

// SecurityMiddleware creates a comprehensive security middleware
func SecurityMiddleware(config *SecurityConfig) func(next http.Handler) http.Handler {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	return func(next http.Handler) http.Handler {
		// Apply security headers
		if config.EnableHSTS || config.EnableCSP || config.EnableXSSProtection || config.EnableFrameOptions {
			next = SecurityHeadersMiddleware()(next)
		}

		// Apply CSRF protection
		if config.EnableCSRF {
			next = CSRFProtectionMiddleware(config.CSRFSecret)(next)
		}

		// Apply request size limit
		if config.MaxRequestSize > 0 {
			next = RequestSizeLimitMiddleware(config.MaxRequestSize)(next)
		}

		// Apply timeout
		if config.RequestTimeout > 0 {
			next = TimeoutMiddleware(config.RequestTimeout)(next)
		}

		// Apply IP whitelist
		if len(config.AllowedIPs) > 0 {
			next = IPWhitelistMiddleware(config.AllowedIPs)(next)
		}

		// Apply IP blacklist
		if len(config.BlockedIPs) > 0 {
			next = IPBlacklistMiddleware(config.BlockedIPs)(next)
		}

		// Apply rate limiting
		if config.RateLimitPerMinute > 0 {
			next = RateLimitByIPMiddleware(config.RateLimitPerMinute)(next)
		}

		return next
	}
}
