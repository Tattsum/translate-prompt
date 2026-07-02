package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCorsMiddleware_AllowAll(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodOptions, "/query", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

	handler := corsMiddleware([]string{"*"}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want *", got)
	}
}

func TestCorsMiddleware_AllowedOrigin(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodOptions, "/query", nil)
	origin := "https://prompt.tattsum.com"
	req.Header.Set("Origin", origin)
	rec := httptest.NewRecorder()

	handler := corsMiddleware([]string{origin}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, origin)
	}
}

func TestCorsMiddleware_DisallowedOriginPreflight(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodOptions, "/query", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()

	handler := corsMiddleware([]string{"https://prompt.tattsum.com"}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

func TestIsOriginAllowed(t *testing.T) {
	t.Parallel()
	allowed := []string{"https://prompt.tattsum.com"}
	if !isOriginAllowed("https://prompt.tattsum.com", allowed) {
		t.Fatal("expected allowed origin")
	}
	if isOriginAllowed("https://evil.example", allowed) {
		t.Fatal("expected disallowed origin")
	}
	if !isOriginAllowed("http://localhost:5173", []string{"*"}) {
		t.Fatal("expected wildcard allow")
	}
}
