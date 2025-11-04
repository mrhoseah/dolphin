package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"go.uber.org/zap"
)

// RecoveryConfig configures recovery middleware behavior
type RecoveryConfig struct {
	Logger      *zap.Logger
	DebugMode   bool
	Environment string
}

// New creates recovery middleware with default config
func New(logger *zap.Logger) func(next http.Handler) http.Handler {
	return NewWithConfig(RecoveryConfig{
		Logger:      logger,
		DebugMode:   false,
		Environment: "production",
	})
}

// NewWithConfig creates recovery middleware with custom config
func NewWithConfig(config RecoveryConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Get stack trace
					stack := make([]byte, 8192)
					length := runtime.Stack(stack, false)
					stackTrace := string(stack[:length])

					// Log the panic with full details
					config.Logger.Error("Panic recovered",
						zap.Any("error", err),
						zap.String("stack", stackTrace),
						zap.String("method", r.Method),
						zap.String("url", r.URL.String()),
						zap.String("path", r.URL.Path),
					)

					// Render error response
					renderErrorResponse(w, r, err, stackTrace, config)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// renderErrorResponse renders an appropriate error response based on request type and environment
func renderErrorResponse(w http.ResponseWriter, r *http.Request, err interface{}, stackTrace string, config RecoveryConfig) {
	// Check if this is an API request
	if strings.HasPrefix(r.URL.Path, "/api/") {
		renderAPIError(w, err, stackTrace, config)
		return
	}

	// Render HTML error page for web requests
	renderHTMLError(w, err, stackTrace, config)
}

// renderAPIError renders JSON error response for API requests
func renderAPIError(w http.ResponseWriter, err interface{}, stackTrace string, config RecoveryConfig) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	response := map[string]interface{}{
		"error":   "Internal server error",
		"message": "An unexpected error occurred",
		"status":  500,
	}

	// Add debug info in development or if debug mode is enabled
	if config.DebugMode || config.Environment == "development" {
		response["error_details"] = fmt.Sprintf("%v", err)
		response["stack_trace"] = stackTrace
	}

	json.NewEncoder(w).Encode(response)
}

// renderHTMLError renders HTML error page for web requests
func renderHTMLError(w http.ResponseWriter, err interface{}, stackTrace string, config RecoveryConfig) {
	debugMode := config.DebugMode || config.Environment == "development"

	errorDetails := ""
	if debugMode {
		errorDetails = fmt.Sprintf(`
		<div class="debug-section bg-gray-50 border border-gray-200 rounded-lg p-6 mt-6">
			<h3 class="text-lg font-semibold text-gray-900 mb-4">Error Details</h3>
			<div class="space-y-3">
				<div>
					<strong class="text-gray-700">Error:</strong>
					<pre class="mt-1 p-3 bg-gray-800 text-red-400 rounded text-sm overflow-x-auto"><code>%v</code></pre>
				</div>
				<div>
					<strong class="text-gray-700">Stack Trace:</strong>
					<pre class="mt-1 p-3 bg-gray-800 text-gray-300 rounded text-xs overflow-x-auto max-h-96 overflow-y-auto"><code>%s</code></pre>
				</div>
			</div>
		</div>`, err, stackTrace)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Internal Server Error - Dolphin Framework</title>
	<script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gradient-to-br from-red-50 to-orange-100 min-h-screen flex items-center justify-center p-4">
	<div class="bg-white rounded-2xl shadow-2xl max-w-3xl w-full p-8 md:p-12">
		<div class="text-center mb-8">
			<div class="text-6xl md:text-8xl mb-4">⚠️</div>
			<h1 class="text-3xl md:text-4xl font-bold text-gray-900 mb-2">Internal Server Error</h1>
			<p class="text-lg text-gray-600">An unexpected error occurred while processing your request.</p>
			%s
		</div>
		
		<div class="border-t border-gray-200 pt-6">
			<div class="flex flex-col sm:flex-row gap-4 justify-center">
				<a href="/" class="inline-flex items-center justify-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors">
					<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"></path>
					</svg>
					Go Home
				</a>
				<button onclick="window.history.back()" class="inline-flex items-center justify-center px-6 py-3 border border-gray-300 text-base font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors">
					<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path>
					</svg>
					Go Back
				</button>
			</div>
		</div>
	</div>
</body>
</html>`, errorDetails)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(html))
}
