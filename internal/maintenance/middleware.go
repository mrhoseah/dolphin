package maintenance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Middleware provides maintenance mode middleware
type Middleware struct {
	manager *Manager
}

// NewMiddleware creates a new maintenance middleware
func NewMiddleware(manager *Manager) *Middleware {
	return &Middleware{
		manager: manager,
	}
}

// Handle returns the maintenance middleware handler
func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if maintenance mode is enabled
		if !m.manager.IsEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		// Get client IP
		clientIP := m.getClientIP(r)

		// Check if IP is allowed
		if m.manager.IsIPAllowed(clientIP) {
			next.ServeHTTP(w, r)
			return
		}

		// Check for bypass secret
		if m.checkBypassSecret(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Return maintenance response
		m.returnMaintenanceResponse(w, r)
	})
}

// getClientIP extracts the real client IP from request
func (m *Middleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	return ip
}

// checkBypassSecret checks if the request has a valid bypass secret
func (m *Middleware) checkBypassSecret(r *http.Request) bool {
	// Check query parameter
	if secret := r.URL.Query().Get("bypass"); secret != "" {
		return m.manager.IsBypassSecretValid(secret)
	}

	// Check header
	if secret := r.Header.Get("X-Bypass-Secret"); secret != "" {
		return m.manager.IsBypassSecretValid(secret)
	}

	// Check cookie
	if cookie, err := r.Cookie("bypass_secret"); err == nil {
		return m.manager.IsBypassSecretValid(cookie.Value)
	}

	return false
}

// returnMaintenanceResponse returns the maintenance mode response
func (m *Middleware) returnMaintenanceResponse(w http.ResponseWriter, r *http.Request) {
	// Set headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Set retry-after header
	if retryAfter := m.manager.GetRetryAfter(); retryAfter > 0 {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
	}

	// Set status code
	w.WriteHeader(http.StatusServiceUnavailable)

	// Prepare response
	response := map[string]interface{}{
		"error":   "Service Unavailable",
		"message": m.manager.GetMessage(),
		"status":  "maintenance",
		"code":    503,
	}

	// Add additional info for API requests
	if strings.HasPrefix(r.URL.Path, "/api/") {
		response["retry_after"] = m.manager.GetRetryAfter()
		response["timestamp"] = time.Now().Unix()
	}

	// Encode and send response
	json.NewEncoder(w).Encode(response)
}

// HTMLResponse returns an HTML maintenance page
func (m *Middleware) HTMLResponse(w http.ResponseWriter, r *http.Request) {
	// Set headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Set retry-after header
	if retryAfter := m.manager.GetRetryAfter(); retryAfter > 0 {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
	}

	// Set status code
	w.WriteHeader(http.StatusServiceUnavailable)

	// HTML template
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Maintenance Mode - Dolphin Framework</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
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
        .retry-info {
            background: #f8f9fa;
            border-radius: 8px;
            padding: 1rem;
            margin-top: 2rem;
            font-size: 0.9rem;
            color: #555;
        }
        .bypass-form {
            margin-top: 2rem;
            padding-top: 2rem;
            border-top: 1px solid #eee;
        }
        .bypass-form input {
            padding: 0.75rem;
            border: 1px solid #ddd;
            border-radius: 6px;
            margin-right: 0.5rem;
            width: 200px;
        }
        .bypass-form button {
            padding: 0.75rem 1.5rem;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 6px;
            cursor: pointer;
        }
        .bypass-form button:hover {
            background: #5a6fd8;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">🐬</div>
        <h1>Maintenance Mode</h1>
        <p>%s</p>
        
        <div class="retry-info">
            <strong>We're currently performing scheduled maintenance.</strong><br>
            Please check back in a few minutes.
        </div>
        
        <div class="bypass-form">
            <p style="font-size: 0.9rem; color: #888; margin-bottom: 1rem;">
                Have a bypass secret? Enter it below:
            </p>
            <form method="GET">
                <input type="text" name="bypass" placeholder="Enter bypass secret" required>
                <button type="submit">Access</button>
            </form>
        </div>
    </div>
    
    <script>
        // Auto-refresh every 30 seconds
        setTimeout(function() {
            window.location.reload();
        }, 30000);
    </script>
</body>
</html>`, m.manager.GetMessage())

	w.Write([]byte(html))
}
