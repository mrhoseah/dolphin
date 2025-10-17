package ratelimit

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(manager *RateLimitManager, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get route pattern from context (set by chi router)
			routePattern := chi.RouteContext(ctx).RoutePattern()
			if routePattern == "" {
				routePattern = r.URL.Path
			}

			// Check rate limit
			allowed, remaining, err := manager.CheckRateLimit(ctx, routePattern, r)
			if err != nil {
				logger.Error("Rate limit check failed", zap.Error(err))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, map[string]string{
					"error": "Rate limit check failed",
				})
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(manager.GetConfig(routePattern).Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(manager.GetConfig(routePattern).Window).Unix(), 10))

			if !allowed {
				logger.Warn("Rate limit exceeded",
					zap.String("route", routePattern),
					zap.String("ip", r.RemoteAddr),
					zap.Int("remaining", remaining))

				render.Status(r, http.StatusTooManyRequests)
				render.JSON(w, r, map[string]interface{}{
					"error":       "Rate limit exceeded",
					"message":     "Too many requests. Please try again later.",
					"retry_after": manager.GetConfig(routePattern).Window.Seconds(),
					"limit":       manager.GetConfig(routePattern).Limit,
					"remaining":   remaining,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IPBasedRateLimitMiddleware creates a rate limiting middleware based on IP
func IPBasedRateLimitMiddleware(limit int, window time.Duration, limiter RateLimiter, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get client IP
			ip := getClientIP(r)

			// Check rate limit
			allowed, err := limiter.Allow(ctx, ip, limit, window)
			if err != nil {
				logger.Error("Rate limit check failed", zap.Error(err))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, map[string]string{
					"error": "Rate limit check failed",
				})
				return
			}

			// Get remaining requests
			remaining, err := limiter.Remaining(ctx, ip, limit, window)
			if err != nil {
				remaining = 0
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(window).Unix(), 10))

			if !allowed {
				logger.Warn("Rate limit exceeded",
					zap.String("ip", ip),
					zap.Int("limit", limit),
					zap.Int("remaining", remaining))

				render.Status(r, http.StatusTooManyRequests)
				render.JSON(w, r, map[string]interface{}{
					"error":       "Rate limit exceeded",
					"message":     "Too many requests. Please try again later.",
					"retry_after": window.Seconds(),
					"limit":       limit,
					"remaining":   remaining,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// UserBasedRateLimitMiddleware creates a rate limiting middleware based on authenticated user
func UserBasedRateLimitMiddleware(limit int, window time.Duration, limiter RateLimiter, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get user ID from context (set by auth middleware)
			userID, ok := ctx.Value("user_id").(string)
			if !ok {
				// Fall back to IP if no user ID
				userID = getClientIP(r)
			}

			// Check rate limit
			allowed, err := limiter.Allow(ctx, "user:"+userID, limit, window)
			if err != nil {
				logger.Error("Rate limit check failed", zap.Error(err))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, map[string]string{
					"error": "Rate limit check failed",
				})
				return
			}

			// Get remaining requests
			remaining, err := limiter.Remaining(ctx, "user:"+userID, limit, window)
			if err != nil {
				remaining = 0
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(window).Unix(), 10))

			if !allowed {
				logger.Warn("Rate limit exceeded",
					zap.String("user_id", userID),
					zap.Int("limit", limit),
					zap.Int("remaining", remaining))

				render.Status(r, http.StatusTooManyRequests)
				render.JSON(w, r, map[string]interface{}{
					"error":       "Rate limit exceeded",
					"message":     "Too many requests. Please try again later.",
					"retry_after": window.Seconds(),
					"limit":       limit,
					"remaining":   remaining,
				})
				return
			}

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
