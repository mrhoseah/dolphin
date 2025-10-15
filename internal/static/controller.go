package static

import (
	"net/http"
	"strings"
)

// Controller handles static page requests
type Controller struct {
	service *Service
}

// NewController creates a new static controller
func NewController(service *Service) *Controller {
	return &Controller{
		service: service,
	}
}

// Serve handles static page serving
func (c *Controller) Serve(w http.ResponseWriter, r *http.Request) {
	// Extract page name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index"
	}

	// Remove trailing slash
	path = strings.TrimSuffix(path, "/")

	// Prepare page data
	data := PageData{
		Title:       c.getTitleFromPath(path),
		Description: "Dolphin Framework - Enterprise-grade Go web framework",
		Keywords:    "golang, web framework, rapid development, enterprise",
		Author:      "Dolphin Framework",
		Data:        make(map[string]interface{}),
		Meta:        make(map[string]string),
		Assets:      make(map[string]string),
	}

	// Add common data
	data.Data["current_path"] = path
	data.Data["request_url"] = r.URL.String()
	data.Data["user_agent"] = r.UserAgent()
	data.Data["timestamp"] = r.Header.Get("X-Request-Time")

	// Add meta tags
	data.Meta["viewport"] = "width=device-width, initial-scale=1.0"
	data.Meta["robots"] = "index, follow"

	// Add asset URLs
	data.Assets["css"] = "/static/css/app.css"
	data.Assets["js"] = "/static/js/app.js"
	data.Assets["favicon"] = "/static/favicon.ico"

	// Serve the page
	if err := c.service.ServePage(w, r, path, data); err != nil {
		c.handleError(w, r, err)
	}
}

// ServeTemplate serves a specific template
func (c *Controller) ServeTemplate(w http.ResponseWriter, r *http.Request, templateName string, data PageData) {
	if err := c.service.ServePage(w, r, templateName, data); err != nil {
		c.handleError(w, r, err)
	}
}

// handleError handles errors when serving pages
func (c *Controller) handleError(w http.ResponseWriter, r *http.Request, err error) {
	// Try to serve a custom error page
	errorData := PageData{
		Title:       "Page Not Found",
		Description: "The requested page could not be found",
		Data: map[string]interface{}{
			"error":      err.Error(),
			"path":       r.URL.Path,
			"status":     404,
			"timestamp":  r.Header.Get("X-Request-Time"),
		},
	}

	if serveErr := c.service.ServePage(w, r, "404", errorData); serveErr != nil {
		// Fallback to default error page
		c.serveDefaultError(w, r, err)
	}
}

// serveDefaultError serves a default error page
func (c *Controller) serveDefaultError(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)

	html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Page Not Found - Dolphin Framework</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            margin: 0;
            padding: 0;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            padding: 3rem;
            max-width: 500px;
            text-align: center;
            margin: 2rem;
        }
        .icon {
            font-size: 4rem;
            margin-bottom: 1rem;
        }
        h1 {
            color: #333;
            margin-bottom: 1rem;
            font-size: 2rem;
        }
        p {
            color: #666;
            line-height: 1.6;
            margin-bottom: 2rem;
        }
        .error-details {
            background: #f8f9fa;
            border-radius: 8px;
            padding: 1rem;
            margin-top: 2rem;
            font-size: 0.9rem;
            color: #555;
            text-align: left;
        }
        .back-button {
            display: inline-block;
            padding: 0.75rem 1.5rem;
            background: #667eea;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            margin-top: 1rem;
        }
        .back-button:hover {
            background: #5a6fd8;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">🐬</div>
        <h1>Page Not Found</h1>
        <p>The page you're looking for doesn't exist or has been moved.</p>
        
        <div class="error-details">
            <strong>Error:</strong> ` + err.Error() + `<br>
            <strong>Path:</strong> ` + r.URL.Path + `<br>
            <strong>Method:</strong> ` + r.Method + `
        </div>
        
        <a href="/" class="back-button">Go Home</a>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

// getTitleFromPath generates a title from the path
func (c *Controller) getTitleFromPath(path string) string {
	if path == "index" {
		return "Dolphin Framework - Home"
	}

	// Convert path to title case
	parts := strings.Split(path, "/")
	var titleParts []string

	for _, part := range parts {
		if part != "" {
			titleParts = append(titleParts, strings.Title(part))
		}
	}

	title := strings.Join(titleParts, " - ")
	return title + " - Dolphin Framework"
}

// ServeAPI serves static page API endpoints
func (c *Controller) ServeAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		c.handleListPages(w, r)
	case "POST":
		c.handleCreatePage(w, r)
	case "DELETE":
		c.handleDeletePage(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListPages handles listing pages
func (c *Controller) handleListPages(w http.ResponseWriter, r *http.Request) {
	pages, err := c.service.ListPages()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Simple JSON response
	response := `{"pages": [`
	for i, page := range pages {
		if i > 0 {
			response += ","
		}
		response += `"` + page + `"`
	}
	response += `]}`

	w.Write([]byte(response))
}

// handleCreatePage handles creating pages
func (c *Controller) handleCreatePage(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	content := r.FormValue("content")

	if name == "" || content == "" {
		http.Error(w, "Name and content are required", http.StatusBadRequest)
		return
	}

	if err := c.service.CreatePage(name, content); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Page created successfully", "name": "` + name + `"}`))
}

// handleDeletePage handles deleting pages
func (c *Controller) handleDeletePage(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name parameter is required", http.StatusBadRequest)
		return
	}

	if err := c.service.DeletePage(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Page deleted successfully", "name": "` + name + `"}`))
}
