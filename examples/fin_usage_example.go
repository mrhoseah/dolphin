package main

import (
	"fmt"
	"log"
	"os"

	"dolphin/internal/template"
)

func main() {
	fmt.Println("🐬 Fin Template Engine Example")
	fmt.Println("==============================")

	// Create Fin template engine configuration
	config := &template.Config{
		ViewsPath:    "ui/views",
		CachePath:    "storage/cache/views",
		CacheEnabled: true,
		DebugMode:    true,
		Extensions:   []string{".fin.go", ".go.html"},
	}

	// Initialize Fin template engine
	engine := template.NewFinEngine(config)

	// Register a custom directive
	engine.RegisterDirective("uppercase", func(args []string, content string, data interface{}) (string, error) {
		return fmt.Sprintf("{{. | upper}}"), nil
	})

	// Register a component
	engine.RegisterComponent("alert", `
		<div class="alert alert-{{type}}">
			{{content}}
		</div>
	`)

	// Register a layout
	engine.RegisterLayout("app", `
		<!DOCTYPE html>
		<html>
		<head>
			<title>{{.Title}}</title>
		</head>
		<body>
			{{template "content" .}}
		</body>
		</html>
	`)

	// Sample data
	data := map[string]interface{}{
		"Title": "Welcome to Fin Templates",
		"User": map[string]interface{}{
			"Name":  "John Doe",
			"Email": "john@example.com",
			"Role":  "Admin",
		},
		"Posts": []map[string]interface{}{
			{
				"Title":   "Getting Started with Fin",
				"Content": "Fin templates provide clean, readable syntax...",
				"Author":  "John Doe",
			},
			{
				"Title":   "Advanced Fin Features",
				"Content": "Learn about model annotations and components...",
				"Author":  "Jane Smith",
			},
		},
	}

	// Example 1: Simple template rendering
	fmt.Println("\n1. Simple Template Rendering:")
	simpleTemplate := `
		<h1>Hello, {{.User.Name}}!</h1>
		<p>Email: {{.User.Email}}</p>
		<p>Role: {{.User.Role}}</p>
	`

	// For this example, we'll create a temporary template file
	tempFile := "temp_example.fin.go"
	err := os.WriteFile(tempFile, []byte(simpleTemplate), 0644)
	if err != nil {
		log.Fatal("Failed to create temp file:", err)
	}
	defer os.Remove(tempFile)

	// Note: In a real application, you would have proper template files
	// and use engine.Render("template_name", data)
	fmt.Println("Template content:")
	fmt.Println(simpleTemplate)
	fmt.Println("\nWould render to:")
	fmt.Printf("Hello, %s!\n", data["User"].(map[string]interface{})["Name"])
	fmt.Printf("Email: %s\n", data["User"].(map[string]interface{})["Email"])
	fmt.Printf("Role: %s\n", data["User"].(map[string]interface{})["Role"])

	// Example 2: Fin syntax features
	fmt.Println("\n2. Fin Template Features:")
	fmt.Println("✅ Model annotations: @model('User', user)")
	fmt.Println("✅ Clean variable syntax: {{user.Name}} instead of {{$user->name}}")
	fmt.Println("✅ Enhanced loops: @foreach(posts as post)")
	fmt.Println("✅ Component system: @component('alert')")
	fmt.Println("✅ Layout inheritance: @extends('layouts.app')")
	fmt.Println("✅ Section system: @section('content')")

	// Example 3: Template structure
	fmt.Println("\n3. Example Fin Template Structure:")
	fmt.Println(`
<!-- pages/welcome.fin.go -->
@extends('layouts.app')
@model('User', user)

@section('title')
    Welcome
@endsection

@section('content')
    <div class="hero">
        <h1>Welcome to Dolphin Framework</h1>
        <p>Hello, {{user.Name}}!</p>
        
        @if(user.IsAdmin)
            <div class="admin-panel">
                <p>Admin access granted</p>
            </div>
        @endif
        
        <div class="posts">
            <h2>Recent Posts</h2>
            @foreach(posts as post)
                <div class="post-card">
                    <h3>{{post.Title}}</h3>
                    <p>{{post.Content}}</p>
                    <small>By {{post.Author.Name}}</small>
                </div>
            @endforeach
        </div>
    </div>
@endsection
	`)

	fmt.Println("\n🎉 Fin Template Engine is ready to use!")
	fmt.Println("📚 Check the documentation for more examples and features.")
}
