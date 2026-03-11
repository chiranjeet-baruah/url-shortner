package driver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestHealthEndpoint(t *testing.T) {
	// Use an in-memory FS so the static file handler doesn't panic.
	// The health route is registered before the catch-all, so no file is needed.
	stubFS := fstest.MapFS{
		"index.html": {Data: []byte("<html></html>")},
	}

	router := NewRouter(nil, stubFS)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	want := `{"status":"ok"}`
	if got := rec.Body.String(); got != want {
		t.Fatalf("expected body %q, got %q", want, got)
	}
}
