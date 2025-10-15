package static

import (
	"os"
	"path/filepath"
)

// DefaultTemplates contains default template content
var DefaultTemplates = map[string]string{
	"layout": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <meta name="description" content="{{.Description}}">
    <meta name="keywords" content="{{.Keywords}}">
    <meta name="author" content="{{.Author}}">
    
    {{range $key, $value := .Meta}}
    <meta name="{{$key}}" content="{{$value}}">
    {{end}}
    
    <link rel="stylesheet" href="{{.Assets.css}}">
    <link rel="icon" href="{{.Assets.favicon}}">
</head>
<body>
    <header class="header">
        <nav class="nav">
            <div class="nav-brand">
                <a href="/">🐬 Dolphin</a>
            </div>
            <div class="nav-menu">
                <a href="/" class="nav-link">Home</a>
                <a href="/docs" class="nav-link">Documentation</a>
                <a href="/examples" class="nav-link">Examples</a>
                <a href="/api" class="nav-link">API</a>
            </div>
        </nav>
    </header>
    
    <main class="main">
        {{template "content" .}}
    </main>
    
    <footer class="footer">
        <div class="footer-content">
            <p>&copy; 2024 Dolphin Framework. Built with ❤️ in Go.</p>
        </div>
    </footer>
    
    <script src="{{.Assets.js}}"></script>
</body>
</html>`,

	"index": `{{define "content"}}
<section class="hero">
    <div class="hero-content">
        <h1 class="hero-title">Welcome to Dolphin Framework</h1>
        <p class="hero-subtitle">Enterprise-grade Go web framework for rapid development</p>
        <div class="hero-actions">
            <a href="/docs" class="btn btn-primary">Get Started</a>
            <a href="/examples" class="btn btn-secondary">View Examples</a>
        </div>
    </div>
</section>

<section class="features">
    <div class="features-grid">
        <div class="feature-card">
            <div class="feature-icon">🚀</div>
            <h3>Rapid Development</h3>
            <p>Built-in scaffolding and code generation tools</p>
        </div>
        <div class="feature-card">
            <div class="feature-icon">⚡</div>
            <h3>High Performance</h3>
            <p>Built on Go's concurrency and performance</p>
        </div>
        <div class="feature-card">
            <div class="feature-icon">🔧</div>
            <h3>CLI Tools</h3>
            <p>Powerful command-line interface</p>
        </div>
        <div class="feature-card">
            <div class="feature-icon">📚</div>
            <h3>Documentation</h3>
            <p>Automatic API documentation</p>
        </div>
    </div>
</section>
{{end}}`,

	"404": `{{define "content"}}
<section class="error-page">
    <div class="error-content">
        <div class="error-icon">🐬</div>
        <h1 class="error-title">Page Not Found</h1>
        <p class="error-message">The page you're looking for doesn't exist.</p>
        <div class="error-details">
            <p><strong>Path:</strong> {{.Data.path}}</p>
            <p><strong>Error:</strong> {{.Data.error}}</p>
        </div>
        <a href="/" class="btn btn-primary">Go Home</a>
    </div>
</section>
{{end}}`,

	"docs": `{{define "content"}}
<section class="docs">
    <div class="docs-content">
        <h1>Documentation</h1>
        <p>Learn how to build amazing applications with Dolphin Framework.</p>
        
        <div class="docs-sections">
            <div class="docs-section">
                <h2>Quick Start</h2>
                <p>Get up and running in minutes</p>
                <a href="/docs/quick-start" class="btn btn-outline">Read More</a>
            </div>
            
            <div class="docs-section">
                <h2>API Reference</h2>
                <p>Complete API documentation</p>
                <a href="/api" class="btn btn-outline">View API</a>
            </div>
            
            <div class="docs-section">
                <h2>Examples</h2>
                <p>Real-world examples and tutorials</p>
                <a href="/examples" class="btn btn-outline">View Examples</a>
            </div>
        </div>
    </div>
</section>
{{end}}`,

	"examples": `{{define "content"}}
<section class="examples">
    <div class="examples-content">
        <h1>Examples</h1>
        <p>See Dolphin Framework in action with these real-world examples.</p>
        
        <div class="examples-grid">
            <div class="example-card">
                <h3>Blog Application</h3>
                <p>Complete blog with posts, comments, and user management</p>
                <a href="/examples/blog" class="btn btn-outline">View Example</a>
            </div>
            
            <div class="example-card">
                <h3>E-commerce API</h3>
                <p>RESTful API for e-commerce with authentication</p>
                <a href="/examples/ecommerce" class="btn btn-outline">View Example</a>
            </div>
            
            <div class="example-card">
                <h3>Real-time Chat</h3>
                <p>WebSocket-based chat application</p>
                <a href="/examples/chat" class="btn btn-outline">View Example</a>
            </div>
        </div>
    </div>
</section>
{{end}}`,
}

// DefaultCSS contains default CSS styles
var DefaultCSS = `
/* Reset and base styles */
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    line-height: 1.6;
    color: #333;
    background-color: #f8f9fa;
}

/* Header */
.header {
    background: white;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    position: sticky;
    top: 0;
    z-index: 100;
}

.nav {
    max-width: 1200px;
    margin: 0 auto;
    padding: 1rem 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.nav-brand a {
    font-size: 1.5rem;
    font-weight: bold;
    text-decoration: none;
    color: #667eea;
}

.nav-menu {
    display: flex;
    gap: 2rem;
}

.nav-link {
    text-decoration: none;
    color: #666;
    font-weight: 500;
    transition: color 0.3s;
}

.nav-link:hover {
    color: #667eea;
}

/* Main content */
.main {
    min-height: calc(100vh - 200px);
}

/* Hero section */
.hero {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 4rem 2rem;
    text-align: center;
}

.hero-content {
    max-width: 800px;
    margin: 0 auto;
}

.hero-title {
    font-size: 3rem;
    margin-bottom: 1rem;
    font-weight: 700;
}

.hero-subtitle {
    font-size: 1.25rem;
    margin-bottom: 2rem;
    opacity: 0.9;
}

.hero-actions {
    display: flex;
    gap: 1rem;
    justify-content: center;
    flex-wrap: wrap;
}

/* Buttons */
.btn {
    display: inline-block;
    padding: 0.75rem 1.5rem;
    border-radius: 6px;
    text-decoration: none;
    font-weight: 500;
    transition: all 0.3s;
    border: 2px solid transparent;
}

.btn-primary {
    background: white;
    color: #667eea;
}

.btn-primary:hover {
    background: #f8f9fa;
    transform: translateY(-2px);
}

.btn-secondary {
    background: transparent;
    color: white;
    border-color: white;
}

.btn-secondary:hover {
    background: white;
    color: #667eea;
}

.btn-outline {
    background: transparent;
    color: #667eea;
    border-color: #667eea;
}

.btn-outline:hover {
    background: #667eea;
    color: white;
}

/* Features */
.features {
    padding: 4rem 2rem;
    background: white;
}

.features-grid {
    max-width: 1200px;
    margin: 0 auto;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 2rem;
}

.feature-card {
    text-align: center;
    padding: 2rem;
    border-radius: 12px;
    background: #f8f9fa;
    transition: transform 0.3s;
}

.feature-card:hover {
    transform: translateY(-4px);
}

.feature-icon {
    font-size: 3rem;
    margin-bottom: 1rem;
}

.feature-card h3 {
    margin-bottom: 1rem;
    color: #333;
}

.feature-card p {
    color: #666;
}

/* Error page */
.error-page {
    padding: 4rem 2rem;
    text-align: center;
}

.error-content {
    max-width: 500px;
    margin: 0 auto;
}

.error-icon {
    font-size: 4rem;
    margin-bottom: 1rem;
}

.error-title {
    font-size: 2rem;
    margin-bottom: 1rem;
    color: #333;
}

.error-message {
    color: #666;
    margin-bottom: 2rem;
}

.error-details {
    background: #f8f9fa;
    padding: 1rem;
    border-radius: 8px;
    margin-bottom: 2rem;
    text-align: left;
}

/* Docs and Examples */
.docs, .examples {
    padding: 4rem 2rem;
    background: white;
}

.docs-content, .examples-content {
    max-width: 1200px;
    margin: 0 auto;
}

.docs-content h1, .examples-content h1 {
    font-size: 2.5rem;
    margin-bottom: 1rem;
    text-align: center;
}

.docs-content p, .examples-content p {
    font-size: 1.125rem;
    color: #666;
    text-align: center;
    margin-bottom: 3rem;
}

.docs-sections, .examples-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 2rem;
}

.docs-section, .example-card {
    padding: 2rem;
    border: 1px solid #e9ecef;
    border-radius: 12px;
    text-align: center;
    transition: all 0.3s;
}

.docs-section:hover, .example-card:hover {
    border-color: #667eea;
    transform: translateY(-2px);
}

.docs-section h2, .example-card h3 {
    margin-bottom: 1rem;
    color: #333;
}

.docs-section p, .example-card p {
    color: #666;
    margin-bottom: 1.5rem;
}

/* Footer */
.footer {
    background: #333;
    color: white;
    padding: 2rem;
    text-align: center;
}

.footer-content {
    max-width: 1200px;
    margin: 0 auto;
}

/* Responsive */
@media (max-width: 768px) {
    .nav {
        flex-direction: column;
        gap: 1rem;
    }
    
    .nav-menu {
        gap: 1rem;
    }
    
    .hero-title {
        font-size: 2rem;
    }
    
    .hero-actions {
        flex-direction: column;
        align-items: center;
    }
    
    .features-grid {
        grid-template-columns: 1fr;
    }
    
    .docs-sections, .examples-grid {
        grid-template-columns: 1fr;
    }
}
`

// CreateDefaultTemplates creates default templates and static files
func CreateDefaultTemplates(baseDir string) error {
	// Create directories
	dirs := []string{
		baseDir,
		filepath.Join(baseDir, "templates"),
		filepath.Join(baseDir, "css"),
		filepath.Join(baseDir, "js"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create CSS file
	cssPath := filepath.Join(baseDir, "css", "app.css")
	if err := os.WriteFile(cssPath, []byte(DefaultCSS), 0644); err != nil {
		return err
	}

	// Create JavaScript file
	jsPath := filepath.Join(baseDir, "js", "app.js")
	jsContent := `
// Dolphin Framework JavaScript
document.addEventListener('DOMContentLoaded', function() {
    // Add smooth scrolling to anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth'
                });
            }
        });
    });
    
    // Add loading states to buttons
    document.querySelectorAll('.btn').forEach(btn => {
        btn.addEventListener('click', function() {
            if (this.href && !this.href.startsWith('#')) {
                this.style.opacity = '0.7';
                this.style.pointerEvents = 'none';
            }
        });
    });
});
`
	if err := os.WriteFile(jsPath, []byte(jsContent), 0644); err != nil {
		return err
	}

	// Create templates
	for name, content := range DefaultTemplates {
		templatePath := filepath.Join(baseDir, "templates", name+".html")
		if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
			return err
		}
	}

	// Create favicon placeholder
	faviconPath := filepath.Join(baseDir, "favicon.ico")
	faviconContent := []byte{0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x10, 0x10, 0x00, 0x00, 0x01, 0x00, 0x20, 0x00, 0x68, 0x04, 0x00, 0x00, 0x16, 0x00, 0x00, 0x00}
	if err := os.WriteFile(faviconPath, faviconContent, 0644); err != nil {
		return err
	}

	return nil
}
