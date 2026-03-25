"""
Code quality analyzer for the PoSA AI engine.

Performs pattern-based static analysis on source code to detect security
vulnerabilities, quality issues, and style problems. Supports Go, Python,
and JavaScript. Returns a score (0-100), list of issues, and suggestions.

The analyzer detects the language from the file extension, applies the
corresponding rule set line-by-line, deduplicates findings, and computes
a weighted score based on issue severity.
"""

import os

from ai.rules.patterns import (
    EXTENSION_MAP,
    LANGUAGE_RULES,
    SEVERITY_WEIGHT,
    SUPPORTED_LANGUAGES,
)


def detect_language(filename: str) -> str | None:
    """
    Detect the programming language from a filename's extension.

    Returns the language identifier ("go", "python", "javascript")
    or None if the extension is not recognized.
    """
    ext = os.path.splitext(filename)[1].lower()
    return EXTENSION_MAP.get(ext)


def analyze(content: str, filename: str) -> dict:
    """
    Analyze source code content and return a structured evaluation.

    Args:
        content: The source code text to analyze.
        filename: Original filename (used for language detection).

    Returns:
        dict with keys:
          - score (int): 0-100 quality score
          - issues (list[dict]): found issues with type, message, line
          - suggestions (list[str]): improvement recommendations
          - language (str|None): detected language
    """
    language = detect_language(filename)

    # If the language is not supported, return a neutral result
    # with a suggestion to submit a supported file type.
    if language is None or language not in SUPPORTED_LANGUAGES:
        return {
            "score": 0,
            "issues": [],
            "suggestions": [
                f"Language not supported for analysis (file: {filename}). "
                f"Supported: Go (.go), Python (.py), JavaScript (.js/.ts)."
            ],
            "language": None,
        }

    rules = LANGUAGE_RULES.get(language, [])
    issues = _scan_content(content, rules)
    score = _calculate_score(issues)
    suggestions = _generate_suggestions(issues, language)

    return {
        "score": score,
        "issues": issues,
        "suggestions": suggestions,
        "language": language,
    }


def _scan_content(content: str, rules: list[dict]) -> list[dict]:
    """
    Scan source code line-by-line against a set of rules.

    Each rule's regex pattern is tested against every line. When a match
    is found, an issue is recorded with the line number. Duplicate issues
    (same message on the same line) are skipped.
    """
    issues = []
    seen = set()  # (message, line) pairs to deduplicate.
    lines = content.splitlines()

    for line_num, line in enumerate(lines, start=1):
        # Skip empty lines and comment-only lines for performance.
        stripped = line.strip()
        if not stripped:
            continue

        for rule in rules:
            if rule["pattern"].search(line):
                key = (rule["message"], line_num)
                if key not in seen:
                    seen.add(key)
                    issues.append(
                        {
                            "issue_type": rule["issue_type"],
                            "message": rule["message"],
                            "line": line_num,
                            "severity": rule["severity"],
                        }
                    )

    return issues


def _calculate_score(issues: list[dict]) -> int:
    """
    Calculate a quality score (0-100) based on found issues.

    Each issue deducts points based on its severity:
      - high: -15 points
      - medium: -8 points
      - low: -3 points

    The score is clamped to [0, 100]. A file with no issues scores 100.
    """
    deduction = sum(SEVERITY_WEIGHT.get(i["severity"], 0) for i in issues)
    return max(0, 100 - deduction)


def _generate_suggestions(issues: list[dict], language: str) -> list[str]:
    """
    Generate actionable suggestions based on the types of issues found.

    Groups issues by type and produces one suggestion per category
    rather than repeating per-issue advice.
    """
    suggestions = []
    issue_types = {i["issue_type"] for i in issues}

    if "security" in issue_types:
        suggestions.append(
            "Security issues detected — review and fix before deployment"
        )

    if "quality" in issue_types:
        suggestions.append(
            "Code quality issues found — consider refactoring for maintainability"
        )

    if "style" in issue_types:
        suggestions.append(
            "Style issues found — apply consistent formatting and remove debug statements"
        )

    if not issues:
        suggestions.append("No issues detected — code looks clean")

    return suggestions
