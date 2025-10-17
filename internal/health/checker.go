package health

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// HealthChecker defines the interface for health checks
type HealthChecker interface {
	Check(ctx context.Context) HealthStatus
	GetName() string
}

// HealthStatus represents the status of a health check
type HealthStatus struct {
	Name      string            `json:"name"`
	Status    string            `json:"status"` // "healthy", "unhealthy", "degraded"
	Message   string            `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Duration  time.Duration     `json:"duration"`
}

// HealthResponse represents the overall health response
type HealthResponse struct {
	Status    string                    `json:"status"` // "healthy", "unhealthy", "degraded"
	Timestamp time.Time                 `json:"timestamp"`
	Duration  time.Duration             `json:"duration"`
	Checks    map[string]HealthStatus   `json:"checks"`
	Version   string                    `json:"version,omitempty"`
	Uptime    time.Duration             `json:"uptime,omitempty"`
}

// DatabaseHealthChecker checks database connectivity
type DatabaseHealthChecker struct {
	db     *sql.DB
	name   string
	logger *zap.Logger
}

// NewDatabaseHealthChecker creates a new database health checker
func NewDatabaseHealthChecker(db *sql.DB, name string, logger *zap.Logger) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{
		db:     db,
		name:   name,
		logger: logger,
	}
}

func (d *DatabaseHealthChecker) GetName() string {
	return d.name
}

func (d *DatabaseHealthChecker) Check(ctx context.Context) HealthStatus {
	start := time.Now()
	
	// Set a timeout for the database check
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	// Ping the database
	err := d.db.PingContext(checkCtx)
	duration := time.Since(start)
	
	status := HealthStatus{
		Name:      d.name,
		Timestamp: time.Now(),
		Duration:  duration,
		Details:   make(map[string]interface{}),
	}
	
	if err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("Database connection failed: %v", err)
		d.logger.Error("Database health check failed", zap.Error(err))
	} else {
		status.Status = "healthy"
		status.Message = "Database connection successful"
		
		// Get additional database stats
		stats := d.db.Stats()
		status.Details["open_connections"] = stats.OpenConnections
		status.Details["in_use"] = stats.InUse
		status.Details["idle"] = stats.Idle
		status.Details["wait_count"] = stats.WaitCount
		status.Details["wait_duration"] = stats.WaitDuration.String()
	}
	
	return status
}

// RedisHealthChecker checks Redis connectivity
type RedisHealthChecker struct {
	client *redis.Client
	name   string
	logger *zap.Logger
}

// NewRedisHealthChecker creates a new Redis health checker
func NewRedisHealthChecker(client *redis.Client, name string, logger *zap.Logger) *RedisHealthChecker {
	return &RedisHealthChecker{
		client: client,
		name:   name,
		logger: logger,
	}
}

func (r *RedisHealthChecker) GetName() string {
	return r.name
}

func (r *RedisHealthChecker) Check(ctx context.Context) HealthStatus {
	start := time.Now()
	
	// Set a timeout for the Redis check
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	// Ping Redis
	err := r.client.Ping(checkCtx).Err()
	duration := time.Since(start)
	
	status := HealthStatus{
		Name:      r.name,
		Timestamp: time.Now(),
		Duration:  duration,
		Details:   make(map[string]interface{}),
	}
	
	if err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("Redis connection failed: %v", err)
		r.logger.Error("Redis health check failed", zap.Error(err))
	} else {
		status.Status = "healthy"
		status.Message = "Redis connection successful"
		
		// Get Redis info
		info, err := r.client.Info(checkCtx).Result()
		if err == nil {
			status.Details["info"] = info
		}
	}
	
	return status
}

// HTTPHealthChecker checks external HTTP service
type HTTPHealthChecker struct {
	url    string
	name   string
	client *http.Client
	logger *zap.Logger
}

// NewHTTPHealthChecker creates a new HTTP health checker
func NewHTTPHealthChecker(url, name string, logger *zap.Logger) *HTTPHealthChecker {
	return &HTTPHealthChecker{
		url:    url,
		name:   name,
		client: &http.Client{Timeout: 5 * time.Second},
		logger: logger,
	}
}

func (h *HTTPHealthChecker) GetName() string {
	return h.name
}

func (h *HTTPHealthChecker) Check(ctx context.Context) HealthStatus {
	start := time.Now()
	
	// Make HTTP request
	resp, err := h.client.Get(h.url)
	duration := time.Since(start)
	
	status := HealthStatus{
		Name:      h.name,
		Timestamp: time.Now(),
		Duration:  duration,
		Details:   make(map[string]interface{}),
	}
	
	if err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("HTTP check failed: %v", err)
		h.logger.Error("HTTP health check failed", zap.Error(err))
	} else {
		defer resp.Body.Close()
		
		status.Details["status_code"] = resp.StatusCode
		status.Details["response_time"] = duration.String()
		
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			status.Status = "healthy"
			status.Message = "HTTP service is responding"
		} else {
			status.Status = "unhealthy"
			status.Message = fmt.Sprintf("HTTP service returned status %d", resp.StatusCode)
		}
	}
	
	return status
}

// HealthManager manages all health checks
type HealthManager struct {
	checkers []HealthChecker
	startTime time.Time
	version   string
	logger    *zap.Logger
}

// NewHealthManager creates a new health manager
func NewHealthManager(version string, logger *zap.Logger) *HealthManager {
	return &HealthManager{
		checkers:  make([]HealthChecker, 0),
		startTime: time.Now(),
		version:   version,
		logger:    logger,
	}
}

// AddChecker adds a health checker
func (h *HealthManager) AddChecker(checker HealthChecker) {
	h.checkers = append(h.checkers, checker)
}

// CheckAll performs all health checks
func (h *HealthManager) CheckAll(ctx context.Context) HealthResponse {
	start := time.Now()
	checks := make(map[string]HealthStatus)
	
	// Run all checks concurrently
	results := make(chan HealthStatus, len(h.checkers))
	
	for _, checker := range h.checkers {
		go func(c HealthChecker) {
			results <- c.Check(ctx)
		}(checker)
	}
	
	// Collect results
	for i := 0; i < len(h.checkers); i++ {
		status := <-results
		checks[status.Name] = status
	}
	
	duration := time.Since(start)
	
	// Determine overall status
	overallStatus := "healthy"
	for _, status := range checks {
		if status.Status == "unhealthy" {
			overallStatus = "unhealthy"
			break
		} else if status.Status == "degraded" && overallStatus == "healthy" {
			overallStatus = "degraded"
		}
	}
	
	return HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Duration:  duration,
		Checks:    checks,
		Version:   h.version,
		Uptime:    time.Since(h.startTime),
	}
}

// CheckLiveness performs a basic liveness check
func (h *HealthManager) CheckLiveness(ctx context.Context) HealthResponse {
	return HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Duration:  0,
		Checks:    make(map[string]HealthStatus),
		Version:   h.version,
		Uptime:    time.Since(h.startTime),
	}
}

// SetupHealthRoutes sets up health check routes
func SetupHealthRoutes(r chi.Router, manager *HealthManager) {
	r.Get("/health/live", func(w http.ResponseWriter, r *http.Request) {
		response := manager.CheckLiveness(r.Context())
		
		statusCode := http.StatusOK
		if response.Status != "healthy" {
			statusCode = http.StatusServiceUnavailable
		}
		
		render.Status(r, statusCode)
		render.JSON(w, r, response)
	})
	
	r.Get("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		response := manager.CheckAll(r.Context())
		
		statusCode := http.StatusOK
		if response.Status == "unhealthy" {
			statusCode = http.StatusServiceUnavailable
		} else if response.Status == "degraded" {
			statusCode = http.StatusOK // Still OK but degraded
		}
		
		render.Status(r, statusCode)
		render.JSON(w, r, response)
	})
	
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response := manager.CheckAll(r.Context())
		
		statusCode := http.StatusOK
		if response.Status == "unhealthy" {
			statusCode = http.StatusServiceUnavailable
		} else if response.Status == "degraded" {
			statusCode = http.StatusOK // Still OK but degraded
		}
		
		render.Status(r, statusCode)
		render.JSON(w, r, response)
	})
}
