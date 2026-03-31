"""
Tests for the GitHub repository analyzer.

Tests cover URL parsing, file filtering, analysis aggregation, and
error handling. Network calls are mocked to avoid GitHub API dependency.
"""

import pytest
from unittest.mock import patch, MagicMock
import base64

from ai.repo_analyzer import (
    parse_repo_url,
    _should_skip,
    _is_analyzable,
    analyze_repo,
    fetch_repo_files,
    MAX_FILES,
    MAX_FILE_SIZE,
)


# --- URL Parsing ---


class TestParseRepoUrl:
    def test_standard_url(self):
        assert parse_repo_url("https://github.com/owner/repo") == ("owner", "repo")

    def test_url_with_git_suffix(self):
        assert parse_repo_url("https://github.com/owner/repo.git") == ("owner", "repo")

    def test_url_with_trailing_slash(self):
        assert parse_repo_url("https://github.com/owner/repo/") == ("owner", "repo")

    def test_url_with_tree_path(self):
        assert parse_repo_url("https://github.com/owner/repo/tree/main/src") == ("owner", "repo")

    def test_url_with_whitespace(self):
        assert parse_repo_url("  https://github.com/owner/repo  ") == ("owner", "repo")

    def test_not_github(self):
        with pytest.raises(ValueError, match="not a GitHub URL"):
            parse_repo_url("https://gitlab.com/owner/repo")

    def test_missing_repo(self):
        with pytest.raises(ValueError, match="must include owner/repo"):
            parse_repo_url("https://github.com/owner")

    def test_empty_owner(self):
        with pytest.raises(ValueError, match="must include owner/repo"):
            parse_repo_url("https://github.com//repo")

    def test_http_not_https(self):
        with pytest.raises(ValueError, match="not a GitHub URL"):
            parse_repo_url("http://github.com/owner/repo")


# --- File Filtering ---


class TestShouldSkip:
    def test_node_modules(self):
        assert _should_skip("node_modules/package/index.js") is True

    def test_vendor(self):
        assert _should_skip("vendor/github.com/pkg/errors/errors.go") is True

    def test_nested_skip(self):
        assert _should_skip("src/node_modules/dep/file.js") is True

    def test_normal_path(self):
        assert _should_skip("src/main.go") is False

    def test_root_file(self):
        assert _should_skip("main.go") is False

    def test_venv(self):
        assert _should_skip(".venv/lib/python3.10/site.py") is True

    def test_git(self):
        assert _should_skip(".git/config") is True


class TestIsAnalyzable:
    def test_go_file(self):
        assert _is_analyzable("main.go") is True

    def test_python_file(self):
        assert _is_analyzable("app.py") is True

    def test_javascript_file(self):
        assert _is_analyzable("index.js") is True

    def test_markdown_file(self):
        assert _is_analyzable("README.md") is True

    def test_no_extension(self):
        assert _is_analyzable("Makefile") is False

    def test_unknown_extension(self):
        assert _is_analyzable("data.xyz") is False

    def test_nested_path(self):
        assert _is_analyzable("src/handlers/submit.go") is True


# --- Mocked Analysis ---


def _mock_tree_response(files):
    """Build a mock GitHub tree API response."""
    tree = []
    for path, size in files:
        tree.append({
            "path": path,
            "type": "blob",
            "sha": f"sha_{path.replace('/', '_')}",
            "size": size,
        })
    return {"tree": tree, "truncated": False}


def _mock_blob_response(content):
    """Build a mock GitHub blob API response."""
    encoded = base64.b64encode(content.encode()).decode()
    return {"content": encoded, "encoding": "base64", "size": len(content)}


class TestAnalyzeRepo:
    @patch("ai.repo_analyzer.httpx.get")
    def test_analyze_simple_repo(self, mock_get):
        """Analyze a repo with one Go file containing a security issue."""
        go_code = 'package main\n\nimport "os/exec"\n\nfunc run(cmd string) {\n\texec.Command(cmd)\n}\n'

        def side_effect(url, **kwargs):
            resp = MagicMock()
            resp.status_code = 200
            resp.raise_for_status = MagicMock()
            if "/repos/" in url and "/git/trees/" not in url and "/git/blobs/" not in url:
                resp.json.return_value = {"default_branch": "main"}
            elif "/git/trees/" in url:
                resp.json.return_value = _mock_tree_response([("main.go", 80)])
            elif "/git/blobs/" in url:
                resp.json.return_value = _mock_blob_response(go_code)
            return resp

        mock_get.side_effect = side_effect

        result = analyze_repo("https://github.com/test/repo")
        assert result["score"] < 100
        assert result["files_analyzed"] == 1
        assert len(result["issues"]) > 0
        assert any("exec.Command" in i["message"] for i in result["issues"])
        assert any("[main.go]" in i["message"] for i in result["issues"])

    @patch("ai.repo_analyzer.httpx.get")
    def test_analyze_empty_repo(self, mock_get):
        """Repo with no analyzable files returns score 0."""
        def side_effect(url, **kwargs):
            resp = MagicMock()
            resp.status_code = 200
            resp.raise_for_status = MagicMock()
            if "/git/trees/" not in url and "/git/blobs/" not in url:
                resp.json.return_value = {"default_branch": "main"}
            elif "/git/trees/" in url:
                resp.json.return_value = _mock_tree_response([("README", 10)])
            return resp

        mock_get.side_effect = side_effect

        result = analyze_repo("https://github.com/test/empty")
        assert result["score"] == 0
        assert result["files_analyzed"] == 0
        assert any("No analyzable files" in s for s in result["suggestions"])

    @patch("ai.repo_analyzer.httpx.get")
    def test_analyze_multi_file_repo(self, mock_get):
        """Repo with multiple files aggregates scores correctly."""
        go_clean = 'package main\n\nimport "net/http"\n\nfunc h(w http.ResponseWriter, r *http.Request) {\n\tw.WriteHeader(200)\n}\n'
        py_vuln = "import os\n\ndef run(cmd):\n    os.system(cmd)\n"

        def side_effect(url, **kwargs):
            resp = MagicMock()
            resp.status_code = 200
            resp.raise_for_status = MagicMock()
            if "/git/trees/" not in url and "/git/blobs/" not in url:
                resp.json.return_value = {"default_branch": "main"}
            elif "/git/trees/" in url:
                resp.json.return_value = _mock_tree_response([
                    ("handler.go", 90),
                    ("script.py", 50),
                ])
            elif "sha_handler.go" in url:
                resp.json.return_value = _mock_blob_response(go_clean)
            elif "sha_script.py" in url:
                resp.json.return_value = _mock_blob_response(py_vuln)
            return resp

        mock_get.side_effect = side_effect

        result = analyze_repo("https://github.com/test/multi")
        assert result["files_analyzed"] == 2
        assert len(result["file_scores"]) == 2
        # Score should be between the two individual scores.
        scores = [fs["score"] for fs in result["file_scores"]]
        assert result["score"] >= min(scores)
        assert result["score"] <= max(scores)

    @patch("ai.repo_analyzer.httpx.get")
    def test_skips_vendored_files(self, mock_get):
        """Files in vendor/node_modules directories are skipped."""
        go_code = "package main\n\nfunc main() {}\n"

        def side_effect(url, **kwargs):
            resp = MagicMock()
            resp.status_code = 200
            resp.raise_for_status = MagicMock()
            if "/git/trees/" not in url and "/git/blobs/" not in url:
                resp.json.return_value = {"default_branch": "main"}
            elif "/git/trees/" in url:
                resp.json.return_value = _mock_tree_response([
                    ("main.go", 30),
                    ("vendor/dep/dep.go", 30),
                    ("node_modules/pkg/index.js", 50),
                ])
            elif "/git/blobs/" in url:
                resp.json.return_value = _mock_blob_response(go_code)
            return resp

        mock_get.side_effect = side_effect

        result = analyze_repo("https://github.com/test/vendor")
        assert result["files_analyzed"] == 1
        assert result["file_scores"][0]["path"] == "main.go"

    @patch("ai.repo_analyzer.httpx.get")
    def test_skips_large_files(self, mock_get):
        """Files exceeding MAX_FILE_SIZE are skipped."""
        go_code = "package main\n\nfunc main() {}\n"

        def side_effect(url, **kwargs):
            resp = MagicMock()
            resp.status_code = 200
            resp.raise_for_status = MagicMock()
            if "/git/trees/" not in url and "/git/blobs/" not in url:
                resp.json.return_value = {"default_branch": "main"}
            elif "/git/trees/" in url:
                resp.json.return_value = _mock_tree_response([
                    ("small.go", 30),
                    ("huge.go", MAX_FILE_SIZE + 1),
                ])
            elif "/git/blobs/" in url:
                resp.json.return_value = _mock_blob_response(go_code)
            return resp

        mock_get.side_effect = side_effect

        result = analyze_repo("https://github.com/test/large")
        assert result["files_analyzed"] == 1
        assert result["file_scores"][0]["path"] == "small.go"

    @patch("ai.repo_analyzer.httpx.get")
    def test_repo_not_found(self, mock_get):
        """404 from GitHub raises ValueError."""
        def side_effect(url, **kwargs):
            resp = MagicMock()
            resp.status_code = 404
            return resp

        mock_get.side_effect = side_effect

        with pytest.raises(ValueError, match="not found"):
            analyze_repo("https://github.com/nonexistent/repo")

    @patch("ai.repo_analyzer.httpx.get")
    def test_rate_limit_error(self, mock_get):
        """403 from GitHub raises ValueError about rate limit."""
        def side_effect(url, **kwargs):
            resp = MagicMock()
            resp.status_code = 403
            return resp

        mock_get.side_effect = side_effect

        with pytest.raises(ValueError, match="rate limit"):
            analyze_repo("https://github.com/test/repo")

    @patch("ai.repo_analyzer.httpx.get")
    def test_file_scores_have_language(self, mock_get):
        """Per-file scores include detected language."""
        py_code = "x = 1\n"

        def side_effect(url, **kwargs):
            resp = MagicMock()
            resp.status_code = 200
            resp.raise_for_status = MagicMock()
            if "/git/trees/" not in url and "/git/blobs/" not in url:
                resp.json.return_value = {"default_branch": "main"}
            elif "/git/trees/" in url:
                resp.json.return_value = _mock_tree_response([("app.py", 10)])
            elif "/git/blobs/" in url:
                resp.json.return_value = _mock_blob_response(py_code)
            return resp

        mock_get.side_effect = side_effect

        result = analyze_repo("https://github.com/test/pyrepo")
        assert result["file_scores"][0]["language"] == "python"

    @patch("ai.repo_analyzer.httpx.get")
    def test_suggestions_include_language_summary(self, mock_get):
        """Suggestions include a summary of analyzed languages."""
        go_code = "package main\n\nfunc main() {}\n"

        def side_effect(url, **kwargs):
            resp = MagicMock()
            resp.status_code = 200
            resp.raise_for_status = MagicMock()
            if "/git/trees/" not in url and "/git/blobs/" not in url:
                resp.json.return_value = {"default_branch": "main"}
            elif "/git/trees/" in url:
                resp.json.return_value = _mock_tree_response([("main.go", 30)])
            elif "/git/blobs/" in url:
                resp.json.return_value = _mock_blob_response(go_code)
            return resp

        mock_get.side_effect = side_effect

        result = analyze_repo("https://github.com/test/repo")
        assert any("Analyzed" in s and "go" in s for s in result["suggestions"])

    @patch("ai.repo_analyzer.httpx.get")
    def test_caps_at_max_files(self, mock_get):
        """Only MAX_FILES files are analyzed even if more are available."""
        files = [(f"file{i}.py", 10) for i in range(MAX_FILES + 10)]
        py_code = "x = 1\n"

        def side_effect(url, **kwargs):
            resp = MagicMock()
            resp.status_code = 200
            resp.raise_for_status = MagicMock()
            if "/git/trees/" not in url and "/git/blobs/" not in url:
                resp.json.return_value = {"default_branch": "main"}
            elif "/git/trees/" in url:
                resp.json.return_value = _mock_tree_response(files)
            elif "/git/blobs/" in url:
                resp.json.return_value = _mock_blob_response(py_code)
            return resp

        mock_get.side_effect = side_effect

        result = analyze_repo("https://github.com/test/big")
        assert result["files_analyzed"] == MAX_FILES


# --- Evaluator Integration ---


class TestEvaluatorRepoEndpoint:
    """Test the /evaluate endpoint with repo submissions."""

    @pytest.fixture
    def client(self):
        from fastapi.testclient import TestClient
        from ai.evaluator import app
        return TestClient(app)

    @patch("ai.evaluator.analyze_repo")
    def test_repo_evaluation(self, mock_analyze, client):
        """Repo submission returns aggregated results."""
        mock_analyze.return_value = {
            "score": 75,
            "issues": [
                {"issue_type": "security", "message": "[main.go] exec.Command", "line": 5}
            ],
            "suggestions": ["Security issues detected"],
            "files_analyzed": 3,
            "file_scores": [
                {"path": "main.go", "score": 60, "issues": 1, "language": "go"},
                {"path": "lib.py", "score": 85, "issues": 0, "language": "python"},
                {"path": "README.md", "score": 80, "issues": 0, "language": None},
            ],
        }

        resp = client.post("/evaluate", json={
            "submission_type": "repo",
            "name": "https://github.com/test/repo",
            "content": "",
            "mime": "",
        })
        assert resp.status_code == 200
        data = resp.json()
        assert data["score"] == 75
        assert data["files_analyzed"] == 3
        assert len(data["file_scores"]) == 3
        assert len(data["issues"]) == 1

    @patch("ai.evaluator.analyze_repo")
    def test_repo_not_found_returns_400(self, mock_analyze, client):
        """Invalid repo URL returns 400."""
        mock_analyze.side_effect = ValueError("repository not found: bad/repo")

        resp = client.post("/evaluate", json={
            "submission_type": "repo",
            "name": "https://github.com/bad/repo",
            "content": "",
            "mime": "",
        })
        assert resp.status_code == 400
        assert "not found" in resp.json()["detail"]

    @patch("ai.evaluator.analyze_repo")
    def test_repo_network_error_returns_502(self, mock_analyze, client):
        """Network failure returns 502."""
        mock_analyze.side_effect = Exception("connection timeout")

        resp = client.post("/evaluate", json={
            "submission_type": "repo",
            "name": "https://github.com/test/repo",
            "content": "",
            "mime": "",
        })
        assert resp.status_code == 502
        assert "failed to fetch" in resp.json()["detail"]
