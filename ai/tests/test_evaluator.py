"""
Tests for the PoSA AI evaluation engine.

Uses FastAPI's TestClient (backed by httpx) to test endpoints
without starting a real server.
"""

import pytest
from fastapi.testclient import TestClient

from ai.evaluator import app


@pytest.fixture
def client():
    """Create a test client for the FastAPI app."""
    return TestClient(app)


# --- Health Check Tests ---


def test_health(client):
    """Health endpoint returns 200 with status ok."""
    resp = client.get("/health")
    assert resp.status_code == 200
    assert resp.json() == {"status": "ok"}


def test_health_post_not_allowed(client):
    """POST to health endpoint is rejected (GET only)."""
    resp = client.post("/health")
    assert resp.status_code == 405


# --- Evaluation Endpoint Tests ---


def test_evaluate_file_submission(client):
    """Valid file submission returns a score and response structure."""
    payload = {
        "submission_type": "file",
        "name": "main.go",
        "content": "package main\n\nfunc main() {}\n",
        "mime": "text/plain",
    }
    resp = client.post("/evaluate", json=payload)
    assert resp.status_code == 200

    data = resp.json()
    assert "score" in data
    assert 0 <= data["score"] <= 100
    assert isinstance(data["issues"], list)
    assert isinstance(data["suggestions"], list)


def test_evaluate_repo_submission(client):
    """Valid repo submission returns a score (content can be empty)."""
    payload = {
        "submission_type": "repo",
        "name": "https://github.com/user/project",
        "content": "",
        "mime": "",
    }
    resp = client.post("/evaluate", json=payload)
    assert resp.status_code == 200

    data = resp.json()
    assert 0 <= data["score"] <= 100


def test_evaluate_file_empty_content(client):
    """File submission with empty content is rejected."""
    payload = {
        "submission_type": "file",
        "name": "empty.go",
        "content": "",
        "mime": "text/plain",
    }
    resp = client.post("/evaluate", json=payload)
    assert resp.status_code == 400


def test_evaluate_file_whitespace_content(client):
    """File submission with whitespace-only content is rejected."""
    payload = {
        "submission_type": "file",
        "name": "blank.py",
        "content": "   \n\t\n  ",
        "mime": "text/plain",
    }
    resp = client.post("/evaluate", json=payload)
    assert resp.status_code == 400


def test_evaluate_invalid_submission_type(client):
    """Invalid submission_type is rejected by Pydantic validation."""
    payload = {
        "submission_type": "invalid",
        "name": "test.go",
        "content": "data",
    }
    resp = client.post("/evaluate", json=payload)
    assert resp.status_code == 422  # Pydantic validation error


def test_evaluate_missing_required_fields(client):
    """Missing required fields are rejected by Pydantic validation."""
    resp = client.post("/evaluate", json={})
    assert resp.status_code == 422


def test_evaluate_missing_name(client):
    """Missing name field is rejected."""
    payload = {"submission_type": "file", "content": "data"}
    resp = client.post("/evaluate", json=payload)
    assert resp.status_code == 422


def test_evaluate_name_too_long(client):
    """Name exceeding 500 characters is rejected."""
    payload = {
        "submission_type": "file",
        "name": "a" * 501,
        "content": "data",
    }
    resp = client.post("/evaluate", json=payload)
    assert resp.status_code == 422


def test_evaluate_empty_body(client):
    """Empty request body is rejected."""
    resp = client.post("/evaluate", content=b"", headers={"Content-Type": "application/json"})
    assert resp.status_code == 422


def test_evaluate_invalid_json(client):
    """Malformed JSON is rejected."""
    resp = client.post("/evaluate", content=b"{bad json", headers={"Content-Type": "application/json"})
    assert resp.status_code == 422


def test_evaluate_response_structure(client):
    """Response matches the EvaluationResponse schema exactly."""
    payload = {
        "submission_type": "file",
        "name": "test.py",
        "content": "print('hello')\n",
        "mime": "text/plain",
    }
    resp = client.post("/evaluate", json=payload)
    data = resp.json()

    # Verify all expected keys are present.
    assert set(data.keys()) == {"score", "issues", "suggestions"}

    # Verify types.
    assert isinstance(data["score"], int)
    assert isinstance(data["issues"], list)
    assert isinstance(data["suggestions"], list)
