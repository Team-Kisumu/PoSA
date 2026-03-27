#!/usr/bin/env python3
"""
End-to-end test for the PoSA AI evaluation pipeline.

Tests:
1. Local pattern-based analyzer (no API key needed)
2. Impulse AI integration (requires AI_API_KEY in .env)
3. Combined analysis (local rules + Impulse AI enhancement)

Run from project root:
    source ai/.venv/bin/activate
    PYTHONPATH=. python scripts/test_e2e_ai.py
"""

import os
import sys

# Ensure project root is in path.
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from ai.analyzer import analyze  # noqa: E402
from ai.impulse_client import analyze_with_impulse, get_api_key  # noqa: E402

# --- Test Cases ---

VULNERABLE_PYTHON = """import os
import pickle

def run_command(cmd):
    os.system(cmd)

def load_data(data):
    return pickle.loads(data)

def compute(expr):
    return eval(expr)
"""

CLEAN_GO = """package main

import (
    "log"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/health", handler)
    log.Fatal(http.ListenAndServe(":8080", mux))
}
"""

VULNERABLE_JS = """function render(data) {
    document.getElementById("output").innerHTML = data;
    var x = eval(userInput);
    console.log(x);
}
"""

WRITING_SAMPLE = """# Project Overview

This project implements a decentralized verification system.
It uses artificial intelligence to evaluate submitted work
and blockchain technology to issue tamper-proof credentials.

## Architecture

The system consists of three main components.
The backend handles API requests and orchestration.
The AI engine performs code and writing analysis.
The blockchain layer anchors proofs on-chain.

## Getting Started

Clone the repository and install dependencies.
Run the backend server on port 8080.
Start the AI engine on port 8000.
"""


def separator(title):
    print(f"\n{'='*60}")
    print(f"  {title}")
    print(f"{'='*60}\n")


def print_result(result):
    print(f"  Score: {result['score']}/100")
    print(f"  Language: {result.get('language', 'N/A')}")
    if result["issues"]:
        print(f"  Issues ({len(result['issues'])}):")
        for i in result["issues"][:5]:
            line = f" (line {i['line']})" if i.get("line") else ""
            print(f"    [{i['issue_type']}] {i['message']}{line}")
        if len(result["issues"]) > 5:
            print(f"    ... and {len(result['issues']) - 5} more")
    else:
        print("  Issues: None")
    if result["suggestions"]:
        print("  Suggestions:")
        for s in result["suggestions"][:3]:
            print(f"    - {s}")


def test_local_analyzer():
    """Test the local pattern-based analyzer."""
    separator("TEST 1: Local Pattern-Based Analyzer")

    print("[1a] Vulnerable Python code:")
    result = analyze(VULNERABLE_PYTHON, "vulnerable.py")
    print_result(result)
    assert result["score"] < 60, f"Expected low score, got {result['score']}"
    assert len(result["issues"]) >= 3, f"Expected 3+ issues, got {len(result['issues'])}"
    print("  PASS\n")

    print("[1b] Clean Go code:")
    result = analyze(CLEAN_GO, "server.go")
    print_result(result)
    assert result["score"] >= 80, f"Expected high score, got {result['score']}"
    print("  PASS\n")

    print("[1c] Vulnerable JavaScript:")
    result = analyze(VULNERABLE_JS, "render.js")
    print_result(result)
    assert result["score"] < 70, f"Expected low score, got {result['score']}"
    assert any("innerHTML" in i["message"] for i in result["issues"])
    assert any("eval" in i["message"] for i in result["issues"])
    print("  PASS\n")

    print("[1d] Writing sample (Markdown):")
    result = analyze(WRITING_SAMPLE, "README.md")
    print_result(result)
    assert result["language"] == "text"
    print("  PASS")


def test_impulse_ai():
    """Test the Impulse AI integration."""
    separator("TEST 2: Impulse AI Integration")

    api_key = get_api_key()
    if not api_key:
        print("  SKIPPED: AI_API_KEY not set in environment or .env")
        return False

    print(f"  API key found: {api_key[:10]}...")
    print()

    print("[2a] Impulse AI - Vulnerable Python:")
    result = analyze_with_impulse(VULNERABLE_PYTHON, "vulnerable.py")
    if result is None:
        print("  FAILED: No response from Impulse AI")
        return False
    print_result(result)
    assert result["score"] < 60, f"Expected low score, got {result['score']}"
    assert len(result["issues"]) >= 1, "Expected at least 1 issue"
    print("  PASS\n")

    print("[2b] Impulse AI - Clean Go:")
    result = analyze_with_impulse(CLEAN_GO, "server.go")
    if result is None:
        print("  FAILED: No response from Impulse AI")
        return False
    print_result(result)
    assert result["score"] >= 50, f"Expected decent score, got {result['score']}"
    print("  PASS")

    return True


def test_combined():
    """Test combined local + Impulse AI analysis."""
    separator("TEST 3: Combined Analysis (Local + Impulse AI)")

    api_key = get_api_key()
    if not api_key:
        print("  SKIPPED: AI_API_KEY not set")
        return

    print("[3a] Local analysis of vulnerable Python:")
    local = analyze(VULNERABLE_PYTHON, "vulnerable.py")
    print(f"  Local score: {local['score']}/100 ({len(local['issues'])} issues)")

    print("\n[3b] Impulse AI analysis of same code:")
    impulse = analyze_with_impulse(VULNERABLE_PYTHON, "vulnerable.py")
    if impulse:
        print(f"  Impulse score: {impulse['score']}/100 ({len(impulse['issues'])} issues)")

        # Merge: take the lower score (stricter), combine unique issues.
        combined_score = min(local["score"], impulse["score"])
        local_msgs = {i["message"] for i in local["issues"]}
        combined_issues = local["issues"] + [i for i in impulse["issues"] if i["message"] not in local_msgs]
        combined_suggestions = list(set(local["suggestions"] + impulse["suggestions"]))

        print("\n[3c] Combined result:")
        print(f"  Combined score: {combined_score}/100")
        print(f"  Total unique issues: {len(combined_issues)}")
        print(f"  Total suggestions: {len(combined_suggestions)}")
        print("  PASS")
    else:
        print("  Impulse AI unavailable, using local only")


if __name__ == "__main__":
    print("PoSA AI Evaluation Pipeline - End-to-End Test")
    print("=" * 60)

    passed = 0
    failed = 0

    try:
        test_local_analyzer()
        passed += 1
    except AssertionError as e:
        print(f"  FAILED: {e}")
        failed += 1

    try:
        impulse_ok = test_impulse_ai()
        if impulse_ok:
            passed += 1
        elif impulse_ok is False:  # noqa: F541
            failed += 1
    except AssertionError as e:
        print(f"  FAILED: {e}")
        failed += 1

    try:
        test_combined()
        passed += 1
    except AssertionError as e:
        print(f"  FAILED: {e}")
        failed += 1

    separator("RESULTS")
    print(f"  Passed: {passed}")
    print(f"  Failed: {failed}")
    sys.exit(1 if failed > 0 else 0)
