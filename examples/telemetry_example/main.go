package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"dolphin/internal/telemetry"
)

func main() {
	fmt.Println("📊 Dolphin Framework - Telemetry System Example")
	fmt.Println("================================================")
	fmt.Println("")

	// Initialize telemetry system
	configPath := telemetry.GetConfigPath()
	storage := telemetry.NewFileStorage(configPath)
	sender := telemetry.NewHTTPSender("https://telemetry.dolphin-framework.dev/api/v1/events")
	manager := telemetry.NewTelemetryManager(storage, sender)

	// Add collectors
	systemCollector := telemetry.NewSystemCollector()
	performanceCollector := telemetry.NewPerformanceCollector()
	errorCollector := telemetry.NewErrorCollector()
	featureCollector := telemetry.NewFeatureCollector()

	manager.AddCollector("system", systemCollector)
	manager.AddCollector("performance", performanceCollector)
	manager.AddCollector("errors", errorCollector)
	manager.AddCollector("features", featureCollector)

	// Add observers
	consoleObserver := telemetry.NewConsoleObserver("console")
	logObserver := telemetry.NewLogObserver("logger")

	manager.AddObserver(consoleObserver)
	manager.AddObserver(logObserver)

	// Start telemetry manager
	if err := manager.Start(); err != nil {
		log.Fatalf("Failed to start telemetry manager: %v", err)
	}
	defer manager.Stop()

	fmt.Println("🚀 Telemetry system initialized!")
	fmt.Println("")

	// Demonstrate different telemetry events
	fmt.Println("📈 Collecting telemetry data...")

	// 1. System startup event
	startupData := map[string]interface{}{
		"startup_time": time.Now().Unix(),
		"version":      "1.0.0",
		"environment":  "development",
	}
	manager.CollectEvent(context.Background(), telemetry.EventTypeStartup, startupData)

	// 2. Feature usage tracking
	featureCollector.TrackFeature("cli_command")
	featureCollector.TrackFeature("database_migration")
	featureCollector.TrackFeature("api_endpoint")

	// 3. Performance metrics
	performanceCollector.AddMetric("response_time", 150)
	performanceCollector.AddMetric("memory_usage", 1024*1024*50) // 50MB
	performanceCollector.AddMetric("cpu_usage", 25.5)

	// 4. Error tracking
	errorCollector.AddError(fmt.Errorf("database connection timeout"), "database")
	errorCollector.AddError(fmt.Errorf("invalid JSON format"), "api")

	// 5. Custom events
	customData := map[string]interface{}{
		"user_action": "button_click",
		"page":        "/dashboard",
		"timestamp":   time.Now().Unix(),
	}
	manager.CollectEvent(context.Background(), telemetry.EventTypeCustom, customData)

	// 6. Collect from all collectors
	if err := manager.CollectFromCollectors(context.Background()); err != nil {
		log.Printf("Failed to collect from collectors: %v", err)
	}

	fmt.Println("✅ Telemetry data collected!")
	fmt.Println("")

	// Show configuration
	fmt.Println("⚙️ Current Configuration:")
	config := manager.GetConfig()
	fmt.Printf("  Enabled: %t\n", config.Enabled)
	fmt.Printf("  Endpoint: %s\n", config.Endpoint)
	fmt.Printf("  Privacy Mode: %t\n", config.PrivacyMode)
	fmt.Printf("  Batch Size: %d\n", config.BatchSize)
	fmt.Printf("  Flush Interval: %s\n", config.FlushInterval)
	fmt.Println("")

	// Demonstrate CLI usage
	fmt.Println("🖥️ CLI Commands Available:")
	fmt.Println("  dolphin telemetry enable     # Enable telemetry")
	fmt.Println("  dolphin telemetry disable    # Disable telemetry")
	fmt.Println("  dolphin telemetry status      # Show status")
	fmt.Println("  dolphin telemetry config      # Show configuration")
	fmt.Println("  dolphin telemetry test        # Send test event")
	fmt.Println("  dolphin telemetry privacy     # Show privacy info")
	fmt.Println("  dolphin telemetry reset       # Reset to defaults")
	fmt.Println("")

	// Privacy information
	fmt.Println("🔒 Privacy Information:")
	fmt.Println("  • No personal data is collected")
	fmt.Println("  • No application data is collected")
	fmt.Println("  • Only anonymous usage statistics")
	fmt.Println("  • Data retention: 30 days")
	fmt.Println("  • You can disable at any time")
	fmt.Println("")

	fmt.Println("🎯 Telemetry system example completed!")
	fmt.Println("📊 Check the console output above for collected events")
}
