package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Translator represents a translation interface
type Translator interface {
	Translate(key string, params map[string]interface{}) string
	TranslatePlural(key string, count int, params map[string]interface{}) string
	SetLocale(locale string) error
	GetLocale() string
	HasTranslation(key string) bool
}

// Manager manages internationalization
type Manager struct {
	translators    map[string]Translator
	defaultLocale  string
	currentLocale  string
	fallbackLocale string
	mu             sync.RWMutex
}

// NewManager creates a new i18n manager
func NewManager(defaultLocale, fallbackLocale string) *Manager {
	return &Manager{
		translators:    make(map[string]Translator),
		defaultLocale:  defaultLocale,
		currentLocale:  defaultLocale,
		fallbackLocale: fallbackLocale,
	}
}

// RegisterTranslator registers a translator for a locale
func (m *Manager) RegisterTranslator(locale string, translator Translator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.translators[locale] = translator
}

// SetLocale sets the current locale
func (m *Manager) SetLocale(locale string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.translators[locale]; !exists {
		return fmt.Errorf("translator for locale '%s' not found", locale)
	}

	m.currentLocale = locale
	return nil
}

// GetLocale returns the current locale
func (m *Manager) GetLocale() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentLocale
}

// Translate translates a key using the current locale
func (m *Manager) Translate(key string, params map[string]interface{}) string {
	m.mu.RLock()
	translator, exists := m.translators[m.currentLocale]
	m.mu.RUnlock()

	if !exists {
		// Try fallback locale
		m.mu.RLock()
		translator, exists = m.translators[m.fallbackLocale]
		m.mu.RUnlock()

		if !exists {
			return key // Return key if no translation found
		}
	}

	return translator.Translate(key, params)
}

// TranslatePlural translates a plural key using the current locale
func (m *Manager) TranslatePlural(key string, count int, params map[string]interface{}) string {
	m.mu.RLock()
	translator, exists := m.translators[m.currentLocale]
	m.mu.RUnlock()

	if !exists {
		// Try fallback locale
		m.mu.RLock()
		translator, exists = m.translators[m.fallbackLocale]
		m.mu.RUnlock()

		if !exists {
			return key // Return key if no translation found
		}
	}

	return translator.TranslatePlural(key, count, params)
}

// HasTranslation checks if a translation exists
func (m *Manager) HasTranslation(key string) bool {
	m.mu.RLock()
	translator, exists := m.translators[m.currentLocale]
	m.mu.RUnlock()

	if !exists {
		return false
	}

	return translator.HasTranslation(key)
}

// FileTranslator implements Translator using JSON files
type FileTranslator struct {
	locale       string
	translations map[string]interface{}
	pluralRules  map[string]PluralRule
	mu           sync.RWMutex
}

// PluralRule represents pluralization rules
type PluralRule struct {
	Zero  string
	One   string
	Two   string
	Few   string
	Many  string
	Other string
}

// NewFileTranslator creates a new file translator
func NewFileTranslator(locale string) *FileTranslator {
	return &FileTranslator{
		locale:       locale,
		translations: make(map[string]interface{}),
		pluralRules:  make(map[string]PluralRule),
	}
}

// LoadTranslations loads translations from a JSON file
func (ft *FileTranslator) LoadTranslations(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var translations map[string]interface{}
	err = json.Unmarshal(data, &translations)
	if err != nil {
		return err
	}

	ft.mu.Lock()
	defer ft.mu.Unlock()
	ft.translations = translations

	return nil
}

// LoadTranslationsFromDir loads all translation files from a directory
func (ft *FileTranslator) LoadTranslationsFromDir(dirPath string) error {
	files, err := filepath.Glob(filepath.Join(dirPath, "*.json"))
	if err != nil {
		return err
	}

	for _, file := range files {
		err = ft.LoadTranslations(file)
		if err != nil {
			return err
		}
	}

	return nil
}

// Translate translates a key
func (ft *FileTranslator) Translate(key string, params map[string]interface{}) string {
	ft.mu.RLock()
	defer ft.mu.RUnlock()

	translation := ft.getTranslation(key)
	if translation == "" {
		return key
	}

	// Replace parameters
	if params != nil {
		for paramKey, paramValue := range params {
			placeholder := fmt.Sprintf(":%s", paramKey)
			translation = strings.ReplaceAll(translation, placeholder, fmt.Sprintf("%v", paramValue))
		}
	}

	return translation
}

// TranslatePlural translates a plural key
func (ft *FileTranslator) TranslatePlural(key string, count int, params map[string]interface{}) string {
	ft.mu.RLock()
	defer ft.mu.RUnlock()

	// Get plural form
	pluralForm := ft.getPluralForm(key, count)
	if pluralForm == "" {
		return key
	}

	// Replace parameters
	translation := pluralForm
	if params != nil {
		params["count"] = count
		for paramKey, paramValue := range params {
			placeholder := fmt.Sprintf(":%s", paramKey)
			translation = strings.ReplaceAll(translation, placeholder, fmt.Sprintf("%v", paramValue))
		}
	}

	return translation
}

// SetLocale sets the locale
func (ft *FileTranslator) SetLocale(locale string) error {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	ft.locale = locale
	return nil
}

// GetLocale returns the locale
func (ft *FileTranslator) GetLocale() string {
	ft.mu.RLock()
	defer ft.mu.RUnlock()
	return ft.locale
}

// HasTranslation checks if a translation exists
func (ft *FileTranslator) HasTranslation(key string) bool {
	ft.mu.RLock()
	defer ft.mu.RUnlock()

	translation := ft.getTranslation(key)
	return translation != ""
}

// getTranslation gets a translation by key
func (ft *FileTranslator) getTranslation(key string) string {
	keys := strings.Split(key, ".")
	current := ft.translations

	for _, k := range keys {
		if val, ok := current[k]; ok {
			if str, ok := val.(string); ok {
				return str
			} else if mapVal, ok := val.(map[string]interface{}); ok {
				current = mapVal
			} else {
				return ""
			}
		} else {
			return ""
		}
	}

	return ""
}

// getPluralForm gets the appropriate plural form
func (ft *FileTranslator) getPluralForm(key string, count int) string {
	// Simple pluralization rules for English
	if count == 0 {
		return ft.getTranslation(key + ".zero")
	} else if count == 1 {
		return ft.getTranslation(key + ".one")
	} else {
		return ft.getTranslation(key + ".other")
	}
}

// Formatter formats values according to locale
type Formatter struct {
	locale string
}

// NewFormatter creates a new formatter
func NewFormatter(locale string) *Formatter {
	return &Formatter{locale: locale}
}

// FormatNumber formats a number according to locale
func (f *Formatter) FormatNumber(number float64) string {
	// Simple formatting for now
	return fmt.Sprintf("%.2f", number)
}

// FormatCurrency formats a currency amount
func (f *Formatter) FormatCurrency(amount float64, currency string) string {
	switch currency {
	case "USD":
		return fmt.Sprintf("$%.2f", amount)
	case "EUR":
		return fmt.Sprintf("€%.2f", amount)
	case "GBP":
		return fmt.Sprintf("£%.2f", amount)
	default:
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
}

// FormatDate formats a date according to locale
func (f *Formatter) FormatDate(date interface{}) string {
	// Simple date formatting
	return fmt.Sprintf("%v", date)
}

// FormatDateTime formats a datetime according to locale
func (f *Formatter) FormatDateTime(dateTime interface{}) string {
	// Simple datetime formatting
	return fmt.Sprintf("%v", dateTime)
}

// Middleware represents i18n middleware
type Middleware struct {
	manager *Manager
}

// NewMiddleware creates new i18n middleware
func NewMiddleware(manager *Manager) *Middleware {
	return &Middleware{manager: manager}
}

// Handle handles the middleware
func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Detect locale from header, query param, or cookie
		locale := m.detectLocale(r)

		// Set locale in context
		ctx := context.WithValue(r.Context(), "locale", locale)
		m.manager.SetLocale(locale)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// detectLocale detects the locale from request
func (m *Middleware) detectLocale(r *http.Request) string {
	// Check query parameter
	if locale := r.URL.Query().Get("locale"); locale != "" {
		return locale
	}

	// Check cookie
	if cookie, err := r.Cookie("locale"); err == nil {
		return cookie.Value
	}

	// Check Accept-Language header
	acceptLanguage := r.Header.Get("Accept-Language")
	if acceptLanguage != "" {
		locales := strings.Split(acceptLanguage, ",")
		if len(locales) > 0 {
			locale := strings.Split(locales[0], ";")[0]
			locale = strings.Split(locale, "-")[0] // Get language part
			return locale
		}
	}

	// Return default locale
	return m.manager.defaultLocale
}

// Helper functions for templates

// T translates a key in templates
func T(key string, params ...map[string]interface{}) string {
	// This would be called from templates
	// In a real implementation, this would access the global i18n manager
	if len(params) > 0 {
		_ = params[0] // Use the parameter
	}

	// For now, return the key
	return key
}

// TPlural translates a plural key in templates
func TPlural(key string, count int, params ...map[string]interface{}) string {
	// This would be called from templates
	if len(params) > 0 {
		_ = params[0] // Use the parameter
	}

	// For now, return the key
	return key
}

// FormatNumber formats a number in templates
func FormatNumber(number float64) string {
	formatter := NewFormatter("en")
	return formatter.FormatNumber(number)
}

// FormatCurrency formats currency in templates
func FormatCurrency(amount float64, currency string) string {
	formatter := NewFormatter("en")
	return formatter.FormatCurrency(amount, currency)
}

// FormatDate formats a date in templates
func FormatDate(date interface{}) string {
	formatter := NewFormatter("en")
	return formatter.FormatDate(date)
}

// FormatDateTime formats a datetime in templates
func FormatDateTime(dateTime interface{}) string {
	formatter := NewFormatter("en")
	return formatter.FormatDateTime(dateTime)
}

// Common translation keys
const (
	WelcomeMessage   = "messages.welcome"
	GoodbyeMessage   = "messages.goodbye"
	ErrorOccurred    = "errors.occurred"
	ValidationFailed = "validation.failed"
	UserNotFound     = "users.not_found"
	AccessDenied     = "auth.access_denied"
	LoginRequired    = "auth.login_required"
	EmailSent        = "mail.email_sent"
	FileUploaded     = "files.uploaded"
	RecordCreated    = "records.created"
	RecordUpdated    = "records.updated"
	RecordDeleted    = "records.deleted"
)

// Default translations
var DefaultTranslations = map[string]map[string]string{
	"en": {
		"messages.welcome":    "Welcome!",
		"messages.goodbye":    "Goodbye!",
		"errors.occurred":     "An error occurred",
		"validation.failed":   "Validation failed",
		"users.not_found":     "User not found",
		"auth.access_denied":  "Access denied",
		"auth.login_required": "Login required",
		"mail.email_sent":     "Email sent successfully",
		"files.uploaded":      "File uploaded successfully",
		"records.created":     "Record created successfully",
		"records.updated":     "Record updated successfully",
		"records.deleted":     "Record deleted successfully",
	},
	"es": {
		"messages.welcome":    "¡Bienvenido!",
		"messages.goodbye":    "¡Adiós!",
		"errors.occurred":     "Ocurrió un error",
		"validation.failed":   "La validación falló",
		"users.not_found":     "Usuario no encontrado",
		"auth.access_denied":  "Acceso denegado",
		"auth.login_required": "Inicio de sesión requerido",
		"mail.email_sent":     "Correo enviado exitosamente",
		"files.uploaded":      "Archivo subido exitosamente",
		"records.created":     "Registro creado exitosamente",
		"records.updated":     "Registro actualizado exitosamente",
		"records.deleted":     "Registro eliminado exitosamente",
	},
	"fr": {
		"messages.welcome":    "Bienvenue!",
		"messages.goodbye":    "Au revoir!",
		"errors.occurred":     "Une erreur s'est produite",
		"validation.failed":   "La validation a échoué",
		"users.not_found":     "Utilisateur non trouvé",
		"auth.access_denied":  "Accès refusé",
		"auth.login_required": "Connexion requise",
		"mail.email_sent":     "Email envoyé avec succès",
		"files.uploaded":      "Fichier téléchargé avec succès",
		"records.created":     "Enregistrement créé avec succès",
		"records.updated":     "Enregistrement mis à jour avec succès",
		"records.deleted":     "Enregistrement supprimé avec succès",
	},
}
