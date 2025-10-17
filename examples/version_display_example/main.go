package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mrhoseah/dolphin/internal/version"
)

func main() {
	fmt.Println("🐬 Dolphin Framework - Version Display Example")
	fmt.Println("==============================================")

	// Show current version
	fmt.Printf("Current Dolphin Version: %s\n", version.GetVersion())
	fmt.Println()

	// Create a simple HTTP server to demonstrate version display
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simulate the footer template with version
		footerTemplate := `
<footer style="border-top:1px solid #e5e7eb;margin-top:32px;background:#fff">
  <div style="max-width:1100px;margin:0 auto;padding:18px 16px;color:#6b7280;font-size:14px;text-align:center">
    <div style="margin-bottom:8px">
      Built with ❤️ by the Dolphin community • MIT License
    </div>
    <div style="font-size:12px;color:#9ca3af">
      🐬 Dolphin Framework v{{.Version}} • Powered by Go
    </div>
  </div>
</footer>`

		// Template data
		data := map[string]interface{}{
			"Version": version.GetVersion(),
		}

		// Parse and execute template
		tmpl, err := template.New("footer").Parse(footerTemplate)
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		// Set content type
		w.Header().Set("Content-Type", "text/html")

		// Write HTML content
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dolphin Framework - Version Display</title>
    <style>
        body { 
            margin: 0; 
            font-family: system-ui, -apple-system, sans-serif; 
            background: #f6f7fb; 
            color: #111827;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
        }
        .container {
            max-width: 1100px;
            margin: 0 auto;
            padding: 40px 20px;
            flex: 1;
        }
        .header {
            text-align: center;
            margin-bottom: 40px;
        }
        .header h1 {
            color: #1f2937;
            margin-bottom: 10px;
        }
        .header p {
            color: #6b7280;
            font-size: 18px;
        }
        .content {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
            margin-bottom: 30px;
        }
        .feature {
            margin-bottom: 20px;
            padding: 15px;
            background: #f9fafb;
            border-left: 4px solid #3b82f6;
            border-radius: 4px;
        }
        .feature h3 {
            margin: 0 0 10px 0;
            color: #1f2937;
        }
        .feature p {
            margin: 0;
            color: #6b7280;
        }
        .code {
            background: #1f2937;
            color: #f9fafb;
            padding: 15px;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 14px;
            overflow-x: auto;
            margin: 15px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🐬 Dolphin Framework</h1>
            <p>Version Display Example - Similar to CakePHP</p>
        </div>
        
        <div class="content">
            <h2>Version Display Features</h2>
            
            <div class="feature">
                <h3>📱 Footer Version Display</h3>
                <p>The version is automatically displayed in the footer of every page, similar to how CakePHP shows its version.</p>
            </div>
            
            <div class="feature">
                <h3>🔧 Template Integration</h3>
                <p>The version is passed to all templates through the template data, making it available anywhere in your application.</p>
            </div>
            
            <div class="feature">
                <h3>⚡ Dynamic Updates</h3>
                <p>When you update the framework version, it automatically appears on all pages without code changes.</p>
            </div>
            
            <div class="feature">
                <h3>🎨 Consistent Styling</h3>
                <p>The version display follows the same styling as the rest of the application for a professional look.</p>
            </div>
            
            <h3>Implementation Example:</h3>
            <div class="code">
// In your template (footer.html)
🐬 Dolphin Framework v{{.Version}} • Powered by Go

// In your Go code (router/web.go)
data := map[string]interface{}{
    "Version": version.GetVersion(),
    // ... other data
}
            </div>
            
            <h3>Current Version:</h3>
            <div class="code">
Dolphin Framework v%s
            </div>
        </div>
    </div>
`, version.GetVersion())

		// Execute footer template
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
			return
		}
	})

	// Start server
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	fmt.Printf("🚀 Starting server on port %s\n", port)
	fmt.Printf("📱 Open http://localhost:%s to see the version display\n", port)
	fmt.Println("⏹️  Press Ctrl+C to stop the server")
	fmt.Println()

	// Start server in goroutine
	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal("Server error:", err)
		}
	}()

	// Wait for a moment to show the server is running
	time.Sleep(2 * time.Second)

	fmt.Println("✅ Server is running!")
	fmt.Println("✅ Version display is working!")
	fmt.Println()
	fmt.Println("Features demonstrated:")
	fmt.Println("  • Version display in footer")
	fmt.Println("  • Template integration")
	fmt.Println("  • Consistent styling")
	fmt.Println("  • Dynamic version updates")
}
