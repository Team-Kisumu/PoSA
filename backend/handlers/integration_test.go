package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Murzuqisah/PoSA/middleware"
)

func newTestMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", Health)
	mux.HandleFunc("POST /api/submit", Submit)
	mux.HandleFunc("GET /api/verify/{cid}", Verify)
	return mux
}

func newTestMuxWithMiddleware() http.Handler {
	return middleware.Recovery(
		middleware.RequestID(
			middleware.SecurityHeaders(newTestMux()),
		),
	)
}

func TestIntegrationHealthEndpoint(t *testing.T) {
	mux := newTestMux()

	t.Run("GET returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assertStatus(t, w.Code, http.StatusOK)
		resp := decodeResponse(t, w)
		assertSuccess(t, resp)
	})

	t.Run("POST not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/health", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			t.Error("POST /health should not return 200")
		}
	})
}

func TestIntegrationSubmitEndpoint(t *testing.T) {
	mux := newTestMux()

	t.Run("POST file upload through mux", func(t *testing.T) {
		body, ct := createMultipartFile(t, "file", "test.py", "print('hello')\n")
		req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
		req.Header.Set("Content-Type", ct)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assertStatus(t, w.Code, http.StatusOK)
		resp := decodeResponse(t, w)
		assertSuccess(t, resp)
	})

	t.Run("POST repo link through mux", func(t *testing.T) {
		payload := `{"repo": "https://github.com/user/project"}`
		req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assertStatus(t, w.Code, http.StatusOK)
		resp := decodeResponse(t, w)
		assertSuccess(t, resp)
	})

	t.Run("GET not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/submit", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			t.Error("GET /api/submit should not return 200")
		}
	})
}

func TestIntegrationVerifyEndpoint(t *testing.T) {
	mux := newTestMux()

	t.Run("GET with valid CID through mux", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/verify/QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assertStatus(t, w.Code, http.StatusOK)
		resp := decodeResponse(t, w)
		assertSuccess(t, resp)
	})

	t.Run("POST not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/verify/QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			t.Error("POST /api/verify should not be handled")
		}
	})
}

func TestIntegrationUnknownRoute(t *testing.T) {
	mux := newTestMux()

	req := httptest.NewRequest(http.MethodGet, "/api/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Error("unknown route should not return 200")
	}
}

func TestIntegrationSecurityHeaders(t *testing.T) {
	handler := newTestMuxWithMiddleware()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assertStatus(t, w.Code, http.StatusOK)

	headers := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"Content-Security-Policy": "default-src 'none'",
		"Referrer-Policy":         "no-referrer",
		"Cache-Control":           "no-store",
	}
	for key, want := range headers {
		got := w.Header().Get(key)
		if got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestIntegrationRequestID(t *testing.T) {
	handler := newTestMuxWithMiddleware()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	rid := w.Header().Get("X-Request-ID")
	if rid == "" {
		t.Error("expected X-Request-ID header")
	}
	if len(rid) != 16 {
		t.Errorf("X-Request-ID length = %d, want 16 hex chars", len(rid))
	}
}

func TestIntegrationResponseEnvelope(t *testing.T) {
	handler := newTestMuxWithMiddleware()

	t.Run("success envelope", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var raw map[string]json.RawMessage
		json.NewDecoder(w.Body).Decode(&raw)

		if _, ok := raw["success"]; !ok {
			t.Error("response missing 'success' field")
		}
		if _, ok := raw["data"]; !ok {
			t.Error("success response missing 'data' field")
		}
	})

	t.Run("error envelope", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader("data"))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var raw map[string]json.RawMessage
		json.NewDecoder(w.Body).Decode(&raw)

		if _, ok := raw["success"]; !ok {
			t.Error("response missing 'success' field")
		}
		if _, ok := raw["error"]; !ok {
			t.Error("error response missing 'error' field")
		}
	})
}
