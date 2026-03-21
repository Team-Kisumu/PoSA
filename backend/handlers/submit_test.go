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
	resp := decodeResponse(t, w)
	assertSuccess(t, resp)

	data, _ := json.Marshal(resp.Data)
	var sr SubmitResponse
	json.Unmarshal(data, &sr)

	if sr.Type != "file" {
		t.Errorf("type = %q, want \"file\"", sr.Type)
	}
	if sr.Name != "main.go" {
		t.Errorf("name = %q, want \"main.go\"", sr.Name)
	}
	if sr.Size != int64(len("package main\n")) {
		t.Errorf("size = %d, want %d", sr.Size, len("package main\n"))
	}
}

func TestSubmitFileUploadMissingField(t *testing.T) {
	body, contentType := createMultipartFile(t, "wrong_field", "main.go", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "MISSING_FILE")
}

func TestSubmitFileUploadEmptyFile(t *testing.T) {
	body, contentType := createMultipartFile(t, "file", "empty.txt", "")
	req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "EMPTY_FILE")
}

func TestSubmitFileUploadPathTraversal(t *testing.T) {
	// filepath.Base strips traversal, so ../../etc/passwd becomes "passwd" (safe)
	// The validator catches backslash-based traversal and null bytes at the
	// ValidateFilename level (tested in types_test.go)
	t.Run("backslash traversal", func(t *testing.T) {
		body, contentType := createMultipartFile(t, "file", "..\\windows\\system32\\config", "data")
		req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()

		Submit(w, req)

		assertStatus(t, w.Code, http.StatusBadRequest)
		assertFailure(t, decodeResponse(t, w), "INVALID_FILENAME")
	})

	// filepath.Base("../../etc/passwd") = "passwd" which is safe
	// This verifies the sanitization works (traversal stripped, file accepted)
	t.Run("dot dot slash sanitized by filepath.Base", func(t *testing.T) {
		body, contentType := createMultipartFile(t, "file", "../../etc/passwd", "data")
		req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()

		Submit(w, req)

		// filepath.Base strips to "passwd" — valid filename, accepted
		assertStatus(t, w.Code, http.StatusOK)
		resp := decodeResponse(t, w)
		assertSuccess(t, resp)
	})
}

func TestSubmitFileUploadBlockedExtension(t *testing.T) {
	cases := []string{".exe", ".bat", ".sh", ".ps1", ".dll", ".cmd"}
	for _, ext := range cases {
		t.Run(ext, func(t *testing.T) {
			body, contentType := createMultipartFile(t, "file", "payload"+ext, "data")
			req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
			req.Header.Set("Content-Type", contentType)
			w := httptest.NewRecorder()

			Submit(w, req)

			assertStatus(t, w.Code, http.StatusBadRequest)
			assertFailure(t, decodeResponse(t, w), "INVALID_FILENAME")
		})
	}
}

// --- Repo Link Tests ---

func TestSubmitRepoLink(t *testing.T) {
	payload := `{"repo": "https://github.com/user/project"}`
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusOK)
	resp := decodeResponse(t, w)
	assertSuccess(t, resp)
}

func TestSubmitRepoLinkEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(`{"repo": ""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "INVALID_REPO")
}

func TestSubmitRepoLinkWhitespace(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(`{"repo": "   "}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "INVALID_REPO")
}

func TestSubmitRepoLinkInvalidURL(t *testing.T) {
	cases := []struct {
		name string
		repo string
	}{
		{"gitlab", `{"repo": "https://gitlab.com/user/project"}`},
		{"http", `{"repo": "http://github.com/user/project"}`},
		{"bare string", `{"repo": "not-a-url"}`},
		{"path traversal", `{"repo": "https://github.com/../../etc/passwd"}`},
		{"no repo name", `{"repo": "https://github.com/user"}`},
		{"just domain", `{"repo": "https://github.com/"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(tc.repo))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			Submit(w, req)

			assertStatus(t, w.Code, http.StatusBadRequest)
			assertFailure(t, decodeResponse(t, w), "INVALID_REPO")
		})
	}
}

func TestSubmitInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(`{bad json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "INVALID_JSON")
}

func TestSubmitEmptyJSONBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "INVALID_REPO")
}

func TestSubmitUnknownJSONFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(`{"repo": "https://github.com/user/project", "evil": "payload"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "INVALID_JSON")
}

// --- Content-Type Tests ---

func TestSubmitUnsupportedContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader("plain text"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusUnsupportedMediaType)
	assertFailure(t, decodeResponse(t, w), "UNSUPPORTED_MEDIA_TYPE")
}

func TestSubmitMissingContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader("data"))
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "MISSING_CONTENT_TYPE")
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
