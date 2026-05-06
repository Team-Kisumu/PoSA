package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSecurityHeaders verifies that the SecurityHeaders middleware injects
// all expected defensive headers with the correct values.
func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/submit", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Headers that must be present on every response.
	required := []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
		"Content-Security-Policy",
		"Referrer-Policy",
		"Strict-Transport-Security",
		"Permissions-Policy",
		"Cross-Origin-Opener-Policy",
		"Cross-Origin-Resource-Policy",
	}
	for _, key := range required {
		if w.Header().Get(key) == "" {
			t.Errorf("%s header missing", key)
		}
	}

	// Specific value checks.
	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want \"nosniff\"", got)
	}
	if got := w.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("X-Frame-Options = %q, want \"DENY\"", got)
	}

	// API paths should have no-cache headers.
	if got := w.Header().Get("Cache-Control"); got == "" {
		t.Error("Cache-Control missing on API path")
	}
}

// TestRequestID verifies that the RequestID middleware generates a 16-char
// hex ID and that consecutive requests produce unique IDs.
func TestRequestID(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	rid := w.Header().Get("X-Request-ID")
	if rid == "" {
		t.Error("expected X-Request-ID header")
	}
	if len(rid) != 16 {
		t.Errorf("X-Request-ID length = %d, want 16", len(rid))
	}

	// Verify uniqueness: a second request should produce a different ID.
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	if w2.Header().Get("X-Request-ID") == rid {
		t.Error("request IDs should be unique")
	}
}

// TestRecovery verifies that the Recovery middleware catches panics and
// returns a 500 JSON error response instead of crashing the server.
func TestRecovery(t *testing.T) {
	handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

// TestRecoveryNoPanic verifies that the Recovery middleware passes through
// normally when no panic occurs.
func TestRecoveryNoPanic(t *testing.T) {
	handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}
