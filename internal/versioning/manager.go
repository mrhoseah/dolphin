package versioning

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// Version represents an API version
type Version struct {
	Major int
	Minor int
	Patch int
}

// String returns the version as a string
func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare compares two versions
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		return v.Major - other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor - other.Minor
	}
	return v.Patch - other.Patch
}

// IsCompatible checks if this version is compatible with another
func (v Version) IsCompatible(other Version) bool {
	// Same major version means compatible
	return v.Major == other.Major
}

// ParseVersion parses a version string
func ParseVersion(versionStr string) (Version, error) {
	// Remove 'v' prefix if present
	versionStr = strings.TrimPrefix(versionStr, "v")

	// Split by dots
	parts := strings.Split(versionStr, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version format: %s", versionStr)
	}

	// Parse major version
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %s", parts[0])
	}

	// Parse minor version
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	// Parse patch version
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

// VersioningStrategy defines how API versioning is handled
type VersioningStrategy int

const (
	// HeaderVersioning uses Accept header for versioning
	HeaderVersioning VersioningStrategy = iota
	// URLVersioning uses URL path for versioning
	URLVersioning
	// QueryVersioning uses query parameter for versioning
	QueryVersioning
	// CustomVersioning uses custom logic for versioning
	CustomVersioning
)

// VersioningConfig defines versioning configuration
type VersioningConfig struct {
	Strategy          VersioningStrategy
	DefaultVersion    Version
	SupportedVersions []Version
	HeaderName        string
	QueryParam        string
	URLPrefix         string
	CustomExtractor   func(r *http.Request) (Version, error)
}

// DefaultVersioningConfig returns default versioning configuration
func DefaultVersioningConfig() *VersioningConfig {
	return &VersioningConfig{
		Strategy:       HeaderVersioning,
		DefaultVersion: Version{Major: 1, Minor: 0, Patch: 0},
		SupportedVersions: []Version{
			{Major: 1, Minor: 0, Patch: 0},
			{Major: 1, Minor: 1, Patch: 0},
			{Major: 2, Minor: 0, Patch: 0},
		},
		HeaderName: "Accept",
		QueryParam: "version",
		URLPrefix:  "/api/v",
	}
}

// VersionManager manages API versioning
type VersionManager struct {
	config *VersioningConfig
}

// NewVersionManager creates a new version manager
func NewVersionManager(config *VersioningConfig) *VersionManager {
	if config == nil {
		config = DefaultVersioningConfig()
	}

	return &VersionManager{
		config: config,
	}
}

// ExtractVersion extracts the API version from the request
func (vm *VersionManager) ExtractVersion(r *http.Request) (Version, error) {
	switch vm.config.Strategy {
	case HeaderVersioning:
		return vm.extractFromHeader(r)
	case URLVersioning:
		return vm.extractFromURL(r)
	case QueryVersioning:
		return vm.extractFromQuery(r)
	case CustomVersioning:
		if vm.config.CustomExtractor != nil {
			return vm.config.CustomExtractor(r)
		}
		return vm.config.DefaultVersion, nil
	default:
		return vm.config.DefaultVersion, nil
	}
}

// extractFromHeader extracts version from Accept header
func (vm *VersionManager) extractFromHeader(r *http.Request) (Version, error) {
	accept := r.Header.Get(vm.config.HeaderName)
	if accept == "" {
		return vm.config.DefaultVersion, nil
	}

	// Parse Accept header: application/json; version=v1.0.0
	re := regexp.MustCompile(`version=([^;,\s]+)`)
	matches := re.FindStringSubmatch(accept)
	if len(matches) < 2 {
		return vm.config.DefaultVersion, nil
	}

	version, err := ParseVersion(matches[1])
	if err != nil {
		return vm.config.DefaultVersion, err
	}

	return version, nil
}

// extractFromURL extracts version from URL path
func (vm *VersionManager) extractFromURL(r *http.Request) (Version, error) {
	path := r.URL.Path

	// Look for version in URL: /api/v1.0.0/users
	re := regexp.MustCompile(vm.config.URLPrefix + `(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 2 {
		return vm.config.DefaultVersion, nil
	}

	version, err := ParseVersion(matches[1])
	if err != nil {
		return vm.config.DefaultVersion, err
	}

	return version, nil
}

// extractFromQuery extracts version from query parameter
func (vm *VersionManager) extractFromQuery(r *http.Request) (Version, error) {
	versionStr := r.URL.Query().Get(vm.config.QueryParam)
	if versionStr == "" {
		return vm.config.DefaultVersion, nil
	}

	version, err := ParseVersion(versionStr)
	if err != nil {
		return vm.config.DefaultVersion, err
	}

	return version, nil
}

// IsVersionSupported checks if a version is supported
func (vm *VersionManager) IsVersionSupported(version Version) bool {
	for _, supported := range vm.config.SupportedVersions {
		if version.Major == supported.Major && version.Minor == supported.Minor {
			return true
		}
	}
	return false
}

// GetSupportedVersions returns the list of supported versions
func (vm *VersionManager) GetSupportedVersions() []Version {
	return vm.config.SupportedVersions
}

// GetDefaultVersion returns the default version
func (vm *VersionManager) GetDefaultVersion() Version {
	return vm.config.DefaultVersion
}

// VersionMiddleware creates a middleware that extracts and validates API version
func (vm *VersionManager) VersionMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract version
			version, err := vm.ExtractVersion(r)
			if err != nil {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, map[string]string{
					"error":   "Invalid version format",
					"message": err.Error(),
				})
				return
			}

			// Check if version is supported
			if !vm.IsVersionSupported(version) {
				render.Status(r, http.StatusNotAcceptable)
				render.JSON(w, r, map[string]interface{}{
					"error":              "Unsupported API version",
					"message":            fmt.Sprintf("Version %s is not supported", version.String()),
					"supported_versions": vm.GetSupportedVersions(),
					"default_version":    vm.GetDefaultVersion().String(),
				})
				return
			}

			// Add version to context
			ctx := context.WithValue(r.Context(), "api_version", version)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetVersionFromContext extracts version from context
func GetVersionFromContext(ctx context.Context) (Version, bool) {
	version, ok := ctx.Value("api_version").(Version)
	return version, ok
}

// VersionedHandler defines a handler that can handle different versions
type VersionedHandler interface {
	HandleV1(w http.ResponseWriter, r *http.Request)
	HandleV2(w http.ResponseWriter, r *http.Request)
	// Add more version handlers as needed
}

// VersionedHandlerFunc defines a function that can handle different versions
type VersionedHandlerFunc func(version Version, w http.ResponseWriter, r *http.Request)

// HandleVersioned creates a handler that routes to the appropriate version
func HandleVersioned(handler VersionedHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		version, ok := GetVersionFromContext(r.Context())
		if !ok {
			version = Version{Major: 1, Minor: 0, Patch: 0} // Default version
		}

		handler(version, w, r)
	}
}

// SetupVersionedRoutes sets up versioned API routes
func SetupVersionedRoutes(r chi.Router, manager *VersionManager) {
	// Apply version middleware
	r.Use(manager.VersionMiddleware())

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Users API
		r.Route("/users", func(r chi.Router) {
			r.Get("/", HandleVersioned(handleUsers))
			r.Post("/", HandleVersioned(handleCreateUser))
			r.Get("/{id}", HandleVersioned(handleGetUser))
			r.Put("/{id}", HandleVersioned(handleUpdateUser))
			r.Delete("/{id}", HandleVersioned(handleDeleteUser))
		})

		// Posts API
		r.Route("/posts", func(r chi.Router) {
			r.Get("/", HandleVersioned(handlePosts))
			r.Post("/", HandleVersioned(handleCreatePost))
			r.Get("/{id}", HandleVersioned(handleGetPost))
			r.Put("/{id}", HandleVersioned(handleUpdatePost))
			r.Delete("/{id}", HandleVersioned(handleDeletePost))
		})
	})
}

// Example versioned handlers
func handleUsers(version Version, w http.ResponseWriter, r *http.Request) {
	switch version.Major {
	case 1:
		handleUsersV1(w, r)
	case 2:
		handleUsersV2(w, r)
	default:
		handleUsersV1(w, r)
	}
}

func handleUsersV1(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]interface{}{
		"version": "v1",
		"users": []map[string]interface{}{
			{"id": 1, "name": "John Doe", "email": "john@example.com"},
			{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
		},
	})
}

func handleUsersV2(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]interface{}{
		"version": "v2",
		"data": []map[string]interface{}{
			{"id": 1, "name": "John Doe", "email": "john@example.com", "created_at": "2023-01-01T00:00:00Z"},
			{"id": 2, "name": "Jane Smith", "email": "jane@example.com", "created_at": "2023-01-02T00:00:00Z"},
		},
		"meta": map[string]interface{}{
			"total":    2,
			"page":     1,
			"per_page": 10,
		},
	})
}

// Placeholder handlers for other endpoints
func handleCreateUser(version Version, w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{"message": "Create user endpoint"})
}

func handleGetUser(version Version, w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{"message": "Get user endpoint"})
}

func handleUpdateUser(version Version, w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{"message": "Update user endpoint"})
}

func handleDeleteUser(version Version, w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{"message": "Delete user endpoint"})
}

func handlePosts(version Version, w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{"message": "Posts endpoint"})
}

func handleCreatePost(version Version, w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{"message": "Create post endpoint"})
}

func handleGetPost(version Version, w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{"message": "Get post endpoint"})
}

func handleUpdatePost(version Version, w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{"message": "Update post endpoint"})
}

func handleDeletePost(version Version, w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{"message": "Delete post endpoint"})
}
