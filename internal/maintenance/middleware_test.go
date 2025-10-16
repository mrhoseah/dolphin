package maintenance

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMaintenanceMiddleware_Disabled(t *testing.T) {
	m := NewManager("testdata/maintenance.json")
	_ = m.Disable() // ensure disabled
	mw := NewMiddleware(m)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	mw.Handle(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 when disabled, got %d", w.Code)
	}
}

func TestMaintenanceMiddleware_Enabled(t *testing.T) {
	m := NewManager("testdata/maintenance.json")
	_ = m.Enable("maintenance", 60, nil, "")
	defer m.Disable()

	mw := NewMiddleware(m)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	mw.Handle(next).ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when enabled, got %d", w.Code)
	}
}
