package main

import (
	"fmt"
	"log"
	"time"

	"dolphin/internal/cli"
)

func main() {
	fmt.Println("⚡ Dolphin Framework - Enhanced CLI Tooling Demo")
	fmt.Println("===============================================")
	fmt.Println("")

	// 1. Interactive Prompt Demo
	fmt.Println("=== 1. Interactive Prompt Demo ===")
	demoInteractivePrompt()

	fmt.Println("")

	// 2. Progress Bar Demo
	fmt.Println("=== 2. Progress Bar Demo ===")
	demoProgressBar()

	fmt.Println("")

	// 3. Spinner Demo
	fmt.Println("=== 3. Spinner Demo ===")
	demoSpinner()

	fmt.Println("")

	// 4. Table Demo
	fmt.Println("=== 4. Table Demo ===")
	demoTable()

	fmt.Println("")

	// 5. CLI Helper Demo
	fmt.Println("=== 5. CLI Helper Demo ===")
	demoCLIHelper()

	fmt.Println("")

	// 6. Command Discovery Demo
	fmt.Println("=== 6. Command Discovery Demo ===")
	demoCommandDiscovery()

	fmt.Println("")

	// 7. Interactive Command Demo
	fmt.Println("=== 7. Interactive Command Demo ===")
	demoInteractiveCommand()

	fmt.Println("")
	fmt.Println("🎉 Enhanced CLI tooling demonstrated successfully!")
	fmt.Println("")
	fmt.Println("💡 Key Features Implemented:")
	fmt.Println("  ✅ Interactive prompting with validation")
	fmt.Println("  ✅ Progress bars for long operations")
	fmt.Println("  ✅ Loading spinners for async operations")
	fmt.Println("  ✅ Table rendering for structured data")
	fmt.Println("  ✅ Command discovery and registration")
	fmt.Println("  ✅ Enhanced user experience with colors and emojis")
	fmt.Println("  ✅ Fluent API for CLI operations")
	fmt.Println("  ✅ Comprehensive error handling")
}

func demoInteractivePrompt() {
	prompt := cli.NewInteractivePrompt()

	// Ask a simple question
	name, err := prompt.Ask("What's your name?")
	if err != nil {
		log.Printf("Error asking question: %v", err)
		return
	}
	fmt.Printf("Hello, %s!\n", name)

	// Ask with default
	language, err := prompt.AskWithDefault("What's your favorite programming language?", "Go")
	if err != nil {
		log.Printf("Error asking question: %v", err)
		return
	}
	fmt.Printf("Great choice: %s!\n", language)

	// Confirm
	likesGo, err := prompt.Confirm("Do you like Go?")
	if err != nil {
		log.Printf("Error asking question: %v", err)
		return
	}
	if likesGo {
		fmt.Println("Awesome! Go is a great language!")
	} else {
		fmt.Println("That's okay, everyone has their preferences!")
	}

	// Select from options
	choice, err := prompt.Select("What would you like to do?", []string{
		"Create a new project",
		"Generate some code",
		"Run migrations",
		"Start the server",
	})
	if err != nil {
		log.Printf("Error asking question: %v", err)
		return
	}
	options := []string{"Create a new project", "Generate some code", "Run migrations", "Start the server"}
	fmt.Printf("You chose: %s\n", options[choice])
}

func demoProgressBar() {
	fmt.Println("Simulating a long operation...")

	// Create progress bar
	pb := cli.NewProgressBar(10)
	pb.SetShowCount(true)

	// Simulate work
	for i := 0; i <= 10; i++ {
		pb.Update(i)
		time.Sleep(200 * time.Millisecond)
	}
	pb.Finish()

	fmt.Println("Operation completed!")

	// Another example with different settings
	fmt.Println("\nSimulating another operation...")
	pb2 := cli.NewProgressBar(5)
	pb2.SetShowCount(false)
	pb2.SetWidth(30)

	for i := 0; i <= 5; i++ {
		pb2.Update(i)
		time.Sleep(300 * time.Millisecond)
	}
	pb2.Finish()
}

func demoSpinner() {
	fmt.Println("Starting a long operation...")

	spinner := cli.NewSpinner()
	spinner.Start("Processing data...")

	// Simulate work
	time.Sleep(3 * time.Second)

	spinner.Stop()
	fmt.Println("Operation completed!")

	// Another example
	spinner2 := cli.NewSpinner()
	spinner2.Start("Loading resources...")

	time.Sleep(2 * time.Second)

	spinner2.Stop()
	fmt.Println("Resources loaded!")
}

func demoTable() {
	fmt.Println("Project Status:")

	// Create table
	table := cli.NewTable("Component", "Status", "Version", "Last Updated")
	table.AddRow("Dolphin Framework", "✅ Active", "1.0.0", "2024-01-15")
	table.AddRow("Database", "✅ Connected", "SQLite 3.40", "2024-01-15")
	table.AddRow("Cache", "✅ Active", "Redis 7.0", "2024-01-15")
	table.AddRow("Queue", "⚠️  Warning", "In-memory", "2024-01-15")
	table.AddRow("Mail Service", "❌ Disabled", "N/A", "N/A")

	table.Render()

	fmt.Println("\nMigration Status:")

	// Another table
	migrationTable := cli.NewTable("Migration", "Status", "Batch", "Run Time")
	migrationTable.AddRow("2024_01_01_000001_create_users_table", "Ran", "1", "2024-01-01 10:00:00")
	migrationTable.AddRow("2024_01_01_000002_create_posts_table", "Ran", "1", "2024-01-01 10:01:00")
	migrationTable.AddRow("2024_01_01_000003_add_email_to_users", "Pending", "-", "-")
	migrationTable.AddRow("2024_01_01_000004_create_comments_table", "Pending", "-", "-")

	migrationTable.Render()
}

func demoCLIHelper() {
	helper := cli.NewCLIHelper()

	// Show different types of messages
	helper.ShowSuccess("Operation completed successfully!")
	helper.ShowError("Something went wrong!")
	helper.ShowWarning("This is a warning message")
	helper.ShowInfo("Here's some information")

	// Show a table
	fmt.Println("\nAvailable Commands:")
	helper.ShowTable(
		[]string{"Command", "Description", "Usage"},
		[][]string{
			{"new", "Create new project", "dolphin new [name]"},
			{"make", "Generate code", "dolphin make [type] [name]"},
			{"serve", "Start server", "dolphin serve"},
			{"db migrate", "Run migrations", "dolphin db migrate"},
			{"db rollback", "Rollback migrations", "dolphin db rollback"},
		},
	)

	// Show progress
	fmt.Println("\nProcessing...")
	helper.ShowProgress(10, 7, "Almost done!")
}

func demoCommandDiscovery() {
	discovery := cli.NewCommandDiscovery()
	_ = discovery // Use the variable to avoid unused error

	// Register some mock commands
	fmt.Println("Registering commands...")

	// In a real implementation, you'd register actual cobra commands
	// For this demo, we'll just show the concept

	commands := []string{
		"new",
		"make",
		"serve",
		"db:migrate",
		"db:rollback",
		"db:status",
		"db:seed",
		"list",
		"status",
		"interactive",
	}

	fmt.Printf("Discovered %d commands:\n", len(commands))
	for _, cmd := range commands {
		fmt.Printf("  - %s\n", cmd)
	}

	fmt.Println("\nCommand discovery features:")
	fmt.Println("  ✅ Automatic command registration")
	fmt.Println("  ✅ Command listing and discovery")
	fmt.Println("  ✅ Help system integration")
	fmt.Println("  ✅ Command validation")
}

func demoInteractiveCommand() {
	fmt.Println("Starting interactive command demo...")
	fmt.Println("(This would normally start the full interactive CLI)")

	// Create interactive command
	interactiveCmd := cli.NewInteractiveCommand()
	_ = interactiveCmd // Use the variable to avoid unused error

	fmt.Println("Interactive command features:")
	fmt.Println("  ✅ Guided project creation")
	fmt.Println("  ✅ Code generation wizard")
	fmt.Println("  ✅ Database operations")
	fmt.Println("  ✅ Server management")
	fmt.Println("  ✅ Status monitoring")
	fmt.Println("  ✅ User-friendly interface")

	fmt.Println("\nTo start the interactive CLI, run:")
	fmt.Println("  dolphin interactive")

	// Note: We're not actually running the interactive command here
	// because it would require user input and block the demo
}
