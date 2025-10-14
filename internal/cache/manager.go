package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache interface defines cache operations
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Flush(ctx context.Context) error
}

// RedisCache implements Cache interface using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(host string, port int, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", host, port),
		DB:   db,
	})

	return &RedisCache{
		client: client,
	}
}

// Get retrieves a value from cache
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Set stores a value in cache
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var val string
	
	switch v := value.(type) {
	case string:
		val = v
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			return err
		}
		val = string(jsonData)
	}
	
	return r.client.Set(ctx, key, val, expiration).Err()
}

// Delete removes a value from cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in cache
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result := r.client.Exists(ctx, key)
	return result.Val() > 0, result.Err()
}

// Flush removes all keys from cache
func (r *RedisCache) Flush(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// MemoryCache implements Cache interface using in-memory storage
type MemoryCache struct {
	data map[string]cacheItem
}

type cacheItem struct {
	value      string
	expiration time.Time
}

// NewMemoryCache creates a new memory cache instance
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]cacheItem),
	}
}

// Get retrieves a value from cache
func (m *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	item, exists := m.data[key]
	if !exists {
		return "", fmt.Errorf("key not found")
	}
	
	if time.Now().After(item.expiration) {
		delete(m.data, key)
		return "", fmt.Errorf("key expired")
	}
	
	return item.value, nil
}

// Set stores a value in cache
func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var val string
	
	switch v := value.(type) {
	case string:
		val = v
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			return err
		}
		val = string(jsonData)
	}
	
	m.data[key] = cacheItem{
		value:      val,
		expiration: time.Now().Add(expiration),
	}
	
	return nil
}

// Delete removes a value from cache
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

// Exists checks if a key exists in cache
func (m *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	item, exists := m.data[key]
	if !exists {
		return false, nil
	}
	
	if time.Now().After(item.expiration) {
		delete(m.data, key)
		return false, nil
	}
	
	return true, nil
}

// Flush removes all keys from cache
func (m *MemoryCache) Flush(ctx context.Context) error {
	m.data = make(map[string]cacheItem)
	return nil
}

// CacheManager manages cache operations
type CacheManager struct {
	cache Cache
}

// NewCacheManager creates a new cache manager
func NewCacheManager(cache Cache) *CacheManager {
	return &CacheManager{
		cache: cache,
	}
}

// Get retrieves a value from cache
func (cm *CacheManager) Get(ctx context.Context, key string) (string, error) {
	return cm.cache.Get(ctx, key)
}

// GetJSON retrieves and unmarshals JSON data from cache
func (cm *CacheManager) GetJSON(ctx context.Context, key string, dest interface{}) error {
	value, err := cm.cache.Get(ctx, key)
	if err != nil {
		return err
	}
	
	return json.Unmarshal([]byte(value), dest)
}

// Set stores a value in cache
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return cm.cache.Set(ctx, key, value, expiration)
}

// SetJSON marshals and stores JSON data in cache
func (cm *CacheManager) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return cm.cache.Set(ctx, key, value, expiration)
}

// Delete removes a value from cache
func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	return cm.cache.Delete(ctx, key)
}

// Exists checks if a key exists in cache
func (cm *CacheManager) Exists(ctx context.Context, key string) (bool, error) {
	return cm.cache.Exists(ctx, key)
}

// Flush removes all keys from cache
func (cm *CacheManager) Flush(ctx context.Context) error {
	return cm.cache.Flush(ctx)
}

// Remember retrieves a value from cache or executes a function and caches the result
func (cm *CacheManager) Remember(ctx context.Context, key string, expiration time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, err := cm.cache.Get(ctx, key); err == nil {
		return value, nil
	}
	
	// Execute function
	result, err := fn()
	if err != nil {
		return nil, err
	}
	
	// Cache the result
	if err := cm.cache.Set(ctx, key, result, expiration); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to cache result: %v\n", err)
	}
	
	return result, nil
}

// RememberJSON retrieves JSON data from cache or executes a function and caches the result
func (cm *CacheManager) RememberJSON(ctx context.Context, key string, expiration time.Duration, dest interface{}, fn func() (interface{}, error)) error {
	// Try to get from cache first
	if err := cm.GetJSON(ctx, key, dest); err == nil {
		return nil
	}
	
	// Execute function
	result, err := fn()
	if err != nil {
		return err
	}
	
	// Cache the result
	if err := cm.cache.Set(ctx, key, result, expiration); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to cache result: %v\n", err)
	}
	
	// Unmarshal result into destination
	jsonData, err := json.Marshal(result)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(jsonData, dest)
}

// Increment increments a numeric value in cache
func (cm *CacheManager) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	// This is a simplified implementation
	// In a real implementation, you might want to use Redis INCRBY or similar
	value, err := cm.cache.Get(ctx, key)
	if err != nil {
		// Key doesn't exist, set it to delta
		err = cm.cache.Set(ctx, key, delta, 0)
		return delta, err
	}
	
	var current int64
	if err := json.Unmarshal([]byte(value), &current); err != nil {
		return 0, err
	}
	
	newValue := current + delta
	err = cm.cache.Set(ctx, key, newValue, 0)
	return newValue, err
}

// Decrement decrements a numeric value in cache
func (cm *CacheManager) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return cm.Increment(ctx, key, -delta)
}
