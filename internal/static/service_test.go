package static

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRenderMissingTemplate(t *testing.T) {
	s := NewService(Config{BaseDir: "testdata/static"})
	_ = s.LoadTemplates()
	if _, err := s.Render("does-not-exist", PageData{}); err == nil {
		t.Fatalf("expected error for missing template")
	}
}

func TestServeStaticNotFound(t *testing.T) {
	s := NewService(Config{BaseDir: "testdata/static"})
	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	w := httptest.NewRecorder()
	if err := s.ServeStaticFile(w, req, "missing.html"); err == nil {
		t.Fatalf("expected not found error")
	}
}
