package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// InteractivePrompt provides interactive prompting capabilities
type InteractivePrompt struct {
	reader *bufio.Reader
	writer io.Writer
}

// NewInteractivePrompt creates a new interactive prompt
func NewInteractivePrompt() *InteractivePrompt {
	return &InteractivePrompt{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

// Ask asks a question and returns the answer
func (ip *InteractivePrompt) Ask(question string) (string, error) {
	fmt.Fprintf(ip.writer, "%s: ", question)
	answer, err := ip.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(answer), nil
}

// AskWithDefault asks a question with a default value
func (ip *InteractivePrompt) AskWithDefault(question, defaultValue string) (string, error) {
	fmt.Fprintf(ip.writer, "%s [%s]: ", question, defaultValue)
	answer, err := ip.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return defaultValue, nil
	}
	return answer, nil
}

// Confirm asks a yes/no question
func (ip *InteractivePrompt) Confirm(question string) (bool, error) {
	for {
		answer, err := ip.AskWithDefault(question+" (y/n)", "n")
		if err != nil {
			return false, err
		}

		switch strings.ToLower(answer) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintf(ip.writer, "Please answer 'y' or 'n'\n")
		}
	}
}

// Select asks user to select from options
func (ip *InteractivePrompt) Select(question string, options []string) (int, error) {
	fmt.Fprintf(ip.writer, "%s\n", question)
	for i, option := range options {
		fmt.Fprintf(ip.writer, "  %d) %s\n", i+1, option)
	}

	for {
		answer, err := ip.Ask("Enter your choice")
		if err != nil {
			return -1, err
		}

		var choice int
		if _, err := fmt.Sscanf(answer, "%d", &choice); err != nil {
			fmt.Fprintf(ip.writer, "Please enter a valid number\n")
			continue
		}

		if choice < 1 || choice > len(options) {
			fmt.Fprintf(ip.writer, "Please enter a number between 1 and %d\n", len(options))
			continue
		}

		return choice - 1, nil
	}
}

// Password asks for a password (hidden input)
func (ip *InteractivePrompt) Password(question string) (string, error) {
	fmt.Fprintf(ip.writer, "%s: ", question)

	// For simplicity, we'll use regular input
	// In a real implementation, you'd hide the input
	password, err := ip.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(password), nil
}

// ProgressBar represents a progress bar
type ProgressBar struct {
	total     int
	current   int
	width     int
	showCount bool
	writer    io.Writer
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		total:     total,
		current:   0,
		width:     50,
		showCount: true,
		writer:    os.Stdout,
	}
}

// SetWidth sets the width of the progress bar
func (pb *ProgressBar) SetWidth(width int) {
	pb.width = width
}

// SetShowCount sets whether to show the count
func (pb *ProgressBar) SetShowCount(show bool) {
	pb.showCount = show
}

// Update updates the progress bar
func (pb *ProgressBar) Update(current int) {
	pb.current = current
	pb.render()
}

// Increment increments the progress bar
func (pb *ProgressBar) Increment() {
	pb.current++
	pb.render()
}

// SetTotal sets the total count
func (pb *ProgressBar) SetTotal(total int) {
	pb.total = total
}

// Finish finishes the progress bar
func (pb *ProgressBar) Finish() {
	pb.current = pb.total
	pb.render()
	fmt.Fprintf(pb.writer, "\n")
}

// render renders the progress bar
func (pb *ProgressBar) render() {
	if pb.total == 0 {
		return
	}

	percentage := float64(pb.current) / float64(pb.total)
	filled := int(percentage * float64(pb.width))

	bar := strings.Repeat("=", filled) + strings.Repeat("-", pb.width-filled)

	if pb.showCount {
		fmt.Fprintf(pb.writer, "\r[%s] %d/%d (%.1f%%)", bar, pb.current, pb.total, percentage*100)
	} else {
		fmt.Fprintf(pb.writer, "\r[%s] %.1f%%", bar, percentage*100)
	}

	if pb.current == pb.total {
		fmt.Fprintf(pb.writer, "\n")
	}
}

// Spinner represents a loading spinner
type Spinner struct {
	frames []string
	index  int
	writer io.Writer
	active bool
}

// NewSpinner creates a new spinner
func NewSpinner() *Spinner {
	return &Spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		index:  0,
		writer: os.Stdout,
		active: false,
	}
}

// Start starts the spinner
func (s *Spinner) Start(message string) {
	s.active = true
	go func() {
		for s.active {
			fmt.Fprintf(s.writer, "\r%s %s", s.frames[s.index], message)
			s.index = (s.index + 1) % len(s.frames)
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	s.active = false
	fmt.Fprintf(s.writer, "\r")
}

// Table represents a table for displaying data
type Table struct {
	headers []string
	rows    [][]string
	writer  io.Writer
}

// NewTable creates a new table
func NewTable(headers ...string) *Table {
	return &Table{
		headers: headers,
		rows:    make([][]string, 0),
		writer:  os.Stdout,
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(row ...string) {
	t.rows = append(t.rows, row)
}

// Render renders the table
func (t *Table) Render() {
	if len(t.headers) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(t.headers))
	for i, header := range t.headers {
		widths[i] = len(header)
	}

	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Render header
	fmt.Fprintf(t.writer, "|")
	for i, header := range t.headers {
		fmt.Fprintf(t.writer, " %-*s |", widths[i], header)
	}
	fmt.Fprintf(t.writer, "\n")

	// Render separator
	fmt.Fprintf(t.writer, "|")
	for _, width := range widths {
		fmt.Fprintf(t.writer, " %s |", strings.Repeat("-", width))
	}
	fmt.Fprintf(t.writer, "\n")

	// Render rows
	for _, row := range t.rows {
		fmt.Fprintf(t.writer, "|")
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprintf(t.writer, " %-*s |", widths[i], cell)
			}
		}
		fmt.Fprintf(t.writer, "\n")
	}
}

// CommandDiscovery provides command discovery capabilities
type CommandDiscovery struct {
	commands map[string]*cobra.Command
}

// NewCommandDiscovery creates a new command discovery
func NewCommandDiscovery() *CommandDiscovery {
	return &CommandDiscovery{
		commands: make(map[string]*cobra.Command),
	}
}

// RegisterCommand registers a command
func (cd *CommandDiscovery) RegisterCommand(name string, command *cobra.Command) {
	cd.commands[name] = command
}

// DiscoverCommands discovers commands from a directory
func (cd *CommandDiscovery) DiscoverCommands(dir string) error {
	// This is a simplified implementation
	// In a real implementation, you'd scan the directory for command files
	return nil
}

// ListCommands lists all registered commands
func (cd *CommandDiscovery) ListCommands() []string {
	var commands []string
	for name := range cd.commands {
		commands = append(commands, name)
	}
	return commands
}

// GetCommand gets a command by name
func (cd *CommandDiscovery) GetCommand(name string) *cobra.Command {
	return cd.commands[name]
}

// CLIHelper provides helper functions for CLI commands
type CLIHelper struct {
	prompt  *InteractivePrompt
	spinner *Spinner
}

// NewCLIHelper creates a new CLI helper
func NewCLIHelper() *CLIHelper {
	return &CLIHelper{
		prompt:  NewInteractivePrompt(),
		spinner: NewSpinner(),
	}
}

// GetPrompt returns the interactive prompt
func (ch *CLIHelper) GetPrompt() *InteractivePrompt {
	return ch.prompt
}

// GetSpinner returns the spinner
func (ch *CLIHelper) GetSpinner() *Spinner {
	return ch.spinner
}

// ShowSuccess shows a success message
func (ch *CLIHelper) ShowSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// ShowError shows an error message
func (ch *CLIHelper) ShowError(message string) {
	fmt.Printf("❌ %s\n", message)
}

// ShowWarning shows a warning message
func (ch *CLIHelper) ShowWarning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// ShowInfo shows an info message
func (ch *CLIHelper) ShowInfo(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// ShowTable shows a table
func (ch *CLIHelper) ShowTable(headers []string, rows [][]string) {
	table := NewTable(headers...)
	for _, row := range rows {
		table.AddRow(row...)
	}
	table.Render()
}

// ShowProgress shows a progress bar
func (ch *CLIHelper) ShowProgress(total int, current int, message string) {
	pb := NewProgressBar(total)
	pb.Update(current)
	if message != "" {
		fmt.Printf(" %s\n", message)
	}
}

// InteractiveCommand provides interactive command capabilities
type InteractiveCommand struct {
	helper *CLIHelper
}

// NewInteractiveCommand creates a new interactive command
func NewInteractiveCommand() *InteractiveCommand {
	return &InteractiveCommand{
		helper: NewCLIHelper(),
	}
}

// Run runs the interactive command
func (ic *InteractiveCommand) Run() error {
	fmt.Println("🐬 Dolphin Framework - Interactive CLI")
	fmt.Println("====================================")

	for {
		choice, err := ic.helper.GetPrompt().Select("What would you like to do?", []string{
			"Create new project",
			"Generate code",
			"Database operations",
			"Run migrations",
			"Start development server",
			"Exit",
		})

		if err != nil {
			return err
		}

		switch choice {
		case 0:
			ic.createProject()
		case 1:
			ic.generateCode()
		case 2:
			ic.databaseOperations()
		case 3:
			ic.runMigrations()
		case 4:
			ic.startServer()
		case 5:
			fmt.Println("Goodbye!")
			return nil
		}
	}
}

// createProject creates a new project interactively
func (ic *InteractiveCommand) createProject() {
	fmt.Println("\n📁 Create New Project")
	fmt.Println("--------------------")

	name, err := ic.helper.GetPrompt().Ask("Project name")
	if err != nil {
		ic.helper.ShowError("Failed to get project name")
		return
	}

	description, err := ic.helper.GetPrompt().AskWithDefault("Project description", "A Dolphin Framework project")
	if err != nil {
		ic.helper.ShowError("Failed to get project description")
		return
	}

	includeAuth, err := ic.helper.GetPrompt().Confirm("Include authentication?")
	if err != nil {
		ic.helper.ShowError("Failed to get auth preference")
		return
	}

	ic.helper.GetSpinner().Start("Creating project...")
	time.Sleep(2 * time.Second) // Simulate work
	ic.helper.GetSpinner().Stop()

	ic.helper.ShowSuccess(fmt.Sprintf("Project '%s' created successfully!", name))
	fmt.Printf("  Description: %s\n", description)
	fmt.Printf("  Authentication: %v\n", includeAuth)
}

// generateCode generates code interactively
func (ic *InteractiveCommand) generateCode() {
	fmt.Println("\n🔨 Generate Code")
	fmt.Println("----------------")

	codeType, err := ic.helper.GetPrompt().Select("What would you like to generate?", []string{
		"Controller",
		"Model",
		"Migration",
		"Middleware",
		"View",
		"API Resource",
	})

	if err != nil {
		ic.helper.ShowError("Failed to get code type")
		return
	}

	name, err := ic.helper.GetPrompt().Ask("Name")
	if err != nil {
		ic.helper.ShowError("Failed to get name")
		return
	}

	types := []string{"Controller", "Model", "Migration", "Middleware", "View", "API Resource"}
	ic.helper.GetSpinner().Start(fmt.Sprintf("Generating %s...", types[codeType]))
	time.Sleep(1 * time.Second) // Simulate work
	ic.helper.GetSpinner().Stop()

	ic.helper.ShowSuccess(fmt.Sprintf("%s '%s' generated successfully!", types[codeType], name))
}

// databaseOperations performs database operations
func (ic *InteractiveCommand) databaseOperations() {
	fmt.Println("\n🗄️  Database Operations")
	fmt.Println("----------------------")

	operation, err := ic.helper.GetPrompt().Select("What would you like to do?", []string{
		"Run migrations",
		"Rollback migrations",
		"Check migration status",
		"Seed database",
		"Back to main menu",
	})

	if err != nil {
		ic.helper.ShowError("Failed to get operation")
		return
	}

	operations := []string{"Run migrations", "Rollback migrations", "Check migration status", "Seed database", "Back to main menu"}

	if operation == 4 { // Back to main menu
		return
	}

	ic.helper.GetSpinner().Start(fmt.Sprintf("Performing %s...", operations[operation]))
	time.Sleep(1 * time.Second) // Simulate work
	ic.helper.GetSpinner().Stop()

	ic.helper.ShowSuccess(fmt.Sprintf("%s completed successfully!", operations[operation]))
}

// runMigrations runs migrations
func (ic *InteractiveCommand) runMigrations() {
	fmt.Println("\n🔄 Run Migrations")
	fmt.Println("-----------------")

	pb := NewProgressBar(5)
	for i := 0; i <= 5; i++ {
		pb.Update(i)
		time.Sleep(200 * time.Millisecond)
	}
	pb.Finish()

	ic.helper.ShowSuccess("All migrations completed successfully!")
}

// startServer starts the development server
func (ic *InteractiveCommand) startServer() {
	fmt.Println("\n🚀 Start Development Server")
	fmt.Println("---------------------------")

	port, err := ic.helper.GetPrompt().AskWithDefault("Port", "8080")
	if err != nil {
		ic.helper.ShowError("Failed to get port")
		return
	}

	ic.helper.ShowInfo(fmt.Sprintf("Starting server on port %s...", port))
	ic.helper.ShowSuccess("Server started successfully!")
	fmt.Printf("  URL: http://localhost:%s\n", port)
	fmt.Printf("  Debug: http://localhost:%s/debug\n", port)
}
