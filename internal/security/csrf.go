package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"go.uber.org/zap"
)

// CSRFManager manages CSRF token generation and validation
type CSRFManager struct {
	secret     []byte
	store      sessions.Store
	logger     *zap.Logger
	tokenName  string
	headerName string
	cookieName string
}

// CSRFConfig represents CSRF configuration
type CSRFConfig struct {
	Secret      string        `yaml:"secret" json:"secret"`
	TokenName   string        `yaml:"token_name" json:"token_name"`
	HeaderName  string        `yaml:"header_name" json:"header_name"`
	CookieName  string        `yaml:"cookie_name" json:"cookie_name"`
	MaxAge      int           `yaml:"max_age" json:"max_age"`
	Secure      bool          `yaml:"secure" json:"secure"`
	HttpOnly    bool          `yaml:"http_only" json:"http_only"`
	SameSite    http.SameSite `yaml:"same_site" json:"same_site"`
	ExemptPaths []string      `yaml:"exempt_paths" json:"exempt_paths"`
}

// DefaultCSRFConfig returns a default CSRF configuration
func DefaultCSRFConfig() *CSRFConfig {
	return &CSRFConfig{
		Secret:      "dolphin-csrf-secret-key-change-in-production",
		TokenName:   "csrf_token",
		HeaderName:  "X-CSRF-Token",
		CookieName:  "dolphin_csrf",
		MaxAge:      3600,  // 1 hour
		Secure:      false, // Set to true in production with HTTPS
		HttpOnly:    false, // Set to true for better security
		SameSite:    http.SameSiteStrictMode,
		ExemptPaths: []string{"/health", "/metrics", "/api/webhooks"},
	}
}

// NewCSRFManager creates a new CSRF manager
func NewCSRFManager(config *CSRFConfig, store sessions.Store, logger *zap.Logger) (*CSRFManager, error) {
	if config == nil {
		config = DefaultCSRFConfig()
	}

	secret := []byte(config.Secret)
	if len(secret) < 32 {
		// Generate a random secret if too short
		secret = make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return nil, fmt.Errorf("failed to generate secret: %w", err)
		}
	}

	return &CSRFManager{
		secret:     secret,
		store:      store,
		logger:     logger,
		tokenName:  config.TokenName,
		headerName: config.HeaderName,
		cookieName: config.CookieName,
	}, nil
}

// GenerateToken generates a new CSRF token
func (cm *CSRFManager) GenerateToken(sessionID string) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Create timestamp
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// Create payload: sessionID + timestamp + random token
	payload := fmt.Sprintf("%s:%s:%s", sessionID, timestamp, hex.EncodeToString(tokenBytes))

	// Create HMAC signature
	h := hmac.New(sha256.New, cm.secret)
	h.Write([]byte(payload))
	signature := hex.EncodeToString(h.Sum(nil))

	// Combine payload and signature
	token := fmt.Sprintf("%s:%s", payload, signature)

	// Encode as base64
	encodedToken := base64.URLEncoding.EncodeToString([]byte(token))

	return encodedToken, nil
}

// ValidateToken validates a CSRF token
func (cm *CSRFManager) ValidateToken(sessionID, token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("empty token")
	}

	// Decode base64
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return false, fmt.Errorf("invalid token format: %w", err)
	}

	// Split payload and signature
	parts := strings.Split(string(decoded), ":")
	if len(parts) != 4 {
		return false, fmt.Errorf("invalid token structure")
	}

	tokenSessionID, timestamp, randomToken, signature := parts[0], parts[1], parts[2], parts[3]

	// Verify session ID matches
	if tokenSessionID != sessionID {
		cm.logger.Debug("CSRF token session ID mismatch",
			zap.String("expected", sessionID),
			zap.String("actual", tokenSessionID))
		return false, nil
	}

	// Check token age (optional - tokens expire after 1 hour by default)
	if cm.isTokenExpired(timestamp) {
		cm.logger.Debug("CSRF token expired", zap.String("timestamp", timestamp))
		return false, nil
	}

	// Recreate payload for signature verification
	payload := fmt.Sprintf("%s:%s:%s", tokenSessionID, timestamp, randomToken)

	// Verify HMAC signature
	h := hmac.New(sha256.New, cm.secret)
	h.Write([]byte(payload))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		cm.logger.Debug("CSRF token signature mismatch")
		return false, nil
	}

	return true, nil
}

// isTokenExpired checks if a token is expired based on timestamp
func (cm *CSRFManager) isTokenExpired(timestamp string) bool {
	var ts int64
	if _, err := fmt.Sscanf(timestamp, "%d", &ts); err != nil {
		return true
	}

	// Tokens expire after 1 hour
	return time.Now().Unix()-ts > 3600
}

// GetTokenFromRequest extracts CSRF token from request
func (cm *CSRFManager) GetTokenFromRequest(r *http.Request) string {
	// Try header first
	if token := r.Header.Get(cm.headerName); token != "" {
		return token
	}

	// Try form data
	if token := r.FormValue(cm.tokenName); token != "" {
		return token
	}

	// Try query parameter
	if token := r.URL.Query().Get(cm.tokenName); token != "" {
		return token
	}

	return ""
}

// SetTokenCookie sets CSRF token in cookie
func (cm *CSRFManager) SetTokenCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     cm.cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   3600,  // 1 hour
		Secure:   false, // Set to true in production with HTTPS
		HttpOnly: false, // Set to true for better security
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}

// GetTokenFromCookie gets CSRF token from cookie
func (cm *CSRFManager) GetTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(cm.cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// CSRFMiddleware creates CSRF protection middleware
func CSRFMiddleware(manager *CSRFManager, config *CSRFConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CSRF check for exempt paths
			if isExemptPath(r.URL.Path, config.ExemptPaths) {
				next.ServeHTTP(w, r)
				return
			}

			// Skip CSRF check for safe methods
			if isSafeMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			// Get session ID
			session, err := manager.store.Get(r, "dolphin_session")
			if err != nil {
				manager.logger.Error("Failed to get session", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			sessionID := session.ID
			if sessionID == "" {
				// Generate new session ID if none exists
				sessionID = generateSessionID()
				session.ID = sessionID
			}

			// Get token from request
			token := manager.GetTokenFromRequest(r)
			if token == "" {
				manager.logger.Warn("CSRF token missing",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("ip", r.RemoteAddr))
				http.Error(w, "CSRF token missing", http.StatusForbidden)
				return
			}

			// Validate token
			valid, err := manager.ValidateToken(sessionID, token)
			if err != nil {
				manager.logger.Error("CSRF token validation error", zap.Error(err))
				http.Error(w, "CSRF token validation failed", http.StatusForbidden)
				return
			}

			if !valid {
				manager.logger.Warn("CSRF token invalid",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("ip", r.RemoteAddr))
				http.Error(w, "CSRF token invalid", http.StatusForbidden)
				return
			}

			// Token is valid, continue
			next.ServeHTTP(w, r)
		})
	}
}

// CSRFHandler handles CSRF token generation and validation
type CSRFHandler struct {
	manager *CSRFManager
	config  *CSRFConfig
}

// NewCSRFHandler creates a new CSRF handler
func NewCSRFHandler(manager *CSRFManager, config *CSRFConfig) *CSRFHandler {
	return &CSRFHandler{
		manager: manager,
		config:  config,
	}
}

// GenerateTokenHandler generates and returns a CSRF token
func (ch *CSRFHandler) GenerateTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Get or create session
	session, err := ch.manager.store.Get(r, "dolphin_session")
	if err != nil {
		ch.manager.logger.Error("Failed to get session", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sessionID := session.ID
	if sessionID == "" {
		sessionID = generateSessionID()
		session.ID = sessionID
	}

	// Generate token
	token, err := ch.manager.GenerateToken(sessionID)
	if err != nil {
		ch.manager.logger.Error("Failed to generate CSRF token", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set token in session
	session.Values[ch.manager.tokenName] = token

	// Save session
	if err := session.Save(r, w); err != nil {
		ch.manager.logger.Error("Failed to save session", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set token in cookie
	ch.manager.SetTokenCookie(w, token)

	// Return token as JSON
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"token":"%s"}`, token)
}

// ValidateTokenHandler validates a CSRF token
func (ch *CSRFHandler) ValidateTokenHandler(w http.ResponseWriter, r *http.Request) {
	token := ch.manager.GetTokenFromRequest(r)
	if token == "" {
		http.Error(w, "Token required", http.StatusBadRequest)
		return
	}

	// Get session
	session, err := ch.manager.store.Get(r, "dolphin_session")
	if err != nil {
		ch.manager.logger.Error("Failed to get session", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sessionID := session.ID
	if sessionID == "" {
		http.Error(w, "Session required", http.StatusUnauthorized)
		return
	}

	// Validate token
	valid, err := ch.manager.ValidateToken(sessionID, token)
	if err != nil {
		ch.manager.logger.Error("CSRF token validation error", zap.Error(err))
		http.Error(w, "Token validation failed", http.StatusInternalServerError)
		return
	}

	// Return validation result
	w.Header().Set("Content-Type", "application/json")
	if valid {
		fmt.Fprintf(w, `{"valid":true}`)
	} else {
		fmt.Fprintf(w, `{"valid":false}`)
	}
}

// Template helpers for CSRF protection
type CSRFTemplateHelper struct {
	manager *CSRFManager
}

// NewCSRFTemplateHelper creates a new CSRF template helper
func NewCSRFTemplateHelper(manager *CSRFManager) *CSRFTemplateHelper {
	return &CSRFTemplateHelper{manager: manager}
}

// Token returns the CSRF token for templates
func (cth *CSRFTemplateHelper) Token(sessionID string) string {
	token, err := cth.manager.GenerateToken(sessionID)
	if err != nil {
		cth.manager.logger.Error("Failed to generate CSRF token for template", zap.Error(err))
		return ""
	}
	return token
}

// TokenField returns a hidden input field with CSRF token
func (cth *CSRFTemplateHelper) TokenField(sessionID string) string {
	token := cth.Token(sessionID)
	return fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, cth.manager.tokenName, token)
}

// MetaTag returns a meta tag with CSRF token
func (cth *CSRFTemplateHelper) MetaTag(sessionID string) string {
	token := cth.Token(sessionID)
	return fmt.Sprintf(`<meta name="%s" content="%s">`, cth.manager.tokenName, token)
}

// HeaderName returns the CSRF header name
func (cth *CSRFTemplateHelper) HeaderName() string {
	return cth.manager.headerName
}

// TokenName returns the CSRF token name
func (cth *CSRFTemplateHelper) TokenName() string {
	return cth.manager.tokenName
}

// Helper functions

// isExemptPath checks if a path is exempt from CSRF protection
func isExemptPath(path string, exemptPaths []string) bool {
	for _, exempt := range exemptPaths {
		if strings.HasPrefix(path, exempt) {
			return true
		}
	}
	return false
}

// isSafeMethod checks if an HTTP method is safe (doesn't require CSRF protection)
func isSafeMethod(method string) bool {
	safeMethods := []string{"GET", "HEAD", "OPTIONS", "TRACE"}
	for _, safe := range safeMethods {
		if method == safe {
			return true
		}
	}
	return false
}

// generateSessionID generates a random session ID
func generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CSRFConfigFromEnv creates CSRF config from environment variables
func CSRFConfigFromEnv() *CSRFConfig {
	config := DefaultCSRFConfig()

	// Override with environment variables if present
	if secret := getEnv("CSRF_SECRET", ""); secret != "" {
		config.Secret = secret
	}
	if tokenName := getEnv("CSRF_TOKEN_NAME", ""); tokenName != "" {
		config.TokenName = tokenName
	}
	if headerName := getEnv("CSRF_HEADER_NAME", ""); headerName != "" {
		config.HeaderName = headerName
	}
	if cookieName := getEnv("CSRF_COOKIE_NAME", ""); cookieName != "" {
		config.CookieName = cookieName
	}
	if secure := getEnv("CSRF_SECURE", ""); secure == "true" {
		config.Secure = true
	}
	if httpOnly := getEnv("CSRF_HTTP_ONLY", ""); httpOnly == "true" {
		config.HttpOnly = true
	}

	return config
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
