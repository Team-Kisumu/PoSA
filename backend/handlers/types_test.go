package handlers

import "testing"

// TestValidateFilename verifies the filename validator accepts safe filenames
// and rejects dangerous ones (empty, null bytes, path traversal, blocked extensions,
// special characters, and names that don't match the whitelist pattern).
func TestValidateFilename(t *testing.T) {
	// Valid filenames: alphanumeric start, safe characters, allowed extensions.
	valid := []string{
		"main.go", "test.py", "README.md", "file-name.txt",
		"code_v2.js", "a.c",
	}
	for _, name := range valid {
		if err := ValidateFilename(name); err != nil {
			t.Errorf("ValidateFilename(%q) = %v, want nil", name, err)
		}
	}

	// Invalid filenames: each case targets a specific validation rule.
	invalid := []struct {
		name string
		file string
	}{
		{"empty", ""},                         // empty filename
		{"null byte", "file\x00.txt"},         // null byte injection
		{"forward slash", "path/file.txt"},    // path separator
		{"backslash", "path\\file.txt"},       // Windows path separator
		{"dot dot slash", "../../etc/passwd"}, // directory traversal
		{"exe", "payload.exe"},                // blocked extension
		{"bat", "script.bat"},                 // blocked extension
		{"sh", "run.sh"},                      // blocked extension
		{"ps1", "script.ps1"},                 // blocked extension
		{"dll", "lib.dll"},                    // blocked extension
		{"starts with dot", ".hidden"},        // fails whitelist (must start alphanumeric)
		{"starts with dash", "-flag.txt"},     // fails whitelist (must start alphanumeric)
		{"special chars", "file<>.txt"},       // fails whitelist (angle brackets)
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateFilename(tc.file); err == nil {
				t.Errorf("ValidateFilename(%q) = nil, want error", tc.file)
			}
		})
	}
}

// TestValidateRepoURL verifies the repo URL validator accepts valid GitHub HTTPS
// URLs and rejects invalid ones (empty, HTTP, non-GitHub hosts, missing repo path,
// path traversal, bare strings).
func TestValidateRepoURL(t *testing.T) {
	// Valid URLs: HTTPS, github.com, at least owner/repo segments.
	valid := []string{
		"https://github.com/user/project",
		"https://github.com/org/repo-name",
		"https://github.com/user/repo/tree/main",
	}
	for _, url := range valid {
		if err := ValidateRepoURL(url); err != nil {
			t.Errorf("ValidateRepoURL(%q) = %v, want nil", url, err)
		}
	}

	// Invalid URLs: each case targets a specific validation rule.
	invalid := []struct {
		name string
		url  string
	}{
		{"empty", ""},                                             // empty string
		{"whitespace", "   "},                                     // whitespace-only
		{"http", "http://github.com/user/repo"},                   // non-HTTPS scheme
		{"gitlab", "https://gitlab.com/user/repo"},                // non-GitHub host
		{"no repo", "https://github.com/user"},                    // missing repo segment
		{"just domain", "https://github.com/"},                    // no path segments
		{"path traversal", "https://github.com/../../etc/passwd"}, // ".." in URL
		{"bare string", "not-a-url"},                              // not a valid URL
		{"empty owner", "https://github.com//repo"},               // empty owner segment
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateRepoURL(tc.url); err == nil {
				t.Errorf("ValidateRepoURL(%q) = nil, want error", tc.url)
			}
		})
	}
}

// TestValidateCID verifies the CID validator accepts valid IPFS CIDs and
// rejects invalid ones (empty, too short, injection characters like semicolons,
// pipes, angle brackets, quotes, slashes, null bytes).
func TestValidateCID(t *testing.T) {
	// Valid CIDs: alphanumeric, 46-59 characters (CIDv0 and CIDv1 formats).
	valid := []string{
		"QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco",
		"QmT5NvUtoM5nWFfrQdVrFtvGfKFmG7AHE8P34isapyhCxX",
	}
	for _, cid := range valid {
		if err := ValidateCID(cid); err != nil {
			t.Errorf("ValidateCID(%q) = %v, want nil", cid, err)
		}
	}

	// Invalid CIDs: each case targets a specific injection or format violation.
	invalid := []struct {
		name string
		cid  string
	}{
		{"empty", ""},                           // empty string
		{"too short", "Qm123"},                  // below 46-char minimum
		{"semicolon injection", "abc;rm -rf /"}, // command injection via semicolon
		{"null byte", "Qm\x00abc"},              // null byte injection
		{"pipe", "abc|cat /etc/passwd"},         // command injection via pipe
		{"angle brackets", "abc<script>"},       // HTML/XSS injection
		{"quotes", `abc"onload=`},               // attribute injection
		{"slash", "abc/def"},                    // path separator
		{"backslash", "abc\\def"},               // Windows path separator
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateCID(tc.cid); err == nil {
				t.Errorf("ValidateCID(%q) = nil, want error", tc.cid)
			}
		})
	}
}
