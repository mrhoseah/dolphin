package main

import (
	"fmt"
	"time"

	"dolphin/internal/template"
)

func main() {
	fmt.Println("🚀 Dolphin Framework - Template Engine Demo")
	fmt.Println("===========================================")
	fmt.Println("")

	// Example 1: Basic Template Engine
	fmt.Println("=== Example 1: Basic Template Engine ===")

	// Create template engine configuration
	config := template.DefaultConfig()
	config.ViewsPath = "ui/views"
	config.CachePath = "storage/cache/views"
	config.CacheEnabled = true
	config.DebugMode = true
	config.Extensions = []string{".html", ".blade.go"}

	// Create template engine
	engine := template.NewFinEngine(config)
	fmt.Println("✅ Template engine created successfully")

	// Example 2: Template Data Preparation
	fmt.Println("\n=== Example 2: Template Data Preparation ===")

	// Prepare template data
	data := map[string]interface{}{
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

	fmt.Println("Template data prepared:")
	fmt.Printf("  Title: %s\n", data["title"])
	fmt.Printf("  Message: %s\n", data["message"])
	fmt.Printf("  User: %v\n", data["user"])
	fmt.Printf("  Items: %v\n", data["items"])
	fmt.Printf("  Count: %v\n", data["count"])
	fmt.Printf("  Price: %v\n", data["price"])
	fmt.Printf("  Date: %v\n", data["date"])

	// Example 3: Custom Directives
	fmt.Println("\n=== Example 3: Custom Directives ===")

	// Register custom directive
	engine.RegisterDirective("greeting", func(args []string, content string, data interface{}) (string, error) {
		if len(args) == 0 {
			return "Hello", nil
		}
		name := args[0]
		return fmt.Sprintf("Hello, %s!", name), nil
	})

	fmt.Println("✅ Custom directive 'greeting' registered")

	// Example 4: Template System Features
	fmt.Println("\n=== Example 4: Template System Features ===")

	// Template system features
	fmt.Println("Template system features:")
	fmt.Println("  - Directive system")
	fmt.Printf("  - Available directives: %d\n", len(engine.GetDirectives()))
	fmt.Printf("  - Available components: %d\n", len(engine.GetComponents()))
	fmt.Printf("  - Available layouts: %d\n", len(engine.GetLayouts()))

	// Example 5: Layout System
	fmt.Println("\n=== Example 5: Layout System ===")

	// Layout system features
	fmt.Println("Layout system features:")
	fmt.Println("  - Template inheritance")
	fmt.Println("  - Block system")
	fmt.Println("  - Component composition")
	fmt.Println("  - Dynamic content injection")

	// Example 6: Component System
	fmt.Println("\n=== Example 6: Component System ===")

	// Component system features
	fmt.Println("Component system features:")
	fmt.Println("  - Reusable UI components")
	fmt.Println("  - Props and events")
	fmt.Println("  - Styling and theming")
	fmt.Println("  - Dynamic rendering")

	// Example 7: Template Rendering
	fmt.Println("\n=== Example 7: Template Rendering ===")

	// Render with layout
	layoutData := map[string]interface{}{
		"title":  "Dolphin Framework",
		"layout": "<h1>Welcome to Dolphin!</h1><p>This is the main content.</p>",
	}

	fmt.Println("Layout rendering prepared:")
	fmt.Printf("  Title: %s\n", layoutData["title"])
	fmt.Printf("  Layout: %s\n", layoutData["layout"])

	// Example 8: Error Handling
	fmt.Println("\n=== Example 8: Error Handling ===")

	// Test error handling with invalid template
	_, err := engine.Render("nonexistent", data)
	if err != nil {
		fmt.Printf("✅ Expected error for nonexistent template: %v\n", err)
	}

	// Example 9: Configuration
	fmt.Println("\n=== Example 9: Configuration ===")

	// Show configuration
	fmt.Printf("Views Path: %s\n", config.ViewsPath)
	fmt.Printf("Cache Path: %s\n", config.CachePath)
	fmt.Printf("Cache Enabled: %v\n", config.CacheEnabled)
	fmt.Printf("Debug Mode: %v\n", config.DebugMode)
	fmt.Printf("Extensions: %v\n", config.Extensions)

	// Example 10: Template Helpers Demo
	fmt.Println("\n=== Example 10: Template Helpers Demo ===")

	// Demo available helpers
	fmt.Println("Available template helpers:")
	fmt.Println("  String helpers: upper, lower, title, capitalize, trim, replace, truncate, slug")
	fmt.Println("  Number helpers: add, subtract, multiply, divide, modulo, round, ceil, floor, abs")
	fmt.Println("  Date/Time helpers: now, formatDate, formatTime, formatDateTime, timeAgo, timeUntil")
	fmt.Println("  Array helpers: join, split, first, last, length, contains, index, slice, reverse, sort")
	fmt.Println("  Object helpers: keys, values, hasKey, get")
	fmt.Println("  HTML helpers: escape, unescape, stripTags, linkify, nl2br, br2nl")
	fmt.Println("  URL helpers: url, asset, route, query, fragment")
	fmt.Println("  Security helpers: csrf, hash, random, uuid")
	fmt.Println("  Conditional helpers: if, unless, eq, ne, gt, gte, lt, lte, and, or, not")
	fmt.Println("  Loop helpers: range, times, each")
	fmt.Println("  Utility helpers: default, coalesce, empty, present, blank, nil")

	fmt.Println("\n🎉 All template engine examples completed successfully!")
	fmt.Println("\n📚 Next Steps:")
	fmt.Println("1. Create your own templates in the template directories")
	fmt.Println("2. Use helpers in your templates for dynamic content")
	fmt.Println("3. Implement layout inheritance for consistent design")
	fmt.Println("4. Build reusable components for common UI elements")
	fmt.Println("5. Register custom directives for specialized functionality")
	fmt.Println("6. Use the template engine in your web applications")
	fmt.Println("7. Explore advanced features like caching and debugging")
	fmt.Println("8. Integrate with the Dolphin Framework's routing system")
}
