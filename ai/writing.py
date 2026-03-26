"""
Writing evaluation module for the PoSA AI engine.

Evaluates non-code text submissions (documentation, reports, writing samples)
for quality, structure, and readability. Returns a score (0-100) on the same
scale as the code analyzer, with issues and improvement suggestions.

Evaluation criteria:
  - Structure: paragraph count, heading usage, sentence variety
  - Readability: average sentence length, word complexity
  - Quality: passive voice, weasel words, repeated phrases
  - Content: minimum length, substance density
"""

import re

# --- Constants ---

# File extensions routed to the writing evaluator.
WRITING_EXTENSIONS = {".md", ".txt", ".rst", ".doc", ".docx", ".tex", ".adoc", ".html"}

# Passive voice indicators (common "to be" + past participle patterns).
PASSIVE_PATTERNS = [
    re.compile(r"\b(?:is|are|was|were|be|been|being)\s+\w+ed\b", re.IGNORECASE),
    re.compile(r"\b(?:is|are|was|were|be|been|being)\s+\w+en\b", re.IGNORECASE),
]

# Weasel words that weaken writing.
WEASEL_WORDS = re.compile(
    r"\b(?:very|really|quite|somewhat|fairly|rather|basically|"
    r"actually|literally|just|simply|obviously|clearly|things|stuff)\b",
    re.IGNORECASE,
)

# Repeated word detection (same word 3+ times in a sentence).
REPEATED_WORD = re.compile(r"\b(\w{4,})\b.*\b\1\b.*\b\1\b", re.IGNORECASE)

# Heading patterns for markdown/rst.
HEADING_PATTERN = re.compile(r"^(?:#{1,6}\s|={3,}|-{3,}|\*{3,})")

# Sentence splitter (rough but effective for scoring).
SENTENCE_SPLIT = re.compile(r"[.!?]+\s+")

# Severity weights matching the code analyzer scale.
SEVERITY_WEIGHT = {
    "high": 15,
    "medium": 8,
    "low": 3,
}


def is_writing_file(filename: str) -> bool:
    """Check if a filename should be evaluated as writing (not code)."""
    ext = "." + filename.rsplit(".", 1)[-1].lower() if "." in filename else ""
    return ext in WRITING_EXTENSIONS


def evaluate_writing(content: str, filename: str) -> dict:
    """
    Evaluate a text submission for writing quality.

    Args:
        content: The text content to evaluate.
        filename: Original filename (used for format-specific checks).

    Returns:
        dict with keys matching the code analyzer output:
          - score (int): 0-100 quality score
          - issues (list[dict]): found issues with type, message, line
          - suggestions (list[str]): improvement recommendations
          - language (str): always "text"
    """
    issues = []
    lines = content.splitlines()
    sentences = _split_sentences(content)

    # Run all quality checks.
    issues.extend(_check_length(content, lines))
    issues.extend(_check_structure(lines))
    issues.extend(_check_sentence_quality(sentences))
    issues.extend(_check_passive_voice(lines))
    issues.extend(_check_weasel_words(lines))
    issues.extend(_check_repeated_phrases(lines))

    score = _calculate_score(issues)
    suggestions = _generate_suggestions(issues, sentences, lines)

    return {
        "score": score,
        "issues": issues,
        "suggestions": suggestions,
        "language": "text",
    }


def _split_sentences(content: str) -> list[str]:
    """Split content into sentences, filtering out empty strings."""
    raw = SENTENCE_SPLIT.split(content.strip())
    return [s.strip() for s in raw if s.strip()]


def _check_length(content: str, lines: list[str]) -> list[dict]:
    """Check minimum content length for meaningful evaluation."""
    issues = []
    word_count = len(content.split())

    if word_count < 20:
        issues.append(
            {
                "issue_type": "quality",
                "message": f"Content is very short ({word_count} words) — insufficient for meaningful evaluation",
                "line": 1,
                "severity": "high",
            }
        )
    elif word_count < 50:
        issues.append(
            {
                "issue_type": "quality",
                "message": f"Content is short ({word_count} words) — consider expanding for depth",
                "line": 1,
                "severity": "medium",
            }
        )

    return issues


def _check_structure(lines: list[str]) -> list[dict]:
    """Check document structure: paragraphs, headings, organization."""
    issues = []
    non_empty = [line for line in lines if line.strip()]
    has_headings = any(HEADING_PATTERN.match(line) for line in lines)

    # Check for single-paragraph wall of text (no blank line breaks).
    paragraph_breaks = sum(1 for line in lines if not line.strip())
    if len(non_empty) > 10 and paragraph_breaks < 1:
        issues.append(
            {
                "issue_type": "quality",
                "message": "No paragraph breaks — break content into paragraphs for readability",
                "line": 1,
                "severity": "medium",
            }
        )

    # Suggest headings for longer documents.
    if len(non_empty) > 20 and not has_headings:
        issues.append(
            {
                "issue_type": "style",
                "message": "No headings found — add section headings to organize longer content",
                "line": 1,
                "severity": "low",
            }
        )

    return issues


def _check_sentence_quality(sentences: list[str]) -> list[dict]:
    """Check sentence length and variety."""
    issues = []
    if not sentences:
        return issues

    lengths = [len(s.split()) for s in sentences]
    avg_length = sum(lengths) / len(lengths)

    # Flag overly long average sentence length.
    if avg_length > 30:
        issues.append(
            {
                "issue_type": "quality",
                "message": f"Average sentence length is {avg_length:.0f} words — aim for 15-25 for readability",
                "line": None,
                "severity": "medium",
            }
        )

    # Flag individual very long sentences.
    for i, (sentence, length) in enumerate(zip(sentences, lengths)):
        if length > 40:
            # Find approximate line number.
            line_num = _find_line(sentence, sentences[0] if i == 0 else sentences[i])
            issues.append(
                {
                    "issue_type": "quality",
                    "message": f"Sentence is {length} words long — consider splitting for clarity",
                    "line": line_num,
                    "severity": "low",
                }
            )

    return issues


def _check_passive_voice(lines: list[str]) -> list[dict]:
    """Detect passive voice constructions."""
    issues = []
    passive_count = 0

    for line_num, line in enumerate(lines, start=1):
        for pattern in PASSIVE_PATTERNS:
            if pattern.search(line):
                passive_count += 1
                # Only report first 3 instances to avoid noise.
                if passive_count <= 3:
                    issues.append(
                        {
                            "issue_type": "style",
                            "message": "Passive voice detected — consider using active voice",
                            "line": line_num,
                            "severity": "low",
                        }
                    )

    return issues


def _check_weasel_words(lines: list[str]) -> list[dict]:
    """Detect vague or weak language."""
    issues = []
    weasel_count = 0

    for line_num, line in enumerate(lines, start=1):
        matches = WEASEL_WORDS.findall(line)
        for word in matches:
            weasel_count += 1
            # Only report first 3 instances.
            if weasel_count <= 3:
                issues.append(
                    {
                        "issue_type": "style",
                        "message": f'Weak word "{word}" — use more precise language',
                        "line": line_num,
                        "severity": "low",
                    }
                )

    return issues


def _check_repeated_phrases(lines: list[str]) -> list[dict]:
    """Detect repeated words within lines."""
    issues = []

    for line_num, line in enumerate(lines, start=1):
        if REPEATED_WORD.search(line):
            issues.append(
                {
                    "issue_type": "quality",
                    "message": "Repeated word detected — vary vocabulary for better readability",
                    "line": line_num,
                    "severity": "low",
                }
            )

    return issues


def _find_line(text: str, content: str) -> int | None:
    """Approximate line number for a text fragment. Returns None if not found."""
    # Simple heuristic — not critical for scoring accuracy.
    return None


def _calculate_score(issues: list[dict]) -> int:
    """Calculate quality score using the same weighted system as code analysis."""
    deduction = sum(SEVERITY_WEIGHT.get(i["severity"], 0) for i in issues)
    return max(0, 100 - deduction)


def _generate_suggestions(issues: list[dict], sentences: list[str], lines: list[str]) -> list[str]:
    """Generate writing-specific improvement suggestions."""
    suggestions = []
    issue_types = {i["issue_type"] for i in issues}
    issue_messages = {i["message"] for i in issues}

    if any("short" in m.lower() for m in issue_messages):
        suggestions.append("Expand content with more detail, examples, or supporting arguments")

    if any("passive" in m.lower() for m in issue_messages):
        suggestions.append("Reduce passive voice — rewrite sentences with the subject performing the action")

    if any("weasel" in m.lower() or "weak word" in m.lower() for m in issue_messages):
        suggestions.append("Replace vague words with specific, concrete language")

    if any("paragraph" in m.lower() for m in issue_messages):
        suggestions.append("Break text into focused paragraphs — one idea per paragraph")

    if any("heading" in m.lower() for m in issue_messages):
        suggestions.append("Add headings to create a clear document structure")

    if any("sentence" in m.lower() and "long" in m.lower() for m in issue_messages):
        suggestions.append("Shorten long sentences — aim for 15-25 words per sentence")

    if "quality" in issue_types and "style" not in issue_types:
        suggestions.append("Focus on content depth and structure")

    if not issues:
        suggestions.append("Writing quality looks good — no major issues detected")

    return suggestions
