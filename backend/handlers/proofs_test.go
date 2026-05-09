package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Murzuqisah/PoSA/auth"
	"github.com/Murzuqisah/PoSA/database"
)

// --- QueryInt Helper Tests ---

func TestProofsListEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/proofs?limit=10&offset=0", nil)
	if got := queryInt(req, "limit", 20); got != 10 {
		t.Errorf("queryInt(limit) = %d, want 10", got)
	}
	if got := queryInt(req, "offset", 0); got != 0 {
		t.Errorf("queryInt(offset) = %d, want 0", got)
	}
	if got := queryInt(req, "missing", 42); got != 42 {
		t.Errorf("queryInt(missing) = %d, want 42", got)
	}
}

func TestQueryIntNegative(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/proofs?limit=-5", nil)
	if got := queryInt(req, "limit", 20); got != 20 {
		t.Errorf("queryInt(negative) = %d, want 20 (default)", got)
	}
}

func TestQueryIntInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/proofs?limit=abc", nil)
	if got := queryInt(req, "limit", 20); got != 20 {
		t.Errorf("queryInt(invalid) = %d, want 20 (default)", got)
	}
}

// --- Submissions Handler Tests ---

func TestSubmissionsCreateNoAuth(t *testing.T) {
	handler := &SubmissionsHandler{DB: nil}
	body := `{"type":"file","name":"test.go","score":85}`
	req := httptest.NewRequest(http.MethodPost, "/api/submissions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestSubmissionsCreateInvalidJSON(t *testing.T) {
	handler := &SubmissionsHandler{DB: nil}
	req := httptest.NewRequest(http.MethodPost, "/api/submissions", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	req = auth.SetUserForTest(req, &database.User{ID: 1, Username: "test", Role: "user"})
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error == nil || resp.Error.Code != "INVALID_JSON" {
		t.Errorf("error code = %v, want INVALID_JSON", resp.Error)
	}
}

func TestSubmissionsCreateMissingName(t *testing.T) {
	handler := &SubmissionsHandler{DB: nil}
	body := `{"type":"file","name":"","score":85}`
	req := httptest.NewRequest(http.MethodPost, "/api/submissions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = auth.SetUserForTest(req, &database.User{ID: 1, Username: "test", Role: "user"})
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestSubmissionsCreateInvalidScore(t *testing.T) {
	handler := &SubmissionsHandler{DB: nil}
	body := `{"type":"file","name":"test.go","score":150}`
	req := httptest.NewRequest(http.MethodPost, "/api/submissions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = auth.SetUserForTest(req, &database.User{ID: 1, Username: "test", Role: "user"})
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestSubmissionsCreateNegativeScore(t *testing.T) {
	handler := &SubmissionsHandler{DB: nil}
	body := `{"type":"file","name":"test.go","score":-1}`
	req := httptest.NewRequest(http.MethodPost, "/api/submissions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = auth.SetUserForTest(req, &database.User{ID: 1, Username: "test", Role: "user"})
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
