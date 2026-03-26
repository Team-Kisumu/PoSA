"""
Tests for the PoSA writing evaluation module.

Covers file type detection, quality checks (length, structure, sentences,
passive voice, weasel words, repeated phrases), scoring, and suggestions.
"""

from ai.writing import evaluate_writing, is_writing_file

# --- File Type Detection Tests ---


def test_is_writing_md():
    assert is_writing_file("README.md") is True


def test_is_writing_txt():
    assert is_writing_file("notes.txt") is True


def test_is_writing_rst():
    assert is_writing_file("docs.rst") is True


def test_is_writing_html():
    assert is_writing_file("page.html") is True


def test_is_writing_tex():
    assert is_writing_file("paper.tex") is True


def test_is_not_writing_go():
    assert is_writing_file("main.go") is False


def test_is_not_writing_py():
    assert is_writing_file("script.py") is False


def test_is_not_writing_no_ext():
    assert is_writing_file("Makefile") is False


# --- Quality Scoring Tests ---


def test_good_writing():
    """Well-structured writing with paragraphs should score high."""
    content = (
        "# Project Overview\n"
        "\n"
        "This project implements a decentralized verification system. "
        "It uses artificial intelligence to evaluate submitted work "
        "and blockchain technology to issue tamper-proof credentials.\n"
        "\n"
        "## Architecture\n"
        "\n"
        "The system consists of three main components. "
        "The backend handles API requests and orchestration. "
        "The AI engine performs code and writing analysis. "
        "The blockchain layer anchors proofs on-chain.\n"
        "\n"
        "## Getting Started\n"
        "\n"
        "Clone the repository and install dependencies. "
        "Run the backend server on port 8080. "
        "Start the AI engine on port 8000. "
        "Open the frontend at localhost:3000.\n"
    )
    result = evaluate_writing(content, "README.md")
    assert result["score"] >= 80
    assert result["language"] == "text"


def test_very_short_content():
    """Content under 20 words should get a high-severity issue."""
    result = evaluate_writing("Hello world.", "note.txt")
    assert result["score"] < 100
    assert any("short" in i["message"].lower() for i in result["issues"])


def test_short_content():
    """Content between 20-50 words should get a medium-severity issue."""
    content = " ".join(["word"] * 30) + "."
    result = evaluate_writing(content, "note.txt")
    assert any("short" in i["message"].lower() for i in result["issues"])


def test_adequate_length():
    """Content over 50 words should not get a length issue."""
    content = " ".join(["The system works well."] * 15)
    result = evaluate_writing(content, "doc.md")
    length_issues = [i for i in result["issues"] if "short" in i["message"].lower()]
    assert len(length_issues) == 0


# --- Structure Tests ---


def test_no_paragraphs():
    """Long text without paragraph breaks should be flagged."""
    content = "\n".join(["This is a line of text." for _ in range(15)])
    result = evaluate_writing(content, "wall.txt")
    assert any("paragraph" in i["message"].lower() for i in result["issues"])


def test_with_paragraphs():
    """Text with paragraph breaks should not get a paragraph issue."""
    content = (
        "First paragraph with enough content here.\n"
        "\n"
        "Second paragraph with more content here.\n"
        "\n"
        "Third paragraph wrapping things up.\n"
    )
    result = evaluate_writing(content, "doc.txt")
    para_issues = [i for i in result["issues"] if "paragraph" in i["message"].lower()]
    assert len(para_issues) == 0


def test_no_headings_long_doc():
    """Long documents without headings should get a suggestion."""
    lines = ["This is a line of content for the document."] * 25
    content = "\n\n".join(lines)
    result = evaluate_writing(content, "long.txt")
    assert any("heading" in i["message"].lower() for i in result["issues"])


def test_with_headings():
    """Documents with headings should not get a heading issue."""
    content = "# Introduction\n\nSome intro text here.\n\n" "# Details\n\nMore detailed content here.\n"
    result = evaluate_writing(content, "doc.md")
    heading_issues = [i for i in result["issues"] if "heading" in i["message"].lower()]
    assert len(heading_issues) == 0


# --- Passive Voice Tests ---


def test_passive_voice_detected():
    """Passive voice constructions should be flagged."""
    content = (
        "The report was written by the team. "
        "The code was reviewed by the lead. "
        "The tests were executed automatically.\n"
    )
    result = evaluate_writing(content, "report.txt")
    assert any("passive" in i["message"].lower() for i in result["issues"])


def test_active_voice_clean():
    """Active voice writing should not trigger passive voice issues."""
    content = (
        "The team wrote the report. " "The lead reviewed the code. " "The CI pipeline ran the tests automatically.\n"
    )
    result = evaluate_writing(content, "report.txt")
    passive_issues = [i for i in result["issues"] if "passive" in i["message"].lower()]
    assert len(passive_issues) == 0


# --- Weasel Words Tests ---


def test_weasel_words_detected():
    """Vague language should be flagged."""
    content = "This is very important and really quite obvious to basically everyone.\n"
    result = evaluate_writing(content, "doc.txt")
    assert any("weak word" in i["message"].lower() for i in result["issues"])


def test_precise_language_clean():
    """Precise language should not trigger weasel word issues."""
    content = (
        "The system processes 1000 requests per second. "
        "Response latency averages 12 milliseconds. "
        "Memory usage stays below 256 megabytes.\n"
    )
    result = evaluate_writing(content, "doc.txt")
    weasel_issues = [i for i in result["issues"] if "weak word" in i["message"].lower()]
    assert len(weasel_issues) == 0


# --- Sentence Length Tests ---


def test_long_sentence_flagged():
    """Sentences over 40 words should be flagged."""
    long_sentence = " ".join(["word"] * 45) + "."
    content = long_sentence + " Short sentence here.\n"
    result = evaluate_writing(content, "doc.txt")
    assert any("sentence" in i["message"].lower() and "long" in i["message"].lower() for i in result["issues"])


# --- Suggestion Tests ---


def test_suggestions_for_short_content():
    """Short content should produce an expand suggestion."""
    result = evaluate_writing("Brief note.", "note.txt")
    assert any("expand" in s.lower() for s in result["suggestions"])


def test_suggestions_for_passive():
    """Passive voice should produce a rewrite suggestion."""
    content = "The code was written by the developer. The tests were run.\n"
    result = evaluate_writing(content, "doc.txt")
    assert any("passive" in s.lower() for s in result["suggestions"])


def test_suggestions_for_clean_writing():
    """Clean writing should get a positive suggestion."""
    content = (
        "# Overview\n\n"
        "This document describes the system architecture in detail. "
        "The backend handles API requests and orchestrates the pipeline. "
        "The frontend renders the user interface and displays results.\n\n"
        "# Components\n\n"
        "Each component operates independently and communicates via HTTP. "
        "The services use structured JSON for all data exchange. "
        "Data flows from the user through the backend to the blockchain. "
        "The AI engine evaluates submissions and returns scored results. "
        "The storage layer persists evaluation reports on IPFS.\n"
    )
    result = evaluate_writing(content, "doc.md")
    assert any("good" in s.lower() or "no major" in s.lower() for s in result["suggestions"])


# --- Integration via Analyzer Tests ---


def test_analyzer_routes_md_to_writing():
    """Markdown files should be routed to the writing evaluator."""
    from ai.analyzer import analyze

    content = (
        "# Title\n\nThis is a well-structured document "
        "with enough content for evaluation.\n\n"
        "## Section\n\nMore content in a second section "
        "to ensure adequate length for scoring.\n"
    )
    result = analyze(content, "README.md")
    assert result["language"] == "text"
    assert result["score"] >= 0


def test_analyzer_routes_txt_to_writing():
    """Text files should be routed to the writing evaluator."""
    from ai.analyzer import analyze

    result = analyze("Some text content for evaluation purposes here.\n", "notes.txt")
    assert result["language"] == "text"


def test_analyzer_unsupported_still_works():
    """Truly unsupported files should still return score 0."""
    from ai.analyzer import analyze

    result = analyze("body { color: red; }", "style.css")
    assert result["score"] == 0
    assert result["language"] is None


# --- Edge Cases ---


def test_empty_content():
    """Empty content should still return a valid result."""
    result = evaluate_writing("", "empty.txt")
    assert result["score"] <= 100
    assert result["language"] == "text"


def test_only_whitespace():
    """Whitespace-only content should flag as very short."""
    result = evaluate_writing("   \n\n   \n", "blank.txt")
    assert any("short" in i["message"].lower() for i in result["issues"])
