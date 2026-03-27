"""
Cross-language quality rules for the PoSA code quality engine.

These rules detect issues that apply across most programming languages:
  - Formatting: inconsistent whitespace, trailing spaces, mixed indentation
  - Incompleteness: stub functions, placeholder code, missing implementations
  - Wrong code: unreachable code, dead code, logic errors
  - Outdated/Deprecated: legacy patterns, deprecated APIs
  - Obsolete: superseded approaches, end-of-life references
"""

import re

# --- Formatting Rules ---
# Applied to all code files to catch whitespace and structure issues.

FORMATTING_RULES = [
    {
        "pattern": re.compile(r"\t |\t  | \t"),
        "issue_type": "formatting",
        "severity": "low",
        "message": "Mixed tabs and spaces — use consistent indentation",
    },
    {
        "pattern": re.compile(r"[ \t]+$"),
        "issue_type": "formatting",
        "severity": "low",
        "message": "Trailing whitespace — remove trailing spaces/tabs",
    },
    {
        "pattern": re.compile(r"^.{200,}$"),
        "issue_type": "formatting",
        "severity": "low",
        "message": "Line exceeds 200 characters — consider breaking into shorter lines",
    },
    {
        "pattern": re.compile(r"^\s*\n\s*\n\s*\n"),
        "issue_type": "formatting",
        "severity": "low",
        "message": "Multiple consecutive blank lines — reduce to a single blank line",
    },
]

# --- Incompleteness Rules ---
# Detect placeholder code, stubs, and missing implementations.

INCOMPLETENESS_RULES = [
    {
        "pattern": re.compile(
            r"(?:pass\s*$|return\s+None\s*$|\{\s*\}|NotImplementedError|"
            r"raise\s+NotImplementedError|todo!|unimplemented!|panic!\(\"not implemented)",
            re.IGNORECASE,
        ),
        "issue_type": "incompleteness",
        "severity": "medium",
        "message": "Stub or placeholder implementation — complete before shipping",
    },
    {
        "pattern": re.compile(
            r"(?:placeholder|dummy|fake|mock|temp|tmp|xxx|fixme)\s*(?:=|:|\()",
            re.IGNORECASE,
        ),
        "issue_type": "incompleteness",
        "severity": "medium",
        "message": "Placeholder variable or function name — replace with meaningful implementation",
    },
    {
        "pattern": re.compile(r"(?://|#|--|;)\s*(?:HACK|WORKAROUND|KLUDGE|BODGE)", re.IGNORECASE),
        "issue_type": "incompleteness",
        "severity": "medium",
        "message": "Workaround/hack comment — refactor to a proper solution",
    },
    {
        "pattern": re.compile(
            r'(?:return|print|echo|puts|console\.log)\s*["\'](?:test|debug|hello|foo|bar)', re.IGNORECASE
        ),
        "issue_type": "incompleteness",
        "severity": "low",
        "message": "Debug/test output left in code — remove before shipping",
    },
]

# --- Wrong Code / Logic Error Rules ---
# Detect patterns that are almost certainly bugs.

WRONG_CODE_RULES = [
    {
        "pattern": re.compile(r"if\s*\(\s*\w+\s*=\s*(?!=)\w+"),
        "issue_type": "logic",
        "severity": "high",
        "message": "Assignment in conditional (= instead of ==) — likely a bug",
    },
    {
        "pattern": re.compile(r"(?:return|break|exit|throw|raise)\s+.+\n\s*\S"),
        "issue_type": "logic",
        "severity": "medium",
        "message": "Code after return/break/exit — unreachable code detected",
    },
    {
        "pattern": re.compile(r"(?:if|while|for)\s*\(\s*(?:true|false|1|0)\s*\)"),
        "issue_type": "logic",
        "severity": "medium",
        "message": "Constant condition in control flow — likely a bug or leftover debug code",
    },
    {
        "pattern": re.compile(r"\b(\w+)\s*==\s*\1\b"),
        "issue_type": "logic",
        "severity": "medium",
        "message": "Comparing variable to itself — always true, likely a copy-paste error",
    },
    {
        "pattern": re.compile(r"/\s*0(?:\s*[;,)\]]|$)"),
        "issue_type": "logic",
        "severity": "high",
        "message": "Division by zero — will cause runtime error",
    },
]

# --- Outdated / Deprecated Rules ---
# Detect usage of deprecated APIs and patterns across languages.

DEPRECATED_RULES = [
    # Python
    {
        "pattern": re.compile(r"from\s+distutils\b"),
        "issue_type": "deprecated",
        "severity": "medium",
        "message": "distutils is deprecated (removed in Python 3.12) — use setuptools",
    },
    {
        "pattern": re.compile(r"\boptparse\b"),
        "issue_type": "deprecated",
        "severity": "low",
        "message": "optparse is deprecated — use argparse instead",
    },
    {
        "pattern": re.compile(r"\basyncio\.coroutine\b"),
        "issue_type": "deprecated",
        "severity": "medium",
        "message": "asyncio.coroutine is deprecated — use async/await syntax",
    },
    {
        "pattern": re.compile(r"\.format\s*\(.*\).*\.format\s*\("),
        "issue_type": "deprecated",
        "severity": "low",
        "message": "Chained .format() calls — use f-strings for cleaner string formatting",
    },
    # JavaScript
    {
        "pattern": re.compile(r"__proto__\s*="),
        "issue_type": "deprecated",
        "severity": "high",
        "message": "__proto__ assignment is deprecated — use Object.create() or Object.setPrototypeOf()",
    },
    {
        "pattern": re.compile(r"arguments\.callee\b"),
        "issue_type": "deprecated",
        "severity": "medium",
        "message": "arguments.callee is deprecated — use named function expressions",
    },
    {
        "pattern": re.compile(r"document\.all\b"),
        "issue_type": "deprecated",
        "severity": "medium",
        "message": "document.all is deprecated — use document.getElementById or querySelector",
    },
    # Java
    {
        "pattern": re.compile(r"new\s+Date\s*\(\s*\d+\s*,"),
        "issue_type": "deprecated",
        "severity": "medium",
        "message": "Date constructor with year/month is deprecated — use java.time API",
    },
    {
        "pattern": re.compile(r"\.finalize\s*\(\s*\)"),
        "issue_type": "deprecated",
        "severity": "medium",
        "message": "finalize() is deprecated since Java 9 — use Cleaner or try-with-resources",
    },
    # Go
    {
        "pattern": re.compile(r"ioutil\.\w+"),
        "issue_type": "deprecated",
        "severity": "medium",
        "message": "ioutil package is deprecated since Go 1.16 — use io and os packages",
    },
    # C/C++
    {
        "pattern": re.compile(r"\bgets\s*\("),
        "issue_type": "deprecated",
        "severity": "high",
        "message": "gets() was removed in C11 — use fgets() instead",
    },
    {
        "pattern": re.compile(r"\bauto_ptr\b"),
        "issue_type": "deprecated",
        "severity": "medium",
        "message": "auto_ptr is deprecated in C++11 — use unique_ptr or shared_ptr",
    },
    # Ruby
    {
        "pattern": re.compile(r"require\s+['\"]rubygems['\"]"),
        "issue_type": "deprecated",
        "severity": "low",
        "message": "require 'rubygems' is unnecessary since Ruby 1.9",
    },
    # PHP
    {
        "pattern": re.compile(r"\bmysql_\w+\s*\("),
        "issue_type": "deprecated",
        "severity": "high",
        "message": "mysql_* functions removed in PHP 7 — use mysqli or PDO",
    },
    {
        "pattern": re.compile(r"\bereg\s*\("),
        "issue_type": "deprecated",
        "severity": "medium",
        "message": "ereg() removed in PHP 7 — use preg_match()",
    },
]

# --- Obsolete Codebase Rules ---
# Detect patterns indicating an outdated or unmaintained codebase.

OBSOLETE_RULES = [
    {
        "pattern": re.compile(r"Copyright\s+(?:19\d{2}|200\d|201[0-7])\b"),
        "issue_type": "obsolete",
        "severity": "low",
        "message": "Outdated copyright year — update to current year",
    },
    {
        "pattern": re.compile(
            r"(?:python|node|ruby|java|go)\s*(?:>=?\s*)?(?:2\.[0-6]|1\.[0-8]|0\.\d+)",
            re.IGNORECASE,
        ),
        "issue_type": "obsolete",
        "severity": "medium",
        "message": "Reference to end-of-life language version — update to a supported version",
    },
    {
        "pattern": re.compile(r"(?:bower|grunt|gulp)(?:file|\.json)", re.IGNORECASE),
        "issue_type": "obsolete",
        "severity": "low",
        "message": "Legacy build tool reference — consider migrating to modern tooling",
    },
    {
        "pattern": re.compile(r"(?:jQuery|\$)\s*\.\s*(?:ajax|get|post)\s*\("),
        "issue_type": "obsolete",
        "severity": "low",
        "message": "jQuery AJAX — consider using fetch() or axios for modern projects",
    },
    {
        "pattern": re.compile(r"var\s+React\s*=\s*require\s*\("),
        "issue_type": "obsolete",
        "severity": "low",
        "message": "CommonJS require for React — use ES module import syntax",
    },
    {
        "pattern": re.compile(r"componentWillMount|componentWillReceiveProps|componentWillUpdate"),
        "issue_type": "obsolete",
        "severity": "medium",
        "message": "Deprecated React lifecycle method — use modern alternatives (useEffect, getDerivedStateFromProps)",
    },
]
