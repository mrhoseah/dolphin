package assets

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// AssetStats represents asset pipeline statistics
type AssetStats struct {
	// Processing statistics
	ProcessCount    int64         `json:"process_count"`
	LastProcess     time.Time     `json:"last_process"`
	ProcessDuration time.Duration `json:"process_duration"`

	// File statistics
	FileChanges    int64               `json:"file_changes"`
	FilesProcessed int64               `json:"files_processed"`
	FilesByType    map[AssetType]int64 `json:"files_by_type"`
	FilesByBundle  map[string]int64    `json:"files_by_bundle"`

	// Bundle statistics
	BundleCount   int64 `json:"bundle_count"`
	BundleSize    int64 `json:"bundle_size"`
	CombinedFiles int64 `json:"combined_files"`

	// Performance statistics
	TotalSize      int64         `json:"total_size"`
	AverageSize    int64         `json:"average_size"`
	ProcessingTime time.Duration `json:"processing_time"`

	// Cache statistics
	CacheHits      int64 `json:"cache_hits"`
	CacheMisses    int64 `json:"cache_misses"`
	CacheEvictions int64 `json:"cache_evictions"`

	// Timing
	StartTime time.Time     `json:"start_time"`
	Uptime    time.Duration `json:"uptime"`

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewAssetStats creates new asset statistics
func NewAssetStats() *AssetStats {
	return &AssetStats{
		FilesByType:   make(map[AssetType]int64),
		FilesByBundle: make(map[string]int64),
		StartTime:     time.Now(),
	}
}

// RecordProcess records a processing operation
func (as *AssetStats) RecordProcess() {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.ProcessCount++
	as.LastProcess = time.Now()
}

// RecordProcessDuration records processing duration
func (as *AssetStats) RecordProcessDuration(duration time.Duration) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.ProcessDuration = duration
	as.ProcessingTime += duration
}

// RecordFileChange records a file change
func (as *AssetStats) RecordFileChange(path string, op fsnotify.Op) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.FileChanges++
}

// RecordFileProcessed records a processed file
func (as *AssetStats) RecordFileProcessed(assetType AssetType, bundle string, size int64) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.FilesProcessed++
	as.FilesByType[assetType]++
	as.FilesByBundle[bundle]++
	as.TotalSize += size

	// Update average size
	if as.FilesProcessed > 0 {
		as.AverageSize = as.TotalSize / as.FilesProcessed
	}
}

// RecordBundleCreated records a created bundle
func (as *AssetStats) RecordBundleCreated(bundleSize int64, combined bool) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.BundleCount++
	as.BundleSize += bundleSize

	if combined {
		as.CombinedFiles++
	}
}

// RecordCacheHit records a cache hit
func (as *AssetStats) RecordCacheHit() {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.CacheHits++
}

// RecordCacheMiss records a cache miss
func (as *AssetStats) RecordCacheMiss() {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.CacheMisses++
}

// RecordCacheEviction records a cache eviction
func (as *AssetStats) RecordCacheEviction() {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.CacheEvictions++
}

// GetStats returns current statistics
func (as *AssetStats) GetStats() map[string]interface{} {
	as.mu.RLock()
	defer as.mu.RUnlock()

	uptime := time.Since(as.StartTime)

	// Calculate hit rate
	totalCacheRequests := as.CacheHits + as.CacheMisses
	hitRate := float64(0)
	if totalCacheRequests > 0 {
		hitRate = float64(as.CacheHits) / float64(totalCacheRequests) * 100
	}

	// Convert files by type to string map
	filesByTypeStr := make(map[string]int64)
	for assetType, count := range as.FilesByType {
		filesByTypeStr[assetType.String()] = count
	}

	return map[string]interface{}{
		"process_count":    as.ProcessCount,
		"last_process":     as.LastProcess,
		"process_duration": as.ProcessDuration,
		"file_changes":     as.FileChanges,
		"files_processed":  as.FilesProcessed,
		"files_by_type":    filesByTypeStr,
		"files_by_bundle":  as.FilesByBundle,
		"bundle_count":     as.BundleCount,
		"bundle_size":      as.BundleSize,
		"combined_files":   as.CombinedFiles,
		"total_size":       as.TotalSize,
		"average_size":     as.AverageSize,
		"processing_time":  as.ProcessingTime,
		"cache_hits":       as.CacheHits,
		"cache_misses":     as.CacheMisses,
		"cache_evictions":  as.CacheEvictions,
		"cache_hit_rate":   hitRate,
		"start_time":       as.StartTime,
		"uptime":           uptime,
	}
}

// GetFileChangeRate returns the file change rate per minute
func (as *AssetStats) GetFileChangeRate() float64 {
	as.mu.RLock()
	defer as.mu.RUnlock()

	uptime := time.Since(as.StartTime)
	if uptime.Minutes() == 0 {
		return 0
	}

	return float64(as.FileChanges) / uptime.Minutes()
}

// GetProcessingRate returns the processing rate per minute
func (as *AssetStats) GetProcessingRate() float64 {
	as.mu.RLock()
	defer as.mu.RUnlock()

	uptime := time.Since(as.StartTime)
	if uptime.Minutes() == 0 {
		return 0
	}

	return float64(as.FilesProcessed) / uptime.Minutes()
}

// GetMostUsedTypes returns the most used asset types
func (as *AssetStats) GetMostUsedTypes(limit int) []TypeUsage {
	as.mu.RLock()
	defer as.mu.RUnlock()

	// Convert to slice and sort
	types := make([]TypeUsage, 0, len(as.FilesByType))
	for assetType, count := range as.FilesByType {
		types = append(types, TypeUsage{
			Type:  assetType,
			Count: count,
		})
	}

	// Sort by count (descending)
	for i := 0; i < len(types); i++ {
		for j := i + 1; j < len(types); j++ {
			if types[i].Count < types[j].Count {
				types[i], types[j] = types[j], types[i]
			}
		}
	}

	// Return limited results
	if limit > 0 && limit < len(types) {
		types = types[:limit]
	}

	return types
}

// GetMostUsedBundles returns the most used bundles
func (as *AssetStats) GetMostUsedBundles(limit int) []BundleUsage {
	as.mu.RLock()
	defer as.mu.RUnlock()

	// Convert to slice and sort
	bundles := make([]BundleUsage, 0, len(as.FilesByBundle))
	for bundle, count := range as.FilesByBundle {
		bundles = append(bundles, BundleUsage{
			Bundle: bundle,
			Count:  count,
		})
	}

	// Sort by count (descending)
	for i := 0; i < len(bundles); i++ {
		for j := i + 1; j < len(bundles); j++ {
			if bundles[i].Count < bundles[j].Count {
				bundles[i], bundles[j] = bundles[j], bundles[i]
			}
		}
	}

	// Return limited results
	if limit > 0 && limit < len(bundles) {
		bundles = bundles[:limit]
	}

	return bundles
}

// Reset resets all statistics
func (as *AssetStats) Reset() {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.ProcessCount = 0
	as.LastProcess = time.Time{}
	as.ProcessDuration = 0
	as.FileChanges = 0
	as.FilesProcessed = 0
	as.FilesByType = make(map[AssetType]int64)
	as.FilesByBundle = make(map[string]int64)
	as.BundleCount = 0
	as.BundleSize = 0
	as.CombinedFiles = 0
	as.TotalSize = 0
	as.AverageSize = 0
	as.ProcessingTime = 0
	as.CacheHits = 0
	as.CacheMisses = 0
	as.CacheEvictions = 0
	as.StartTime = time.Now()
}

// TypeUsage represents asset type usage statistics
type TypeUsage struct {
	Type  AssetType `json:"type"`
	Count int64     `json:"count"`
}

// BundleUsage represents bundle usage statistics
type BundleUsage struct {
	Bundle string `json:"bundle"`
	Count  int64  `json:"count"`
}

// AssetSummary represents an asset summary
type AssetSummary struct {
	TotalAssets    int64            `json:"total_assets"`
	TotalSize      int64            `json:"total_size"`
	AverageSize    int64            `json:"average_size"`
	FilesByType    map[string]int64 `json:"files_by_type"`
	FilesByBundle  map[string]int64 `json:"files_by_bundle"`
	BundleCount    int64            `json:"bundle_count"`
	CombinedFiles  int64            `json:"combined_files"`
	CacheHitRate   float64          `json:"cache_hit_rate"`
	ProcessingTime time.Duration    `json:"processing_time"`
	Uptime         time.Duration    `json:"uptime"`
}
