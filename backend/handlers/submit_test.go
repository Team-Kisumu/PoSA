package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- File Upload Tests ---

func TestSubmitFileUpload(t *testing.T) {
	body, contentType := createMultipartFile(t, "file", "main.go", "package main\n")
	req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusOK)
	assertContentType(t, w)

	var resp SubmitResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Type != "file" {
		t.Errorf("type = %q, want \"file\"", resp.Type)
	}
	if resp.Name != "main.go" {
		t.Errorf("name = %q, want \"main.go\"", resp.Name)
	}
	if resp.Size != len("package main\n") {
		t.Errorf("size = %d, want %d", resp.Size, len("package main\n"))
	}
	if resp.Message == "" {
		t.Error("expected non-empty message")
	}
}

func TestSubmitFileUploadMissingField(t *testing.T) {
	body, contentType := createMultipartFile(t, "wrong_field", "main.go", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertErrorField(t, w)
}

func TestSubmitFileUploadEmptyFile(t *testing.T) {
	body, contentType := createMultipartFile(t, "file", "empty.txt", "")
	req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertErrorField(t, w)
}

// --- Repo Link Tests ---

func TestSubmitRepoLink(t *testing.T) {
	payload := `{"repo": "https://github.com/user/project"}`
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusOK)
	assertContentType(t, w)

	var resp SubmitResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Type != "repo" {
		t.Errorf("type = %q, want \"repo\"", resp.Type)
	}
	if resp.Name != "https://github.com/user/project" {
		t.Errorf("name = %q, want repo URL", resp.Name)
	}
	if resp.Message == "" {
		t.Error("expected non-empty message")
	}
}

func TestSubmitRepoLinkEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(`{"repo": ""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertErrorField(t, w)
}

func TestSubmitRepoLinkWhitespace(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(`{"repo": "   "}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertErrorField(t, w)
}

func TestSubmitRepoLinkInvalidURL(t *testing.T) {
	cases := []struct {
		name string
		repo string
	}{
		{"gitlab", `{"repo": "https://gitlab.com/user/project"}`},
		{"http", `{"repo": "http://github.com/user/project"}`},
		{"bare string", `{"repo": "not-a-url"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(tc.repo))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			Submit(w, req)

			assertStatus(t, w.Code, http.StatusBadRequest)
			assertErrorField(t, w)
		})
	}
}

func TestSubmitInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(`{bad json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertErrorField(t, w)
}

func TestSubmitEmptyJSONBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertErrorField(t, w)
}

// --- Content-Type Tests ---

func TestSubmitUnsupportedContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader("plain text"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusUnsupportedMediaType)
	assertErrorField(t, w)
}

func TestSubmitMissingContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader("data"))
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertErrorField(t, w)
}

// --- Helpers ---

func createMultipartFile(t *testing.T, field, filename, content string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile(field, filename)
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte(content))
	writer.Close()
	return &buf, writer.FormDataContentType()
}

func assertErrorField(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	body := decodeBody(t, w)
	if body["error"] == "" {
		t.Error("expected non-empty 'error' field in response")
	}
}
