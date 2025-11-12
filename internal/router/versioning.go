package router

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// APIVersion represents an API version
type APIVersion string

const (
	// APIVersionV1 represents API version 1
	APIVersionV1 APIVersion = "v1"
	// APIVersionV2 represents API version 2
	APIVersionV2 APIVersion = "v2"
)

// VersionFromRequest extracts API version from request
func VersionFromRequest(r *http.Request) APIVersion {
	// Check URL path for version (e.g., /api/v1/users)
	path := r.URL.Path
	if strings.HasPrefix(path, "/api/v2/") {
		return APIVersionV2
	}
	if strings.HasPrefix(path, "/api/v1/") {
		return APIVersionV1
	}
	
	// Check Accept header (e.g., application/vnd.api+json;version=2)
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "version=2") {
		return APIVersionV2
	}
	if strings.Contains(accept, "version=1") {
		return APIVersionV1
	}
	
	// Check custom header
	if version := r.Header.Get("X-API-Version"); version != "" {
		return APIVersion(version)
	}
	
	// Default to v1
	return APIVersionV1
}

// VersionMiddleware creates middleware to handle API versioning
func VersionMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			version := VersionFromRequest(r)
			w.Header().Set("X-API-Version", string(version))
			next.ServeHTTP(w, r)
		})
	}
}

// VersionRouter creates a router group for a specific API version
func (r *Router) VersionRouter(version APIVersion, fn func(chi.Router)) {
	versionPath := "/api/" + string(version)
	r.router.Route(versionPath, fn)
}

// IsVersion checks if the request is for a specific version
func IsVersion(r *http.Request, version APIVersion) bool {
	return VersionFromRequest(r) == version
}

