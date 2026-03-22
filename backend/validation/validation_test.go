package validation

import "testing"

// --- Filename Validation Tests ---

// TestFilename verifies the filename validator accepts safe filenames
// and rejects dangerous ones (empty, null bytes, path traversal, blocked extensions,
// special characters, and names that don't match the whitelist pattern).
func TestFilename(t *testing.T) {
	// Valid filenames: alphanumeric start, safe characters, allowed extensions.
	valid := []string{
		"main.go", "test.py", "README.md", "file-name.txt",
		"code_v2.js", "a.c",
	}
	for _, name := range valid {
		if err := Filename(name); err != nil {
			t.Errorf("Filename(%q) = %v, want nil", name, err)
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
		{"bin", "firmware.bin"},               // blocked extension (new)
		{"elf", "binary.elf"},                 // blocked extension (new)
		{"class", "Main.class"},               // blocked extension (new)
		{"jar", "app.jar"},                    // blocked extension (new)
		{"starts with dot", ".hidden"},        // fails whitelist (must start alphanumeric)
		{"starts with dash", "-flag.txt"},     // fails whitelist (must start alphanumeric)
		{"special chars", "file<>.txt"},       // fails whitelist (angle brackets)
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if err := Filename(tc.file); err == nil {
				t.Errorf("Filename(%q) = nil, want error", tc.file)
			}
		})
	}
}

// --- Content Type (MIME) Validation Tests ---

// TestContentType verifies MIME sniffing accepts text-based content and
// rejects binary formats regardless of file extension.
func TestContentType(t *testing.T) {
	t.Run("plain text", func(t *testing.T) {
		mime, err := ContentType([]byte("package main\n\nfunc main() {}\n"))
		if err != nil {
			t.Errorf("ContentType(go source) = %v, want nil", err)
		}
		if mime != "text/plain" {
			t.Errorf("mime = %q, want \"text/plain\"", mime)
		}
	})

	t.Run("html content", func(t *testing.T) {
		mime, err := ContentType([]byte("<html><body>hello</body></html>"))
		if err != nil {
			t.Errorf("ContentType(html) = %v, want nil", err)
		}
		if mime != "text/html" {
			t.Errorf("mime = %q, want \"text/html\"", mime)
		}
	})

	t.Run("json content", func(t *testing.T) {
		// http.DetectContentType returns "text/plain" for JSON, not "application/json".
		// This is expected behavior — JSON is text-based and passes the whitelist.
		_, err := ContentType([]byte(`{"key": "value"}`))
		if err != nil {
			t.Errorf("ContentType(json) = %v, want nil", err)
		}
	})

	t.Run("empty content", func(t *testing.T) {
		_, err := ContentType([]byte{})
		if err == nil {
			t.Error("ContentType(empty) = nil, want error")
		}
	})

	t.Run("PNG image", func(t *testing.T) {
		// PNG magic bytes: 0x89 P N G
		png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
		_, err := ContentType(png)
		if err == nil {
			t.Error("ContentType(PNG) = nil, want error")
		}
	})

	t.Run("JPEG image", func(t *testing.T) {
		// JPEG magic bytes: 0xFF 0xD8 0xFF
		jpeg := []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10}
		_, err := ContentType(jpeg)
		if err == nil {
			t.Error("ContentType(JPEG) = nil, want error")
		}
	})

	t.Run("GIF image", func(t *testing.T) {
		gif := []byte("GIF89a" + "\x00\x00\x00\x00")
		_, err := ContentType(gif)
		if err == nil {
			t.Error("ContentType(GIF) = nil, want error")
		}
	})

	t.Run("ZIP archive", func(t *testing.T) {
		// ZIP magic bytes: PK\x03\x04
		zip := []byte{0x50, 0x4b, 0x03, 0x04, 0x00, 0x00}
		_, err := ContentType(zip)
		if err == nil {
			t.Error("ContentType(ZIP) = nil, want error")
		}
	})

	t.Run("gzip archive", func(t *testing.T) {
		gz := []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00}
		_, err := ContentType(gz)
		if err == nil {
			t.Error("ContentType(gzip) = nil, want error")
		}
	})

	t.Run("PDF document", func(t *testing.T) {
		pdf := []byte("%PDF-1.4 fake pdf content here")
		_, err := ContentType(pdf)
		if err == nil {
			t.Error("ContentType(PDF) = nil, want error")
		}
	})
}

// --- Deep Content Scan Tests ---

// TestContent verifies the content scanner detects binary magic bytes,
// null bytes, and shebangs in inappropriate file types.
func TestContent(t *testing.T) {
	t.Run("valid go source", func(t *testing.T) {
		if err := Content([]byte("package main\n\nfunc main() {}\n"), "main.go"); err != nil {
			t.Errorf("Content(go source) = %v, want nil", err)
		}
	})

	t.Run("valid python with shebang", func(t *testing.T) {
		// Python files may legitimately start with a shebang.
		if err := Content([]byte("#!/usr/bin/env python3\nprint('hello')\n"), "script.py"); err != nil {
			t.Errorf("Content(python shebang) = %v, want nil", err)
		}
	})

	t.Run("valid ruby with shebang", func(t *testing.T) {
		if err := Content([]byte("#!/usr/bin/env ruby\nputs 'hello'\n"), "script.rb"); err != nil {
			t.Errorf("Content(ruby shebang) = %v, want nil", err)
		}
	})

	t.Run("empty content", func(t *testing.T) {
		if err := Content([]byte{}, "file.go"); err == nil {
			t.Error("Content(empty) = nil, want error")
		}
	})

	// Binary magic byte detection tests.
	t.Run("ELF binary", func(t *testing.T) {
		elf := append([]byte{0x7f, 'E', 'L', 'F'}, make([]byte, 20)...)
		if err := Content(elf, "main.go"); err == nil {
			t.Error("Content(ELF) = nil, want error")
		}
	})

	t.Run("PE executable (MZ header)", func(t *testing.T) {
		pe := append([]byte{0x4d, 0x5a}, make([]byte, 20)...)
		if err := Content(pe, "program.go"); err == nil {
			t.Error("Content(PE) = nil, want error")
		}
	})

	t.Run("Mach-O binary", func(t *testing.T) {
		macho := append([]byte{0xcf, 0xfa, 0xed, 0xfe}, make([]byte, 20)...)
		if err := Content(macho, "binary.go"); err == nil {
			t.Error("Content(Mach-O) = nil, want error")
		}
	})

	t.Run("Java class file", func(t *testing.T) {
		class := append([]byte{0xca, 0xfe, 0xba, 0xbe}, make([]byte, 20)...)
		if err := Content(class, "Main.go"); err == nil {
			t.Error("Content(Java class) = nil, want error")
		}
	})

	t.Run("ZIP archive bytes", func(t *testing.T) {
		zip := append([]byte{0x50, 0x4b, 0x03, 0x04}, make([]byte, 20)...)
		if err := Content(zip, "archive.txt"); err == nil {
			t.Error("Content(ZIP) = nil, want error")
		}
	})

	t.Run("gzip archive bytes", func(t *testing.T) {
		gz := append([]byte{0x1f, 0x8b}, make([]byte, 20)...)
		if err := Content(gz, "data.txt"); err == nil {
			t.Error("Content(gzip) = nil, want error")
		}
	})

	t.Run("PDF header", func(t *testing.T) {
		pdf := []byte("%PDF-1.4 fake content")
		if err := Content(pdf, "doc.txt"); err == nil {
			t.Error("Content(PDF) = nil, want error")
		}
	})

	t.Run("RAR archive", func(t *testing.T) {
		rar := append([]byte{0x52, 0x61, 0x72, 0x21, 0x1a}, make([]byte, 20)...)
		if err := Content(rar, "archive.txt"); err == nil {
			t.Error("Content(RAR) = nil, want error")
		}
	})

	// Null byte detection tests.
	t.Run("null bytes in content", func(t *testing.T) {
		data := []byte("normal text\x00hidden binary")
		if err := Content(data, "file.go"); err == nil {
			t.Error("Content(null bytes) = nil, want error")
		}
	})

	t.Run("null byte at start", func(t *testing.T) {
		data := []byte{0x00, 'h', 'e', 'l', 'l', 'o'}
		if err := Content(data, "file.txt"); err == nil {
			t.Error("Content(null at start) = nil, want error")
		}
	})

	// Shebang detection tests.
	t.Run("shebang in go file", func(t *testing.T) {
		// Go files should not have shebangs.
		data := []byte("#!/bin/bash\necho 'pwned'\n")
		if err := Content(data, "main.go"); err == nil {
			t.Error("Content(shebang in .go) = nil, want error")
		}
	})

	t.Run("shebang in txt file", func(t *testing.T) {
		data := []byte("#!/usr/bin/env bash\nrm -rf /\n")
		if err := Content(data, "readme.txt"); err == nil {
			t.Error("Content(shebang in .txt) = nil, want error")
		}
	})

	t.Run("shebang in js file allowed", func(t *testing.T) {
		// Node.js files may use shebangs.
		data := []byte("#!/usr/bin/env node\nconsole.log('hello')\n")
		if err := Content(data, "cli.js"); err != nil {
			t.Errorf("Content(shebang in .js) = %v, want nil", err)
		}
	})

	t.Run("shebang in perl file allowed", func(t *testing.T) {
		data := []byte("#!/usr/bin/perl\nprint 'hello';\n")
		if err := Content(data, "script.pl"); err != nil {
			t.Errorf("Content(shebang in .pl) = %v, want nil", err)
		}
	})
}

// --- Repo URL Validation Tests ---

// TestRepoURL verifies the repo URL validator accepts valid GitHub HTTPS
// URLs and rejects invalid ones.
func TestRepoURL(t *testing.T) {
	// Valid URLs: HTTPS, github.com, at least owner/repo segments.
	valid := []string{
		"https://github.com/user/project",
		"https://github.com/org/repo-name",
		"https://github.com/user/repo/tree/main",
	}
	for _, u := range valid {
		if err := RepoURL(u); err != nil {
			t.Errorf("RepoURL(%q) = %v, want nil", u, err)
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
			if err := RepoURL(tc.url); err == nil {
				t.Errorf("RepoURL(%q) = nil, want error", tc.url)
			}
		})
	}
}

// --- CID Validation Tests ---

// TestCID verifies the CID validator accepts valid IPFS CIDs and
// rejects invalid ones (empty, too short, injection characters).
func TestCID(t *testing.T) {
	// Valid CIDs: alphanumeric, 46-59 characters (CIDv0 and CIDv1 formats).
	valid := []string{
		"QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco",
		"QmT5NvUtoM5nWFfrQdVrFtvGfKFmG7AHE8P34isapyhCxX",
	}
	for _, cid := range valid {
		if err := CID(cid); err != nil {
			t.Errorf("CID(%q) = %v, want nil", cid, err)
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
			if err := CID(tc.cid); err == nil {
				t.Errorf("CID(%q) = nil, want error", tc.cid)
			}
		})
	}
}
