package assets

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AssetType represents the type of asset
type AssetType int

const (
	TypeCSS AssetType = iota
	TypeJS
	TypeImage
	TypeFont
	TypeOther
)

func (at AssetType) String() string {
	switch at {
	case TypeCSS:
		return "css"
	case TypeJS:
		return "js"
	case TypeImage:
		return "image"
	case TypeFont:
		return "font"
	case TypeOther:
		return "other"
	default:
		return "unknown"
	}
}

// BundleType represents the type of bundle
type BundleType int

const (
	BundleTypeApp BundleType = iota
	BundleTypeVendor
	BundleTypeCommon
	BundleTypePage
)

func (bt BundleType) String() string {
	switch bt {
	case BundleTypeApp:
		return "app"
	case BundleTypeVendor:
		return "vendor"
	case BundleTypeCommon:
		return "common"
	case BundleTypePage:
		return "page"
	default:
		return "unknown"
	}
}

// Config represents asset pipeline configuration
type Config struct {
	// Source and output directories
	SourceDir    string `yaml:"source_dir" json:"source_dir"`
	OutputDir    string `yaml:"output_dir" json:"output_dir"`
	PublicDir    string `yaml:"public_dir" json:"public_dir"`
	
	// Bundling configuration
	EnableBundling    bool     `yaml:"enable_bundling" json:"enable_bundling"`
	BundleTypes       []string `yaml:"bundle_types" json:"bundle_types"`
	MinifyAssets      bool     `yaml:"minify_assets" json:"minify_assets"`
	CombineAssets     bool     `yaml:"combine_assets" json:"combine_assets"`
	
	// Versioning configuration
	EnableVersioning  bool   `yaml:"enable_versioning" json:"enable_versioning"`
	VersionStrategy   string `yaml:"version_strategy" json:"version_strategy"` // hash, timestamp, manual
	VersionLength     int    `yaml:"version_length" json:"version_length"`
	
	// Optimization configuration
	EnableOptimization bool     `yaml:"enable_optimization" json:"enable_optimization"`
	OptimizeImages     bool     `yaml:"optimize_images" json:"optimize_images"`
	OptimizeCSS        bool     `yaml:"optimize_css" json:"optimize_css"`
	OptimizeJS         bool     `yaml:"optimize_js" json:"optimize_js"`
	
	// Cache configuration
	EnableCache       bool          `yaml:"enable_cache" json:"enable_cache"`
	CacheDir          string        `yaml:"cache_dir" json:"cache_dir"`
	CacheExpiry       time.Duration `yaml:"cache_expiry" json:"cache_expiry"`
	
	// Watch configuration
	EnableWatch       bool     `yaml:"enable_watch" json:"enable_watch"`
	WatchExtensions   []string `yaml:"watch_extensions" json:"watch_extensions"`
	
	// CDN configuration
	CDNUrl            string `yaml:"cdn_url" json:"cdn_url"`
	CDNEnabled        bool   `yaml:"cdn_enabled" json:"cdn_enabled"`
	
	// Logging
	EnableLogging     bool `yaml:"enable_logging" json:"enable_logging"`
	VerboseLogging    bool `yaml:"verbose_logging" json:"verbose_logging"`
}

// DefaultConfig returns default asset pipeline configuration
func DefaultConfig() *Config {
	return &Config{
		SourceDir:         "resources/assets",
		OutputDir:         "public/assets",
		PublicDir:         "public",
		EnableBundling:    true,
		BundleTypes:       []string{"app", "vendor", "common"},
		MinifyAssets:      true,
		CombineAssets:     true,
		EnableVersioning:  true,
		VersionStrategy:   "hash",
		VersionLength:     8,
		EnableOptimization: true,
		OptimizeImages:    true,
		OptimizeCSS:       true,
		OptimizeJS:        true,
		EnableCache:       true,
		CacheDir:          "storage/cache/assets",
		CacheExpiry:       24 * time.Hour,
		EnableWatch:       true,
		WatchExtensions:   []string{".css", ".js", ".scss", ".sass", ".less", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".woff", ".woff2", ".ttf", ".eot"},
		CDNUrl:            "",
		CDNEnabled:        false,
		EnableLogging:     true,
		VerboseLogging:    false,
	}
}

// Asset represents a single asset
type Asset struct {
	Path        string    `json:"path"`
	Type        AssetType `json:"type"`
	Bundle      string    `json:"bundle"`
	Version     string    `json:"version"`
	Size        int64     `json:"size"`
	Hash        string    `json:"hash"`
	LastModified time.Time `json:"last_modified"`
	CDNUrl      string    `json:"cdn_url,omitempty"`
}

// Bundle represents a collection of assets
type Bundle struct {
	Name        string    `json:"name"`
	Type        BundleType `json:"type"`
	Assets      []*Asset  `json:"assets"`
	CombinedPath string   `json:"combined_path"`
	Version     string    `json:"version"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
}

// AssetManager manages the asset pipeline
type AssetManager struct {
	config *Config
	logger *zap.Logger
	
	// Asset storage
	assets  map[string]*Asset
	bundles map[string]*Bundle
	mu      sync.RWMutex
	
	// File watching
	watcher *AssetWatcher
	
	// Cache
	cache *AssetCache
	
	// Statistics
	stats *AssetStats
}

// NewAssetManager creates a new asset manager
func NewAssetManager(config *Config, logger *zap.Logger) (*AssetManager, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	// Create cache
	cache, err := NewAssetCache(config.CacheDir, config.CacheExpiry, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset cache: %w", err)
	}
	
	// Create watcher if enabled
	var watcher *AssetWatcher
	if config.EnableWatch {
		watcher, err = NewAssetWatcher(config.SourceDir, config.WatchExtensions, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create asset watcher: %w", err)
		}
	}
	
	am := &AssetManager{
		config:  config,
		logger:  logger,
		assets:  make(map[string]*Asset),
		bundles: make(map[string]*Bundle),
		watcher: watcher,
		cache:   cache,
		stats:   NewAssetStats(),
	}
	
	// Start watcher if enabled
	if watcher != nil {
		go am.startWatching()
	}
	
	return am, nil
}

// ProcessAssets processes all assets in the source directory
func (am *AssetManager) ProcessAssets() error {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	if am.config.EnableLogging && am.logger != nil {
		am.logger.Info("Processing assets",
			zap.String("source_dir", am.config.SourceDir),
			zap.String("output_dir", am.config.OutputDir))
	}
	
	// Clear existing assets
	am.assets = make(map[string]*Asset)
	am.bundles = make(map[string]*Bundle)
	
	// Process source directory
	if err := am.processDirectory(am.config.SourceDir); err != nil {
		return fmt.Errorf("failed to process directory: %w", err)
	}
	
	// Create bundles if enabled
	if am.config.EnableBundling {
		if err := am.createBundles(); err != nil {
			return fmt.Errorf("failed to create bundles: %w", err)
		}
	}
	
	// Update statistics
	am.stats.RecordProcess()
	
	if am.config.EnableLogging && am.logger != nil {
		am.logger.Info("Assets processed successfully",
			zap.Int("asset_count", len(am.assets)),
			zap.Int("bundle_count", len(am.bundles)))
	}
	
	return nil
}

// processDirectory processes a directory recursively
func (am *AssetManager) processDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Check if file should be processed
		if !am.shouldProcessFile(path) {
			return nil
		}
		
		// Process the file
		asset, err := am.processFile(path)
		if err != nil {
			if am.config.EnableLogging && am.logger != nil {
				am.logger.Warn("Failed to process file",
					zap.String("file", path),
					zap.Error(err))
			}
			return nil // Continue processing other files
		}
		
		// Store asset
		am.assets[path] = asset
		
		return nil
	})
}

// shouldProcessFile checks if a file should be processed
func (am *AssetManager) shouldProcessFile(path string) bool {
	ext := filepath.Ext(path)
	for _, watchExt := range am.config.WatchExtensions {
		if ext == watchExt {
			return true
		}
	}
	return false
}

// processFile processes a single file
func (am *AssetManager) processFile(path string) (*Asset, error) {
	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	
	// Determine asset type
	assetType := am.getAssetType(path)
	
	// Calculate hash
	hash, err := am.calculateHash(path)
	if err != nil {
		return nil, err
	}
	
	// Generate version
	version := am.generateVersion(hash, info.ModTime())
	
	// Determine bundle
	bundle := am.determineBundle(path, assetType)
	
	// Create asset
	asset := &Asset{
		Path:         path,
		Type:         assetType,
		Bundle:       bundle,
		Version:      version,
		Size:         info.Size(),
		Hash:         hash,
		LastModified: info.ModTime(),
	}
	
	// Add CDN URL if enabled
	if am.config.CDNEnabled && am.config.CDNUrl != "" {
		asset.CDNUrl = am.config.CDNUrl + "/" + am.getOutputPath(asset)
	}
	
	return asset, nil
}

// getAssetType determines the asset type from file extension
func (am *AssetManager) getAssetType(path string) AssetType {
	ext := strings.ToLower(filepath.Ext(path))
	
	switch ext {
	case ".css", ".scss", ".sass", ".less":
		return TypeCSS
	case ".js", ".ts", ".jsx", ".tsx":
		return TypeJS
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp":
		return TypeImage
	case ".woff", ".woff2", ".ttf", ".eot", ".otf":
		return TypeFont
	default:
		return TypeOther
	}
}

// calculateHash calculates the MD5 hash of a file
func (am *AssetManager) calculateHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// generateVersion generates a version string
func (am *AssetManager) generateVersion(hash string, modTime time.Time) string {
	switch am.config.VersionStrategy {
	case "hash":
		if am.config.VersionLength > 0 && am.config.VersionLength < len(hash) {
			return hash[:am.config.VersionLength]
		}
		return hash
	case "timestamp":
		return fmt.Sprintf("%d", modTime.Unix())
	case "manual":
		return "1.0.0" // This would be read from a config file
	default:
		return hash[:8]
	}
}

// determineBundle determines which bundle an asset belongs to
func (am *AssetManager) determineBundle(path string, assetType AssetType) string {
	// Simple bundle determination based on directory structure
	dir := filepath.Dir(path)
	
	if strings.Contains(dir, "vendor") || strings.Contains(dir, "node_modules") {
		return "vendor"
	}
	
	if strings.Contains(dir, "common") || strings.Contains(dir, "shared") {
		return "common"
	}
	
	if strings.Contains(dir, "pages") || strings.Contains(dir, "views") {
		return "page"
	}
	
	return "app"
}

// createBundles creates bundles from assets
func (am *AssetManager) createBundles() error {
	// Group assets by bundle
	bundleAssets := make(map[string][]*Asset)
	for _, asset := range am.assets {
		bundleAssets[asset.Bundle] = append(bundleAssets[asset.Bundle], asset)
	}
	
	// Create bundles
	for bundleName, assets := range bundleAssets {
		bundle, err := am.createBundle(bundleName, assets)
		if err != nil {
			return fmt.Errorf("failed to create bundle %s: %w", bundleName, err)
		}
		am.bundles[bundleName] = bundle
	}
	
	return nil
}

// createBundle creates a single bundle
func (am *AssetManager) createBundle(name string, assets []*Asset) (*Bundle, error) {
	// Sort assets by type and path
	sort.Slice(assets, func(i, j int) bool {
		if assets[i].Type != assets[j].Type {
			return assets[i].Type < assets[j].Type
		}
		return assets[i].Path < assets[j].Path
	})
	
	// Calculate total size
	var totalSize int64
	for _, asset := range assets {
		totalSize += asset.Size
	}
	
	// Generate bundle version
	version := am.generateBundleVersion(assets)
	
	// Create bundle
	bundle := &Bundle{
		Name:        name,
		Type:        am.getBundleType(name),
		Assets:      assets,
		Version:     version,
		Size:        totalSize,
		CreatedAt:   time.Now(),
	}
	
	// Create combined file if enabled
	if am.config.CombineAssets {
		combinedPath, err := am.createCombinedFile(bundle)
		if err != nil {
			return nil, fmt.Errorf("failed to create combined file: %w", err)
		}
		bundle.CombinedPath = combinedPath
	}
	
	return bundle, nil
}

// generateBundleVersion generates a version for a bundle
func (am *AssetManager) generateBundleVersion(assets []*Asset) string {
	// Combine all asset hashes
	var hashes []string
	for _, asset := range assets {
		hashes = append(hashes, asset.Hash)
	}
	
	// Sort hashes for consistency
	sort.Strings(hashes)
	
	// Create combined hash
	combined := strings.Join(hashes, "")
	hash := md5.Sum([]byte(combined))
	hashStr := hex.EncodeToString(hash[:])
	
	if am.config.VersionLength > 0 && am.config.VersionLength < len(hashStr) {
		return hashStr[:am.config.VersionLength]
	}
	
	return hashStr
}

// getBundleType determines the bundle type from name
func (am *AssetManager) getBundleType(name string) BundleType {
	switch name {
	case "vendor":
		return BundleTypeVendor
	case "common":
		return BundleTypeCommon
	case "page":
		return BundleTypePage
	default:
		return BundleTypeApp
	}
}

// createCombinedFile creates a combined file for a bundle
func (am *AssetManager) createCombinedFile(bundle *Bundle) (string, error) {
	// Group assets by type
	cssAssets := make([]*Asset, 0)
	jsAssets := make([]*Asset, 0)
	
	for _, asset := range bundle.Assets {
		switch asset.Type {
		case TypeCSS:
			cssAssets = append(cssAssets, asset)
		case TypeJS:
			jsAssets = append(jsAssets, asset)
		}
	}
	
	// Create combined files
	var combinedPaths []string
	
	if len(cssAssets) > 0 {
		cssPath, err := am.combineAssets(cssAssets, "css")
		if err != nil {
			return "", fmt.Errorf("failed to combine CSS assets: %w", err)
		}
		combinedPaths = append(combinedPaths, cssPath)
	}
	
	if len(jsAssets) > 0 {
		jsPath, err := am.combineAssets(jsAssets, "js")
		if err != nil {
			return "", fmt.Errorf("failed to combine JS assets: %w", err)
		}
		combinedPaths = append(combinedPaths, jsPath)
	}
	
	// Return the first combined path (or empty if none)
	if len(combinedPaths) > 0 {
		return combinedPaths[0], nil
	}
	
	return "", nil
}

// combineAssets combines multiple assets into a single file
func (am *AssetManager) combineAssets(assets []*Asset, ext string) (string, error) {
	// Create output directory
	outputDir := filepath.Join(am.config.OutputDir, "bundles")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", err
	}
	
	// Create output file
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.%s", ext, ext))
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer outputFile.Close()
	
	// Combine assets
	for i, asset := range assets {
		// Add comment
		fmt.Fprintf(outputFile, "/* %s */\n", asset.Path)
		
		// Read and write asset content
		assetFile, err := os.Open(asset.Path)
		if err != nil {
			return "", err
		}
		
		_, err = io.Copy(outputFile, assetFile)
		assetFile.Close()
		
		if err != nil {
			return "", err
		}
		
		// Add newline between assets
		if i < len(assets)-1 {
			fmt.Fprintln(outputFile)
		}
	}
	
	return outputPath, nil
}

// getOutputPath returns the output path for an asset
func (am *AssetManager) getOutputPath(asset *Asset) string {
	// Get relative path from source directory
	relPath, err := filepath.Rel(am.config.SourceDir, asset.Path)
	if err != nil {
		relPath = asset.Path
	}
	
	// Add version to filename
	ext := filepath.Ext(relPath)
	name := strings.TrimSuffix(relPath, ext)
	versionedName := fmt.Sprintf("%s.%s%s", name, asset.Version, ext)
	
	return filepath.Join(am.config.OutputDir, versionedName)
}

// startWatching starts the file watcher
func (am *AssetManager) startWatching() {
	if am.watcher == nil {
		return
	}
	
	for event := range am.watcher.Events() {
		am.handleFileEvent(event)
	}
}

// handleFileEvent handles file system events
func (am *AssetManager) handleFileEvent(event FileEvent) {
	// Check if file should be processed
	if !am.shouldProcessFile(event.Path) {
		return
	}
	
	if am.config.EnableLogging && am.logger != nil {
		am.logger.Debug("File changed",
			zap.String("file", event.Path),
			zap.String("op", event.Op.String()))
	}
	
	// Reprocess the file
	asset, err := am.processFile(event.Path)
	if err != nil {
		if am.config.EnableLogging && am.logger != nil {
			am.logger.Warn("Failed to reprocess file",
				zap.String("file", event.Path),
				zap.Error(err))
		}
		return
	}
	
	// Update asset
	am.mu.Lock()
	am.assets[event.Path] = asset
	am.mu.Unlock()
	
	// Update statistics
	am.stats.RecordFileChange(event.Path, event.Op)
}

// GetAsset returns an asset by path
func (am *AssetManager) GetAsset(path string) (*Asset, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	asset, exists := am.assets[path]
	return asset, exists
}

// GetBundle returns a bundle by name
func (am *AssetManager) GetBundle(name string) (*Bundle, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	bundle, exists := am.bundles[name]
	return bundle, exists
}

// GetAllAssets returns all assets
func (am *AssetManager) GetAllAssets() map[string]*Asset {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	assets := make(map[string]*Asset)
	for path, asset := range am.assets {
		assets[path] = asset
	}
	return assets
}

// GetAllBundles returns all bundles
func (am *AssetManager) GetAllBundles() map[string]*Bundle {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	bundles := make(map[string]*Bundle)
	for name, bundle := range am.bundles {
		bundles[name] = bundle
	}
	return bundles
}

// GetStats returns asset pipeline statistics
func (am *AssetManager) GetStats() *AssetStats {
	return am.stats
}

// Stop stops the asset manager
func (am *AssetManager) Stop() error {
	if am.watcher != nil {
		am.watcher.Stop()
	}
	
	if am.cache != nil {
		am.cache.Close()
	}
	
	return nil
}

// ConfigFromEnv creates config from environment variables
func ConfigFromEnv() *Config {
	config := DefaultConfig()
	
	// Override with environment variables if present
	// This would typically read from environment variables
	// For now, return the default config
	return config
}
