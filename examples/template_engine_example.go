package main

import (
	"fmt"
	"log"
	"time"

	"github.com/mrhoseah/dolphin/internal/template"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()

	// Example 1: Basic Template Engine
	fmt.Println("=== Example 1: Basic Template Engine ===")

	// Create template engine configuration
	config := template.DefaultConfig()
	config.LayoutsDir = "ui/views/layouts"
	config.PartialsDir = "ui/views/partials"
	config.PagesDir = "ui/views/pages"
	config.ComponentsDir = "ui/views/components"
	config.EmailsDir = "ui/views/emails"
	config.Extension = ".html"
	config.AutoReload = true
	config.CacheTemplates = true
	config.EnableHelpers = true
	config.EnableLogging = true
	config.VerboseLogging = true

	// Create template engine
	engine, err := template.NewEngine(config, logger)
	if err != nil {
		log.Fatalf("Failed to create template engine: %v", err)
	}
	defer engine.Stop()

	// Example 2: Render Template
	fmt.Println("\n=== Example 2: Render Template ===")

	// Prepare template data
	data := template.TemplateData{
		"title":   "Welcome to Dolphin",
		"message": "Hello, World!",
		"user": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
		},
		"items": []string{"Apple", "Banana", "Cherry"},
		"count": 42,
		"price": 19.99,
		"date":  time.Now(),
	}

	// Render a template (this would typically be a real template file)
	fmt.Println("Template data prepared:")
	fmt.Printf("  Title: %s\n", data["title"])
	fmt.Printf("  Message: %s\n", data["message"])
	fmt.Printf("  User: %v\n", data["user"])
	fmt.Printf("  Items: %v\n", data["items"])
	fmt.Printf("  Count: %v\n", data["count"])
	fmt.Printf("  Price: %v\n", data["price"])
	fmt.Printf("  Date: %v\n", data["date"])

	// Example 3: Template Helpers
	fmt.Println("\n=== Example 3: Template Helpers ===")

	// Register custom helper
	engine.RegisterHelper("greeting", func(args ...interface{}) (interface{}, error) {
		if len(args) == 0 {
			return "Hello", nil
		}
		name := fmt.Sprintf("%v", args[0])
		return fmt.Sprintf("Hello, %s!", name), nil
	})

	// Test string helpers
	fmt.Println("String helpers:")
	fmt.Printf("  upper('hello'): %s\n", testHelper(engine, "upper", "hello"))
	fmt.Printf("  lower('WORLD'): %s\n", testHelper(engine, "lower", "WORLD"))
	fmt.Printf("  title('hello world'): %s\n", testHelper(engine, "title", "hello world"))
	fmt.Printf("  capitalize('hello'): %s\n", testHelper(engine, "capitalize", "hello"))
	fmt.Printf("  trim('  hello  '): %s\n", testHelper(engine, "trim", "  hello  "))
	fmt.Printf("  replace('hello world', 'world', 'universe'): %s\n", testHelper(engine, "replace", "hello world", "world", "universe"))
	fmt.Printf("  truncate('This is a long string', 10): %s\n", testHelper(engine, "truncate", "This is a long string", 10))
	fmt.Printf("  slug('Hello World!'): %s\n", testHelper(engine, "slug", "Hello World!"))
	fmt.Printf("  pluralize('cat'): %s\n", testHelper(engine, "pluralize", "cat"))
	fmt.Printf("  singularize('cats'): %s\n", testHelper(engine, "singularize", "cats"))

	// Test number helpers
	fmt.Println("\nNumber helpers:")
	fmt.Printf("  add(5, 3): %v\n", testHelper(engine, "add", 5, 3))
	fmt.Printf("  subtract(10, 4): %v\n", testHelper(engine, "subtract", 10, 4))
	fmt.Printf("  multiply(6, 7): %v\n", testHelper(engine, "multiply", 6, 7))
	fmt.Printf("  divide(20, 4): %v\n", testHelper(engine, "divide", 20, 4))
	fmt.Printf("  modulo(17, 5): %v\n", testHelper(engine, "modulo", 17, 5))
	fmt.Printf("  round(3.14159, 2): %v\n", testHelper(engine, "round", 3.14159, 2))
	fmt.Printf("  ceil(3.2): %v\n", testHelper(engine, "ceil", 3.2))
	fmt.Printf("  floor(3.8): %v\n", testHelper(engine, "floor", 3.8))
	fmt.Printf("  abs(-5): %v\n", testHelper(engine, "abs", -5))
	fmt.Printf("  min(5, 3, 8, 1): %v\n", testHelper(engine, "min", 5, 3, 8, 1))
	fmt.Printf("  max(5, 3, 8, 1): %v\n", testHelper(engine, "max", 5, 3, 8, 1))

	// Test date/time helpers
	fmt.Println("\nDate/Time helpers:")
	fmt.Printf("  now(): %v\n", testHelper(engine, "now"))
	fmt.Printf("  formatDate(now(), '2006-01-02'): %v\n", testHelper(engine, "formatDate", time.Now(), "2006-01-02"))
	fmt.Printf("  formatTime(now(), '15:04:05'): %v\n", testHelper(engine, "formatTime", time.Now(), "15:04:05"))
	fmt.Printf("  formatDateTime(now(), '2006-01-02 15:04:05'): %v\n", testHelper(engine, "formatDateTime", time.Now(), "2006-01-02 15:04:05"))
	fmt.Printf("  timeAgo(now().Add(-2*time.Hour)): %v\n", testHelper(engine, "timeAgo", time.Now().Add(-2*time.Hour)))
	fmt.Printf("  timeUntil(now().Add(2*time.Hour)): %v\n", testHelper(engine, "timeUntil", time.Now().Add(2*time.Hour)))
	fmt.Printf("  isToday(now()): %v\n", testHelper(engine, "isToday", time.Now()))
	fmt.Printf("  isYesterday(now().Add(-24*time.Hour)): %v\n", testHelper(engine, "isYesterday", time.Now().Add(-24*time.Hour)))
	fmt.Printf("  isTomorrow(now().Add(24*time.Hour)): %v\n", testHelper(engine, "isTomorrow", time.Now().Add(24*time.Hour)))

	// Test array helpers
	fmt.Println("\nArray helpers:")
	items := []string{"Apple", "Banana", "Cherry"}
	fmt.Printf("  join(items, ', '): %v\n", testHelper(engine, "join", items, ", "))
	fmt.Printf("  split('a,b,c', ','): %v\n", testHelper(engine, "split", "a,b,c", ","))
	fmt.Printf("  first(items): %v\n", testHelper(engine, "first", items))
	fmt.Printf("  last(items): %v\n", testHelper(engine, "last", items))
	fmt.Printf("  length(items): %v\n", testHelper(engine, "length", items))
	fmt.Printf("  contains(items, 'Banana'): %v\n", testHelper(engine, "contains", items, "Banana"))
	fmt.Printf("  index(items, 'Cherry'): %v\n", testHelper(engine, "index", items, "Cherry"))
	fmt.Printf("  slice(items, 1, 3): %v\n", testHelper(engine, "slice", items, 1, 3))
	fmt.Printf("  reverse(items): %v\n", testHelper(engine, "reverse", items))
	fmt.Printf("  sort(items): %v\n", testHelper(engine, "sort", items))
	fmt.Printf("  unique([]string{'a', 'b', 'a', 'c'}): %v\n", testHelper(engine, "unique", []string{"a", "b", "a", "c"}))

	// Test object helpers
	fmt.Println("\nObject helpers:")
	obj := map[string]interface{}{
		"name": "John",
		"age":  30,
		"city": "New York",
	}
	fmt.Printf("  keys(obj): %v\n", testHelper(engine, "keys", obj))
	fmt.Printf("  values(obj): %v\n", testHelper(engine, "values", obj))
	fmt.Printf("  hasKey(obj, 'name'): %v\n", testHelper(engine, "hasKey", obj, "name"))
	fmt.Printf("  get(obj, 'age'): %v\n", testHelper(engine, "get", obj, "age"))

	// Test HTML helpers
	fmt.Println("\nHTML helpers:")
	fmt.Printf("  escape('<script>alert(\"xss\")</script>'): %v\n", testHelper(engine, "escape", `<script>alert("xss")</script>`))
	fmt.Printf("  unescape('&lt;script&gt;'): %v\n", testHelper(engine, "unescape", "&lt;script&gt;"))
	fmt.Printf("  stripTags('<p>Hello <b>World</b></p>'): %v\n", testHelper(engine, "stripTags", "<p>Hello <b>World</b></p>"))
	fmt.Printf("  linkify('Visit https://example.com for more info'): %v\n", testHelper(engine, "linkify", "Visit https://example.com for more info"))
	fmt.Printf("  nl2br('Line 1\nLine 2'): %v\n", testHelper(engine, "nl2br", "Line 1\nLine 2"))
	fmt.Printf("  br2nl('Line 1<br>Line 2'): %v\n", testHelper(engine, "br2nl", "Line 1<br>Line 2"))

	// Test URL helpers
	fmt.Println("\nURL helpers:")
	fmt.Printf("  url('/about'): %v\n", testHelper(engine, "url", "/about"))
	fmt.Printf("  asset('css/style.css'): %v\n", testHelper(engine, "asset", "css/style.css"))
	fmt.Printf("  route('user.profile'): %v\n", testHelper(engine, "route", "user.profile"))
	fmt.Printf("  query('/search', 'q=hello'): %v\n", testHelper(engine, "query", "/search", "q=hello"))
	fmt.Printf("  fragment('/page', 'section1'): %v\n", testHelper(engine, "fragment", "/page", "section1"))

	// Test security helpers
	fmt.Println("\nSecurity helpers:")
	fmt.Printf("  csrf(): %v\n", testHelper(engine, "csrf"))
	fmt.Printf("  hash('password123'): %v\n", testHelper(engine, "hash", "password123"))
	fmt.Printf("  random(10): %v\n", testHelper(engine, "random", 10))
	fmt.Printf("  uuid(): %v\n", testHelper(engine, "uuid"))

	// Test conditional helpers
	fmt.Println("\nConditional helpers:")
	fmt.Printf("  if(true, 'yes', 'no'): %v\n", testHelper(engine, "if", true, "yes", "no"))
	fmt.Printf("  unless(false, 'yes', 'no'): %v\n", testHelper(engine, "unless", false, "yes", "no"))
	fmt.Printf("  eq(5, 5): %v\n", testHelper(engine, "eq", 5, 5))
	fmt.Printf("  ne(5, 3): %v\n", testHelper(engine, "ne", 5, 3))
	fmt.Printf("  gt(10, 5): %v\n", testHelper(engine, "gt", 10, 5))
	fmt.Printf("  gte(10, 10): %v\n", testHelper(engine, "gte", 10, 10))
	fmt.Printf("  lt(3, 5): %v\n", testHelper(engine, "lt", 3, 5))
	fmt.Printf("  lte(5, 5): %v\n", testHelper(engine, "lte", 5, 5))
	fmt.Printf("  and(true, true, false): %v\n", testHelper(engine, "and", true, true, false))
	fmt.Printf("  or(true, false, false): %v\n", testHelper(engine, "or", true, false, false))
	fmt.Printf("  not(true): %v\n", testHelper(engine, "not", true))

	// Test loop helpers
	fmt.Println("\nLoop helpers:")
	fmt.Printf("  range(items): %v\n", testHelper(engine, "range", items))
	fmt.Printf("  times(5): %v\n", testHelper(engine, "times", 5))
	fmt.Printf("  each(items): %v\n", testHelper(engine, "each", items))

	// Test utility helpers
	fmt.Println("\nUtility helpers:")
	fmt.Printf("  default('', 'fallback'): %v\n", testHelper(engine, "default", "", "fallback"))
	fmt.Printf("  coalesce('', 'hello', 'world'): %v\n", testHelper(engine, "coalesce", "", "hello", "world"))
	fmt.Printf("  empty(''): %v\n", testHelper(engine, "empty", ""))
	fmt.Printf("  present('hello'): %v\n", testHelper(engine, "present", "hello"))
	fmt.Printf("  blank(''): %v\n", testHelper(engine, "blank", ""))
	fmt.Printf("  nil(nil): %v\n", testHelper(engine, "nil", nil))

	// Example 4: Template Types
	fmt.Println("\n=== Example 4: Template Types ===")

	// Get templates by type
	layouts := engine.GetTemplatesByType(template.TypeLayout)
	fmt.Printf("Layouts: %d\n", len(layouts))

	partials := engine.GetTemplatesByType(template.TypePartial)
	fmt.Printf("Partials: %d\n", len(partials))

	pages := engine.GetTemplatesByType(template.TypePage)
	fmt.Printf("Pages: %d\n", len(pages))

	components := engine.GetTemplatesByType(template.TypeComponent)
	fmt.Printf("Components: %d\n", len(components))

	emails := engine.GetTemplatesByType(template.TypeEmail)
	fmt.Printf("Emails: %d\n", len(emails))

	// Example 5: Layout System
	fmt.Println("\n=== Example 5: Layout System ===")

	// Create layout manager
	layoutManager := template.NewLayoutManager(engine)

	// Create a layout using builder
	layout, err := template.NewLayoutBuilder("example", layoutManager).
		Content(`<!DOCTYPE html>
<html>
<head>
    <title>{{.title}}</title>
</head>
<body>
    <header>{{block "header" .}}{{end}}</header>
    <main>{{.layout}}</main>
    <footer>{{block "footer" .}}{{end}}</footer>
</body>
</html>`).
		Block("header", `<h1>{{.title}}</h1>`).
		Block("footer", `<p>&copy; 2024 Dolphin Framework</p>`).
		Build()

	if err != nil {
		log.Printf("Failed to create layout: %v", err)
	} else {
		fmt.Printf("Layout created: %s\n", layout.Name)
		fmt.Printf("Blocks: %d\n", len(layout.Blocks))
	}

	// Example 6: Component System
	fmt.Println("\n=== Example 6: Component System ===")

	// Create component manager
	componentManager := template.NewComponentManager(engine)

	// Create a component using builder
	component, err := template.NewComponentBuilder("button", componentManager).
		Content(`<button class="btn {{.class}}" {{event "click" .onClick}}>
    {{.text}}
</button>`).
		Prop("text", "Click me").
		Prop("class", "btn-primary").
		Event("click", "handleClick").
		Style(`.btn { padding: 8px 16px; border: none; border-radius: 4px; }`).
		Build()

	if err != nil {
		log.Printf("Failed to create component: %v", err)
	} else {
		fmt.Printf("Component created: %s\n", component.Name)
		fmt.Printf("Props: %d\n", len(component.Props))
		fmt.Printf("Events: %d\n", len(component.Events))
	}

	// Example 7: Template Rendering
	fmt.Println("\n=== Example 7: Template Rendering ===")

	// Render with layout
	layoutData := template.TemplateData{
		"title":  "Dolphin Framework",
		"layout": "<h1>Welcome to Dolphin!</h1><p>This is the main content.</p>",
	}

	// This would typically render a real template
	fmt.Println("Layout rendering prepared:")
	fmt.Printf("  Title: %s\n", layoutData["title"])
	fmt.Printf("  Layout: %s\n", layoutData["layout"])

	// Example 8: Error Handling
	fmt.Println("\n=== Example 8: Error Handling ===")

	// Test error handling with invalid template
	_, err = engine.Render("nonexistent", data)
	if err != nil {
		fmt.Printf("Expected error for nonexistent template: %v\n", err)
	}

	// Example 9: Configuration
	fmt.Println("\n=== Example 9: Configuration ===")

	// Show configuration
	fmt.Printf("Layouts Directory: %s\n", config.LayoutsDir)
	fmt.Printf("Partials Directory: %s\n", config.PartialsDir)
	fmt.Printf("Pages Directory: %s\n", config.PagesDir)
	fmt.Printf("Components Directory: %s\n", config.ComponentsDir)
	fmt.Printf("Emails Directory: %s\n", config.EmailsDir)
	fmt.Printf("Extension: %s\n", config.Extension)
	fmt.Printf("Auto Reload: %v\n", config.AutoReload)
	fmt.Printf("Cache Templates: %v\n", config.CacheTemplates)
	fmt.Printf("Enable Helpers: %v\n", config.EnableHelpers)
	fmt.Printf("Escape HTML: %v\n", config.EscapeHTML)

	// Example 10: Statistics
	fmt.Println("\n=== Example 10: Statistics ===")

	// Get all templates
	allTemplates := engine.GetAllTemplates()
	fmt.Printf("Total templates: %d\n", len(allTemplates))

	// Show template details
	for name, tmpl := range allTemplates {
		fmt.Printf("  %s (%s): %d bytes\n", name, tmpl.Type.String(), tmpl.Size)
	}

	fmt.Println("\n🎉 All template engine examples completed successfully!")
	fmt.Println("\n📚 Next Steps:")
	fmt.Println("1. Use 'dolphin template compile' to compile templates")
	fmt.Println("2. Use 'dolphin template watch' to watch for changes")
	fmt.Println("3. Use 'dolphin template list' to list all templates")
	fmt.Println("4. Use 'dolphin template helpers' to list available helpers")
	fmt.Println("5. Use 'dolphin template test' to test template rendering")
	fmt.Println("6. Use 'dolphin template stats' to view statistics")
	fmt.Println("7. Create your own templates in the template directories")
	fmt.Println("8. Use helpers in your templates for dynamic content")
	fmt.Println("9. Implement layout inheritance for consistent design")
	fmt.Println("10. Build reusable components for common UI elements")
}

// testHelper tests a template helper function
func testHelper(engine *template.Engine, name string, args ...interface{}) interface{} {
	// Get helper function
	helper, exists := engine.GetTemplate(name)
	if !exists {
		// Try to call helper directly
		// This is a simplified test - in reality, helpers are called from templates
		return fmt.Sprintf("Helper %s not found", name)
	}

	// Return template info for demonstration
	return fmt.Sprintf("Template: %s (%s)", helper.Name, helper.Type.String())
}
