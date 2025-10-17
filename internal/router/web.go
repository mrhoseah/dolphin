package router

import (
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mrhoseah/dolphin/internal/auth"
	dolphinMiddleware "github.com/mrhoseah/dolphin/internal/middleware"
	"github.com/mrhoseah/dolphin/internal/time"
	"github.com/mrhoseah/dolphin/internal/version"
)

// render joins base layout with header/footer partials and the page body.
func render(w http.ResponseWriter, pagePath string) error {
	header, _ := os.ReadFile("ui/views/partials/header.html")
	footer, _ := os.ReadFile("ui/views/partials/footer.html")
	bodyBytes, err := os.ReadFile(pagePath)
	if err != nil {
		return err
	}
	layout := "base"
	body := string(bodyBytes)
	// Layout tag formats supported (first occurrence wins):
	//   {{layout:admin}}  or  <!-- layout: admin -->
	if idx := strings.Index(body, "{{layout:"); idx != -1 {
		end := strings.Index(body[idx:], "}}")
		if end != -1 {
			spec := body[idx+9 : idx+end] // after '{{layout:' to before '}}'
			layout = strings.TrimSpace(spec)
			// remove the tag from body
			body = body[:idx] + body[idx+end+2:]
		}
	} else if idx := strings.Index(body, "<!-- layout:"); idx != -1 {
		end := strings.Index(body[idx:], "-->")
		if end != -1 {
			spec := body[idx+12 : idx+end]
			layout = strings.TrimSpace(spec)
			body = body[:idx] + body[idx+end+3:]
		}
	}

	layoutPath := "ui/views/layouts/" + layout + ".html"
	base, err := os.ReadFile(layoutPath)
	if err != nil {
		// fallback to base layout
		if layout != "base" {
			if fallback, fe := os.ReadFile("ui/views/layouts/base.html"); fe == nil {
				base = fallback
			} else {
				return err
			}
		} else {
			return err
		}
	}

	// Create template data with version information
	data := map[string]interface{}{
		"Version": version.GetVersion(),
		"Header":  string(header),
		"Body":    body,
		"Footer":  string(footer),
	}

	// Parse and execute template with time helpers
	tmpl, err := template.New("layout").Funcs(time.TemplateHelpers()).Parse(string(base))
	if err != nil {
		return err
	}
	
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	return tmpl.Execute(w, data)
}

// setupWebRoutes configures web routes with HTMX support
func (r *Router) setupWebRoutes(router chi.Router) {
	// Setup Dolphin-style authentication for web routes using router's manager
	webAuthMiddleware := dolphinMiddleware.NewAuthMiddleware(r.authManager, r.app.Logger())

	// Home page with HTMX
	router.Get("/", r.handleHome)

	// Authentication pages
	router.Route("/auth", func(auth chi.Router) {
		auth.Get("/login", r.handleLoginPage)
		auth.Post("/login", r.handleLoginSubmit)
		auth.Get("/register", r.handleRegisterPage)
		auth.Post("/register", r.handleRegisterSubmit)
		auth.Post("/logout", webAuthMiddleware.Authenticate(http.HandlerFunc(r.handleLogout)).ServeHTTP)
	})

	// Dashboard (protected)
	router.Route("/dashboard", func(dashboard chi.Router) {
		dashboard.Use(webAuthMiddleware.Authenticate)
		dashboard.Get("/", r.handleDashboard)
	})

	// Admin routes
	router.Route("/admin", func(admin chi.Router) {
		admin.Use(webAuthMiddleware.Authenticate)
		admin.Use(webAuthMiddleware.RoleMiddleware("admin"))

		admin.Get("/", r.handleAdminDashboard)
		admin.Get("/users", r.handleAdminUsers)
		admin.Get("/posts", r.handleAdminPosts)
	})

	// HTMX partial routes
	router.Route("/partials", func(partials chi.Router) {
		partials.Use(webAuthMiddleware.Authenticate)
		partials.Get("/user-menu", r.handleUserMenu)
		partials.Get("/notifications", r.handleNotifications)
		partials.Get("/sidebar", r.handleSidebar)
	})
}

// handleHome renders the home page with HTMX integration
func (r *Router) handleHome(w http.ResponseWriter, req *http.Request) {
	if err := render(w, "ui/views/pages/home.html"); err != nil {
		http.Error(w, "Home view not found", http.StatusInternalServerError)
	}
}

// handleLoginPage renders the login page
func (r *Router) handleLoginPage(w http.ResponseWriter, req *http.Request) {
	if err := render(w, "ui/views/auth/login.html"); err != nil {
		http.Error(w, "Login view not found", http.StatusInternalServerError)
	}
}

// handleLoginSubmit handles login form submission
func (r *Router) handleLoginSubmit(w http.ResponseWriter, req *http.Request) {
	_ = req.ParseForm()
	email := req.FormValue("email")
	password := req.FormValue("password")

	w.Header().Set("Content-Type", "text/html")

	if email == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">Email and password are required.</div>`))
		return
	}

	if err := r.authManager.LoginWithCredentials(map[string]string{"email": email, "password": password}); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">Invalid credentials.</div>`))
		return
	}

	// HTMX-friendly redirect
	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">Login successful.</div>`))
}

// handleRegisterPage renders the register page
func (r *Router) handleRegisterPage(w http.ResponseWriter, req *http.Request) {
	if err := render(w, "ui/views/auth/register.html"); err != nil {
		http.Error(w, "Register view not found", http.StatusInternalServerError)
	}
}

// handleRegisterSubmit handles registration form submission
func (r *Router) handleRegisterSubmit(w http.ResponseWriter, req *http.Request) {
	_ = req.ParseForm()
	first := req.FormValue("firstName")
	last := req.FormValue("lastName")
	email := req.FormValue("email")
	password := req.FormValue("password")

	w.Header().Set("Content-Type", "text/html")

	if first == "" || last == "" || email == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">All fields are required.</div>`))
		return
	}

	// Minimal user create (plaintext password placeholder)
	db := r.app.DB().GetDB()
	u := auth.User{Email: email, Password: password, FirstName: first, LastName: last}
	if err := db.Create(&u).Error; err != nil {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">` + err.Error() + `</div>`))
		return
	}

	// HTMX-friendly redirect
	w.Header().Set("HX-Redirect", "/auth/login")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">Registration successful.</div>`))
}

// handleLogout handles logout
func (r *Router) handleLogout(w http.ResponseWriter, req *http.Request) {
	r.authManager.Logout()
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<div class="bg-blue-100 border border-blue-400 text-blue-700 px-4 py-3 rounded">
    Logged out successfully! Redirecting...
</div>
<script>
    setTimeout(() => {
        window.location.href = '/';
    }, 1000);
</script>
	`))
}

// handleDashboard renders the dashboard with HTMX
func (r *Router) handleDashboard(w http.ResponseWriter, req *http.Request) {
	if err := render(w, "ui/views/pages/dashboard.html"); err != nil {
		http.Error(w, "Dashboard view not found", http.StatusInternalServerError)
	}
}

// handleAdminDashboard renders admin dashboard
func (r *Router) handleAdminDashboard(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Admin Dashboard - Dolphin Framework</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen">
        <nav class="bg-white shadow">
            <div class="max-w-7xl mx-auto px-4">
                <div class="flex justify-between h-16">
                    <div class="flex items-center">
                        <h1 class="text-xl font-semibold">🐬 Admin Dashboard</h1>
                    </div>
                </div>
            </div>
        </nav>
        <div class="max-w-7xl mx-auto py-6 px-4">
            <h2 class="text-2xl font-bold text-gray-900 mb-6">Admin Panel</h2>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <a href="/admin/users" class="bg-white rounded-lg shadow p-6 hover:shadow-lg transition">
                    <h3 class="text-lg font-medium text-gray-900">User Management</h3>
                    <p class="text-gray-600">Manage user accounts and permissions</p>
                </a>
                <a href="/admin/posts" class="bg-white rounded-lg shadow p-6 hover:shadow-lg transition">
                    <h3 class="text-lg font-medium text-gray-900">Content Management</h3>
                    <p class="text-gray-600">Manage posts and content</p>
                </a>
            </div>
        </div>
    </div>
</body>
</html>
	`))
}

// handleAdminUsers renders admin users page
func (r *Router) handleAdminUsers(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>User Management - Dolphin Framework</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen">
        <nav class="bg-white shadow">
            <div class="max-w-7xl mx-auto px-4">
                <div class="flex justify-between h-16">
                    <div class="flex items-center">
                        <h1 class="text-xl font-semibold">🐬 User Management</h1>
                    </div>
                </div>
            </div>
        </nav>
        <div class="max-w-7xl mx-auto py-6 px-4">
            <h2 class="text-2xl font-bold text-gray-900 mb-6">User Management</h2>
            <div class="bg-white rounded-lg shadow">
                <div class="p-6">
                    <p class="text-gray-600">User management interface will be implemented here.</p>
                </div>
            </div>
        </div>
    </div>
</body>
</html>
	`))
}

// handleAdminPosts renders admin posts page
func (r *Router) handleAdminPosts(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Content Management - Dolphin Framework</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen">
        <nav class="bg-white shadow">
            <div class="max-w-7xl mx-auto px-4">
                <div class="flex justify-between h-16">
                    <div class="flex items-center">
                        <h1 class="text-xl font-semibold">🐬 Content Management</h1>
                    </div>
                </div>
            </div>
        </nav>
        <div class="max-w-7xl mx-auto py-6 px-4">
            <h2 class="text-2xl font-bold text-gray-900 mb-6">Content Management</h2>
            <div class="bg-white rounded-lg shadow">
                <div class="p-6">
                    <p class="text-gray-600">Content management interface will be implemented here.</p>
                </div>
            </div>
        </div>
    </div>
</body>
</html>
	`))
}

// HTMX partial handlers
func (r *Router) handleUserMenu(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<div class="flex items-center space-x-4">
    <span class="text-gray-700">Welcome, User!</span>
    <form hx-post="/auth/logout" class="inline">
        <button type="submit" class="text-gray-500 hover:text-gray-700">Logout</button>
    </form>
</div>
	`))
}

func (r *Router) handleNotifications(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<div class="bg-blue-100 border border-blue-400 text-blue-700 px-4 py-3 rounded">
    No new notifications
</div>
	`))
}

func (r *Router) handleSidebar(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<nav class="bg-gray-800 text-white w-64 min-h-screen p-4">
    <ul class="space-y-2">
        <li><a href="/dashboard" class="block py-2 px-4 hover:bg-gray-700 rounded">Dashboard</a></li>
        <li><a href="/admin" class="block py-2 px-4 hover:bg-gray-700 rounded">Admin</a></li>
        <li><a href="/admin/users" class="block py-2 px-4 hover:bg-gray-700 rounded">Users</a></li>
        <li><a href="/admin/posts" class="block py-2 px-4 hover:bg-gray-700 rounded">Posts</a></li>
    </ul>
</nav>
	`))
}
