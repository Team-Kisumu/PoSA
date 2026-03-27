"""
Analysis rules for compiled and systems languages.

Covers: C/C++, Java, C#, Ruby, Rust, PHP, Swift, Shell/Bash.
"""

import re

# --- C / C++ Rules ---

C_RULES = [
    {
        "pattern": re.compile(r"\b(?:gets|scanf)\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of gets/scanf — buffer overflow risk, use fgets/sscanf with size limits",
    },
    {
        "pattern": re.compile(r"\bstrcpy\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of strcpy — buffer overflow risk, use strncpy or strlcpy",
    },
    {
        "pattern": re.compile(r"\bstrcat\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of strcat — buffer overflow risk, use strncat or strlcat",
    },
    {
        "pattern": re.compile(r"\bsprintf\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of sprintf — buffer overflow risk, use snprintf",
    },
    {
        "pattern": re.compile(r"\bsystem\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of system() — command injection risk, use exec family instead",
    },
    {
        "pattern": re.compile(r"\bmalloc\s*\(.*\)(?!.*\bfree\b)"),
        "issue_type": "quality",
        "severity": "medium",
        "message": "malloc without visible free — potential memory leak",
    },
    {
        "pattern": re.compile(r"//\s*TODO|//\s*FIXME|//\s*HACK|//\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
    {
        "pattern": re.compile(r"\bprintf\s*\(\s*[a-zA-Z_]"),
        "issue_type": "security",
        "severity": "medium",
        "message": "printf with variable as format string — potential format string vulnerability",
    },
]

# --- Java Rules ---

JAVA_RULES = [
    {
        "pattern": re.compile(r"Runtime\.getRuntime\(\)\.exec\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Runtime.exec() — command injection risk, validate and sanitize input",
    },
    {
        "pattern": re.compile(r'Statement\s*.*\.\s*execute(?:Query|Update)?\s*\(\s*".*\+'),
        "issue_type": "security",
        "severity": "high",
        "message": "SQL string concatenation — use PreparedStatement to prevent SQL injection",
    },
    {
        "pattern": re.compile(r"ObjectInputStream\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "ObjectInputStream deserialization — untrusted data can execute arbitrary code",
    },
    {
        "pattern": re.compile(r"\.printStackTrace\s*\("),
        "issue_type": "quality",
        "severity": "medium",
        "message": "printStackTrace() — use a logging framework instead of printing stack traces",
    },
    {
        "pattern": re.compile(r"catch\s*\(\s*Exception\s+\w+\s*\)\s*\{?\s*\}"),
        "issue_type": "quality",
        "severity": "medium",
        "message": "Empty catch block — handle or log the exception",
    },
    {
        "pattern": re.compile(r"System\.out\.print"),
        "issue_type": "style",
        "severity": "low",
        "message": "System.out.print — use a logging framework for production code",
    },
    {
        "pattern": re.compile(r"//\s*TODO|//\s*FIXME|//\s*HACK|//\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
    {
        "pattern": re.compile(r"\bnew\s+Random\s*\("),
        "issue_type": "security",
        "severity": "medium",
        "message": "java.util.Random is not cryptographically secure — use SecureRandom for security contexts",
    },
]

# --- C# Rules ---

CSHARP_RULES = [
    {
        "pattern": re.compile(r"Process\.Start\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Process.Start — command injection risk, validate input",
    },
    {
        "pattern": re.compile(r'SqlCommand\s*\(.*".*\+'),
        "issue_type": "security",
        "severity": "high",
        "message": "SQL string concatenation — use parameterized queries",
    },
    {
        "pattern": re.compile(r"BinaryFormatter\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "BinaryFormatter deserialization — security risk, use System.Text.Json",
    },
    {
        "pattern": re.compile(r"Console\.Write"),
        "issue_type": "style",
        "severity": "low",
        "message": "Console.Write — use ILogger for production code",
    },
    {
        "pattern": re.compile(r"catch\s*\(\s*Exception\s*\)\s*\{?\s*\}"),
        "issue_type": "quality",
        "severity": "medium",
        "message": "Empty catch block — handle or log the exception",
    },
    {
        "pattern": re.compile(r"//\s*TODO|//\s*FIXME|//\s*HACK|//\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
]

# --- Ruby Rules ---

RUBY_RULES = [
    {
        "pattern": re.compile(r"\beval\s*[\(]"),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of eval — arbitrary code execution risk",
    },
    {
        "pattern": re.compile(r"\bsystem\s*[\(]"),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of system() — command injection risk, use Open3 or shellescape",
    },
    {
        "pattern": re.compile(r"`[^`]*`"),
        "issue_type": "security",
        "severity": "high",
        "message": "Backtick command execution — command injection risk",
    },
    {
        "pattern": re.compile(r"Marshal\.load\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Marshal.load — deserialization of untrusted data can execute arbitrary code",
    },
    {
        "pattern": re.compile(r"\bputs\s+"),
        "issue_type": "style",
        "severity": "low",
        "message": "Use of puts — consider using a logger for production code",
    },
    {
        "pattern": re.compile(r"#\s*TODO|#\s*FIXME|#\s*HACK|#\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
    {
        "pattern": re.compile(r"rescue\s*$", re.MULTILINE),
        "issue_type": "quality",
        "severity": "medium",
        "message": "Bare rescue — catch specific exceptions instead",
    },
]

# --- Rust Rules ---

RUST_RULES = [
    {
        "pattern": re.compile(r"\bunsafe\s*\{"),
        "issue_type": "security",
        "severity": "medium",
        "message": "unsafe block — review carefully for memory safety violations",
    },
    {
        "pattern": re.compile(r"\.unwrap\s*\("),
        "issue_type": "quality",
        "severity": "medium",
        "message": "unwrap() — will panic on None/Err, use match or ? operator",
    },
    {
        "pattern": re.compile(r"\.expect\s*\("),
        "issue_type": "quality",
        "severity": "low",
        "message": "expect() — will panic with message on None/Err, consider proper error handling",
    },
    {
        "pattern": re.compile(r"\bprintln!\s*\("),
        "issue_type": "style",
        "severity": "low",
        "message": "println! — consider using a logging crate (log, tracing) for production",
    },
    {
        "pattern": re.compile(r"//\s*TODO|//\s*FIXME|//\s*HACK|//\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
    {
        "pattern": re.compile(r"std::process::Command::new\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Command::new — verify input is sanitized to prevent command injection",
    },
]

# --- PHP Rules ---

PHP_RULES = [
    {
        "pattern": re.compile(r"\beval\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of eval() — arbitrary code execution risk",
    },
    {
        "pattern": re.compile(r"\bexec\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of exec() — command injection risk",
    },
    {
        "pattern": re.compile(r"\bshell_exec\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of shell_exec() — command injection risk",
    },
    {
        "pattern": re.compile(r"\bsystem\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of system() — command injection risk",
    },
    {
        "pattern": re.compile(r"\bmysqli?_query\s*\(.*\$"),
        "issue_type": "security",
        "severity": "high",
        "message": "SQL query with variable interpolation — use prepared statements",
    },
    {
        "pattern": re.compile(r"\bunserialize\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "unserialize() — deserialization of untrusted data can execute arbitrary code",
    },
    {
        "pattern": re.compile(r"\b(?:echo|print)\s+\$_(?:GET|POST|REQUEST|COOKIE)"),
        "issue_type": "security",
        "severity": "high",
        "message": "Outputting unsanitized user input — XSS risk, use htmlspecialchars()",
    },
    {
        "pattern": re.compile(r"//\s*TODO|//\s*FIXME|//\s*HACK|//\s*XXX|#\s*TODO|#\s*FIXME"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
]

# --- Swift Rules ---

SWIFT_RULES = [
    {
        "pattern": re.compile(r"try!\s+"),
        "issue_type": "quality",
        "severity": "medium",
        "message": "Force try (try!) — will crash on error, use do/catch or try?",
    },
    {
        "pattern": re.compile(r"!\s*$|!\.", re.MULTILINE),
        "issue_type": "quality",
        "severity": "medium",
        "message": "Force unwrap (!) — will crash on nil, use optional binding or guard",
    },
    {
        "pattern": re.compile(r"\bprint\s*\("),
        "issue_type": "style",
        "severity": "low",
        "message": "print() — use os_log or a logging framework for production",
    },
    {
        "pattern": re.compile(r"Process\s*\(\)"),
        "issue_type": "security",
        "severity": "high",
        "message": "Process() — command execution, verify input is sanitized",
    },
    {
        "pattern": re.compile(r"//\s*TODO|//\s*FIXME|//\s*HACK|//\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
]

# --- Shell / Bash Rules ---

SHELL_RULES = [
    {
        "pattern": re.compile(r"\beval\s+"),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of eval — arbitrary code execution risk",
    },
    {
        "pattern": re.compile(r"\$\{?\w+\}?\s*(?:&&|\|\||;)"),
        "issue_type": "security",
        "severity": "medium",
        "message": "Unquoted variable in command chain — potential injection if variable is user-controlled",
    },
    {
        "pattern": re.compile(r"chmod\s+777\b"),
        "issue_type": "security",
        "severity": "high",
        "message": "chmod 777 — world-writable permissions, use more restrictive permissions",
    },
    {
        "pattern": re.compile(r"curl\s+.*\|\s*(?:bash|sh)"),
        "issue_type": "security",
        "severity": "high",
        "message": "Piping curl to shell — remote code execution risk",
    },
    {
        "pattern": re.compile(r"#\s*TODO|#\s*FIXME|#\s*HACK|#\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
    {
        "pattern": re.compile(r"^\s*cd\s+[^&|;]+(?:&&|\|\|)", re.MULTILINE),
        "issue_type": "quality",
        "severity": "low",
        "message": "cd in command chain — use subshell (cd dir && cmd) or check exit status",
    },
]
