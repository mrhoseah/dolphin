package assets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CacheEntry represents a cache entry
type CacheEntry struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	ExpiresAt   time.Time   `json:"expires_at"`
	CreatedAt   time.Time   `json:"created_at"`
	AccessCount int64       `json:"access_count"`
}

// AssetCache provides caching for assets
type AssetCache struct {
	cacheDir string
	expiry   time.Duration
	logger   *zap.Logger

	// In-memory cache
	memory   map[string]*CacheEntry
	memoryMu sync.RWMutex

	// Control
	stopChan chan struct{}
	doneChan chan struct{}

	// Statistics
	hits      int64
	misses    int64
	evictions int64
}

// NewAssetCache creates a new asset cache
func NewAssetCache(cacheDir string, expiry time.Duration, logger *zap.Logger) (*AssetCache, error) {
	// Create cache directory
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	ac := &AssetCache{
		cacheDir: cacheDir,
		expiry:   expiry,
		logger:   logger,
		memory:   make(map[string]*CacheEntry),
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}

	// Load existing cache
	if err := ac.loadCache(); err != nil {
		if logger != nil {
			logger.Warn("Failed to load cache", zap.Error(err))
		}
	}

	// Start cleanup goroutine
	go ac.cleanup()

	return ac, nil
}

// Get retrieves a value from the cache
func (ac *AssetCache) Get(key string) (interface{}, bool) {
	ac.memoryMu.RLock()
	entry, exists := ac.memory[key]
	ac.memoryMu.RUnlock()

	if !exists {
		ac.misses++
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		ac.memoryMu.Lock()
		delete(ac.memory, key)
		ac.memoryMu.Unlock()
		ac.misses++
		ac.evictions++
		return nil, false
	}

	// Update access count
	ac.memoryMu.Lock()
	entry.AccessCount++
	ac.memoryMu.Unlock()

	ac.hits++
	return entry.Value, true
}

// Set stores a value in the cache
func (ac *AssetCache) Set(key string, value interface{}) {
	ac.memoryMu.Lock()
	defer ac.memoryMu.Unlock()

	entry := &CacheEntry{
		Key:         key,
		Value:       value,
		ExpiresAt:   time.Now().Add(ac.expiry),
		CreatedAt:   time.Now(),
		AccessCount: 0,
	}

	ac.memory[key] = entry
}

// Delete removes a value from the cache
func (ac *AssetCache) Delete(key string) {
	ac.memoryMu.Lock()
	defer ac.memoryMu.Unlock()

	delete(ac.memory, key)
}

// Clear clears all cache entries
func (ac *AssetCache) Clear() {
	ac.memoryMu.Lock()
	defer ac.memoryMu.Unlock()

	ac.memory = make(map[string]*CacheEntry)
}

// GetStats returns cache statistics
func (ac *AssetCache) GetStats() map[string]interface{} {
	ac.memoryMu.RLock()
	defer ac.memoryMu.RUnlock()

	total := ac.hits + ac.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(ac.hits) / float64(total) * 100
	}

	return map[string]interface{}{
		"hits":      ac.hits,
		"misses":    ac.misses,
		"evictions": ac.evictions,
		"hit_rate":  hitRate,
		"entries":   len(ac.memory),
		"expiry":    ac.expiry.String(),
	}
}

// loadCache loads cache from disk
func (ac *AssetCache) loadCache() error {
	cacheFile := filepath.Join(ac.cacheDir, "cache.json")

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No cache file exists
		}
		return err
	}

	var entries map[string]*CacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}

	// Filter out expired entries
	now := time.Now()
	ac.memoryMu.Lock()
	for key, entry := range entries {
		if now.Before(entry.ExpiresAt) {
			ac.memory[key] = entry
		}
	}
	ac.memoryMu.Unlock()

	return nil
}

// saveCache saves cache to disk
func (ac *AssetCache) saveCache() error {
	ac.memoryMu.RLock()
	entries := make(map[string]*CacheEntry)
	for key, entry := range ac.memory {
		entries[key] = entry
	}
	ac.memoryMu.RUnlock()

	cacheFile := filepath.Join(ac.cacheDir, "cache.json")

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}

// cleanup runs periodic cleanup
func (ac *AssetCache) cleanup() {
	defer close(ac.doneChan)

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ac.stopChan:
			// Save cache before stopping
			if err := ac.saveCache(); err != nil && ac.logger != nil {
				ac.logger.Error("Failed to save cache", zap.Error(err))
			}
			return
		case <-ticker.C:
			ac.cleanupExpired()
		}
	}
}

// cleanupExpired removes expired entries
func (ac *AssetCache) cleanupExpired() {
	now := time.Now()

	ac.memoryMu.Lock()
	defer ac.memoryMu.Unlock()

	var expiredKeys []string
	for key, entry := range ac.memory {
		if now.After(entry.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(ac.memory, key)
		ac.evictions++
	}

	if len(expiredKeys) > 0 && ac.logger != nil {
		ac.logger.Debug("Cleaned up expired cache entries",
			zap.Int("count", len(expiredKeys)))
	}
}

// Close closes the cache
func (ac *AssetCache) Close() error {
	close(ac.stopChan)
	<-ac.doneChan

	return ac.saveCache()
}
