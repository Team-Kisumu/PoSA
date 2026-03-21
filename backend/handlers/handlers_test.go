package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHealth verifies the health endpoint returns a 200 JSON response
// with success=true and status="ok".
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

// TestVerify tests the verify endpoint with valid CIDs, empty CIDs,
// injection attempts, and malformed CID formats.
func TestVerify(t *testing.T) {
	// Use a real mux to test path parameter extraction for {cid}.
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
		// Call Verify directly (mux would 404 on empty path segment).
		req := httptest.NewRequest(http.MethodGet, "/api/verify/", nil)
		w := httptest.NewRecorder()
		Verify(w, req)

		assertStatus(t, w.Code, http.StatusBadRequest)
		resp := decodeResponse(t, w)
		assertFailure(t, resp, "INVALID_CID")
	})

	t.Run("cid with injection chars", func(t *testing.T) {
		// Semicolon in CID — potential command injection attempt.
		req := httptest.NewRequest(http.MethodGet, "/api/verify/abc;rm+-rf", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assertStatus(t, w.Code, http.StatusBadRequest)
		resp := decodeResponse(t, w)
		assertFailure(t, resp, "INVALID_CID")
	})

	t.Run("cid too short", func(t *testing.T) {
		// CID below the 46-character minimum length.
		req := httptest.NewRequest(http.MethodGet, "/api/verify/abc", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assertStatus(t, w.Code, http.StatusBadRequest)
		resp := decodeResponse(t, w)
		assertFailure(t, resp, "INVALID_CID")
	})
}

// TestWriteJSON verifies the JSON serializer sets the correct Content-Type,
// status code, and produces a decodable APIResponse envelope.
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
// Shared assertion functions used across all handler test files.

// assertStatus checks that the HTTP status code matches the expected value.
func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
}

// assertContentType checks that the response Content-Type is application/json.
func assertContentType(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want \"application/json\"", ct)
	}
}

// decodeResponse parses the response body into an APIResponse struct.
// Fails the test immediately if JSON decoding fails.
func decodeResponse(t *testing.T, w *httptest.ResponseRecorder) APIResponse {
	t.Helper()
	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return resp
}

// assertSuccess verifies the response envelope indicates success.
func assertSuccess(t *testing.T, resp APIResponse) {
	t.Helper()
	if !resp.Success {
		t.Errorf("expected success=true, got error: %+v", resp.Error)
	}
}

// assertFailure verifies the response envelope indicates failure with
// the expected error code.
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
