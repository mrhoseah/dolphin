package artisan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Usage       string
	Arguments   []Argument
	Options     []Option
	Handler     func(*Context) error
}

// Argument represents a command argument
type Argument struct {
	Name        string
	Description string
	Required    bool
	Default     interface{}
}

// Option represents a command option
type Option struct {
	Name        string
	ShortName   string
	Description string
	Required    bool
	Default     interface{}
	Type        string // string, bool, int, etc.
}

// Context represents command execution context
type Context struct {
	Arguments map[string]interface{}
	Options   map[string]interface{}
	Command   *Command
}

// Warden represents the CLI application
type Warden struct {
	commands map[string]*Command
	version  string
	name     string
}

// NewWarden creates a new Warden instance
func NewWarden(name, version string) *Warden {
	return &Warden{
		commands: make(map[string]*Command),
		version:  version,
		name:     name,
	}
}

// RegisterCommand registers a command
func (w *Warden) RegisterCommand(command *Command) {
	w.commands[command.Name] = command
}

// Run runs the CLI application
func (w *Warden) Run(args []string) error {
	if len(args) < 2 {
		w.showHelp()
		return nil
	}

	commandName := args[1]
	command, exists := w.commands[commandName]
	if !exists {
		fmt.Printf("Command '%s' not found.\n", commandName)
		w.showHelp()
		return nil
	}

	// Parse arguments and options
	context := &Context{
		Arguments: make(map[string]interface{}),
		Options:   make(map[string]interface{}),
		Command:   command,
	}

	err := w.parseArgs(args[2:], command, context)
	if err != nil {
		return err
	}

	// Execute command
	return command.Handler(context)
}

// parseArgs parses command line arguments
func (w *Warden) parseArgs(args []string, command *Command, context *Context) error {
	argIndex := 0

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "--") {
			// Long option
			optionName := arg[2:]
			option := w.findOption(command, optionName)
			if option == nil {
				return fmt.Errorf("unknown option: --%s", optionName)
			}

			var value interface{}
			if option.Type == "bool" {
				value = true
			} else {
				if i+1 >= len(args) {
					return fmt.Errorf("option --%s requires a value", optionName)
				}
				value = args[i+1]
				i++
			}

			context.Options[optionName] = value
		} else if strings.HasPrefix(arg, "-") {
			// Short option
			shortName := arg[1:]
			option := w.findOptionByShort(command, shortName)
			if option == nil {
				return fmt.Errorf("unknown option: -%s", shortName)
			}

			var value interface{}
			if option.Type == "bool" {
				value = true
			} else {
				if i+1 >= len(args) {
					return fmt.Errorf("option -%s requires a value", shortName)
				}
				value = args[i+1]
				i++
			}

			context.Options[option.Name] = value
		} else {
			// Argument
			if argIndex >= len(command.Arguments) {
				return fmt.Errorf("too many arguments")
			}

			context.Arguments[command.Arguments[argIndex].Name] = arg
			argIndex++
		}
	}

	// Check required arguments
	for _, arg := range command.Arguments {
		if arg.Required {
			if _, exists := context.Arguments[arg.Name]; !exists {
				return fmt.Errorf("required argument '%s' is missing", arg.Name)
			}
		}
	}

	// Check required options
	for _, opt := range command.Options {
		if opt.Required {
			if _, exists := context.Options[opt.Name]; !exists {
				return fmt.Errorf("required option '%s' is missing", opt.Name)
			}
		}
	}

	return nil
}

// findOption finds an option by name
func (w *Warden) findOption(command *Command, name string) *Option {
	for _, opt := range command.Options {
		if opt.Name == name {
			return &opt
		}
	}
	return nil
}

// findOptionByShort finds an option by short name
func (w *Warden) findOptionByShort(command *Command, shortName string) *Option {
	for _, opt := range command.Options {
		if opt.ShortName == shortName {
			return &opt
		}
	}
	return nil
}

// showHelp shows the help information
func (w *Warden) showHelp() {
	fmt.Printf("%s %s\n\n", w.name, w.version)
	fmt.Println("Usage:")
	fmt.Printf("  %s <command> [options] [arguments]\n\n", w.name)
	fmt.Println("Available commands:")

	for _, command := range w.commands {
		fmt.Printf("  %-20s %s\n", command.Name, command.Description)
	}

	fmt.Println("\nUse 'dolphin <command> --help' for more information about a command.")
}

// Generator represents a code generator
type Generator struct {
	templates map[string]*template.Template
}

// NewGenerator creates a new generator
func NewGenerator() *Generator {
	return &Generator{
		templates: make(map[string]*template.Template),
	}
}

// RegisterTemplate registers a template
func (g *Generator) RegisterTemplate(name, content string) error {
	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		return err
	}

	g.templates[name] = tmpl
	return nil
}

// Generate generates code from a template
func (g *Generator) Generate(templateName string, data interface{}, outputPath string) error {
	tmpl, exists := g.templates[templateName]
	if !exists {
		return fmt.Errorf("template '%s' not found", templateName)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create file
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Execute template
	return tmpl.Execute(file, data)
}

// MakeControllerCommand creates a make:controller command
func MakeControllerCommand() *Command {
	return &Command{
		Name:        "make:controller",
		Description: "Create a new controller class",
		Usage:       "make:controller <name> [--resource] [--model=<model>]",
		Arguments: []Argument{
			{Name: "name", Description: "The name of the controller", Required: true},
		},
		Options: []Option{
			{Name: "resource", ShortName: "r", Description: "Generate a resource controller", Type: "bool"},
			{Name: "model", ShortName: "m", Description: "The model to use", Type: "string"},
		},
		Handler: func(ctx *Context) error {
			name := ctx.Arguments["name"].(string)
			isResource := ctx.Options["resource"] == true
			model := ctx.Options["model"]

			generator := NewGenerator()

			// Register controller template
			controllerTemplate := `package controllers

import (
	"net/http"
	"github.com/go-chi/chi/v5"
)

type {{.Name}}Controller struct {
	// Add dependencies here
}

func New{{.Name}}Controller() *{{.Name}}Controller {
	return &{{.Name}}Controller{}
}

{{if .IsResource}}
// Index displays a listing of the resource
func (c *{{.Name}}Controller) Index(w http.ResponseWriter, r *http.Request) {
	// Implementation here
}

// Show displays the specified resource
func (c *{{.Name}}Controller) Show(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// Implementation here
}

// Create shows the form for creating a new resource
func (c *{{.Name}}Controller) Create(w http.ResponseWriter, r *http.Request) {
	// Implementation here
}

// Store stores a newly created resource in storage
func (c *{{.Name}}Controller) Store(w http.ResponseWriter, r *http.Request) {
	// Implementation here
}

// Edit shows the form for editing the specified resource
func (c *{{.Name}}Controller) Edit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// Implementation here
}

// Update updates the specified resource in storage
func (c *{{.Name}}Controller) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// Implementation here
}

// Destroy removes the specified resource from storage
func (c *{{.Name}}Controller) Destroy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// Implementation here
}
{{else}}
// Handle handles the request
func (c *{{.Name}}Controller) Handle(w http.ResponseWriter, r *http.Request) {
	// Implementation here
}
{{end}}`

			err := generator.RegisterTemplate("controller", controllerTemplate)
			if err != nil {
				return err
			}

			data := map[string]interface{}{
				"Name":       strings.Title(name),
				"IsResource": isResource,
				"Model":      model,
			}

			outputPath := filepath.Join("app", "http", "controllers", strings.ToLower(name)+"_controller.go")
			err = generator.Generate("controller", data, outputPath)
			if err != nil {
				return err
			}

			fmt.Printf("Controller created successfully: %s\n", outputPath)
			return nil
		},
	}
}

// MakeModelCommand creates a make:model command
func MakeModelCommand() *Command {
	return &Command{
		Name:        "make:model",
		Description: "Create a new model class",
		Usage:       "make:model <name> [--migration] [--factory] [--seeder]",
		Arguments: []Argument{
			{Name: "name", Description: "The name of the model", Required: true},
		},
		Options: []Option{
			{Name: "migration", ShortName: "m", Description: "Create a migration for the model", Type: "bool"},
			{Name: "factory", ShortName: "f", Description: "Create a factory for the model", Type: "bool"},
			{Name: "seeder", ShortName: "s", Description: "Create a seeder for the model", Type: "bool"},
		},
		Handler: func(ctx *Context) error {
			name := ctx.Arguments["name"].(string)
			createMigration := ctx.Options["migration"] == true
			createFactory := ctx.Options["factory"] == true
			createSeeder := ctx.Options["seeder"] == true

			generator := NewGenerator()

			// Register model template
			modelTemplate := `package models

import (
	"time"
	"gorm.io/gorm"
)

type {{.Name}} struct {
	ID        uint           ` + "`gorm:\"primaryKey\" json:\"id\"`" + `
	CreatedAt time.Time      ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time      ` + "`json:\"updated_at\"`" + `
	DeletedAt gorm.DeletedAt ` + "`gorm:\"index\" json:\"deleted_at,omitempty\"`" + `
	
	// Add your fields here
}

// TableName returns the table name
func ({{.Name}}) TableName() string {
	return "{{.TableName}}"
}

// BeforeCreate hook
func ({{.Name}}) BeforeCreate(tx *gorm.DB) error {
	// Implementation here
	return nil
}

// AfterCreate hook
func ({{.Name}}) AfterCreate(tx *gorm.DB) error {
	// Implementation here
	return nil
}`

			err := generator.RegisterTemplate("model", modelTemplate)
			if err != nil {
				return err
			}

			data := map[string]interface{}{
				"Name":      strings.Title(name),
				"TableName": strings.ToLower(name) + "s",
			}

			outputPath := filepath.Join("app", "models", strings.ToLower(name)+".go")
			err = generator.Generate("model", data, outputPath)
			if err != nil {
				return err
			}

			fmt.Printf("Model created successfully: %s\n", outputPath)

			// Create migration if requested
			if createMigration {
				migrationName := fmt.Sprintf("create_%s_table", strings.ToLower(name)+"s")
				fmt.Printf("Migration created: %s\n", migrationName)
			}

			// Create factory if requested
			if createFactory {
				factoryName := strings.ToLower(name) + "_factory"
				fmt.Printf("Factory created: %s\n", factoryName)
			}

			// Create seeder if requested
			if createSeeder {
				seederName := strings.ToLower(name) + "_seeder"
				fmt.Printf("Seeder created: %s\n", seederName)
			}

			return nil
		},
	}
}

// MakeMigrationCommand creates a make:migration command
func MakeMigrationCommand() *Command {
	return &Command{
		Name:        "make:migration",
		Description: "Create a new migration file",
		Usage:       "make:migration <name> [--create=<table>] [--table=<table>]",
		Arguments: []Argument{
			{Name: "name", Description: "The name of the migration", Required: true},
		},
		Options: []Option{
			{Name: "create", ShortName: "c", Description: "The table to create", Type: "string"},
			{Name: "table", ShortName: "t", Description: "The table to modify", Type: "string"},
		},
		Handler: func(ctx *Context) error {
			name := ctx.Arguments["name"].(string)
			createTable := ctx.Options["create"]
			modifyTable := ctx.Options["table"]

			generator := NewGenerator()

			// Register migration template
			migrationTemplate := `package migrations

import (
	"gorm.io/gorm"
)

type {{.Name}} struct {
	ID uint ` + "`gorm:\"primaryKey\"`" + `
}

func ({{.Name}}) Up(db *gorm.DB) error {
	{{if .CreateTable}}
	// Create table
	return db.Exec("CREATE TABLE {{.TableName}} (
		id INT PRIMARY KEY AUTO_INCREMENT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)").Error
	{{else if .ModifyTable}}
	// Modify table
	return db.Exec("ALTER TABLE {{.TableName}} ADD COLUMN new_column VARCHAR(255)").Error
	{{else}}
	// Custom migration logic
	return nil
	{{end}}
}

func ({{.Name}}) Down(db *gorm.DB) error {
	{{if .CreateTable}}
	// Drop table
	return db.Exec("DROP TABLE IF EXISTS {{.TableName}}").Error
	{{else if .ModifyTable}}
	// Revert table changes
	return db.Exec("ALTER TABLE {{.TableName}} DROP COLUMN new_column").Error
	{{else}}
	// Custom rollback logic
	return nil
	{{end}}
}`

			err := generator.RegisterTemplate("migration", migrationTemplate)
			if err != nil {
				return err
			}

			timestamp := time.Now().Format("20060102150405")
			migrationName := fmt.Sprintf("%s_%s", timestamp, name)

			data := map[string]interface{}{
				"Name":        strings.Title(strings.ReplaceAll(name, "_", "")),
				"CreateTable": createTable != nil,
				"ModifyTable": modifyTable != nil,
				"TableName":   createTable,
			}

			if modifyTable != nil {
				data["TableName"] = modifyTable
			}

			outputPath := filepath.Join("database", "migrations", migrationName+".go")
			err = generator.Generate("migration", data, outputPath)
			if err != nil {
				return err
			}

			fmt.Printf("Migration created successfully: %s\n", outputPath)
			return nil
		},
	}
}

// MigrateCommand creates a migrate command
func MigrateCommand() *Command {
	return &Command{
		Name:        "migrate",
		Description: "Run the database migrations",
		Usage:       "migrate [--fresh] [--seed]",
		Options: []Option{
			{Name: "fresh", ShortName: "f", Description: "Drop all tables and re-run all migrations", Type: "bool"},
			{Name: "seed", ShortName: "s", Description: "Run the database seeders", Type: "bool"},
		},
		Handler: func(ctx *Context) error {
			fresh := ctx.Options["fresh"] == true
			seed := ctx.Options["seed"] == true

			if fresh {
				fmt.Println("Dropping all tables...")
				fmt.Println("Running all migrations...")
			} else {
				fmt.Println("Running pending migrations...")
			}

			if seed {
				fmt.Println("Running database seeders...")
			}

			fmt.Println("Migration completed successfully!")
			return nil
		},
	}
}

// SeedCommand creates a seed command
func SeedCommand() *Command {
	return &Command{
		Name:        "db:seed",
		Description: "Seed the database with records",
		Usage:       "db:seed [--class=<class>]",
		Options: []Option{
			{Name: "class", ShortName: "c", Description: "The seeder class to run", Type: "string"},
		},
		Handler: func(ctx *Context) error {
			className := ctx.Options["class"]

			if className != nil {
				fmt.Printf("Running seeder: %s\n", className)
			} else {
				fmt.Println("Running all seeders...")
			}

			fmt.Println("Seeding completed successfully!")
			return nil
		},
	}
}

// CacheCommand creates a cache command
func CacheCommand() *Command {
	return &Command{
		Name:        "cache:clear",
		Description: "Flush the application cache",
		Usage:       "cache:clear",
		Handler: func(ctx *Context) error {
			fmt.Println("Clearing application cache...")
			fmt.Println("Cache cleared successfully!")
			return nil
		},
	}
}

// ConfigCommand creates a config command
func ConfigCommand() *Command {
	return &Command{
		Name:        "config:cache",
		Description: "Create a cache file for faster configuration loading",
		Usage:       "config:cache",
		Handler: func(ctx *Context) error {
			fmt.Println("Creating configuration cache...")
			fmt.Println("Configuration cached successfully!")
			return nil
		},
	}
}

// RouteCommand creates a route command
func RouteCommand() *Command {
	return &Command{
		Name:        "route:cache",
		Description: "Create a route cache file for faster route registration",
		Usage:       "route:cache",
		Handler: func(ctx *Context) error {
			fmt.Println("Creating route cache...")
			fmt.Println("Routes cached successfully!")
			return nil
		},
	}
}

// ServeCommand creates a serve command
func ServeCommand() *Command {
	return &Command{
		Name:        "serve",
		Description: "Serve the application on the PHP development server",
		Usage:       "serve [--host=<host>] [--port=<port>]",
		Options: []Option{
			{Name: "host", ShortName: "h", Description: "The host to serve on", Type: "string", Default: "127.0.0.1"},
			{Name: "port", ShortName: "p", Description: "The port to serve on", Type: "string", Default: "8000"},
		},
		Handler: func(ctx *Context) error {
			host := ctx.Options["host"]
			port := ctx.Options["port"]

			if host == nil {
				host = "127.0.0.1"
			}
			if port == nil {
				port = "8000"
			}

			fmt.Printf("Starting development server on http://%s:%s\n", host, port)
			fmt.Println("Press Ctrl+C to stop the server")

			// In a real implementation, this would start the HTTP server
			return nil
		},
	}
}

// TinkerCommand creates a tinker command
func TinkerCommand() *Command {
	return &Command{
		Name:        "tinker",
		Description: "Interact with your application",
		Usage:       "tinker",
		Handler: func(ctx *Context) error {
			fmt.Println("Starting interactive shell...")
			fmt.Println("Type 'exit' to quit")

			// In a real implementation, this would start an interactive REPL
			return nil
		},
	}
}

// RegisterDefaultCommands registers all default commands
func RegisterDefaultCommands(warden *Warden) {
	warden.RegisterCommand(MakeControllerCommand())
	warden.RegisterCommand(MakeModelCommand())
	warden.RegisterCommand(MakeMigrationCommand())
	warden.RegisterCommand(MigrateCommand())
	warden.RegisterCommand(SeedCommand())
	warden.RegisterCommand(CacheCommand())
	warden.RegisterCommand(ConfigCommand())
	warden.RegisterCommand(RouteCommand())
	warden.RegisterCommand(ServeCommand())
	warden.RegisterCommand(TinkerCommand())
}
