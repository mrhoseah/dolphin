package template

// DefaultLayout returns the default Fin layout template with HTMX and TailwindCSS
func DefaultLayout() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>@yield('title')</title>
    
    <!-- TailwindCSS -->
    <script src="https://cdn.tailwindcss.com"></script>
    
    <!-- HTMX -->
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    
    <!-- HTMX Extensions (optional) -->
    <script src="https://unpkg.com/htmx.org/dist/ext/json-enc.js"></script>
    <script src="https://unpkg.com/htmx.org/dist/ext/ws.js"></script>
    <script src="https://unpkg.com/htmx.org/dist/ext/sse.js"></script>
    
    @stack('styles')
</head>
<body class="bg-gray-50">
    <!-- Navigation -->
    <nav class="bg-white shadow-sm">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="flex justify-between h-16">
                <div class="flex">
                    <div class="flex-shrink-0 flex items-center">
                        <a href="/" class="text-xl font-bold text-gray-900">@yield('app_name', 'Dolphin App')</a>
                    </div>
                    <div class="hidden sm:ml-6 sm:flex sm:space-x-8">
                        <a href="/" class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                            Home
                        </a>
                        @auth
                            <a href="/dashboard" class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                                Dashboard
                            </a>
                        @endif
                    </div>
                </div>
                <div class="hidden sm:ml-6 sm:flex sm:items-center">
                    @auth
                        <span class="text-gray-700 mr-4">{{.User.Name}}</span>
                        <a href="/logout" class="text-gray-500 hover:text-gray-700">Logout</a>
                    @else
                        <a href="/login" class="text-gray-500 hover:text-gray-700 mr-4">Login</a>
                        <a href="/register" class="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700">Register</a>
                    @endif
                </div>
            </div>
        </div>
    </nav>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        @yield('content')
    </main>

    <!-- Footer -->
    <footer class="bg-white border-t mt-auto">
        <div class="max-w-7xl mx-auto py-4 px-4 sm:px-6 lg:px-8">
            <p class="text-center text-gray-500 text-sm">
                &copy; {{date "2006"}} @yield('app_name', 'Dolphin App'). All rights reserved.
            </p>
        </div>
    </footer>

    @stack('scripts')
</body>
</html>`
}

// MinimalLayout returns a minimal layout with HTMX and TailwindCSS
func MinimalLayout() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>@yield('title')</title>
    
    <!-- TailwindCSS -->
    <script src="https://cdn.tailwindcss.com"></script>
    
    <!-- HTMX -->
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    
    @stack('styles')
</head>
<body>
    @yield('content')
    @stack('scripts')
</body>
</html>`
}

// HTMXLayout returns a layout optimized for HTMX with TailwindCSS
func HTMXLayout() string {
	return `<!DOCTYPE html>
<html lang="en" hx-boost="true">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>@yield('title')</title>
    
    <!-- TailwindCSS -->
    <script src="https://cdn.tailwindcss.com"></script>
    
    <!-- HTMX -->
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    
    <!-- HTMX Extensions -->
    <script src="https://unpkg.com/htmx.org/dist/ext/json-enc.js"></script>
    <script src="https://unpkg.com/htmx.org/dist/ext/ws.js"></script>
    <script src="https://unpkg.com/htmx.org/dist/ext/sse.js"></script>
    
    <!-- HTMX Indicators -->
    <style>
        .htmx-indicator {
            display: none;
        }
        .htmx-request .htmx-indicator {
            display: inline;
        }
        .htmx-request.htmx-indicator {
            display: inline;
        }
    </style>
    
    @stack('styles')
</head>
<body class="bg-gray-50">
    <!-- Loading Indicator -->
    <div id="htmx-indicator" class="htmx-indicator fixed top-0 left-0 right-0 bg-blue-600 h-1 z-50">
        <div class="bg-blue-400 h-full" style="width: 0%; animation: progress 1s ease-in-out infinite;"></div>
    </div>
    
    @yield('content')
    
    @stack('scripts')
</body>
</html>`
}

