package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/mrhoseah/dolphin/internal/assets"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()

	// Example 1: Basic Asset Pipeline
	fmt.Println("=== Example 1: Basic Asset Pipeline ===")

	// Create asset pipeline configuration
	config := assets.DefaultConfig()
	config.SourceDir = "resources/assets"
	config.OutputDir = "public/assets"
	config.PublicDir = "public"
	config.EnableBundling = true
	config.EnableVersioning = true
	config.EnableOptimization = true
	config.EnableWatch = true
	config.EnableLogging = true
	config.VerboseLogging = true

	// Create asset manager
	manager, err := assets.NewAssetManager(config, logger)
	if err != nil {
		log.Fatalf("Failed to create asset manager: %v", err)
	}
	defer manager.Stop()

	// Example 2: Process Assets
	fmt.Println("\n=== Example 2: Process Assets ===")

	// Process all assets
	if err := manager.ProcessAssets(); err != nil {
		log.Printf("Failed to process assets: %v", err)
	} else {
		fmt.Println("Assets processed successfully!")
	}

	// Example 3: List Assets
	fmt.Println("\n=== Example 3: List Assets ===")

	// Get all assets
	allAssets := manager.GetAllAssets()
	fmt.Printf("Total assets: %d\n", len(allAssets))

	// Group by type
	assetsByType := make(map[assets.AssetType][]*assets.Asset)
	for _, asset := range allAssets {
		assetsByType[asset.Type] = append(assetsByType[asset.Type], asset)
	}

	for assetType, typeAssets := range assetsByType {
		fmt.Printf("%s: %d files\n", assetType.String(), len(typeAssets))
	}

	// Example 4: List Bundles
	fmt.Println("\n=== Example 4: List Bundles ===")

	// Get all bundles
	allBundles := manager.GetAllBundles()
	fmt.Printf("Total bundles: %d\n", len(allBundles))

	for bundleName, bundle := range allBundles {
		fmt.Printf("Bundle %s: %d assets, %d bytes\n",
			bundleName, len(bundle.Assets), bundle.Size)
	}

	// Example 5: Asset Statistics
	fmt.Println("\n=== Example 5: Asset Statistics ===")

	// Get statistics
	stats := manager.GetStats()
	fmt.Printf("File Change Rate: %.2f/min\n", stats.GetFileChangeRate())
	fmt.Printf("Processing Rate: %.2f/min\n", stats.GetProcessingRate())

	// Get most used types
	mostUsedTypes := stats.GetMostUsedTypes(3)
	fmt.Println("Most used asset types:")
	for _, typeUsage := range mostUsedTypes {
		fmt.Printf("  %s: %d files\n", typeUsage.Type.String(), typeUsage.Count)
	}

	// Get most used bundles
	mostUsedBundles := stats.GetMostUsedBundles(3)
	fmt.Println("Most used bundles:")
	for _, bundleUsage := range mostUsedBundles {
		fmt.Printf("  %s: %d files\n", bundleUsage.Bundle, bundleUsage.Count)
	}

	// Example 6: Custom Configuration
	fmt.Println("\n=== Example 6: Custom Configuration ===")

	// Create custom configuration
	customConfig := &assets.Config{
		SourceDir:          "resources/assets",
		OutputDir:          "public/assets",
		PublicDir:          "public",
		EnableBundling:     true,
		BundleTypes:        []string{"app", "vendor", "common"},
		MinifyAssets:       true,
		CombineAssets:      true,
		EnableVersioning:   true,
		VersionStrategy:    "hash",
		VersionLength:      8,
		EnableOptimization: true,
		OptimizeImages:     true,
		OptimizeCSS:        true,
		OptimizeJS:         true,
		EnableCache:        true,
		CacheDir:           "storage/cache/assets",
		CacheExpiry:        24 * time.Hour,
		EnableWatch:        true,
		WatchExtensions:    []string{".css", ".js", ".scss", ".sass", ".less", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".woff", ".woff2", ".ttf", ".eot"},
		CDNUrl:             "",
		CDNEnabled:         false,
		EnableLogging:      true,
		VerboseLogging:     false,
	}

	customManager, err := assets.NewAssetManager(customConfig, logger)
	if err != nil {
		log.Printf("Failed to create custom asset manager: %v", err)
	} else {
		fmt.Printf("Created custom asset manager with %d watch extensions\n", len(customConfig.WatchExtensions))
		customManager.Stop()
	}

	// Example 7: Asset Optimization
	fmt.Println("\n=== Example 7: Asset Optimization ===")

	// Create optimizer
	optimizer := assets.NewOptimizer(config, logger)

	// Get optimization stats
	optStats := optimizer.GetOptimizationStats()
	fmt.Println("Optimization configuration:")
	for key, value := range optStats {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Example 8: File Watching
	fmt.Println("\n=== Example 8: File Watching ===")

	// Check if watching is enabled
	if config.EnableWatch {
		fmt.Println("File watching is enabled")
		fmt.Printf("Watching extensions: %v\n", config.WatchExtensions)
		fmt.Printf("Source directory: %s\n", config.SourceDir)
	} else {
		fmt.Println("File watching is disabled")
	}

	// Example 9: Versioning
	fmt.Println("\n=== Example 9: Versioning ===")

	// Show versioning configuration
	fmt.Printf("Versioning enabled: %v\n", config.EnableVersioning)
	fmt.Printf("Version strategy: %s\n", config.VersionStrategy)
	fmt.Printf("Version length: %d\n", config.VersionLength)

	// Show some asset versions
	for path, asset := range allAssets {
		if len(asset.Version) > 0 {
			fmt.Printf("  %s → %s\n", filepath.Base(path), asset.Version)
			break // Just show one example
		}
	}

	// Example 10: CDN Integration
	fmt.Println("\n=== Example 10: CDN Integration ===")

	// Show CDN configuration
	fmt.Printf("CDN enabled: %v\n", config.CDNEnabled)
	if config.CDNEnabled {
		fmt.Printf("CDN URL: %s\n", config.CDNUrl)
	} else {
		fmt.Println("CDN integration is disabled")
	}

	// Example 11: Cache Management
	fmt.Println("\n=== Example 11: Cache Management ===")

	// Show cache configuration
	fmt.Printf("Cache enabled: %v\n", config.EnableCache)
	if config.EnableCache {
		fmt.Printf("Cache directory: %s\n", config.CacheDir)
		fmt.Printf("Cache expiry: %v\n", config.CacheExpiry)
	}

	// Example 12: Bundle Management
	fmt.Println("\n=== Example 12: Bundle Management ===")

	// Show bundle configuration
	fmt.Printf("Bundling enabled: %v\n", config.EnableBundling)
	fmt.Printf("Bundle types: %v\n", config.BundleTypes)
	fmt.Printf("Combine assets: %v\n", config.CombineAssets)

	// Show bundle details
	for bundleName, bundle := range allBundles {
		fmt.Printf("Bundle %s:\n", bundleName)
		fmt.Printf("  Type: %s\n", bundle.Type.String())
		fmt.Printf("  Assets: %d\n", len(bundle.Assets))
		fmt.Printf("  Size: %d bytes\n", bundle.Size)
		fmt.Printf("  Version: %s\n", bundle.Version)
		if bundle.CombinedPath != "" {
			fmt.Printf("  Combined: %s\n", bundle.CombinedPath)
		}
	}

	// Example 13: Performance Monitoring
	fmt.Println("\n=== Example 13: Performance Monitoring ===")

	// Get detailed statistics
	detailedStats := stats.GetStats()
	fmt.Println("Detailed statistics:")
	for key, value := range detailedStats {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Example 14: Error Handling
	fmt.Println("\n=== Example 14: Error Handling ===")

	// Test error handling with invalid configuration
	invalidConfig := &assets.Config{
		SourceDir: "/nonexistent/directory",
		OutputDir: "/nonexistent/output",
	}

	_, err = assets.NewAssetManager(invalidConfig, logger)
	if err != nil {
		fmt.Printf("Expected error with invalid config: %v\n", err)
	}

	// Example 15: Cleanup
	fmt.Println("\n=== Example 15: Cleanup ===")

	// Stop the manager
	if err := manager.Stop(); err != nil {
		log.Printf("Error stopping asset manager: %v", err)
	} else {
		fmt.Println("Asset manager stopped successfully")
	}

	fmt.Println("\n🎉 All asset pipeline examples completed successfully!")
	fmt.Println("\n📚 Next Steps:")
	fmt.Println("1. Use 'dolphin asset build' to build assets")
	fmt.Println("2. Use 'dolphin asset watch' to watch for changes")
	fmt.Println("3. Use 'dolphin asset list' to list all assets")
	fmt.Println("4. Use 'dolphin asset stats' to view statistics")
	fmt.Println("5. Use 'dolphin asset optimize' to optimize assets")
	fmt.Println("6. Use 'dolphin asset version' to view asset versions")
	fmt.Println("7. Use 'dolphin asset clean' to clean built assets")
	fmt.Println("8. Configure CDN integration for production")
	fmt.Println("9. Set up asset caching for better performance")
	fmt.Println("10. Monitor asset pipeline performance in production")
}
