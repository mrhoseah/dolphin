package http

import (
	"context"
	"time"
)

// RequestOption represents a request option function
type RequestOption func(*Request)

// WithContext sets the request context
func WithContext(ctx context.Context) RequestOption {
	return func(req *Request) {
		req.Context = ctx
	}
}

// WithHeaders sets request headers
func WithHeaders(headers map[string]string) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		for key, value := range headers {
			req.Headers[key] = value
		}
	}
}

// WithHeader sets a single request header
func WithHeader(key, value string) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers[key] = value
	}
}

// WithQueryParams sets query parameters
func WithQueryParams(params map[string]interface{}) RequestOption {
	return func(req *Request) {
		if req.QueryParams == nil {
			req.QueryParams = make(map[string]interface{})
		}
		for key, value := range params {
			req.QueryParams[key] = value
		}
	}
}

// WithQueryParam sets a single query parameter
func WithQueryParam(key string, value interface{}) RequestOption {
	return func(req *Request) {
		if req.QueryParams == nil {
			req.QueryParams = make(map[string]interface{})
		}
		req.QueryParams[key] = value
	}
}

// WithBody sets the request body
func WithBody(body interface{}) RequestOption {
	return func(req *Request) {
		req.Body = body
	}
}

// WithTimeout sets the request timeout
func WithTimeout(timeout time.Duration) RequestOption {
	return func(req *Request) {
		req.Timeout = timeout
	}
}

// WithRetries sets the number of retries
func WithRetries(retries int) RequestOption {
	return func(req *Request) {
		req.Retries = retries
	}
}

// WithCorrelationID sets the correlation ID
func WithCorrelationID(correlationID string) RequestOption {
	return func(req *Request) {
		req.CorrelationID = correlationID
	}
}

// WithContentType sets the content type header
func WithContentType(contentType string) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["Content-Type"] = contentType
	}
}

// WithAccept sets the accept header
func WithAccept(accept string) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["Accept"] = accept
	}
}

// WithAuthorization sets the authorization header
func WithAuthorization(auth string) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["Authorization"] = auth
	}
}

// WithBearerToken sets the bearer token
func WithBearerToken(token string) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["Authorization"] = "Bearer " + token
	}
}

// WithAPIKey sets the API key header
func WithAPIKey(key, value string) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers[key] = value
	}
}

// WithUserAgent sets the user agent header
func WithUserAgent(userAgent string) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["User-Agent"] = userAgent
	}
}

// WithCustomHeader sets a custom header
func WithCustomHeader(key, value string) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers[key] = value
	}
}

// WithJSON sets the body as JSON and content type
func WithJSON(body interface{}) RequestOption {
	return func(req *Request) {
		req.Body = body
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["Content-Type"] = "application/json"
	}
}

// WithXML sets the body as XML and content type
func WithXML(body interface{}) RequestOption {
	return func(req *Request) {
		req.Body = body
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["Content-Type"] = "application/xml"
	}
}

// WithFormData sets the body as form data and content type
func WithFormData(data map[string]interface{}) RequestOption {
	return func(req *Request) {
		req.Body = data
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	}
}

// WithMultipartFormData sets the body as multipart form data and content type
func WithMultipartFormData(data map[string]interface{}) RequestOption {
	return func(req *Request) {
		req.Body = data
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["Content-Type"] = "multipart/form-data"
	}
}

// WithBasicAuth sets basic authentication
func WithBasicAuth(username, password string) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		// This would typically encode the basic auth
		// For now, we'll set it as a custom header
		req.Headers["Authorization"] = "Basic " + username + ":" + password
	}
}

// WithDigestAuth sets digest authentication
func WithDigestAuth(username, password string) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["Authorization"] = "Digest " + username + ":" + password
	}
}

// WithFollowRedirects enables following redirects
func WithFollowRedirects(follow bool) RequestOption {
	return func(req *Request) {
		// This would typically be handled by the client configuration
		// For now, we'll store it as a custom header
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		if follow {
			req.Headers["X-Follow-Redirects"] = "true"
		} else {
			req.Headers["X-Follow-Redirects"] = "false"
		}
	}
}

// WithMaxRedirects sets the maximum number of redirects to follow
func WithMaxRedirects(max int) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["X-Max-Redirects"] = string(rune(max))
	}
}

// WithTimeout sets the request timeout
func WithRequestTimeout(timeout time.Duration) RequestOption {
	return func(req *Request) {
		req.Timeout = timeout
	}
}

// WithRetryDelay sets the retry delay
func WithRetryDelay(delay time.Duration) RequestOption {
	return func(req *Request) {
		// This would typically be handled by the client configuration
		// For now, we'll store it as a custom header
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["X-Retry-Delay"] = delay.String()
	}
}

// WithRetryBackoff sets the retry backoff multiplier
func WithRetryBackoff(backoff float64) RequestOption {
	return func(req *Request) {
		// This would typically be handled by the client configuration
		// For now, we'll store it as a custom header
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["X-Retry-Backoff"] = string(rune(int(backoff * 100)))
	}
}

// WithMaxRetryDelay sets the maximum retry delay
func WithMaxRetryDelay(maxDelay time.Duration) RequestOption {
	return func(req *Request) {
		// This would typically be handled by the client configuration
		// For now, we'll store it as a custom header
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["X-Max-Retry-Delay"] = maxDelay.String()
	}
}

// WithRetryOnStatus sets the status codes to retry on
func WithRetryOnStatus(statusCodes []int) RequestOption {
	return func(req *Request) {
		// This would typically be handled by the client configuration
		// For now, we'll store it as a custom header
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["X-Retry-On-Status"] = "true"
	}
}

// WithCircuitBreaker enables circuit breaker for this request
func WithCircuitBreaker(enabled bool) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		if enabled {
			req.Headers["X-Circuit-Breaker"] = "true"
		} else {
			req.Headers["X-Circuit-Breaker"] = "false"
		}
	}
}

// WithRateLimit enables rate limiting for this request
func WithRateLimit(enabled bool) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		if enabled {
			req.Headers["X-Rate-Limit"] = "true"
		} else {
			req.Headers["X-Rate-Limit"] = "false"
		}
	}
}

// WithMetrics enables metrics for this request
func WithMetrics(enabled bool) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		if enabled {
			req.Headers["X-Metrics"] = "true"
		} else {
			req.Headers["X-Metrics"] = "false"
		}
	}
}

// WithLogging enables logging for this request
func WithLogging(enabled bool) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		if enabled {
			req.Headers["X-Logging"] = "true"
		} else {
			req.Headers["X-Logging"] = "false"
		}
	}
}

// WithVerboseLogging enables verbose logging for this request
func WithVerboseLogging(enabled bool) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		if enabled {
			req.Headers["X-Verbose-Logging"] = "true"
		} else {
			req.Headers["X-Verbose-Logging"] = "false"
		}
	}
}

// WithLogRequestBody enables logging of request body
func WithLogRequestBody(enabled bool) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		if enabled {
			req.Headers["X-Log-Request-Body"] = "true"
		} else {
			req.Headers["X-Log-Request-Body"] = "false"
		}
	}
}

// WithLogResponseBody enables logging of response body
func WithLogResponseBody(enabled bool) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		if enabled {
			req.Headers["X-Log-Response-Body"] = "true"
		} else {
			req.Headers["X-Log-Response-Body"] = "false"
		}
	}
}

// WithCustomOption sets a custom option
func WithCustomOption(key, value string) RequestOption {
	return func(req *Request) {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["X-Custom-"+key] = value
	}
}
