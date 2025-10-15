package debug

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Debugger provides debugging capabilities
type Debugger struct {
	enabled   bool
	profiler  *Profiler
	logger    *Logger
	tracer    *Tracer
	inspector *Inspector
	mu        sync.RWMutex
	requests  map[string]*RequestInfo
	stats     *Stats
}

// RequestInfo holds information about a request
type RequestInfo struct {
	ID         string
	Method     string
	URL        string
	Headers    map[string]string
	Body       string
	Response   *ResponseInfo
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	Status     int
	UserAgent  string
	RemoteAddr string
	Stack      []byte
}

// ResponseInfo holds response information
type ResponseInfo struct {
	Status      int
	Headers     map[string]string
	Body        string
	Size        int
	ContentType string
}

// Stats holds application statistics
type Stats struct {
	RequestsTotal   int64
	RequestsPerSec  float64
	AvgResponseTime time.Duration
	MaxResponseTime time.Duration
	MinResponseTime time.Duration
	ErrorCount      int64
	MemoryUsage     uint64
	GoroutineCount  int
	GCStats         debug.GCStats
	LastUpdated     time.Time
}

// Config holds debug configuration
type Config struct {
	Enabled         bool
	Port            int
	ProfilerPort    int
	LogLevel        string
	MaxRequests     int
	EnableProfiler  bool
	EnableTracer    bool
	EnableInspector bool
}

// NewDebugger creates a new debugger instance
func NewDebugger(config Config) *Debugger {
	if config.Port == 0 {
		config.Port = 8082
	}
	if config.ProfilerPort == 0 {
		config.ProfilerPort = 8083
	}
	if config.MaxRequests == 0 {
		config.MaxRequests = 1000
	}

	d := &Debugger{
		enabled:  config.Enabled,
		requests: make(map[string]*RequestInfo),
		stats:    &Stats{},
	}

	if config.EnableProfiler {
		d.profiler = NewProfiler(config.ProfilerPort)
	}
	if config.EnableTracer {
		d.tracer = NewTracer()
	}
	if config.EnableInspector {
		d.inspector = NewInspector()
	}

	d.logger = NewLogger(config.LogLevel)

	return d
}

// Middleware returns the debug middleware
func (d *Debugger) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !d.enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Create request info
			reqID := middleware.GetReqID(r.Context())
			if reqID == "" {
				reqID = fmt.Sprintf("%d", time.Now().UnixNano())
			}

			reqInfo := &RequestInfo{
				ID:         reqID,
				Method:     r.Method,
				URL:        r.URL.String(),
				Headers:    make(map[string]string),
				StartTime:  time.Now(),
				UserAgent:  r.UserAgent(),
				RemoteAddr: r.RemoteAddr,
			}

			// Copy headers
			for name, values := range r.Header {
				reqInfo.Headers[name] = strings.Join(values, ", ")
			}

			// Capture stack trace for debugging
			if d.enabled {
				reqInfo.Stack = debug.Stack()
			}

			// Wrap response writer to capture response
			wrapped := &responseWriter{
				ResponseWriter: w,
				status:         200,
				body:           &strings.Builder{},
			}

			// Execute request
			next.ServeHTTP(wrapped, r)

			// Finalize request info
			reqInfo.EndTime = time.Now()
			reqInfo.Duration = reqInfo.EndTime.Sub(reqInfo.StartTime)
			reqInfo.Status = wrapped.status
			reqInfo.Response = &ResponseInfo{
				Status:      wrapped.status,
				Body:        wrapped.body.String(),
				Size:        wrapped.body.Len(),
				ContentType: w.Header().Get("Content-Type"),
				Headers:     make(map[string]string),
			}

			// Copy response headers
			for name, values := range w.Header() {
				reqInfo.Response.Headers[name] = strings.Join(values, ", ")
			}

			// Store request info
			d.mu.Lock()
			d.requests[reqID] = reqInfo

			// Clean up old requests
			if len(d.requests) > int(d.stats.RequestsTotal) {
				d.cleanupOldRequests()
			}

			d.updateStats(reqInfo)
			d.mu.Unlock()

			// Log request
			d.logger.LogRequest(reqInfo)
		})
	}
}

// Router returns debug routes
func (d *Debugger) Router() chi.Router {
	if !d.enabled {
		return nil
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Debug dashboard
	r.Get("/", d.dashboard)

	// Request information
	r.Get("/requests", d.listRequests)
	r.Get("/requests/{id}", d.getRequest)

	// Statistics
	r.Get("/stats", d.getStats)
	r.Get("/stats/reset", d.resetStats)

	// Memory information
	r.Get("/memory", d.getMemoryInfo)
	r.Get("/memory/gc", d.forceGC)

	// Goroutine information
	r.Get("/goroutines", d.getGoroutines)

	// Profiling
	if d.profiler != nil {
		r.Get("/profile/cpu", d.cpuProfile)
		r.Get("/profile/memory", d.memoryProfile)
		r.Get("/profile/goroutine", d.goroutineProfile)
		r.Get("/profile/block", d.blockProfile)
	}

	// Tracing
	if d.tracer != nil {
		r.Get("/trace", d.getTrace)
		r.Post("/trace/start", d.startTrace)
		r.Post("/trace/stop", d.stopTrace)
	}

	// Inspector
	if d.inspector != nil {
		r.Get("/inspect", d.inspect)
		r.Get("/inspect/{type}", d.inspectType)
	}

	return r
}

// dashboard serves the debug dashboard
func (d *Debugger) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dolphin Debug Dashboard</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
        }
        .card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .card h3 {
            margin-top: 0;
            color: #333;
        }
        .stat {
            display: flex;
            justify-content: space-between;
            margin: 10px 0;
            padding: 10px;
            background: #f8f9fa;
            border-radius: 4px;
        }
        .stat-label {
            font-weight: 500;
        }
        .stat-value {
            font-weight: bold;
            color: #667eea;
        }
        .btn {
            display: inline-block;
            padding: 10px 20px;
            background: #667eea;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            margin: 5px;
        }
        .btn:hover {
            background: #5a6fd8;
        }
        .status {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: bold;
        }
        .status-success {
            background: #d4edda;
            color: #155724;
        }
        .status-error {
            background: #f8d7da;
            color: #721c24;
        }
        .status-warning {
            background: #fff3cd;
            color: #856404;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🐬 Dolphin Debug Dashboard</h1>
            <p>Real-time debugging and monitoring for your Dolphin application</p>
        </div>
        
        <div class="grid">
            <div class="card">
                <h3>📊 Statistics</h3>
                <div class="stat">
                    <span class="stat-label">Total Requests:</span>
                    <span class="stat-value" id="total-requests">-</span>
                </div>
                <div class="stat">
                    <span class="stat-label">Requests/sec:</span>
                    <span class="stat-value" id="requests-per-sec">-</span>
                </div>
                <div class="stat">
                    <span class="stat-label">Avg Response Time:</span>
                    <span class="stat-value" id="avg-response-time">-</span>
                </div>
                <div class="stat">
                    <span class="stat-label">Error Count:</span>
                    <span class="stat-value" id="error-count">-</span>
                </div>
                <a href="/debug/stats" class="btn">View Details</a>
            </div>
            
            <div class="card">
                <h3>💾 Memory</h3>
                <div class="stat">
                    <span class="stat-label">Memory Usage:</span>
                    <span class="stat-value" id="memory-usage">-</span>
                </div>
                <div class="stat">
                    <span class="stat-label">Goroutines:</span>
                    <span class="stat-value" id="goroutines">-</span>
                </div>
                <div class="stat">
                    <span class="stat-label">GC Runs:</span>
                    <span class="stat-value" id="gc-runs">-</span>
                </div>
                <a href="/debug/memory" class="btn">Memory Details</a>
                <a href="/debug/memory/gc" class="btn">Force GC</a>
            </div>
            
            <div class="card">
                <h3>🔍 Requests</h3>
                <div class="stat">
                    <span class="stat-label">Recent Requests:</span>
                    <span class="stat-value" id="recent-requests">-</span>
                </div>
                <a href="/debug/requests" class="btn">View All</a>
            </div>
            
            <div class="card">
                <h3>📈 Profiling</h3>
                <p>CPU and memory profiling tools</p>
                <a href="/debug/profile/cpu" class="btn">CPU Profile</a>
                <a href="/debug/profile/memory" class="btn">Memory Profile</a>
                <a href="/debug/profile/goroutine" class="btn">Goroutine Profile</a>
            </div>
            
            <div class="card">
                <h3>🔧 Inspector</h3>
                <p>Application inspection tools</p>
                <a href="/debug/inspect" class="btn">Inspect App</a>
                <a href="/debug/trace" class="btn">Trace</a>
            </div>
        </div>
    </div>
    
    <script>
        // Auto-refresh stats every 5 seconds
        function updateStats() {
            fetch('/debug/stats')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('total-requests').textContent = data.requests_total || 0;
                    document.getElementById('requests-per-sec').textContent = (data.requests_per_sec || 0).toFixed(2);
                    document.getElementById('avg-response-time').textContent = data.avg_response_time || '0ms';
                    document.getElementById('error-count').textContent = data.error_count || 0;
                    document.getElementById('memory-usage').textContent = data.memory_usage || '0MB';
                    document.getElementById('goroutines').textContent = data.goroutine_count || 0;
                    document.getElementById('gc-runs').textContent = data.gc_stats.num_gc || 0;
                })
                .catch(error => console.error('Error updating stats:', error));
        }
        
        // Update stats on load and every 5 seconds
        updateStats();
        setInterval(updateStats, 5000);
    </script>
</body>
</html>`

	w.Write([]byte(html))
}

// listRequests lists all requests
func (d *Debugger) listRequests(w http.ResponseWriter, r *http.Request) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	requests := make([]*RequestInfo, 0, len(d.requests))
	for _, req := range d.requests {
		requests = append(requests, req)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// getRequest gets a specific request
func (d *Debugger) getRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	d.mu.RLock()
	req, exists := d.requests[id]
	d.mu.RUnlock()

	if !exists {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

// getStats returns current statistics
func (d *Debugger) getStats(w http.ResponseWriter, r *http.Request) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Update runtime stats
	d.updateRuntimeStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d.stats)
}

// resetStats resets statistics
func (d *Debugger) resetStats(w http.ResponseWriter, r *http.Request) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.stats = &Stats{}
	d.requests = make(map[string]*RequestInfo)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Stats reset"})
}

// getMemoryInfo returns memory information
func (d *Debugger) getMemoryInfo(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	info := map[string]interface{}{
		"alloc":          m.Alloc,
		"total_alloc":    m.TotalAlloc,
		"sys":            m.Sys,
		"lookups":        m.Lookups,
		"mallocs":        m.Mallocs,
		"frees":          m.Frees,
		"heap_alloc":     m.HeapAlloc,
		"heap_sys":       m.HeapSys,
		"heap_idle":      m.HeapIdle,
		"heap_inuse":     m.HeapInuse,
		"heap_released":  m.HeapReleased,
		"heap_objects":   m.HeapObjects,
		"stack_inuse":    m.StackInuse,
		"stack_sys":      m.StackSys,
		"gc_cycles":      m.NumGC,
		"gc_pause_total": m.PauseTotalNs,
		"gc_pause_ns":    m.PauseNs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// forceGC forces garbage collection
func (d *Debugger) forceGC(w http.ResponseWriter, r *http.Request) {
	runtime.GC()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "GC completed",
		"memory":  m.Alloc,
	})
}

// getGoroutines returns goroutine information
func (d *Debugger) getGoroutines(w http.ResponseWriter, r *http.Request) {
	profile := pprof.Lookup("goroutine")
	if profile == nil {
		http.Error(w, "Goroutine profile not available", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	profile.WriteTo(w, 1)
}

// Helper methods

func (d *Debugger) cleanupOldRequests() {
	if len(d.requests) <= 1000 {
		return
	}

	// Remove oldest requests
	count := 0
	for id := range d.requests {
		if count >= len(d.requests)-1000 {
			break
		}
		delete(d.requests, id)
		count++
	}
}

func (d *Debugger) updateStats(reqInfo *RequestInfo) {
	d.stats.RequestsTotal++

	if reqInfo.Status >= 400 {
		d.stats.ErrorCount++
	}

	// Update response time stats
	if d.stats.MinResponseTime == 0 || reqInfo.Duration < d.stats.MinResponseTime {
		d.stats.MinResponseTime = reqInfo.Duration
	}
	if reqInfo.Duration > d.stats.MaxResponseTime {
		d.stats.MaxResponseTime = reqInfo.Duration
	}

	// Calculate average response time
	totalDuration := d.stats.AvgResponseTime * time.Duration(d.stats.RequestsTotal-1)
	d.stats.AvgResponseTime = (totalDuration + reqInfo.Duration) / time.Duration(d.stats.RequestsTotal)

	d.stats.LastUpdated = time.Now()
}

func (d *Debugger) updateRuntimeStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	d.stats.MemoryUsage = m.Alloc
	d.stats.GoroutineCount = runtime.NumGoroutine()
	d.stats.GCStats = debug.GCStats{
		NumGC: int64(m.NumGC), Pause: []time.Duration{time.Duration(m.PauseNs[0])},
		PauseEnd:       []time.Time{time.Now().Add(time.Duration(m.PauseEnd[0]))},
		PauseTotal:     time.Duration(m.PauseTotalNs),
		PauseQuantiles: nil,
	}
}

// responseWriter wraps http.ResponseWriter to capture response data
type responseWriter struct {
	http.ResponseWriter
	status int
	body   *strings.Builder
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}
