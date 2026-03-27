"""
Impulse AI client for the PoSA evaluation engine.

Calls the Impulse Labs API (https://api.impulselabs.ai/api/chat) to get
AI-powered code analysis that supplements the local pattern-based rules.
The API uses SSE streaming; this client collects the full response.
"""

import json
import os
import re

import httpx

IMPULSE_API_URL = "https://api.impulselabs.ai/api/chat"
IMPULSE_TIMEOUT = 60

SYSTEM_PROMPT = (
    "You are a code quality analyzer. Analyze the submitted code and respond with ONLY valid JSON "
    "(no markdown, no code fences). Use this exact format: "
    '{"score": <0-100>, "issues": [{"type": "<security|quality|logic|style>", '
    '"message": "<description>", "line": <number_or_null>}], "suggestions": ["<suggestion>"]}. '
    "Be strict on security. Deduct heavily for command injection, SQL injection, XSS, "
    "and deserialization vulnerabilities."
)


def get_api_key() -> str:
    """Load the Impulse AI API key from environment or .env file."""
    key = os.environ.get("AI_API_KEY", "")
    if key:
        return key
    # Fallback: read from .env in project root.
    env_path = os.path.join(os.path.dirname(__file__), "..", ".env")
    if os.path.exists(env_path):
        with open(env_path) as f:
            for line in f:
                if line.startswith("AI_API_KEY="):
                    return line.strip().split("=", 1)[1]
    return ""


def analyze_with_impulse(content: str, filename: str) -> dict | None:
    """
    Send code to Impulse AI for analysis and return structured results.

    Returns a dict with score, issues, suggestions on success.
    Returns None if the API key is missing, the call fails, or
    the response can't be parsed — the caller should fall back
    to local pattern-based analysis.
    """
    api_key = get_api_key()
    if not api_key:
        return None

    prompt = f"Analyze this file ({filename}) for security, quality, and logic issues:\n\n```\n{content}\n```"

    try:
        collected = _stream_response(api_key, prompt)
        if not collected:
            return None
        return _parse_response(collected)
    except Exception:
        return None


def _stream_response(api_key: str, prompt: str) -> str:
    """Call the Impulse AI SSE endpoint and collect all content deltas."""
    content_parts = []

    with httpx.stream(
        "POST",
        IMPULSE_API_URL,
        headers={"x-api-key": api_key, "Content-Type": "application/json"},
        json={"messages": [{"role": "system", "content": SYSTEM_PROMPT}, {"role": "user", "content": prompt}]},
        timeout=IMPULSE_TIMEOUT,
    ) as resp:
        if resp.status_code != 200:
            return ""
        for line in resp.iter_lines():
            if not line.startswith("data: "):
                continue
            try:
                event = json.loads(line[6:])
            except json.JSONDecodeError:
                continue
            if event.get("type") == "delta":
                text = event.get("data", {}).get("content", "")
                if text:
                    content_parts.append(text)

    return "".join(content_parts)


def _parse_response(text: str) -> dict | None:
    """Extract JSON from the Impulse AI response text."""
    # Strip markdown code fences if present.
    text = re.sub(r"^```json\s*", "", text.strip())
    text = re.sub(r"\s*```$", "", text.strip())

    try:
        data = json.loads(text)
    except json.JSONDecodeError:
        # Try to find JSON object in the text.
        match = re.search(r"\{.*\}", text, re.DOTALL)
        if not match:
            return None
        try:
            data = json.loads(match.group())
        except json.JSONDecodeError:
            return None

    # Validate structure.
    if not isinstance(data.get("score"), (int, float)):
        return None
    score = max(0, min(100, int(data["score"])))

    issues = []
    for issue in data.get("issues", []):
        if isinstance(issue, dict) and "message" in issue:
            issues.append(
                {
                    "issue_type": issue.get("type", "quality"),
                    "message": issue["message"],
                    "line": issue.get("line"),
                    "severity": "high" if issue.get("type") == "security" else "medium",
                }
            )

    suggestions = [s for s in data.get("suggestions", []) if isinstance(s, str)]

    return {"score": score, "issues": issues, "suggestions": suggestions}
