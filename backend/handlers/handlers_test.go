package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	Health(w, req)

	assertStatus(t, w.Code, http.StatusOK)
	assertContentType(t, w)

	body := decodeBody(t, w)
	if body["status"] != "ok" {
		t.Errorf("expected status \"ok\", got %q", body["status"])
	}
}

func TestSubmit(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", nil)
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusNotImplemented)
	assertContentType(t, w)

	body := decodeBody(t, w)
	if body["message"] == "" {
		t.Error("expected non-empty message")
	}
}

func TestVerify(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/verify/{cid}", Verify)

	t.Run("valid cid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/verify/QmTestCid123", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assertStatus(t, w.Code, http.StatusNotImplemented)
		assertContentType(t, w)

		body := decodeBody(t, w)
		if body["cid"] != "QmTestCid123" {
			t.Errorf("expected cid \"QmTestCid123\", got %q", body["cid"])
		}
	})

	t.Run("empty cid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/verify/", nil)
		w := httptest.NewRecorder()

		Verify(w, req)

		assertStatus(t, w.Code, http.StatusBadRequest)

		body := decodeBody(t, w)
		if body["error"] == "" {
			t.Error("expected error message for empty cid")
		}
	})
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	writeJSON(w, http.StatusCreated, data)

	assertStatus(t, w.Code, http.StatusCreated)
	assertContentType(t, w)

	body := decodeBody(t, w)
	if body["key"] != "value" {
		t.Errorf("expected key \"value\", got %q", body["key"])
	}
}

func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
}

func assertContentType(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want \"application/json\"", ct)
	}
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return body
}
