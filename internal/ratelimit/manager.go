package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter defines the interface for rate limiting
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
	Remaining(ctx context.Context, key string, limit int, window time.Duration) (int, error)
	Reset(ctx context.Context, key string) error
}

// RedisRateLimiter implements rate limiting using Redis
type RedisRateLimiter struct {
	client *redis.Client
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
	}
}

// Allow checks if a request is allowed within the rate limit
func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now()
	windowStart := now.Truncate(window)
	redisKey := fmt.Sprintf("ratelimit:%s:%d", key, windowStart.Unix())

	pipe := r.client.Pipeline()

	// Increment counter
	incr := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := incr.Val()
	return count <= int64(limit), nil
}

// Remaining returns the number of remaining requests
func (r *RedisRateLimiter) Remaining(ctx context.Context, key string, limit int, window time.Duration) (int, error) {
	now := time.Now()
	windowStart := now.Truncate(window)
	redisKey := fmt.Sprintf("ratelimit:%s:%d", key, windowStart.Unix())

	count, err := r.client.Get(ctx, redisKey).Int()
	if err == redis.Nil {
		return limit, nil
	}
	if err != nil {
		return 0, err
	}

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// Reset resets the rate limit for a key
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	pattern := fmt.Sprintf("ratelimit:%s:*", key)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}

	return nil
}

// MemoryRateLimiter implements rate limiting using in-memory storage
type MemoryRateLimiter struct {
	store map[string]*rateLimitData
}

type rateLimitData struct {
	count     int
	windowEnd time.Time
}

// NewMemoryRateLimiter creates a new memory-based rate limiter
func NewMemoryRateLimiter() *MemoryRateLimiter {
	return &MemoryRateLimiter{
		store: make(map[string]*rateLimitData),
	}
}

// Allow checks if a request is allowed within the rate limit
func (m *MemoryRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now()
	windowStart := now.Truncate(window)
	windowKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())

	data, exists := m.store[windowKey]
	if !exists || now.After(data.windowEnd) {
		// New window or expired window
		m.store[windowKey] = &rateLimitData{
			count:     1,
			windowEnd: windowStart.Add(window),
		}
		return true, nil
	}

	if data.count >= limit {
		return false, nil
	}

	data.count++
	return true, nil
}

// Remaining returns the number of remaining requests
func (m *MemoryRateLimiter) Remaining(ctx context.Context, key string, limit int, window time.Duration) (int, error) {
	now := time.Now()
	windowStart := now.Truncate(window)
	windowKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())

	data, exists := m.store[windowKey]
	if !exists || now.After(data.windowEnd) {
		return limit, nil
	}

	remaining := limit - data.count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// Reset resets the rate limit for a key
func (m *MemoryRateLimiter) Reset(ctx context.Context, key string) error {
	// Remove all windows for this key
	for k := range m.store {
		if strings.HasPrefix(k, key+":") {
			delete(m.store, k)
		}
	}
	return nil
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	Enabled bool
	Limit   int
	Window  time.Duration
	KeyFunc func(r *http.Request) string
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Enabled: true,
		Limit:   100,         // 100 requests
		Window:  time.Minute, // per minute
		KeyFunc: func(r *http.Request) string {
			// Use IP address as default key
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ip = r.Header.Get("X-Real-IP")
			}
			if ip == "" {
				ip = strings.Split(r.RemoteAddr, ":")[0]
			}
			return ip
		},
	}
}

// RateLimitManager manages rate limiting
type RateLimitManager struct {
	limiter RateLimiter
	configs map[string]*RateLimitConfig
}

// NewRateLimitManager creates a new rate limit manager
func NewRateLimitManager(limiter RateLimiter) *RateLimitManager {
	return &RateLimitManager{
		limiter: limiter,
		configs: make(map[string]*RateLimitConfig),
	}
}

// AddConfig adds a rate limit configuration for a route
func (m *RateLimitManager) AddConfig(route string, config *RateLimitConfig) {
	m.configs[route] = config
}

// GetConfig returns the rate limit configuration for a route
func (m *RateLimitManager) GetConfig(route string) *RateLimitConfig {
	if config, exists := m.configs[route]; exists {
		return config
	}
	return DefaultRateLimitConfig()
}

// CheckRateLimit checks if a request is allowed
func (m *RateLimitManager) CheckRateLimit(ctx context.Context, route string, r *http.Request) (bool, int, error) {
	config := m.GetConfig(route)
	if !config.Enabled {
		return true, 0, nil
	}

	key := config.KeyFunc(r)
	allowed, err := m.limiter.Allow(ctx, key, config.Limit, config.Window)
	if err != nil {
		return false, 0, err
	}

	remaining, err := m.limiter.Remaining(ctx, key, config.Limit, config.Window)
	if err != nil {
		return allowed, 0, err
	}

	return allowed, remaining, nil
}

// ResetRateLimit resets the rate limit for a key
func (m *RateLimitManager) ResetRateLimit(ctx context.Context, key string) error {
	return m.limiter.Reset(ctx, key)
}
