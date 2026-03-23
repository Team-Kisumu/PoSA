"""
Language-specific analysis rules for the PoSA code quality engine.

Each rule is a dict with:
  - pattern: regex pattern to match against source code lines
  - issue_type: category ("security", "quality", "logic", "style")
  - severity: weight used in scoring ("high"=15, "medium"=8, "low"=3)
  - message: human-readable description of the issue

Rules are grouped by language. The analyzer detects the language from
the file extension and applies the corresponding rule set.
"""

import re

# Severity weights used to calculate the final score.
# Each matched rule deducts its severity weight from 100.
SEVERITY_WEIGHT = {
    "high": 15,
    "medium": 8,
    "low": 3,
}

# Map file extensions to language identifiers.
EXTENSION_MAP = {
    ".go": "go",
    ".py": "python",
    ".js": "javascript",
    ".jsx": "javascript",
    ".ts": "javascript",
    ".tsx": "javascript",
    ".mjs": "javascript",
    ".cjs": "javascript",
}

# Supported languages for analysis.
SUPPORTED_LANGUAGES = set(EXTENSION_MAP.values())


# --- Go Rules ---

GO_RULES = [
    # Security
    {
        "pattern": re.compile(r'fmt\.Sprintf\s*\(\s*"[^"]*%s[^"]*"\s*,.*\bRequest\b', re.IGNORECASE),
        "issue_type": "security",
        "severity": "high",
        "message": "Potential format string injection — user input in fmt.Sprintf",
    },
    {
        "pattern": re.compile(r'exec\.Command\s*\('),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of exec.Command — verify input is sanitized to prevent command injection",
    },
    {
        "pattern": re.compile(r'\.(?:Query|Exec|QueryRow)\s*\(.*\+'),
        "issue_type": "security",
        "severity": "high",
        "message": "Potential SQL injection — use parameterized queries instead of string concatenation",
    },
    {
        "pattern": re.compile(r'http\.ListenAndServe\s*\(\s*"[^"]*"\s*,\s*nil\s*\)'),
        "issue_type": "security",
        "severity": "medium",
        "message": "Using default ServeMux (nil handler) — consider using a custom mux for route isolation",
    },
    {
        "pattern": re.compile(r'\.Write\s*\(\s*\[\]byte\s*\(\s*r\.(URL|Body|Form)'),
        "issue_type": "security",
        "severity": "high",
        "message": "Writing unsanitized request data directly to response — potential XSS",
    },
    # Quality
    {
        "pattern": re.compile(r'fmt\.Println\s*\('),
        "issue_type": "quality",
        "severity": "low",
        "message": "Use of fmt.Println — consider using structured logging (log.Printf or slog)",
    },
    {
        "pattern": re.compile(r'panic\s*\('),
        "issue_type": "quality",
        "severity": "medium",
        "message": "Use of panic — return errors instead for graceful error handling",
    },
    {
        "pattern": re.compile(r'//\s*TODO|//\s*FIXME|//\s*HACK|//\s*XXX'),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
    {
        "pattern": re.compile(r'\b_\s*=\s*\w+\.\w+\('),
        "issue_type": "quality",
        "severity": "medium",
        "message": "Discarded error return value — handle or explicitly document why it is safe to ignore",
    },
    {
        "pattern": re.compile(r'defer\s+\w+\.Close\(\)'),
        "issue_type": "quality",
        "severity": "low",
        "message": "Deferred Close without error check — consider defer with error handling",
    },
]

# --- Python Rules ---

PYTHON_RULES = [
    # Security
    {
        "pattern": re.compile(r'\beval\s*\('),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of eval() — arbitrary code execution risk",
    },
    {
        "pattern": re.compile(r'\bexec\s*\('),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of exec() — arbitrary code execution risk",
    },
    {
        "pattern": re.compile(r'subprocess\.\w+\s*\(.*shell\s*=\s*True'),
        "issue_type": "security",
        "severity": "high",
        "message": "subprocess with shell=True — command injection risk",
    },
    {
        "pattern": re.compile(r'os\.system\s*\('),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of os.system() — use subprocess with shell=False instead",
    },
    {
        "pattern": re.compile(r'pickle\.loads?\s*\('),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of pickle.load/loads — deserialization of untrusted data can execute arbitrary code",
    },
    {
        "pattern": re.compile(r'__import__\s*\('),
        "issue_type": "security",
        "severity": "medium",
        "message": "Dynamic import via __import__() — verify input is trusted",
    },
    {
        "pattern": re.compile(r'yaml\.load\s*\((?!.*Loader)'),
        "issue_type": "security",
        "severity": "high",
        "message": "yaml.load without safe Loader — use yaml.safe_load() instead",
    },
    # Quality
    {
        "pattern": re.compile(r'except\s*:\s*$', re.MULTILINE),
        "issue_type": "quality",
        "severity": "medium",
        "message": "Bare except clause — catch specific exceptions instead",
    },
    {
        "pattern": re.compile(r'except\s+Exception\s*:\s*\n\s*pass', re.MULTILINE),
        "issue_type": "quality",
        "severity": "medium",
        "message": "Silenced exception (except Exception: pass) — log or handle the error",
    },
    {
        "pattern": re.compile(r'#\s*TODO|#\s*FIXME|#\s*HACK|#\s*XXX'),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
    {
        "pattern": re.compile(r'import\s+\*'),
        "issue_type": "quality",
        "severity": "low",
        "message": "Wildcard import — import specific names to avoid namespace pollution",
    },
    {
        "pattern": re.compile(r'print\s*\('),
        "issue_type": "style",
        "severity": "low",
        "message": "Use of print() — consider using the logging module for production code",
    },
]

# --- JavaScript Rules ---

JAVASCRIPT_RULES = [
    # Security
    {
        "pattern": re.compile(r'\beval\s*\('),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of eval() — arbitrary code execution risk",
    },
    {
        "pattern": re.compile(r'innerHTML\s*='),
        "issue_type": "security",
        "severity": "high",
        "message": "Direct innerHTML assignment — potential XSS vulnerability, use textContent or sanitize",
    },
    {
        "pattern": re.compile(r'document\.write\s*\('),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of document.write() — XSS risk and performance issues",
    },
    {
        "pattern": re.compile(r'child_process\.\w+\s*\('),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of child_process — verify input is sanitized to prevent command injection",
    },
    {
        "pattern": re.compile(r'new\s+Function\s*\('),
        "issue_type": "security",
        "severity": "high",
        "message": "new Function() constructor — equivalent to eval(), arbitrary code execution risk",
    },
    {
        "pattern": re.compile(r'\.createObjectURL\s*\('),
        "issue_type": "security",
        "severity": "medium",
        "message": "createObjectURL without revocation — potential memory leak and blob URL abuse",
    },
    # Quality
    {
        "pattern": re.compile(r'\bvar\s+'),
        "issue_type": "quality",
        "severity": "low",
        "message": "Use of var — prefer let or const for block scoping",
    },
    {
        "pattern": re.compile(r'==(?!=)'),
        "issue_type": "quality",
        "severity": "low",
        "message": "Loose equality (==) — use strict equality (===) to avoid type coercion",
    },
    {
        "pattern": re.compile(r'console\.log\s*\('),
        "issue_type": "style",
        "severity": "low",
        "message": "console.log left in code — remove or replace with proper logging",
    },
    {
        "pattern": re.compile(r'//\s*TODO|//\s*FIXME|//\s*HACK|//\s*XXX'),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
    {
        "pattern": re.compile(r'setTimeout\s*\(\s*["\']'),
        "issue_type": "security",
        "severity": "medium",
        "message": "setTimeout with string argument — equivalent to eval(), use a function reference",
    },
]

# Map language identifiers to their rule sets.
LANGUAGE_RULES = {
    "go": GO_RULES,
    "python": PYTHON_RULES,
    "javascript": JAVASCRIPT_RULES,
}
