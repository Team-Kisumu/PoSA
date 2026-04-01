package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Murzuqisah/PoSA/validation"
)

// FuzzSubmitFilename fuzzes the filename field of multipart file uploads.
// Targets: path traversal, null bytes, extension bypass, regex edge cases.
func FuzzSubmitFilename(f *testing.F) {
	// Seed corpus: known attack patterns and edge cases.
	seeds := []string{
		"main.go",
		"../../etc/passwd",
		"..\\windows\\system32\\config",
		"file\x00.txt",
		".hidden",
		"-flag.txt",
		"payload.exe",
		"a.bat",
		strings.Repeat("a", 300) + ".go",
		"file<script>.go",
		"file\n.go",
		"file\r\n.go",
		"CON.go",       // Windows reserved name
		"NUL.txt",      // Windows reserved name
		"file .go",     // trailing space
		"file\t.go",    // tab in name
		"%2e%2e%2f.go", // URL-encoded traversal
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, filename string) {
		body, contentType := fuzzMultipartFile(t, filename, "package main\n")
		if body == nil {
			return
		}
		req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()

		// Must not panic regardless of input.
		Submit(w, req)

		// Must return valid JSON.
		var resp APIResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("invalid JSON response for filename %q: %v", filename, err)
		}

		// If accepted, filename must be sanitized (no path separators).
		if resp.Success {
			data, _ := json.Marshal(resp.Data)
			var sr SubmitResponse
			json.Unmarshal(data, &sr)
			if strings.ContainsAny(sr.Name, "/\\") {
				t.Errorf("accepted filename contains path separator: %q", sr.Name)
			}
			if strings.Contains(sr.Name, "\x00") {
				t.Errorf("accepted filename contains null byte: %q", sr.Name)
			}
		}
	})
}

// FuzzSubmitContent fuzzes the file content of multipart uploads.
// Targets: binary detection bypass, null byte injection, magic byte evasion.
func FuzzSubmitContent(f *testing.F) {
	seeds := [][]byte{
		[]byte("package main\n"),
		[]byte("#!/bin/bash\nrm -rf /\n"),
		{0x7f, 'E', 'L', 'F', 0, 0, 0, 0},           // ELF header
		{0x4d, 0x5a, 0x90, 0x00},                    // PE/MZ header
		{0x50, 0x4b, 0x03, 0x04},                    // ZIP header
		{0x89, 0x50, 0x4e, 0x47},                    // PNG header
		[]byte("normal text\x00hidden binary"),      // embedded null
		[]byte("%PDF-1.4 fake pdf"),                 // PDF header
		[]byte("\xff\xfe<script>alert(1)</script>"), // BOM + XSS
		bytes.Repeat([]byte("A"), 1024),             // repetitive content
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, content []byte) {
		if len(content) == 0 {
			return // empty files are rejected by design
		}
		body, contentType := fuzzMultipartFile(t, "test.go", string(content))
		if body == nil {
			return
		}
		req := httptest.NewRequest(http.MethodPost, "/api/submit", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()

		Submit(w, req)

		var resp APIResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("invalid JSON response: %v", err)
		}

		// If accepted, content must have passed all security checks.
		if resp.Success {
			// Verify MIME was detected as text-based.
			data, _ := json.Marshal(resp.Data)
			var sr SubmitResponse
			json.Unmarshal(data, &sr)
			if !strings.HasPrefix(sr.MIME, "text/") && sr.MIME != "application/json" && sr.MIME != "application/xml" {
				t.Errorf("accepted non-text MIME: %q", sr.MIME)
			}
		}
	})
}

// FuzzSubmitRepoURL fuzzes the repo URL field of JSON submissions.
// Targets: SSRF, injection, traversal, protocol smuggling.
func FuzzSubmitRepoURL(f *testing.F) {
	seeds := []string{
		"https://github.com/user/repo",
		"https://github.com/../../etc/passwd",
		"http://github.com/user/repo",
		"https://evil.com/user/repo",
		"javascript:alert(1)",
		"file:///etc/passwd",
		"https://github.com/user/repo\x00.evil.com",
		"https://github.com@evil.com/user/repo",
		"https://github.com/user/repo#<script>",
		"https://github.com/user/repo?redirect=http://evil.com",
		strings.Repeat("a", 10000),
		"",
		"   ",
		"https://github.com//repo",
		"https://github.com/user/",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, repoURL string) {
		payload, err := json.Marshal(map[string]string{"repo": repoURL})
		if err != nil {
			return
		}
		req := httptest.NewRequest(http.MethodPost, "/api/submit", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		Submit(w, req)

		var resp APIResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("invalid JSON response for repo %q: %v", repoURL, err)
		}

		// If accepted, URL must be HTTPS github.com with owner/repo.
		if resp.Success {
			if !strings.HasPrefix(repoURL, "https://github.com/") {
				t.Errorf("accepted non-GitHub URL: %q", repoURL)
			}
		}
	})
}

// FuzzCIDValidation fuzzes the CID validation function directly.
// Targets: injection characters, format bypass, overflow.
func FuzzCIDValidation(f *testing.F) {
	seeds := []string{
		"QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco",
		"",
		"abc;rm -rf /",
		"abc|cat /etc/passwd",
		"abc<script>alert(1)</script>",
		"abc\x00def",
		"abc/../../etc/passwd",
		strings.Repeat("Q", 100),
		"Qm" + strings.Repeat("a", 44),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, cid string) {
		err := validation.CID(cid)

		// If accepted, CID must be safe (no injection chars, correct length).
		if err == nil {
			if strings.ContainsAny(cid, "\x00/\\<>\"';|&") {
				t.Errorf("CID accepted with injection chars: %q", cid)
			}
			if len(cid) < 46 || len(cid) > 59 {
				t.Errorf("CID accepted with invalid length %d: %q", len(cid), cid)
			}
		}
	})
}

// FuzzSubmitJSONBody fuzzes the raw JSON body of repo submissions.
// Targets: JSON parsing exploits, deeply nested objects, huge payloads.
func FuzzSubmitJSONBody(f *testing.F) {
	seeds := []string{
		`{"repo": "https://github.com/user/repo"}`,
		`{}`,
		`{"repo": ""}`,
		`{"repo": "https://github.com/user/repo", "evil": "payload"}`,
		`null`,
		`[]`,
		`{"repo": ` + strings.Repeat("[", 100) + strings.Repeat("]", 100) + `}`,
		strings.Repeat(`{"a":`, 50) + `1` + strings.Repeat(`}`, 50),
		`{"repo": "` + strings.Repeat("a", 2000000) + `"}`,
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, body string) {
		req := httptest.NewRequest(http.MethodPost, "/api/submit", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		Submit(w, req)

		// Must always return valid JSON, never panic.
		var resp APIResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("invalid JSON response for body %q: %v", truncate(body, 100), err)
		}
	})
}

// --- Helpers ---

func fuzzMultipartFile(t *testing.T, filename, content string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, ""
	}
	part.Write([]byte(content))
	writer.Close()
	return &buf, writer.FormDataContentType()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
