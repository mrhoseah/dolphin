package telemetry

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CLI provides command-line interface for telemetry management
type CLI struct {
	manager *TelemetryManager
}

// NewCLI creates a new telemetry CLI
func NewCLI(manager *TelemetryManager) *CLI {
	return &CLI{
		manager: manager,
	}
}

// Enable enables telemetry collection
func (cli *CLI) Enable() error {
	fmt.Println("🔍 Enabling telemetry collection...")

	if err := cli.manager.Enable(); err != nil {
		return fmt.Errorf("failed to enable telemetry: %w", err)
	}

	fmt.Println("✅ Telemetry enabled successfully!")
	fmt.Println("📊 Data will be collected and sent to:", cli.manager.GetConfig().Endpoint)
	fmt.Println("🔒 Privacy mode:", cli.manager.GetConfig().PrivacyMode)

	return nil
}

// Disable disables telemetry collection
func (cli *CLI) Disable() error {
	fmt.Println("🔍 Disabling telemetry collection...")

	if err := cli.manager.Disable(); err != nil {
		return fmt.Errorf("failed to disable telemetry: %w", err)
	}

	fmt.Println("✅ Telemetry disabled successfully!")
	fmt.Println("📊 No data will be collected or sent")

	return nil
}

// Status shows the current telemetry status
func (cli *CLI) Status() error {
	config := cli.manager.GetConfig()

	fmt.Println("📊 Telemetry Status")
	fmt.Println("==================")
	fmt.Printf("Status: %s\n", getStatusText(config.Enabled))
	fmt.Printf("Endpoint: %s\n", config.Endpoint)
	fmt.Printf("Privacy Mode: %s\n", getBoolText(config.PrivacyMode))
	fmt.Printf("Batch Size: %d\n", config.BatchSize)
	fmt.Printf("Flush Interval: %s\n", config.FlushInterval)
	fmt.Printf("Retry Attempts: %d\n", config.RetryAttempts)
	fmt.Printf("Timeout: %s\n", config.Timeout)
	fmt.Printf("Data Retention: %s\n", config.DataRetention)

	fmt.Println("\n📈 Collectors:")
	for name, enabled := range config.Collectors {
		fmt.Printf("  %s: %s\n", name, getBoolText(enabled))
	}

	return nil
}

// Config shows detailed configuration
func (cli *CLI) Config() error {
	config := cli.manager.GetConfig()

	fmt.Println("⚙️ Telemetry Configuration")
	fmt.Println("==========================")
	fmt.Printf("Enabled: %t\n", config.Enabled)
	fmt.Printf("Endpoint: %s\n", config.Endpoint)
	fmt.Printf("Batch Size: %d\n", config.BatchSize)
	fmt.Printf("Flush Interval: %s\n", config.FlushInterval)
	fmt.Printf("Retry Attempts: %d\n", config.RetryAttempts)
	fmt.Printf("Timeout: %s\n", config.Timeout)
	fmt.Printf("Privacy Mode: %t\n", config.PrivacyMode)
	fmt.Printf("Data Retention: %s\n", config.DataRetention)

	fmt.Println("\nCollectors:")
	for name, enabled := range config.Collectors {
		fmt.Printf("  %s: %t\n", name, enabled)
	}

	return nil
}

// Reset resets telemetry configuration to defaults
func (cli *CLI) Reset() error {
	fmt.Println("🔄 Resetting telemetry configuration to defaults...")

	defaultConfig := DefaultConfig()
	if err := cli.manager.SetConfig(defaultConfig); err != nil {
		return fmt.Errorf("failed to reset configuration: %w", err)
	}

	fmt.Println("✅ Configuration reset successfully!")
	fmt.Println("📊 Telemetry is now disabled by default")

	return nil
}

// Test sends a test telemetry event
func (cli *CLI) Test() error {
	fmt.Println("🧪 Sending test telemetry event...")

	if !cli.manager.IsEnabled() {
		fmt.Println("⚠️ Telemetry is disabled. Enable it first with 'dolphin telemetry enable'")
		return nil
	}

	eventData := map[string]interface{}{
		"test":      true,
		"message":   "This is a test telemetry event",
		"timestamp": time.Now().Unix(),
	}

	if err := cli.manager.CollectEvent(nil, EventTypeCustom, eventData); err != nil {
		return fmt.Errorf("failed to send test event: %w", err)
	}

	fmt.Println("✅ Test event sent successfully!")
	fmt.Println("📊 Check your telemetry endpoint for the event")

	return nil
}

// Privacy shows privacy information
func (cli *CLI) Privacy() error {
	fmt.Println("🔒 Telemetry Privacy Information")
	fmt.Println("===============================")
	fmt.Println("")
	fmt.Println("What data is collected:")
	fmt.Println("• Framework version and Go version")
	fmt.Println("• Operating system and architecture")
	fmt.Println("• Feature usage statistics")
	fmt.Println("• Performance metrics (memory, CPU)")
	fmt.Println("• Error information (anonymized)")
	fmt.Println("• CLI command usage")
	fmt.Println("")
	fmt.Println("What data is NOT collected:")
	fmt.Println("• Personal information")
	fmt.Println("• Application data")
	fmt.Println("• Source code")
	fmt.Println("• File contents")
	fmt.Println("• Network traffic")
	fmt.Println("• User credentials")
	fmt.Println("")
	fmt.Println("Data retention: 30 days")
	fmt.Println("Data sharing: Anonymous usage statistics only")
	fmt.Println("")
	fmt.Println("You can disable telemetry at any time with:")
	fmt.Println("  dolphin telemetry disable")

	return nil
}

// getStatusText returns a formatted status text
func getStatusText(enabled bool) string {
	if enabled {
		return "✅ Enabled"
	}
	return "❌ Disabled"
}

// getBoolText returns a formatted boolean text
func getBoolText(value bool) string {
	if value {
		return "✅ Yes"
	}
	return "❌ No"
}

// GetConfigPath returns the default telemetry config path
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".dolphin/telemetry.json"
	}
	return filepath.Join(homeDir, ".dolphin", "telemetry.json")
}
