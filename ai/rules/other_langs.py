"""
Analysis rules for scripting, functional, and data languages.

Covers: Perl, R, Lua, Dart, Elixir, Erlang, Haskell, Clojure,
Scala, Kotlin, Groovy, SQL, CSS, and config/data formats.
"""

import re

# --- Perl Rules ---

PERL_RULES = [
    {
        "pattern": re.compile(r"\beval\s*[\({]"),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of eval — arbitrary code execution risk",
    },
    {
        "pattern": re.compile(r"\bsystem\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of system() — command injection risk",
    },
    {
        "pattern": re.compile(r"`[^`]+`"),
        "issue_type": "security",
        "severity": "high",
        "message": "Backtick command execution — command injection risk",
    },
    {
        "pattern": re.compile(r"#\s*TODO|#\s*FIXME|#\s*HACK|#\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
]

# --- R Rules ---

R_RULES = [
    {
        "pattern": re.compile(r"\beval\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of eval() — arbitrary code execution risk",
    },
    {
        "pattern": re.compile(r"\bsystem\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of system() — command injection risk",
    },
    {
        "pattern": re.compile(r"\bsystem2\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of system2() — command injection risk, validate input",
    },
    {
        "pattern": re.compile(r"#\s*TODO|#\s*FIXME|#\s*HACK|#\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
]

# --- Lua Rules ---

LUA_RULES = [
    {
        "pattern": re.compile(r"\bloadstring\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of loadstring — arbitrary code execution risk",
    },
    {
        "pattern": re.compile(r"\bos\.execute\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of os.execute() — command injection risk",
    },
    {
        "pattern": re.compile(r"\bio\.popen\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of io.popen() — command injection risk",
    },
    {
        "pattern": re.compile(r"--\s*TODO|--\s*FIXME|--\s*HACK|--\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
]

# --- Dart Rules ---

DART_RULES = [
    {
        "pattern": re.compile(r"Process\.run\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Process.run — command injection risk, validate input",
    },
    {
        "pattern": re.compile(r"\bprint\s*\("),
        "issue_type": "style",
        "severity": "low",
        "message": "print() — use a logging package for production code",
    },
    {
        "pattern": re.compile(r"//\s*TODO|//\s*FIXME|//\s*HACK|//\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
    {
        "pattern": re.compile(r"!\s*;", re.MULTILINE),
        "issue_type": "quality",
        "severity": "medium",
        "message": "Force unwrap (!) — will throw on null, use null-aware operators",
    },
]

# --- Elixir Rules ---

ELIXIR_RULES = [
    {
        "pattern": re.compile(r"Code\.eval_string\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Code.eval_string — arbitrary code execution risk",
    },
    {
        "pattern": re.compile(r"System\.cmd\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "System.cmd — command injection risk, validate input",
    },
    {
        "pattern": re.compile(r":erlang\.binary_to_term\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "binary_to_term — deserialization of untrusted data risk",
    },
    {
        "pattern": re.compile(r"#\s*TODO|#\s*FIXME|#\s*HACK|#\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
    {
        "pattern": re.compile(r"IO\.puts\s*\("),
        "issue_type": "style",
        "severity": "low",
        "message": "IO.puts — use Logger for production code",
    },
]

# --- Erlang Rules ---

ERLANG_RULES = [
    {
        "pattern": re.compile(r"\bos:cmd\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "os:cmd — command injection risk, validate input",
    },
    {
        "pattern": re.compile(r"\bbinary_to_term\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "binary_to_term — deserialization of untrusted data risk",
    },
    {
        "pattern": re.compile(r"%\s*TODO|%\s*FIXME|%\s*HACK|%\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
]

# --- Haskell Rules ---

HASKELL_RULES = [
    {
        "pattern": re.compile(r"\bunsafePerformIO\b"),
        "issue_type": "security",
        "severity": "high",
        "message": "unsafePerformIO — breaks referential transparency, use with extreme caution",
    },
    {
        "pattern": re.compile(r"\bhead\s+"),
        "issue_type": "quality",
        "severity": "medium",
        "message": "head on potentially empty list — will throw exception, use pattern matching",
    },
    {
        "pattern": re.compile(r"\bsystem\s+"),
        "issue_type": "security",
        "severity": "high",
        "message": "System.Process — command injection risk, validate input",
    },
    {
        "pattern": re.compile(r"--\s*TODO|--\s*FIXME|--\s*HACK|--\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
]

# --- Clojure Rules ---

CLOJURE_RULES = [
    {
        "pattern": re.compile(r"\(eval\s+"),
        "issue_type": "security",
        "severity": "high",
        "message": "Use of eval — arbitrary code execution risk",
    },
    {
        "pattern": re.compile(r"\(sh\s+"),
        "issue_type": "security",
        "severity": "high",
        "message": "Shell execution via clojure.java.shell/sh — command injection risk",
    },
    {
        "pattern": re.compile(r";\s*TODO|;\s*FIXME|;\s*HACK|;\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
]

# --- Scala Rules ---

SCALA_RULES = [
    {
        "pattern": re.compile(r"Runtime\.getRuntime.*\.exec\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Runtime.exec — command injection risk",
    },
    {
        "pattern": re.compile(r"\bsys\.process\b", re.IGNORECASE),
        "issue_type": "security",
        "severity": "medium",
        "message": "sys.process — command execution, validate input",
    },
    {
        "pattern": re.compile(r"\bprintln\s*\("),
        "issue_type": "style",
        "severity": "low",
        "message": "println — use a logging framework for production code",
    },
    {
        "pattern": re.compile(r"//\s*TODO|//\s*FIXME|//\s*HACK|//\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
]

# --- Kotlin Rules ---

KOTLIN_RULES = [
    {
        "pattern": re.compile(r"Runtime\.getRuntime\(\)\.exec\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "Runtime.exec — command injection risk",
    },
    {
        "pattern": re.compile(r"!!\s*[.;\n]"),
        "issue_type": "quality",
        "severity": "medium",
        "message": "Non-null assertion (!!) — will throw NPE on null, use safe calls",
    },
    {
        "pattern": re.compile(r"\bprintln\s*\("),
        "issue_type": "style",
        "severity": "low",
        "message": "println — use a logging framework for production code",
    },
    {
        "pattern": re.compile(r"//\s*TODO|//\s*FIXME|//\s*HACK|//\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
]

# --- SQL Rules ---

SQL_RULES = [
    {
        "pattern": re.compile(r"GRANT\s+ALL\s+PRIVILEGES", re.IGNORECASE),
        "issue_type": "security",
        "severity": "high",
        "message": "GRANT ALL PRIVILEGES — use least-privilege principle",
    },
    {
        "pattern": re.compile(r"SELECT\s+\*\s+FROM", re.IGNORECASE),
        "issue_type": "quality",
        "severity": "low",
        "message": "SELECT * — specify columns explicitly for clarity and performance",
    },
    {
        "pattern": re.compile(r"DROP\s+(?:TABLE|DATABASE)\s+(?!IF\s+EXISTS)", re.IGNORECASE),
        "issue_type": "quality",
        "severity": "medium",
        "message": "DROP without IF EXISTS — may fail if object doesn't exist",
    },
    {
        "pattern": re.compile(r"--\s*TODO|--\s*FIXME|--\s*HACK|--\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
    {
        "pattern": re.compile(r"PASSWORD\s*=\s*'[^']+'", re.IGNORECASE),
        "issue_type": "security",
        "severity": "high",
        "message": "Hardcoded password in SQL — use environment variables or secrets manager",
    },
]

# --- CSS Rules ---

CSS_RULES = [
    {
        "pattern": re.compile(r"expression\s*\("),
        "issue_type": "security",
        "severity": "high",
        "message": "CSS expression() — can execute JavaScript, XSS risk (IE legacy)",
    },
    {
        "pattern": re.compile(r"!important"),
        "issue_type": "quality",
        "severity": "low",
        "message": "!important — overrides specificity, use more specific selectors instead",
    },
    {
        "pattern": re.compile(r"/\*\s*TODO|/\*\s*FIXME|/\*\s*HACK|/\*\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
    {
        "pattern": re.compile(r"\*\s*\{"),
        "issue_type": "quality",
        "severity": "low",
        "message": "Universal selector (*) — may impact performance on large DOMs",
    },
]

# --- Config / Data Rules (JSON, YAML, TOML, XML, INI, .env) ---

CONFIG_RULES = [
    {
        "pattern": re.compile(
            r'(?:password|passwd|secret|api_key|apikey|token|private_key)\s*[:=]\s*["\']?[^\s"\']{4,}',
            re.IGNORECASE,
        ),
        "issue_type": "security",
        "severity": "high",
        "message": "Potential hardcoded secret — use environment variables or a secrets manager",
    },
    {
        "pattern": re.compile(r"0\.0\.0\.0|localhost:\d+|127\.0\.0\.1:\d+"),
        "issue_type": "quality",
        "severity": "low",
        "message": "Hardcoded local address — ensure this is not used in production",
    },
    {
        "pattern": re.compile(r"#\s*TODO|#\s*FIXME|//\s*TODO|//\s*FIXME|<!--\s*TODO|<!--\s*FIXME"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
]

# --- Infrastructure (Terraform/HCL) Rules ---

TERRAFORM_RULES = [
    {
        "pattern": re.compile(r'cidr_blocks\s*=\s*\[\s*"0\.0\.0\.0/0"\s*\]'),
        "issue_type": "security",
        "severity": "high",
        "message": "Security group open to 0.0.0.0/0 — restrict to specific CIDR ranges",
    },
    {
        "pattern": re.compile(
            r'(?:password|secret|api_key|token)\s*=\s*"[^"]{4,}"',
            re.IGNORECASE,
        ),
        "issue_type": "security",
        "severity": "high",
        "message": "Hardcoded secret in infrastructure code — use variables or secrets manager",
    },
    {
        "pattern": re.compile(r"#\s*TODO|#\s*FIXME|#\s*HACK|#\s*XXX"),
        "issue_type": "quality",
        "severity": "low",
        "message": "TODO/FIXME comment found — address before shipping",
    },
    {
        "pattern": re.compile(r"encrypted\s*=\s*false", re.IGNORECASE),
        "issue_type": "security",
        "severity": "medium",
        "message": "Encryption disabled — enable encryption for data at rest",
    },
]
