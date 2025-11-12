package router

import (
	"context"
	"net/http"
)

type contextKey string

const (
	// UserContextKey is the key for storing authenticated user in context
	UserContextKey contextKey = "user"
	// RequestIDContextKey is the key for storing request ID in context
	RequestIDContextKey contextKey = "request_id"
	// IPAddressContextKey is the key for storing client IP in context
	IPAddressContextKey contextKey = "ip_address"
)

// GetUser retrieves the authenticated user from request context
func GetUser(r *http.Request) interface{} {
	if user := r.Context().Value(UserContextKey); user != nil {
		return user
	}
	return nil
}

// SetUser sets the authenticated user in request context
func SetUser(r *http.Request, user interface{}) *http.Request {
	ctx := context.WithValue(r.Context(), UserContextKey, user)
	return r.WithContext(ctx)
}

// GetRequestID retrieves the request ID from context
func GetRequestID(r *http.Request) string {
	if id := r.Context().Value(RequestIDContextKey); id != nil {
		if str, ok := id.(string); ok {
			return str
		}
	}
	return ""
}

// SetRequestID sets the request ID in context
func SetRequestID(r *http.Request, id string) *http.Request {
	ctx := context.WithValue(r.Context(), RequestIDContextKey, id)
	return r.WithContext(ctx)
}

// GetIPAddress retrieves the client IP address from context
func GetIPAddress(r *http.Request) string {
	if ip := r.Context().Value(IPAddressContextKey); ip != nil {
		if str, ok := ip.(string); ok {
			return str
		}
	}
	return r.RemoteAddr
}

// SetIPAddress sets the client IP address in context
func SetIPAddress(r *http.Request, ip string) *http.Request {
	ctx := context.WithValue(r.Context(), IPAddressContextKey, ip)
	return r.WithContext(ctx)
}

