package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", Health)
	mux.HandleFunc("POST /api/submit", Submit)
	mux.HandleFunc("GET /api/verify/{cid}", Verify)
	return mux
}

func TestIntegrationHealthEndpoint(t *testing.T) {
	mux := newTestMux()

	t.Run("GET returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assertStatus(t, w.Code, http.StatusOK)
		body := decodeBody(t, w)
		if body["status"] != "ok" {
			t.Errorf("status = %q, want \"ok\"", body["status"])
		}
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
		var resp SubmitResponse
		json.NewDecoder(w.Body).Decode(&resp)
		if resp.Type != "file" || resp.Name != "test.py" {
			t.Errorf("unexpected response: %+v", resp)
		}
	})

	t.Run("POST repo link through mux", func(t *testing.T) {
		payload := `{"repo": "https://github.com/user/project"}`
		req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assertStatus(t, w.Code, http.StatusOK)
		var resp SubmitResponse
		json.NewDecoder(w.Body).Decode(&resp)
		if resp.Type != "repo" {
			t.Errorf("type = %q, want \"repo\"", resp.Type)
		}
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
		req := httptest.NewRequest(http.MethodGet, "/api/verify/QmABC123", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assertStatus(t, w.Code, http.StatusNotImplemented)
		body := decodeBody(t, w)
		if body["cid"] != "QmABC123" {
			t.Errorf("cid = %q, want \"QmABC123\"", body["cid"])
		}
	})

	t.Run("POST not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/verify/QmABC123", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code == http.StatusNotImplemented {
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
