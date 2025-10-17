package http

import "errors"

// HTTP client errors
var (
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrRequestTimeout     = errors.New("request timeout")
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")
	ErrInvalidURL         = errors.New("invalid URL")
	ErrInvalidMethod      = errors.New("invalid HTTP method")
	ErrInvalidBody        = errors.New("invalid request body")
	ErrInvalidHeaders     = errors.New("invalid headers")
	ErrInvalidConfig      = errors.New("invalid configuration")
	ErrClientClosed       = errors.New("client is closed")
	ErrContextCancelled   = errors.New("context cancelled")
	ErrContextTimeout     = errors.New("context timeout")
)
