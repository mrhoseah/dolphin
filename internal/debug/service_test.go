package debug

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestDebugger() *Debugger {
	return NewDebugger(Config{Enabled: true})
}

func TestStatsEndpoint(t *testing.T) {
	dbg := newTestDebugger()
	r := dbg.Router()
	if r == nil {
		t.Fatalf("router should not be nil when enabled")
	}

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct == "" {
		t.Fatalf("expected content type header to be set")
	}
}

func TestMemoryGC(t *testing.T) {
	dbg := newTestDebugger()
	r := dbg.Router()

	req := httptest.NewRequest(http.MethodGet, "/memory/gc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
