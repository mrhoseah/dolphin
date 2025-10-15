package debug

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// Profiler handles profiling operations
type Profiler struct {
	port int
}

// Logger handles debug logging
type Logger struct {
	level string
}

// Tracer handles request tracing
type Tracer struct {
	enabled bool
}

// Inspector handles application inspection
type Inspector struct {
	enabled bool
}

// NewProfiler creates a new profiler
func NewProfiler(port int) *Profiler {
	return &Profiler{port: port}
}

// NewLogger creates a new logger
func NewLogger(level string) *Logger {
	return &Logger{level: level}
}

// NewTracer creates a new tracer
func NewTracer() *Tracer {
	return &Tracer{enabled: true}
}

// NewInspector creates a new inspector
func NewInspector() *Inspector {
	return &Inspector{enabled: true}
}

// LogRequest logs request information
func (l *Logger) LogRequest(req *RequestInfo) {
	if l.level == "debug" {
		log.Printf("Request: %s %s - %d (%v)", req.Method, req.URL, req.Status, req.Duration)
	}
}

// CPU profiling methods
func (d *Debugger) cpuProfile(w http.ResponseWriter, r *http.Request) {
	seconds := 30
	if s := r.URL.Query().Get("seconds"); s != "" {
		if sec, err := strconv.Atoi(s); err == nil && sec > 0 && sec <= 300 {
			seconds = sec
		}
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=cpu.prof")

	if err := pprof.StartCPUProfile(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer pprof.StopCPUProfile()

	time.Sleep(time.Duration(seconds) * time.Second)
}

func (d *Debugger) memoryProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=memory.prof")

	if err := pprof.WriteHeapProfile(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (d *Debugger) goroutineProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=goroutine.prof")

	profile := pprof.Lookup("goroutine")
	if profile == nil {
		http.Error(w, "Goroutine profile not available", http.StatusInternalServerError)
		return
	}

	profile.WriteTo(w, 1)
}

func (d *Debugger) blockProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=block.prof")

	profile := pprof.Lookup("block")
	if profile == nil {
		http.Error(w, "Block profile not available", http.StatusInternalServerError)
		return
	}

	profile.WriteTo(w, 1)
}

// Tracing methods
func (d *Debugger) getTrace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Tracing not implemented yet"}`))
}

func (d *Debugger) startTrace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Trace started"}`))
}

func (d *Debugger) stopTrace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Trace stopped"}`))
}

// Inspector methods
func (d *Debugger) inspect(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"goroutines": runtime.NumGoroutine(),
		"version":    runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"cpu_count":  runtime.NumCPU(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (d *Debugger) inspectType(w http.ResponseWriter, r *http.Request) {
	inspectType := chi.URLParam(r, "type")

	var info map[string]interface{}
	switch inspectType {
	case "memory":
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		info = map[string]interface{}{
			"alloc":   m.Alloc,
			"total":   m.TotalAlloc,
			"sys":     m.Sys,
			"heap":    m.HeapAlloc,
			"objects": m.HeapObjects,
		}
	case "goroutines":
		info = map[string]interface{}{
			"count": runtime.NumGoroutine(),
		}
	default:
		info = map[string]interface{}{
			"error": "Unknown inspection type",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
