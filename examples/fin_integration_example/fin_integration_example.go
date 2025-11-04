package main

import (
	"fmt"
	"log"
	"os"

	"dolphin/internal/template"
)

func runFinIntegrationExample() {
	fmt.Println("🐬 Fin Template Integration Example")
	fmt.Println("===================================")

	// Note: In a real application, you would initialize Dolphin app with proper dependencies
	// app := app.New(config, logger, dbManager)

	// Initialize Fin template engine
	finConfig := &template.Config{
		ViewsPath:    "ui/views",
		CachePath:    "storage/cache/views",
		CacheEnabled: true,
		DebugMode:    true,
		Extensions:   []string{".fin.html"}, // Only .fin.html is supported
	}

	finEngine := template.NewFinEngine(finConfig)

	// Register a custom directive
	finEngine.RegisterDirective("uppercase", func(args []string, content string, data interface{}) (string, error) {
		return "{{. | upper}}", nil
	})

	// Register a component
	finEngine.RegisterComponent("alert", `
		<div class="alert alert-{{type}}">
			<strong>{{title}}</strong>
			<p>{{message}}</p>
		</div>
	`)

	// Register a layout
	finEngine.RegisterLayout("app", `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>{{.Title}} - Dolphin Framework</title>
			<script src="https://cdn.tailwindcss.com"></script>
			<script src="https://unpkg.com/htmx.org@1.9.10"></script>
		</head>
		<body class="bg-gray-100">
			<div class="min-h-screen">
				<nav class="bg-white shadow">
					<div class="max-w-7xl mx-auto px-4">
						<div class="flex justify-between h-16">
							<div class="flex items-center">
								<h1 class="text-xl font-semibold">🐬 Dolphin Framework</h1>
							</div>
						</div>
					</div>
				</nav>
				
				<main class="max-w-7xl mx-auto py-6 px-4">
					{{template "content" .}}
				</main>
			</div>
		</body>
		</html>
	`)

	// Sample data for demonstration
	data := map[string]interface{}{
		"Title": "Fin Template Demo",
		"User": map[string]interface{}{
			"Name":    "John Doe",
			"Email":   "john@example.com",
			"Role":    "Admin",
			"IsAdmin": true,
		},
		"Posts": []map[string]interface{}{
			{
				"Title":   "Getting Started with Fin Templates",
				"Content": "Fin templates provide clean, readable syntax for Go web applications...",
				"Author": map[string]interface{}{
					"Name": "John Doe",
				},
				"CreatedAt": "2024-01-15",
			},
			{
				"Title":   "Advanced Fin Features",
				"Content": "Learn about model annotations, components, and more...",
				"Author": map[string]interface{}{
					"Name": "Jane Smith",
				},
				"CreatedAt": "2024-01-16",
			},
		},
		"Version": "1.0.0",
		"AppName": "Dolphin Framework",
	}

	// Example 1: Simple template rendering
	fmt.Println("\n1. Simple Template Rendering:")
	simpleTemplate := `
		@extends('app')
		@model('User', user)
		
		@section('title')
			Welcome
		@endsection
		
		@section('content')
			<div class="hero">
				<h1 class="text-4xl font-bold text-gray-900">Welcome to {{AppName}}!</h1>
				<p class="text-xl text-gray-600">Hello, {{user.Name}}!</p>
				
				@if(user.IsAdmin)
					<div class="mt-4 p-4 bg-blue-100 border border-blue-400 text-blue-700 rounded">
						<p>Admin access granted</p>
					</div>
				@endif
			</div>
		@endsection
	`

	// For demonstration, we'll create a temporary template file
	tempFile := "temp_demo.fin.html"
	err := writeTemplateFile(tempFile, simpleTemplate)
	if err != nil {
		log.Fatal("Failed to create temp template:", err)
	}
	defer removeTemplateFile(tempFile)

	fmt.Println("Template content:")
	fmt.Println(simpleTemplate)
	fmt.Println("\nWould render to:")
	fmt.Printf("Welcome to %s!\n", data["AppName"])
	fmt.Printf("Hello, %s!\n", data["User"].(map[string]interface{})["Name"])
	if data["User"].(map[string]interface{})["IsAdmin"].(bool) {
		fmt.Println("Admin access granted")
	}

	// Example 2: Complex template with loops and conditionals
	fmt.Println("\n2. Complex Template with Loops:")
	complexTemplate := `
		@extends('app')
		@model('User', user)
		
		@section('title')
			Dashboard
		@endsection
		
		@section('content')
			<div class="dashboard">
				<h1 class="text-3xl font-bold mb-6">Dashboard</h1>
				
				<!-- User info -->
				<div class="bg-white rounded-lg shadow p-6 mb-6">
					<h2 class="text-xl font-semibold mb-4">User Information</h2>
					<p><strong>Name:</strong> {{user.Name}}</p>
					<p><strong>Email:</strong> {{user.Email}}</p>
					<p><strong>Role:</strong> {{user.Role}}</p>
				</div>
				
				<!-- Posts list -->
				<div class="bg-white rounded-lg shadow p-6">
					<h2 class="text-xl font-semibold mb-4">Recent Posts</h2>
					@foreach(Posts as post)
						<div class="border-b border-gray-200 pb-4 mb-4 last:border-b-0">
							<h3 class="text-lg font-medium text-gray-900">{{post.Title}}</h3>
							<p class="text-gray-600 mt-2">{{post.Content}}</p>
							<div class="mt-2 text-sm text-gray-500">
								By {{post.Author.Name}} on {{post.CreatedAt}}
							</div>
						</div>
					@endforeach
				</div>
			</div>
		@endsection
	`

	fmt.Println("Complex template content:")
	fmt.Println(complexTemplate)
	fmt.Println("\nWould render to:")
	fmt.Println("Dashboard")
	fmt.Println("User Information")
	fmt.Printf("Name: %s\n", data["User"].(map[string]interface{})["Name"])
	fmt.Printf("Email: %s\n", data["User"].(map[string]interface{})["Email"])
	fmt.Printf("Role: %s\n", data["User"].(map[string]interface{})["Role"])
	fmt.Println("Recent Posts")
	posts := data["Posts"].([]map[string]interface{})
	for _, post := range posts {
		fmt.Printf("- %s\n", post["Title"])
		fmt.Printf("  %s\n", post["Content"])
		fmt.Printf("  By %s on %s\n", post["Author"].(map[string]interface{})["Name"], post["CreatedAt"])
	}

	// Example 3: CLI Commands
	fmt.Println("\n3. Available CLI Commands:")
	fmt.Println("dolphin fin make template welcome --layout=app --model=User")
	fmt.Println("dolphin fin make component alert")
	fmt.Println("dolphin fin make layout admin")
	fmt.Println("dolphin fin make partial header")
	fmt.Println("dolphin fin list")
	fmt.Println("dolphin fin validate pages/welcome")
	fmt.Println("dolphin fin cache")

	// Example 4: Integration with HTTP handlers
	fmt.Println("\n4. HTTP Handler Integration:")
	fmt.Println(`
func (r *Router) handleHome(w http.ResponseWriter, req *http.Request) {
    data := map[string]interface{}{
        "User": getCurrentUser(req),
        "Posts": getRecentPosts(),
    }
    
    if err := r.renderFin(w, "pages/home", data); err != nil {
        http.Error(w, "Template error", http.StatusInternalServerError)
    }
}
	`)

	fmt.Println("\n🎉 Fin Template Integration Complete!")
	fmt.Println("📚 Check the documentation for more examples and features.")
	fmt.Println("🚀 Use 'dolphin fin make' to generate new templates!")
}

func main() {
	runFinIntegrationExample()
}

func writeTemplateFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

func removeTemplateFile(filename string) error {
	return os.Remove(filename)
}
