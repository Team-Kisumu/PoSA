package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Murzuqisah/PoSA/validation"
)

// Submit handles POST /api/submit. It routes to the appropriate sub-handler
// based on Content-Type: multipart/form-data for file uploads, or
// application/json for GitHub repo link submissions.
func Submit(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		respondError(w, http.StatusBadRequest, "MISSING_CONTENT_TYPE", "Content-Type header is required")
		return
	}

	switch {
	case strings.HasPrefix(ct, "multipart/form-data"):
		handleFileUpload(w, r)
	case strings.HasPrefix(ct, "application/json"):
		handleRepoLink(w, r)
	default:
		respondError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE",
			"Content-Type must be multipart/form-data or application/json")
	}
}

// handleFileUpload processes multipart file uploads from the "file" form field.
// Validation pipeline (each step must pass before the next runs):
//  1. Size limit — MaxBytesReader caps the body at 10MB
//  2. Filename — filepath.Base sanitization + validation.Filename whitelist
//  3. Content read — file bytes loaded into memory
//  4. Empty check — zero-byte files rejected
//  5. MIME sniffing — http.DetectContentType checks actual content type
//  6. Content scan — magic bytes, null bytes, and shebang detection
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	// Cap request body size to prevent denial-of-service via large uploads.
	r.Body = http.MaxBytesReader(w, r.Body, validation.MaxUploadSize)

	if err := r.ParseMultipartForm(validation.MaxUploadSize); err != nil {
		respondError(w, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "file exceeds 10MB limit")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "MISSING_FILE", "missing or invalid 'file' field")
		return
	}
	defer file.Close()

	// Step 2: Sanitize and validate filename.
	// filepath.Base strips directory prefixes (e.g. "../../etc/passwd" -> "passwd"),
	// then validation.Filename enforces the character whitelist and extension blocklist.
	filename := filepath.Base(header.Filename)
	if err := validation.Filename(filename); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_FILENAME", err.Error())
		return
	}

	// Step 3: Read file content into memory for inspection.
	content, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "READ_ERROR", "failed to read file")
		return
	}

	// Step 4: Reject empty files — no content for AI evaluation.
	if len(content) == 0 {
		respondError(w, http.StatusBadRequest, "EMPTY_FILE", "file is empty")
		return
	}

	// Step 5: MIME-type sniffing — detect actual content type from bytes.
	// Catches renamed binaries (e.g. an ELF binary named "main.go").
	mime, err := validation.ContentType(content)
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_CONTENT_TYPE", err.Error())
		return
	}

	// Step 6: Deep content scan — magic bytes, null bytes, shebangs.
	if err := validation.Content(content, filename); err != nil {
		respondError(w, http.StatusBadRequest, "MALICIOUS_CONTENT", err.Error())
		return
	}

	respondOK(w, SubmitResponse{
		Type:    "file",
		Name:    filename,
		Size:    int64(len(content)),
		MIME:    mime,
		Message: "file received, pending evaluation",
	})
}

// handleRepoLink processes JSON requests containing a GitHub repository URL.
// Security measures:
//   - MaxBytesReader caps the body at 1MB to prevent oversized JSON payloads.
//   - DisallowUnknownFields rejects unexpected JSON keys (prevents parameter pollution).
//   - validation.RepoURL enforces HTTPS, github.com host, owner/repo path, and blocks traversal.
func handleRepoLink(w http.ResponseWriter, r *http.Request) {
	// Cap JSON body size to prevent resource exhaustion.
	r.Body = http.MaxBytesReader(w, r.Body, validation.MaxJSONBodySize)

	// Strict JSON decoding: reject payloads with unexpected fields.
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req RepoRequest
	if err := decoder.Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON body")
		return
	}

	req.Repo = strings.TrimSpace(req.Repo)
	if err := validation.RepoURL(req.Repo); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REPO", err.Error())
		return
	}

	respondOK(w, SubmitResponse{
		Type:    "repo",
		Name:    req.Repo,
		Size:    0,
		Message: "repo link received, pending evaluation",
	})
}
