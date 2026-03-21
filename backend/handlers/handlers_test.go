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

	resp := decodeResponse(t, w)
	assertSuccess(t, resp)

	data := resp.Data.(map[string]any)
	if data["status"] != "ok" {
		t.Errorf("status = %q, want \"ok\"", data["status"])
	}
}

func TestVerify(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/verify/{cid}", Verify)

	t.Run("valid cid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/verify/QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assertStatus(t, w.Code, http.StatusOK)
		resp := decodeResponse(t, w)
		assertSuccess(t, resp)
	})

	t.Run("empty cid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/verify/", nil)
		w := httptest.NewRecorder()
		Verify(w, req)

		assertStatus(t, w.Code, http.StatusBadRequest)
		resp := decodeResponse(t, w)
		assertFailure(t, resp, "INVALID_CID")
	})

	t.Run("cid with injection chars", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/verify/abc;rm+-rf", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assertStatus(t, w.Code, http.StatusBadRequest)
		resp := decodeResponse(t, w)
		assertFailure(t, resp, "INVALID_CID")
	})

	t.Run("cid too short", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/verify/abc", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assertStatus(t, w.Code, http.StatusBadRequest)
		resp := decodeResponse(t, w)
		assertFailure(t, resp, "INVALID_CID")
	})
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	resp := APIResponse{Success: true, Data: HealthData{Status: "ok"}}

	writeJSON(w, http.StatusCreated, resp)

	assertStatus(t, w.Code, http.StatusCreated)
	assertContentType(t, w)

	decoded := decodeResponse(t, w)
	if !decoded.Success {
		t.Error("expected success=true")
	}
}

// --- Test Helpers ---

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

func decodeResponse(t *testing.T, w *httptest.ResponseRecorder) APIResponse {
	t.Helper()
	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return resp
}

func assertSuccess(t *testing.T, resp APIResponse) {
	t.Helper()
	if !resp.Success {
		t.Errorf("expected success=true, got error: %+v", resp.Error)
	}
}

func assertFailure(t *testing.T, resp APIResponse, code string) {
	t.Helper()
	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Error == nil {
		t.Fatal("expected error object, got nil")
	}
	if resp.Error.Code != code {
		t.Errorf("error code = %q, want %q", resp.Error.Code, code)
	}
}
