// Package handlers implements HTTP route handlers and input validation
// for the PoSA backend API. All responses use a consistent APIResponse
// envelope to simplify client-side parsing and error handling.
package handlers

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// --- Response Types ---

// APIResponse is the standard envelope for all API responses.
// On success: Success=true, Data is populated, Error is nil.
// On failure: Success=false, Data is nil, Error contains the code and message.
type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

// Error represents a structured error with a machine-readable code
// (e.g. "INVALID_CID") and a human-readable message.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// HealthData is the response payload for the GET /health endpoint.
type HealthData struct {
	Status string `json:"status"`
}

// SubmitResponse is the response payload for successful POST /api/submit requests.
// Type indicates the submission kind ("file" or "repo").
// Size is the byte count of the uploaded file (0 for repo submissions).
type SubmitResponse struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Message string `json:"message"`
}

// VerifyResponse is the response payload for GET /api/verify/{cid}.
type VerifyResponse struct {
	CID     string `json:"cid"`
	Message string `json:"message"`
}

// --- Request Types ---

// RepoRequest is the expected JSON body for repo-link submissions.
// Only the "repo" field is accepted; unknown fields are rejected
// via DisallowUnknownFields() in the decoder.
type RepoRequest struct {
	Repo string `json:"repo"`
}

// --- Validation ---

var (
	// cidPattern matches IPFS CIDv0 (Qm..., 46 chars) and CIDv1 (bafy..., up to 59 chars).
	// Only alphanumeric characters are allowed to prevent injection attacks.
	cidPattern = regexp.MustCompile(`^[a-zA-Z0-9]{46,59}$`)

	// safeFilenameRe enforces a whitelist: must start with alphanumeric,
	// followed by alphanumeric, dots, hyphens, or underscores. Max 255 chars.
	safeFilenameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,254}$`)

	// blockedExtensions prevents upload of executable/script file types
	// that could pose a security risk if stored or processed.
	blockedExtensions = map[string]bool{
		".exe": true, ".bat": true, ".cmd": true, ".sh": true,
		".ps1": true, ".msi": true, ".dll": true, ".so": true,
		".com": true, ".scr": true, ".vbs": true, ".wsf": true,
	}
)

const (
	maxJSONBodySize = 1 << 20  // 1MB — limit for JSON request bodies
	maxUploadSize   = 10 << 20 // 10MB — limit for multipart file uploads
)

// ValidateFilename checks that a filename is safe for storage and processing.
// It rejects: empty names, null bytes, path separators, traversal sequences,
// names that don't match the safe character whitelist, and blocked extensions.
func ValidateFilename(name string) error {
	if name == "" {
		return fmt.Errorf("filename is required")
	}

	// Block null bytes and path separators to prevent path injection.
	if strings.ContainsAny(name, "\x00/\\") {
		return fmt.Errorf("filename contains illegal characters")
	}

	// filepath.Base strips directory components; if the result differs
	// from the input, the original contained traversal sequences.
	clean := filepath.Base(name)
	if clean != name {
		return fmt.Errorf("filename contains path traversal")
	}

	// Enforce the alphanumeric-start whitelist pattern.
	if !safeFilenameRe.MatchString(clean) {
		return fmt.Errorf("filename contains invalid characters")
	}

	// Reject dangerous executable/script extensions.
	ext := strings.ToLower(filepath.Ext(clean))
	if blockedExtensions[ext] {
		return fmt.Errorf("file type %q is not allowed", ext)
	}
	return nil
}

// ValidateRepoURL ensures the repo URL is a valid HTTPS GitHub repository link.
// It enforces: non-empty, valid URL format, HTTPS scheme, github.com host,
// at least owner/repo path segments, and no path traversal sequences.
func ValidateRepoURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("'repo' field is required")
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("invalid URL format")
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("repo URL must use HTTPS")
	}
	if parsed.Host != "github.com" {
		return fmt.Errorf("repo must be a GitHub URL")
	}

	// Require at least two non-empty path segments: owner and repo name.
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("repo URL must include owner and repository (https://github.com/owner/repo)")
	}

	// Block ".." sequences that could be used for path traversal.
	if strings.Contains(raw, "..") {
		return fmt.Errorf("repo URL contains path traversal")
	}
	return nil
}

// ValidateCID checks that a CID string is safe and matches the expected
// IPFS CID format. It blocks: empty values, injection characters
// (null, slashes, angle brackets, quotes, semicolons, pipes, ampersands),
// and strings that don't match the alphanumeric 46-59 char pattern.
func ValidateCID(cid string) error {
	if cid == "" {
		return fmt.Errorf("cid is required")
	}

	// Block characters commonly used in command/HTML/SQL injection.
	if strings.ContainsAny(cid, "\x00/\\<>\"';&|") {
		return fmt.Errorf("cid contains illegal characters")
	}

	// Enforce the strict alphanumeric CID pattern.
	if !cidPattern.MatchString(cid) {
		return fmt.Errorf("cid format is invalid")
	}
	return nil
}
