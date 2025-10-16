package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	dolphinMiddleware "github.com/mrhoseah/dolphin/internal/middleware"
)

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
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dolphin Framework</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen flex items-center justify-center p-6">
        <div class="w-full max-w-6xl bg-white rounded-2xl shadow p-8 grid grid-cols-1 md:grid-cols-2 gap-8">
            <div class="flex flex-col justify-center">
                <div class="flex items-center space-x-3 mb-6">
                    <span class="text-3xl">🐬</span>
                    <span class="text-emerald-600 font-bold tracking-wide">DOLPHIN</span>
                </div>
                <h1 class="text-4xl md:text-5xl font-extrabold text-gray-900 leading-tight">Dolphin<br/>Framework</h1>
                <p class="text-gray-600 mt-4">Enterprise-grade Go web framework for rapid development</p>
                <div class="flex items-center gap-4 mt-8">
                    <a href="/auth/register" class="inline-flex items-center bg-emerald-600 text-white px-5 py-2.5 rounded-full shadow hover:bg-emerald-700 transition">Get Started</a>
                    <a href="/auth/login" class="inline-flex items-center border-2 border-emerald-300 text-emerald-700 px-5 py-2.5 rounded-full hover:bg-emerald-50 transition">Learn More</a>
                </div>
                <div class="grid grid-cols-3 gap-6 mt-10 text-sm text-gray-600">
                    <div class="flex items-center space-x-2"><span>🚀</span><span>Rapid Development</span></div>
                    <div class="flex items-center space-x-2"><span>🗄️</span><span>Integrated ORM</span></div>
                    <div class="flex items-center space-x-2"><span>🎨</span><span>Frontend Ready</span></div>
                </div>
            </div>
            <div class="rounded-xl overflow-hidden bg-gray-50 border border-gray-100 p-2">
                <img src="/static/hero.png" alt="Dolphin Framework" class="w-full rounded-lg shadow-sm"/>
            </div>
        </div>
    </div>
</body>
</html>
	`))
}

// handleLoginPage renders the login page
func (r *Router) handleLoginPage(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - Dolphin Framework</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen flex items-center justify-center">
        <div class="max-w-md w-full bg-white rounded-lg shadow-md p-6">
            <div class="flex items-center justify-center mb-4"><span class="text-3xl mr-2">🐬</span><span class="text-emerald-600 font-bold">DOLPHIN</span></div>
            <h2 class="text-2xl font-bold text-gray-900 mb-6 text-center">Login</h2>
            <form hx-post="/auth/login" hx-target="#login-result" class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-700">Email</label>
                    <input type="email" name="email" required class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500">
                </div>
                <div>
                    <label class="block text-sm font-medium text-gray-700">Password</label>
                    <input type="password" name="password" required class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500">
                </div>
                <button type="submit" class="w-full bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-600 transition">Login</button>
            </form>
            <div id="login-result" class="mt-4"></div>
            <div class="mt-4 text-center">
                <a href="/auth/register" class="text-blue-500 hover:text-blue-700">Don't have an account? Register</a>
            </div>
        </div>
    </div>
</body>
</html>
	`))
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

    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`
<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">
    Login successful! Redirecting...
</div>
<script>setTimeout(()=>{ window.location.href='/dashboard'; }, 800);</script>
    `))
}

// handleRegisterPage renders the register page
func (r *Router) handleRegisterPage(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Register - Dolphin Framework</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen flex items-center justify-center">
        <div class="max-w-md w-full bg-white rounded-lg shadow-md p-6">
            <div class="flex items-center justify-center mb-4"><span class="text-3xl mr-2">🐬</span><span class="text-emerald-600 font-bold">DOLPHIN</span></div>
            <h2 class="text-2xl font-bold text-gray-900 mb-6 text-center">Register</h2>
            <form hx-post="/auth/register" hx-target="#register-result" class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-700">First Name</label>
                    <input type="text" name="firstName" required class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500">
                </div>
                <div>
                    <label class="block text-sm font-medium text-gray-700">Last Name</label>
                    <input type="text" name="lastName" required class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500">
                </div>
                <div>
                    <label class="block text-sm font-medium text-gray-700">Email</label>
                    <input type="email" name="email" required class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500">
                </div>
                <div>
                    <label class="block text-sm font-medium text-gray-700">Password</label>
                    <input type="password" name="password" required class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500">
                </div>
                <button type="submit" class="w-full bg-green-500 text-white py-2 px-4 rounded hover:bg-green-600 transition">Register</button>
            </form>
            <div id="register-result" class="mt-4"></div>
            <div class="mt-4 text-center">
                <a href="/auth/login" class="text-blue-500 hover:text-blue-700">Already have an account? Login</a>
            </div>
        </div>
    </div>
</body>
</html>
	`))
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

    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`
<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">
    Registration successful! Redirecting to login...
</div>
<script>setTimeout(()=>{ window.location.href='/auth/login'; }, 800);</script>
    `))
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
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dashboard - Dolphin Framework</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen">
        <nav class="bg-white shadow">
            <div class="max-w-7xl mx-auto px-4">
                <div class="flex justify-between h-16">
                    <div class="flex items-center">
                        <h1 class="text-xl font-semibold">🐬 Dolphin Dashboard</h1>
                    </div>
                    <div class="flex items-center space-x-4">
                        <div hx-get="/partials/user-menu" hx-trigger="load" hx-target="#user-menu"></div>
                        <div id="user-menu"></div>
                    </div>
                </div>
            </div>
        </nav>
        <div class="max-w-7xl mx-auto py-6 px-4">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div class="bg-white rounded-lg shadow p-6">
                    <h3 class="text-lg font-medium text-gray-900">Users</h3>
                    <p class="text-gray-600">Manage user accounts</p>
                </div>
                <div class="bg-white rounded-lg shadow p-6">
                    <h3 class="text-lg font-medium text-gray-900">Posts</h3>
                    <p class="text-gray-600">Manage blog posts</p>
                </div>
                <div class="bg-white rounded-lg shadow p-6">
                    <h3 class="text-lg font-medium text-gray-900">Settings</h3>
                    <p class="text-gray-600">Application settings</p>
                </div>
            </div>
        </div>
    </div>
</body>
</html>
	`))
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
