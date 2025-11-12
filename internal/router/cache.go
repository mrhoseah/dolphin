package router

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// CacheConfig configures API response caching
type CacheConfig struct {
	TTL           time.Duration
	KeyGenerator  func(r *http.Request) string
	VaryHeaders   []string
	Logger        *zap.Logger
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		TTL: 5 * time.Minute,
		KeyGenerator: func(r *http.Request) string {
			return generateCacheKey(r)
		},
		VaryHeaders: []string{"Accept", "Accept-Language"},
	}
}

// CacheMiddleware creates middleware for API response caching
func CacheMiddleware(config *CacheConfig) func(next http.Handler) http.Handler {
	if config == nil {
		config = DefaultCacheConfig()
	}

	// In-memory cache (in production, use Redis or similar)
	cache := make(map[string]*CachedResponse)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only cache GET requests
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			// Generate cache key
			cacheKey := config.KeyGenerator(r)
			
			// Check cache
			if cached, exists := cache[cacheKey]; exists {
				if time.Since(cached.Timestamp) < config.TTL {
					// Set cache headers
					w.Header().Set("X-Cache", "HIT")
					w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(config.TTL.Seconds())))
					
					// Copy cached headers
					for k, v := range cached.Headers {
						w.Header().Set(k, strings.Join(v, ", "))
					}
					
					// Write cached body
					w.WriteHeader(cached.StatusCode)
					w.Write(cached.Body)
					return
				}
				// Remove expired cache entry
				delete(cache, cacheKey)
			}

			// Cache miss - use response writer to capture response
			cw := &cacheWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				headers:        make(http.Header),
				body:           []byte{},
			}

			next.ServeHTTP(cw, r)

			// Only cache successful responses
			if cw.statusCode >= 200 && cw.statusCode < 300 {
				cached := &CachedResponse{
					StatusCode: cw.statusCode,
					Headers:    cw.headers,
					Body:       cw.body,
					Timestamp:  time.Now(),
				}
				cache[cacheKey] = cached
				
				w.Header().Set("X-Cache", "MISS")
			}
		})
	}
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Timestamp  time.Time
}

// cacheWriter captures response for caching
type cacheWriter struct {
	http.ResponseWriter
	statusCode int
	headers    http.Header
	body       []byte
}

func (cw *cacheWriter) WriteHeader(code int) {
	cw.statusCode = code
	cw.ResponseWriter.WriteHeader(code)
}

func (cw *cacheWriter) Write(b []byte) (int, error) {
	cw.body = append(cw.body, b...)
	return cw.ResponseWriter.Write(b)
}

func (cw *cacheWriter) Header() http.Header {
	if cw.headers == nil {
		cw.headers = make(http.Header)
	}
	// Copy headers from original response writer
	for k, v := range cw.ResponseWriter.Header() {
		cw.headers[k] = v
	}
	return cw.ResponseWriter.Header()
}

// generateCacheKey generates a cache key from request
func generateCacheKey(r *http.Request) string {
	key := r.Method + ":" + r.URL.Path + "?" + r.URL.RawQuery
	
	// Include vary headers in key
	for _, header := range []string{"Accept", "Accept-Language", "Authorization"} {
		if val := r.Header.Get(header); val != "" {
			key += ":" + header + "=" + val
		}
	}
	
	// Hash the key
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// SetCacheHeaders sets cache control headers
func SetCacheHeaders(w http.ResponseWriter, maxAge time.Duration, private bool) {
	directive := "public"
	if private {
		directive = "private"
	}
	w.Header().Set("Cache-Control", fmt.Sprintf("%s, max-age=%d", directive, int(maxAge.Seconds())))
}

// NoCache sets no-cache headers
func NoCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

