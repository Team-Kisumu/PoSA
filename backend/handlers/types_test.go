package handlers

import "testing"

func TestValidateFilename(t *testing.T) {
	valid := []string{
		"main.go", "test.py", "README.md", "file-name.txt",
		"code_v2.js", "a.c",
	}
	for _, name := range valid {
		if err := ValidateFilename(name); err != nil {
			t.Errorf("ValidateFilename(%q) = %v, want nil", name, err)
		}
	}

	invalid := []struct {
		name string
		file string
	}{
		{"empty", ""},
		{"null byte", "file\x00.txt"},
		{"forward slash", "path/file.txt"},
		{"backslash", "path\\file.txt"},
		{"dot dot slash", "../../etc/passwd"},
		{"exe", "payload.exe"},
		{"bat", "script.bat"},
		{"sh", "run.sh"},
		{"ps1", "script.ps1"},
		{"dll", "lib.dll"},
		{"starts with dot", ".hidden"},
		{"starts with dash", "-flag.txt"},
		{"special chars", "file<>.txt"},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateFilename(tc.file); err == nil {
				t.Errorf("ValidateFilename(%q) = nil, want error", tc.file)
			}
		})
	}
}

func TestValidateRepoURL(t *testing.T) {
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

	invalid := []struct {
		name string
		url  string
	}{
		{"empty", ""},
		{"whitespace", "   "},
		{"http", "http://github.com/user/repo"},
		{"gitlab", "https://gitlab.com/user/repo"},
		{"no repo", "https://github.com/user"},
		{"just domain", "https://github.com/"},
		{"path traversal", "https://github.com/../../etc/passwd"},
		{"bare string", "not-a-url"},
		{"empty owner", "https://github.com//repo"},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateRepoURL(tc.url); err == nil {
				t.Errorf("ValidateRepoURL(%q) = nil, want error", tc.url)
			}
		})
	}
}

func TestValidateCID(t *testing.T) {
	valid := []string{
		"QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco",
		"QmT5NvUtoM5nWFfrQdVrFtvGfKFmG7AHE8P34isapyhCxX",
	}
	for _, cid := range valid {
		if err := ValidateCID(cid); err != nil {
			t.Errorf("ValidateCID(%q) = %v, want nil", cid, err)
		}
	}

	invalid := []struct {
		name string
		cid  string
	}{
		{"empty", ""},
		{"too short", "Qm123"},
		{"semicolon injection", "abc;rm -rf /"},
		{"null byte", "Qm\x00abc"},
		{"pipe", "abc|cat /etc/passwd"},
		{"angle brackets", "abc<script>"},
		{"quotes", `abc"onload=`},
		{"slash", "abc/def"},
		{"backslash", "abc\\def"},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateCID(tc.cid); err == nil {
				t.Errorf("ValidateCID(%q) = nil, want error", tc.cid)
			}
		})
	}
}
