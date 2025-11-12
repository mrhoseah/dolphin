package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrhoseah/dolphin/internal/livereload"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()

	// Example 1: Basic Live Reload
	fmt.Println("=== Example 1: Basic Live Reload ===")

	// Create live reload configuration
	config := livereload.DefaultConfig()
	config.Strategy = livereload.StrategyRestart
	config.WatchPaths = []string{
		".",
		"cmd",
		"internal",
		"app",
		"ui",
		"public",
	}
	config.IgnorePaths = []string{
		".git",
		"node_modules",
		"vendor",
		"*.log",
		"*.tmp",
		".env",
	}
	config.EnableHotReload = true
	config.HotReloadPort = 35729
	config.EnableLogging = true
	config.VerboseLogging = true

	// Create live reload manager
	manager, err := livereload.NewLiveReloadManager(config, logger)
	if err != nil {
		log.Fatalf("Failed to create live reload manager: %v", err)
	}

	// Example 2: Start Live Reload
	fmt.Println("\n=== Example 2: Start Live Reload ===")

	// Start the live reload manager
	if err := manager.Start(); err != nil {
		log.Fatalf("Failed to start live reload manager: %v", err)
	}

	fmt.Println("Live reload manager started successfully!")
	fmt.Printf("Watching paths: %v\n", manager.GetWatchedPaths())
	fmt.Printf("Process running: %v\n", manager.IsRunning())

	// Example 3: Monitor Statistics
	fmt.Println("\n=== Example 3: Monitor Statistics ===")

	// Monitor statistics for a while
	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second)

		stats := manager.GetStats()
		fmt.Printf("Iteration %d: File Changes=%d, Reloads=%d, Uptime=%v\n",
			i+1, stats.FileChanges, stats.Reloads, time.Since(stats.StartTime))
	}

	// Example 4: Different Strategies
	fmt.Println("\n=== Example 4: Different Strategies ===")

	// Create manager with rebuild strategy
	rebuildConfig := livereload.DefaultConfig()
	rebuildConfig.Strategy = livereload.StrategyRebuild
	rebuildConfig.BuildCommand = "go build -o bin/app cmd/dolphin/main.go"
	rebuildConfig.RunCommand = "./bin/app serve"

	rebuildManager, err := livereload.NewLiveReloadManager(rebuildConfig, logger)
	if err != nil {
		log.Printf("Failed to create rebuild manager: %v", err)
	} else {
		fmt.Printf("Created rebuild manager with strategy: %s\n", rebuildConfig.Strategy.String())
		rebuildManager.Stop()
	}

	// Create manager with hot reload strategy
	hotReloadConfig := livereload.DefaultConfig()
	hotReloadConfig.Strategy = livereload.StrategyHotReload
	hotReloadConfig.EnableHotReload = true
	hotReloadConfig.HotReloadPort = 35730

	hotReloadManager, err := livereload.NewLiveReloadManager(hotReloadConfig, logger)
	if err != nil {
		log.Printf("Failed to create hot reload manager: %v", err)
	} else {
		fmt.Printf("Created hot reload manager with strategy: %s\n", hotReloadConfig.Strategy.String())
		hotReloadManager.Stop()
	}

	// Example 5: Custom Configuration
	fmt.Println("\n=== Example 5: Custom Configuration ===")

	// Create custom configuration
	customConfig := &livereload.Config{
		WatchPaths: []string{
			"cmd",
			"internal",
			"ui",
		},
		IgnorePaths: []string{
			".git",
			"*.log",
		},
		FileExtensions: []string{
			".go",
			".html",
			".css",
		},
		Strategy:        livereload.StrategyRestart,
		BuildCommand:    "go build -o bin/app cmd/dolphin/main.go",
		RunCommand:      "./bin/app serve",
		BuildTimeout:    60 * time.Second,
		RestartDelay:    2 * time.Second,
		EnableHotReload: true,
		HotReloadPort:   35731,
		HotReloadPaths:  []string{"/", "/admin"},
		DebounceDelay:   1 * time.Second,
		MaxDebounce:     10 * time.Second,
		EnableLogging:   true,
		VerboseLogging:  true,
	}

	customManager, err := livereload.NewLiveReloadManager(customConfig, logger)
	if err != nil {
		log.Printf("Failed to create custom manager: %v", err)
	} else {
		fmt.Printf("Created custom manager with %d watch paths\n", len(customConfig.WatchPaths))
		customManager.Stop()
	}

	// Example 6: Statistics and Monitoring
	fmt.Println("\n=== Example 6: Statistics and Monitoring ===")

	// Get detailed statistics
	stats := manager.GetStats()
	fmt.Printf("File Change Rate: %.2f/min\n", stats.GetFileChangeRate())
	fmt.Printf("Reload Rate: %.2f/min\n", stats.GetReloadRate())

	// Get most changed files
	mostChanged := stats.GetMostChangedFiles(5)
	fmt.Println("Most changed files:")
	for _, file := range mostChanged {
		fmt.Printf("  %s: %d changes\n", file.Filename, file.Count)
	}

	// Get change type statistics
	changeTypes := stats.GetChangeTypeStats()
	fmt.Println("Change types:")
	for changeType, count := range changeTypes {
		fmt.Printf("  %s: %d\n", changeType, count)
	}

	// Example 7: Graceful Shutdown
	fmt.Println("\n=== Example 7: Graceful Shutdown ===")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start graceful shutdown in goroutine
	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal, stopping live reload manager...")

		if err := manager.Stop(); err != nil {
			log.Printf("Error stopping live reload manager: %v", err)
		}

		fmt.Println("Live reload manager stopped successfully!")
		os.Exit(0)
	}()

	// Example 8: Hot Reload Server
	fmt.Println("\n=== Example 8: Hot Reload Server ===")

	// Get hot reload server stats if available
	if manager.GetStats().HotReloads > 0 {
		fmt.Printf("Hot reload server is running on port %d\n", config.HotReloadPort)
		fmt.Printf("WebSocket connections: %d\n", 0) // This would be from the server
	} else {
		fmt.Println("Hot reload server is not running")
	}

	// Example 9: File Watching
	fmt.Println("\n=== Example 9: File Watching ===")

	// Show watched paths
	watchedPaths := manager.GetWatchedPaths()
	fmt.Printf("Currently watching %d paths:\n", len(watchedPaths))
	for _, path := range watchedPaths {
		fmt.Printf("  • %s\n", path)
	}

	// Example 10: Process Management
	fmt.Println("\n=== Example 10: Process Management ===")

	// Check if process is running
	if manager.IsRunning() {
		fmt.Println("Main process is running")
	} else {
		fmt.Println("Main process is not running")
	}

	// Get process statistics
	processStats := manager.GetStats()
	fmt.Printf("Process starts: %d\n", processStats.ProcessStarts)
	fmt.Printf("Process stops: %d\n", processStats.ProcessStops)
	fmt.Printf("Last start: %v\n", processStats.LastStart)
	fmt.Printf("Last stop: %v\n", processStats.LastStop)

	// Keep the manager running
	fmt.Println("\n🎉 Live reload examples completed successfully!")
	fmt.Println("\n📚 Next Steps:")
	fmt.Println("1. Use 'dolphin dev start' to start live reload development")
	fmt.Println("2. Use 'dolphin dev status' to view current status")
	fmt.Println("3. Use 'dolphin dev stats' to view detailed statistics")
	fmt.Println("4. Use 'dolphin dev config' to view configuration")
	fmt.Println("5. Use 'dolphin dev test' to test live reload functionality")
	fmt.Println("6. Edit files in watched directories to see live reload in action")
	fmt.Println("7. Monitor the console for reload notifications")
	fmt.Println("8. Use Ctrl+C to stop the live reload manager")

	// Keep running until interrupted
	select {}
}
