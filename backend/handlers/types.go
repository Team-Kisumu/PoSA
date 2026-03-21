package handlers

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// --- Response Types ---

type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type HealthData struct {
	Status string `json:"status"`
}

type SubmitResponse struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Message string `json:"message"`
}

type VerifyResponse struct {
	CID     string `json:"cid"`
	Message string `json:"message"`
}

// --- Request Types ---

type RepoRequest struct {
	Repo string `json:"repo"`
}

// --- Validation ---

var (
	cidPattern        = regexp.MustCompile(`^[a-zA-Z0-9]{46,59}$`)
	safeFilenameRe    = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,254}$`)
	blockedExtensions = map[string]bool{
		".exe": true, ".bat": true, ".cmd": true, ".sh": true,
		".ps1": true, ".msi": true, ".dll": true, ".so": true,
		".com": true, ".scr": true, ".vbs": true, ".wsf": true,
	}
)

const (
	maxJSONBodySize = 1 << 20  // 1MB
	maxUploadSize   = 10 << 20 // 10MB
)

func ValidateFilename(name string) error {
	if name == "" {
		return fmt.Errorf("filename is required")
	}
	if strings.ContainsAny(name, "\x00/\\") {
		return fmt.Errorf("filename contains illegal characters")
	}
	clean := filepath.Base(name)
	if clean != name {
		return fmt.Errorf("filename contains path traversal")
	}
	if !safeFilenameRe.MatchString(clean) {
		return fmt.Errorf("filename contains invalid characters")
	}
	ext := strings.ToLower(filepath.Ext(clean))
	if blockedExtensions[ext] {
		return fmt.Errorf("file type %q is not allowed", ext)
	}
	return nil
}

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
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("repo URL must include owner and repository (https://github.com/owner/repo)")
	}
	if strings.Contains(raw, "..") {
		return fmt.Errorf("repo URL contains path traversal")
	}
	return nil
}

func ValidateCID(cid string) error {
	if cid == "" {
		return fmt.Errorf("cid is required")
	}
	if strings.ContainsAny(cid, "\x00/\\<>\"';&|") {
		return fmt.Errorf("cid contains illegal characters")
	}
	if !cidPattern.MatchString(cid) {
		return fmt.Errorf("cid format is invalid")
	}
	return nil
}
