package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
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
// Security measures:
//   - MaxBytesReader caps the body at 10MB to prevent resource exhaustion.
//   - filepath.Base strips directory components to neutralize path traversal.
//   - ValidateFilename enforces a safe character whitelist and blocks dangerous extensions.
//   - Empty files are rejected to prevent no-op submissions.
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	// Cap request body size to prevent denial-of-service via large uploads.
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "file exceeds 10MB limit")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "MISSING_FILE", "missing or invalid 'file' field")
		return
	}
	defer file.Close()

	// Sanitize filename: filepath.Base strips any directory prefix (e.g. "../../etc/passwd" → "passwd"),
	// then ValidateFilename enforces the safe character whitelist and blocks dangerous extensions.
	filename := filepath.Base(header.Filename)
	if err := ValidateFilename(filename); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_FILENAME", err.Error())
		return
	}

	content, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "READ_ERROR", "failed to read file")
		return
	}

	// Reject empty files — they provide no content for AI evaluation.
	if len(content) == 0 {
		respondError(w, http.StatusBadRequest, "EMPTY_FILE", "file is empty")
		return
	}

	respondOK(w, SubmitResponse{
		Type:    "file",
		Name:    filename,
		Size:    int64(len(content)),
		Message: "file received, pending evaluation",
	})
}

// handleRepoLink processes JSON requests containing a GitHub repository URL.
// Security measures:
//   - MaxBytesReader caps the body at 1MB to prevent oversized JSON payloads.
//   - DisallowUnknownFields rejects unexpected JSON keys (prevents parameter pollution).
//   - ValidateRepoURL enforces HTTPS, github.com host, owner/repo path, and blocks traversal.
func handleRepoLink(w http.ResponseWriter, r *http.Request) {
	// Cap JSON body size to prevent resource exhaustion.
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodySize)

	// Strict JSON decoding: reject payloads with unexpected fields.
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req RepoRequest
	if err := decoder.Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON body")
		return
	}

	req.Repo = strings.TrimSpace(req.Repo)
	if err := ValidateRepoURL(req.Repo); err != nil {
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
