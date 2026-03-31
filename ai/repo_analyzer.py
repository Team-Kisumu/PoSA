"""
GitHub repository analyzer for the PoSA AI engine.

Fetches source files from a public GitHub repository via the GitHub API,
runs each file through the existing pattern-based analyzer, and aggregates
results into a single repo-level evaluation.

GitHub API usage (unauthenticated: 60 req/hr, authenticated: 5000 req/hr):
  - 1 request: GET /repos/{owner}/{repo} (metadata + default branch)
  - 1 request: GET /repos/{owner}/{repo}/git/trees/{branch}?recursive=1
  - N requests: GET /repos/{owner}/{repo}/git/blobs/{sha} (one per file)

To stay within rate limits, file count is capped at MAX_FILES and large
files (>MAX_FILE_SIZE bytes) are skipped.
"""

import base64
import os
from collections import Counter

import httpx

from ai.analyzer import analyze, detect_language
from ai.rules.patterns import EXTENSION_MAP
from ai.writing import is_writing_file

GITHUB_API = "https://api.github.com"
GITHUB_TIMEOUT = 15

# Limits to stay within API rate limits and keep response times reasonable.
MAX_FILES = 20
MAX_FILE_SIZE = 100_000  # 100KB per file

# Extensions worth analyzing (code + writing).
ANALYZABLE_EXTENSIONS = set(EXTENSION_MAP.keys()) | {
    ".md", ".txt", ".rst", ".html",
}

# Directories to skip (dependencies, build artifacts, vendored code).
SKIP_DIRS = {
    "node_modules", "vendor", ".git", "__pycache__", ".next",
    "dist", "build", "out", ".venv", "venv", ".tox",
    "target", "bin", "obj", "packages", "bower_components",
}


def _github_headers() -> dict:
    """Build GitHub API request headers, with optional auth token."""
    headers = {"Accept": "application/vnd.github.v3+json"}
    token = os.environ.get("GITHUB_TOKEN", "")
    if token:
        headers["Authorization"] = f"Bearer {token}"
    return headers


def parse_repo_url(url: str) -> tuple[str, str]:
    """
    Extract owner and repo name from a GitHub URL.

    Accepts formats:
      - https://github.com/owner/repo
      - https://github.com/owner/repo.git
      - https://github.com/owner/repo/tree/branch/...

    Returns (owner, repo) or raises ValueError.
    """
    url = url.strip().rstrip("/")
    if url.endswith(".git"):
        url = url[:-4]

    prefix = "https://github.com/"
    if not url.startswith(prefix):
        raise ValueError(f"not a GitHub URL: {url}")

    path = url[len(prefix):]
    parts = path.split("/")
    if len(parts) < 2 or not parts[0] or not parts[1]:
        raise ValueError(f"URL must include owner/repo: {url}")

    return parts[0], parts[1]


def _should_skip(path: str) -> bool:
    """Check if a file path is in a directory that should be skipped."""
    parts = path.split("/")
    return any(p in SKIP_DIRS for p in parts[:-1])


def _is_analyzable(path: str) -> bool:
    """Check if a file has an extension we can analyze."""
    dot = path.rfind(".")
    if dot == -1:
        return False
    ext = path[dot:].lower()
    return ext in ANALYZABLE_EXTENSIONS


def fetch_repo_files(owner: str, repo: str) -> list[dict]:
    """
    Fetch the file tree and contents of analyzable files from a GitHub repo.

    Returns a list of dicts with keys: path, content, size.
    Files are prioritized by: code files first, then by size (smaller first),
    capped at MAX_FILES.
    """
    headers = _github_headers()

    # Step 1: Get repo metadata to find the default branch.
    meta_resp = httpx.get(
        f"{GITHUB_API}/repos/{owner}/{repo}",
        headers=headers,
        timeout=GITHUB_TIMEOUT,
    )
    if meta_resp.status_code == 404:
        raise ValueError(f"repository not found: {owner}/{repo}")
    if meta_resp.status_code == 403:
        raise ValueError("GitHub API rate limit exceeded — try again later or set GITHUB_TOKEN")
    meta_resp.raise_for_status()
    meta = meta_resp.json()
    default_branch = meta.get("default_branch", "main")

    # Step 2: Get the full file tree recursively.
    tree_resp = httpx.get(
        f"{GITHUB_API}/repos/{owner}/{repo}/git/trees/{default_branch}?recursive=1",
        headers=headers,
        timeout=GITHUB_TIMEOUT,
    )
    tree_resp.raise_for_status()
    tree_data = tree_resp.json()

    # Filter to analyzable files within size limits, skip vendored dirs.
    candidates = [
        item for item in tree_data.get("tree", [])
        if item["type"] == "blob"
        and item.get("size", 0) <= MAX_FILE_SIZE
        and item.get("size", 0) > 0
        and not _should_skip(item["path"])
        and _is_analyzable(item["path"])
    ]

    # Prioritize: code files before writing files, smaller files first.
    def sort_key(item):
        ext = item["path"][item["path"].rfind("."):].lower()
        is_code = ext in EXTENSION_MAP
        return (0 if is_code else 1, item.get("size", 0))

    candidates.sort(key=sort_key)
    candidates = candidates[:MAX_FILES]

    # Step 3: Fetch content for each file via the Blobs API.
    files = []
    for item in candidates:
        try:
            blob_resp = httpx.get(
                f"{GITHUB_API}/repos/{owner}/{repo}/git/blobs/{item['sha']}",
                headers=headers,
                timeout=GITHUB_TIMEOUT,
            )
            if blob_resp.status_code != 200:
                continue
            blob = blob_resp.json()
            if blob.get("encoding") == "base64":
                content = base64.b64decode(blob["content"]).decode("utf-8", errors="replace")
            else:
                content = blob.get("content", "")
            if content.strip():
                files.append({
                    "path": item["path"],
                    "content": content,
                    "size": item.get("size", len(content)),
                })
        except Exception:
            continue

    return files


def analyze_repo(repo_url: str) -> dict:
    """
    Analyze a GitHub repository and return aggregated evaluation results.

    Returns dict with:
      - score (int): weighted average across all analyzed files
      - issues (list): all issues with file path context
      - suggestions (list): deduplicated suggestions
      - files_analyzed (int): number of files successfully analyzed
      - file_scores (list): per-file score breakdown
    """
    owner, repo = parse_repo_url(repo_url)
    files = fetch_repo_files(owner, repo)

    if not files:
        return {
            "score": 0,
            "issues": [],
            "suggestions": [
                f"No analyzable files found in {owner}/{repo}. "
                "Ensure the repository is public and contains supported file types."
            ],
            "files_analyzed": 0,
            "file_scores": [],
        }

    all_issues = []
    all_suggestions = []
    file_scores = []
    total_weight = 0

    for f in files:
        filename = f["path"].split("/")[-1]
        result = analyze(f["content"], filename)
        score = result["score"]
        size = f["size"]

        # Weight by file size so larger files have more impact on the score.
        weight = max(1, size)
        file_scores.append({
            "path": f["path"],
            "score": score,
            "issues": len(result["issues"]),
            "language": result.get("language"),
        })
        total_weight += weight

        # Prefix issue messages with the file path for context.
        for issue in result["issues"]:
            all_issues.append({
                "issue_type": issue["issue_type"],
                "message": f"[{f['path']}] {issue['message']}",
                "line": issue.get("line"),
                "severity": issue.get("severity", "medium"),
            })

        all_suggestions.extend(result.get("suggestions", []))

        # Accumulate weighted score.
        file_scores[-1]["_weighted"] = score * weight

    # Weighted average score.
    if total_weight > 0:
        avg_score = sum(fs["_weighted"] for fs in file_scores) // total_weight
    else:
        avg_score = 0

    # Clean up internal weight field.
    for fs in file_scores:
        fs.pop("_weighted", None)

    # Deduplicate suggestions.
    seen = set()
    unique_suggestions = []
    for s in all_suggestions:
        if s not in seen:
            seen.add(s)
            unique_suggestions.append(s)

    # Add repo-level summary suggestion.
    if file_scores:
        lang_counts = Counter(
            fs["language"] for fs in file_scores if fs.get("language")
        )
        if lang_counts:
            langs = ", ".join(
                f"{lang} ({count})" for lang, count in lang_counts.most_common(5)
            )
            unique_suggestions.insert(
                0, f"Analyzed {len(file_scores)} files across {len(lang_counts)} language(s): {langs}"
            )

    return {
        "score": max(0, min(100, avg_score)),
        "issues": all_issues,
        "suggestions": unique_suggestions,
        "files_analyzed": len(file_scores),
        "file_scores": file_scores,
    }
