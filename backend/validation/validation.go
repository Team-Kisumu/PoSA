// Package validation provides input validation and content filtering for
// the PoSA backend. All validators are pure functions with no side effects,
// making them independently testable and reusable across handlers.
//
// Validation layers (defense-in-depth):
//  1. Filename validation — character whitelist, extension blocklist, traversal prevention
//  2. MIME-type validation — sniffs actual file bytes to detect true content type
//  3. Content sanitization — scans for embedded threats (null bytes, polyglot headers, shebangs)
//  4. Repo URL validation — scheme, host, path structure, traversal prevention
//  5. CID validation — format and injection character blocking
package validation

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// --- Compiled Patterns ---

var (
	// cidPattern matches IPFS CIDv0 (Qm..., 46 chars) and CIDv1 (bafy..., up to 59 chars).
	// Only alphanumeric characters are allowed to prevent injection attacks.
	cidPattern = regexp.MustCompile(`^[a-zA-Z0-9]{46,59}$`)

	// safeFilenameRe enforces a whitelist: must start with alphanumeric,
	// followed by alphanumeric, dots, hyphens, or underscores. Max 255 chars.
	safeFilenameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,254}$`)
)

// --- Extension & MIME Policies ---

// blockedExtensions prevents upload of executable/script file types
// that could pose a security risk if stored or processed.
var blockedExtensions = map[string]bool{
	".exe": true, ".bat": true, ".cmd": true, ".sh": true,
	".ps1": true, ".msi": true, ".dll": true, ".so": true,
	".com": true, ".scr": true, ".vbs": true, ".wsf": true,
	".bin": true, ".elf": true, ".class": true, ".jar": true,
	".war": true, ".ear": true, ".deb": true, ".rpm": true,
}

// allowedMIMETypes is a whitelist of MIME types that are valid for code
// and writing submissions. http.DetectContentType sniffs the first 512
// bytes and returns one of these. Only text-based content is accepted.
var allowedMIMETypes = map[string]bool{
	"text/plain":               true,  // source code, plain text, markdown, etc.
	"text/html":                true,  // HTML files
	"text/xml":                 true,  // XML, SVG, config files
	"application/json":         true,  // JSON config/data files
	"application/xml":          true,  // XML variants
	"application/octet-stream": false, // generic binary — rejected
	"application/x-gzip":       false, // compressed archives
	"application/zip":          false, // zip archives
	"application/pdf":          false, // PDFs
	"image/png":                false, // images
	"image/jpeg":               false, // images
	"image/gif":                false, // images
	"audio/mpeg":               false, // audio
	"video/mp4":                false, // video
}

// dangerousHeaders are magic byte sequences found at the start of binary
// or executable file formats. If any of these appear at the beginning of
// an uploaded file, the content is rejected regardless of extension or MIME.
var dangerousHeaders = []struct {
	Name   string // human-readable format name for error messages
	Prefix []byte // magic bytes to match against file start
}{
	{"ELF binary", []byte{0x7f, 'E', 'L', 'F'}},           // Linux executables
	{"PE executable", []byte{0x4d, 0x5a}},                 // Windows .exe/.dll (MZ header)
	{"Mach-O binary", []byte{0xcf, 0xfa, 0xed, 0xfe}},     // macOS executables
	{"Java class", []byte{0xca, 0xfe, 0xba, 0xbe}},        // .class files
	{"gzip archive", []byte{0x1f, 0x8b}},                  // .gz/.tar.gz
	{"ZIP archive", []byte{0x50, 0x4b, 0x03, 0x04}},       // .zip/.jar/.war
	{"PDF document", []byte{0x25, 0x50, 0x44, 0x46}},      // %PDF
	{"RAR archive", []byte{0x52, 0x61, 0x72, 0x21, 0x1a}}, // Rar!
}

// --- Size Limits ---

const (
	MaxJSONBodySize = 1 << 20  // 1MB — limit for JSON request bodies
	MaxUploadSize   = 10 << 20 // 10MB — limit for multipart file uploads
)

// --- Filename Validation ---

// Filename checks that a filename is safe for storage and processing.
// It rejects: empty names, null bytes, path separators, traversal sequences,
// names that don't match the safe character whitelist, and blocked extensions.
func Filename(name string) error {
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

// --- Content Validation ---

// ContentType sniffs the first 512 bytes of content using http.DetectContentType
// and checks the result against the allowed MIME whitelist. This catches renamed
// binaries (e.g. an ELF binary named "main.go") that would pass filename validation.
func ContentType(content []byte) (string, error) {
	if len(content) == 0 {
		return "", fmt.Errorf("file is empty")
	}

	// http.DetectContentType reads up to 512 bytes to determine the MIME type.
	detected := http.DetectContentType(content)

	// Strip parameters (e.g. "text/plain; charset=utf-8" → "text/plain").
	mime := strings.SplitN(detected, ";", 2)[0]
	mime = strings.TrimSpace(mime)

	// Check against the whitelist. Unknown MIME types (not in the map) are
	// rejected by default — only explicitly allowed types pass.
	allowed, known := allowedMIMETypes[mime]
	if !known || !allowed {
		return mime, fmt.Errorf("file type %q is not allowed", mime)
	}
	return mime, nil
}

// Content performs deep inspection of file bytes for malicious patterns.
// It checks for:
//   - Binary magic bytes (ELF, PE, Mach-O, Java class, archives, PDF)
//   - Embedded null bytes (common in binary files, not in source code)
//   - Shell shebangs (#!/bin/sh, #!/usr/bin/env) in non-shell files
//
// The filename parameter is used to allow shebangs in appropriate file types
// (e.g. .py files may legitimately start with #!/usr/bin/env python).
func Content(content []byte, filename string) error {
	if len(content) == 0 {
		return fmt.Errorf("file is empty")
	}

	// Check for dangerous binary format magic bytes at the file start.
	for _, hdr := range dangerousHeaders {
		if len(content) >= len(hdr.Prefix) && matchPrefix(content, hdr.Prefix) {
			return fmt.Errorf("file contains %s header", hdr.Name)
		}
	}

	// Null bytes are present in binary files but not in legitimate source code
	// or text. Scan a reasonable prefix (first 8KB) to avoid scanning huge files.
	scanLen := len(content)
	if scanLen > 8192 {
		scanLen = 8192
	}
	for i := 0; i < scanLen; i++ {
		if content[i] == 0x00 {
			return fmt.Errorf("file contains null bytes (binary content)")
		}
	}

	// Check for shell shebangs in files that shouldn't have them.
	// Python, Ruby, Perl, and Node files may legitimately use shebangs.
	if len(content) >= 2 && content[0] == '#' && content[1] == '!' {
		ext := strings.ToLower(filepath.Ext(filename))
		shebangAllowed := map[string]bool{
			".py": true, ".rb": true, ".pl": true, ".js": true,
		}
		if !shebangAllowed[ext] {
			return fmt.Errorf("file contains shell shebang in non-script file type")
		}
	}

	return nil
}

// matchPrefix compares the start of data against a prefix byte slice.
func matchPrefix(data, prefix []byte) bool {
	for i, b := range prefix {
		if data[i] != b {
			return false
		}
	}
	return true
}

// --- Repo URL Validation ---

// RepoURL ensures the repo URL is a valid HTTPS GitHub repository link.
// It enforces: non-empty, valid URL format, HTTPS scheme, github.com host,
// at least owner/repo path segments, and no path traversal sequences.
func RepoURL(raw string) error {
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

// --- CID Validation ---

// CID checks that a CID string is safe and matches the expected IPFS CID format.
// It blocks: empty values, injection characters (null, slashes, angle brackets,
// quotes, semicolons, pipes, ampersands), and strings that don't match the
// alphanumeric 46-59 char pattern.
func CID(cid string) error {
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
