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

// TestSubmitFileUpload verifies a valid multipart file upload returns 200
// with the correct type, filename, byte size, and detected MIME type.
func TestSubmitFileUpload(t *testing.T) {
	body, contentType := createMultipartFile(t, "file", "main.go", "package main\n")
	req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusOK)
	resp := decodeResponse(t, w)
	assertSuccess(t, resp)

	// Verify the response payload matches the uploaded file metadata.
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
	if sr.MIME != "text/plain" {
		t.Errorf("mime = %q, want \"text/plain\"", sr.MIME)
	}
}

// TestSubmitFileUploadMissingField verifies that using the wrong form field
// name (not "file") returns a MISSING_FILE error.
func TestSubmitFileUploadMissingField(t *testing.T) {
	body, contentType := createMultipartFile(t, "wrong_field", "main.go", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "MISSING_FILE")
}

// TestSubmitFileUploadEmptyFile verifies that uploading a zero-byte file
// returns an EMPTY_FILE error.
func TestSubmitFileUploadEmptyFile(t *testing.T) {
	body, contentType := createMultipartFile(t, "file", "empty.txt", "")
	req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "EMPTY_FILE")
}

// TestSubmitFileUploadPathTraversal tests that path traversal attempts in
// filenames are handled safely by the filepath.Base sanitization layer.
func TestSubmitFileUploadPathTraversal(t *testing.T) {
	// Backslash-based traversal is caught by validation.Filename's illegal char check.
	t.Run("backslash traversal", func(t *testing.T) {
		body, contentType := createMultipartFile(t, "file", "..\\windows\\system32\\config", "data")
		req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()

		Submit(w, req)

		assertStatus(t, w.Code, http.StatusBadRequest)
		assertFailure(t, decodeResponse(t, w), "INVALID_FILENAME")
	})

	// Forward-slash traversal: filepath.Base("../../etc/passwd") -> "passwd",
	// which is a safe filename. This confirms sanitization works correctly.
	t.Run("dot dot slash sanitized by filepath.Base", func(t *testing.T) {
		body, contentType := createMultipartFile(t, "file", "../../etc/passwd", "data")
		req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()

		Submit(w, req)

		// filepath.Base strips to "passwd" — valid filename, accepted.
		assertStatus(t, w.Code, http.StatusOK)
		resp := decodeResponse(t, w)
		assertSuccess(t, resp)
	})
}

// TestSubmitFileUploadBlockedExtension verifies that files with dangerous
// extensions are rejected at the filename validation layer.
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

// --- MIME Type Filtering Tests ---

// TestSubmitFileUploadBinaryContent verifies that binary files disguised with
// safe extensions are rejected by the MIME sniffing layer.
func TestSubmitFileUploadBinaryContent(t *testing.T) {
	t.Run("PNG disguised as .go", func(t *testing.T) {
		// PNG magic bytes in a file named "main.go".
		png := string([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}) + "fake png data"
		body, contentType := createMultipartFile(t, "file", "main.go", png)
		req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()

		Submit(w, req)

		assertStatus(t, w.Code, http.StatusBadRequest)
		resp := decodeResponse(t, w)
		// Could be INVALID_CONTENT_TYPE (MIME) or MALICIOUS_CONTENT (magic bytes).
		if resp.Success {
			t.Error("expected rejection for PNG disguised as .go")
		}
	})

	t.Run("JPEG disguised as .py", func(t *testing.T) {
		jpeg := string([]byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10}) + "fake jpeg"
		body, contentType := createMultipartFile(t, "file", "script.py", jpeg)
		req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()

		Submit(w, req)

		assertStatus(t, w.Code, http.StatusBadRequest)
		if decodeResponse(t, w).Success {
			t.Error("expected rejection for JPEG disguised as .py")
		}
	})

	t.Run("ZIP disguised as .txt", func(t *testing.T) {
		zip := string([]byte{0x50, 0x4b, 0x03, 0x04}) + "fake zip content padding"
		body, contentType := createMultipartFile(t, "file", "readme.txt", zip)
		req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()

		Submit(w, req)

		assertStatus(t, w.Code, http.StatusBadRequest)
		if decodeResponse(t, w).Success {
			t.Error("expected rejection for ZIP disguised as .txt")
		}
	})
}

// --- Content Scan Tests (through handler) ---

// TestSubmitFileUploadMaliciousContent verifies that the deep content scanner
// catches embedded threats even when filename and MIME checks pass.
func TestSubmitFileUploadMaliciousContent(t *testing.T) {
	t.Run("null bytes in source code", func(t *testing.T) {
		// Embedded null byte causes http.DetectContentType to return
		// "application/octet-stream" (binary), so the MIME check rejects
		// it before the content scanner runs.
		data := "package main\n\x00func evil() {}\n"
		body, contentType := createMultipartFile(t, "file", "main.go", data)
		req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()

		Submit(w, req)

		assertStatus(t, w.Code, http.StatusBadRequest)
		resp := decodeResponse(t, w)
		if resp.Success {
			t.Error("expected rejection for null bytes in content")
		}
		// Rejected by MIME check (octet-stream) or content scan (null bytes).
		code := resp.Error.Code
		if code != "INVALID_CONTENT_TYPE" && code != "MALICIOUS_CONTENT" {
			t.Errorf("error code = %q, want INVALID_CONTENT_TYPE or MALICIOUS_CONTENT", code)
		}
	})

	t.Run("shebang in go file", func(t *testing.T) {
		// Shell shebang in a .go file — suspicious, likely a disguised script.
		data := "#!/bin/bash\necho 'pwned'\n"
		body, contentType := createMultipartFile(t, "file", "main.go", data)
		req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()

		Submit(w, req)

		assertStatus(t, w.Code, http.StatusBadRequest)
		assertFailure(t, decodeResponse(t, w), "MALICIOUS_CONTENT")
	})

	t.Run("shebang in python file allowed", func(t *testing.T) {
		// Python files legitimately use shebangs.
		data := "#!/usr/bin/env python3\nprint('hello')\n"
		body, contentType := createMultipartFile(t, "file", "script.py", data)
		req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()

		Submit(w, req)

		assertStatus(t, w.Code, http.StatusOK)
		assertSuccess(t, decodeResponse(t, w))
	})
}

// --- Repo Link Tests ---

// TestSubmitRepoLink verifies a valid GitHub repo URL is accepted.
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

// TestSubmitRepoLinkEmpty verifies that an empty repo field is rejected.
func TestSubmitRepoLinkEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(`{"repo": ""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "INVALID_REPO")
}

// TestSubmitRepoLinkWhitespace verifies that a whitespace-only repo field is rejected.
func TestSubmitRepoLinkWhitespace(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(`{"repo": "   "}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "INVALID_REPO")
}

// TestSubmitRepoLinkInvalidURL tests various invalid repo URL formats.
func TestSubmitRepoLinkInvalidURL(t *testing.T) {
	cases := []struct {
		name string
		repo string
	}{
		{"gitlab", `{"repo": "https://gitlab.com/user/project"}`},             // wrong host
		{"http", `{"repo": "http://github.com/user/project"}`},                // non-HTTPS
		{"bare string", `{"repo": "not-a-url"}`},                              // not a URL
		{"path traversal", `{"repo": "https://github.com/../../etc/passwd"}`}, // ".." in path
		{"no repo name", `{"repo": "https://github.com/user"}`},               // missing repo
		{"just domain", `{"repo": "https://github.com/"}`},                    // no path segments
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

// TestSubmitInvalidJSON verifies that malformed JSON bodies are rejected.
func TestSubmitInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(`{bad json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "INVALID_JSON")
}

// TestSubmitEmptyJSONBody verifies that an empty JSON object (no repo field)
// is rejected with INVALID_REPO.
func TestSubmitEmptyJSONBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "INVALID_REPO")
}

// TestSubmitUnknownJSONFields verifies that JSON payloads with unexpected
// fields are rejected (DisallowUnknownFields enforcement).
func TestSubmitUnknownJSONFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(`{"repo": "https://github.com/user/project", "evil": "payload"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "INVALID_JSON")
}

// --- Content-Type Tests ---

// TestSubmitUnsupportedContentType verifies that non-multipart/non-JSON
// content types are rejected with 415 Unsupported Media Type.
func TestSubmitUnsupportedContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader("plain text"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusUnsupportedMediaType)
	assertFailure(t, decodeResponse(t, w), "UNSUPPORTED_MEDIA_TYPE")
}

// TestSubmitMissingContentType verifies that requests without a Content-Type
// header are rejected with MISSING_CONTENT_TYPE.
func TestSubmitMissingContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader("data"))
	w := httptest.NewRecorder()

	Submit(w, req)

	assertStatus(t, w.Code, http.StatusBadRequest)
	assertFailure(t, decodeResponse(t, w), "MISSING_CONTENT_TYPE")
}

// --- Helpers ---

// createMultipartFile builds a multipart/form-data request body with a single
// file field. Returns the body buffer and the Content-Type header value
// (which includes the multipart boundary).
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
