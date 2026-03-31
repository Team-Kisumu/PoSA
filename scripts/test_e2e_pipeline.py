#!/usr/bin/env python3
"""
End-to-end pipeline test for PoSA.

Tests the full flow: submission → AI evaluation → storage → verification.
Requires backend (:8080) and AI engine (:8000) to be running.

Run: PYTHONPATH=. python scripts/test_e2e_pipeline.py
"""

import json
import os
import sys
import time

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import httpx  # noqa: E402

BACKEND = os.environ.get("BACKEND_URL", "http://localhost:8080")
AI_ENGINE = os.environ.get("AI_ENGINE_URL", "http://localhost:8000")
TIMEOUT = 30


def separator(title):
    print(f"\n{'='*60}")
    print(f"  {title}")
    print(f"{'='*60}\n")


def check_services():
    """Verify backend and AI engine are running."""
    separator("SERVICE HEALTH CHECK")

    try:
        resp = httpx.get(f"{BACKEND}/health", timeout=5)
        data = resp.json()
        assert data["success"] and data["data"]["status"] == "ok"
        print(f"  Backend ({BACKEND}): OK")
    except Exception as e:
        print(f"  Backend ({BACKEND}): FAILED - {e}")
        print("  Start with: cd backend && go run main.go")
        return False

    try:
        resp = httpx.get(f"{AI_ENGINE}/health", timeout=5)
        data = resp.json()
        assert data["status"] == "ok"
        print(f"  AI Engine ({AI_ENGINE}): OK")
    except Exception as e:
        print(f"  AI Engine ({AI_ENGINE}): FAILED - {e}")
        print("  Start with: cd ai && PYTHONPATH=.. python evaluator.py")
        return False

    return True


def test_file_upload():
    """E2E: Upload a file → backend validates → returns success."""
    separator("TEST 1: File Upload Flow")

    code = 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("hello")\n}\n'
    files = {"file": ("main.go", code.encode(), "text/plain")}

    resp = httpx.post(f"{BACKEND}/api/submit", files=files, timeout=TIMEOUT)
    data = resp.json()

    assert resp.status_code == 200, f"Expected 200, got {resp.status_code}"
    assert data["success"], f"Expected success, got: {data}"
    assert data["data"]["type"] == "file"
    assert data["data"]["name"] == "main.go"
    assert data["data"]["size"] > 0

    print(f"  Upload: {data['data']['name']} ({data['data']['size']} bytes)")
    print(f"  MIME: {data['data']['mime']}")
    print("  PASS")


def test_file_upload_validation():
    """E2E: Upload invalid files → backend rejects."""
    separator("TEST 2: File Upload Validation")

    # Empty file
    files = {"file": ("empty.txt", b"", "text/plain")}
    resp = httpx.post(f"{BACKEND}/api/submit", files=files, timeout=TIMEOUT)
    data = resp.json()
    assert not data["success"]
    assert data["error"]["code"] == "EMPTY_FILE"
    print("  Empty file rejected: PASS")

    # Blocked extension
    files = {"file": ("payload.exe", b"MZ\x90\x00", "application/octet-stream")}
    resp = httpx.post(f"{BACKEND}/api/submit", files=files, timeout=TIMEOUT)
    data = resp.json()
    assert not data["success"]
    assert data["error"]["code"] == "INVALID_FILENAME"
    print("  Blocked extension rejected: PASS")

    # Missing file field
    resp = httpx.post(
        f"{BACKEND}/api/submit",
        headers={"Content-Type": "multipart/form-data; boundary=---"},
        content=b"-----\r\n",
        timeout=TIMEOUT,
    )
    assert resp.status_code != 200
    print("  Missing file field rejected: PASS")


def test_repo_submission():
    """E2E: Submit a GitHub repo URL → backend validates."""
    separator("TEST 3: Repo Submission Flow")

    resp = httpx.post(
        f"{BACKEND}/api/submit",
        json={"repo": "https://github.com/Team-Kisumu/PoSA"},
        timeout=TIMEOUT,
    )
    data = resp.json()

    assert resp.status_code == 200
    assert data["success"]
    assert data["data"]["type"] == "repo"
    print(f"  Repo: {data['data']['name']}")
    print("  PASS")

    # Invalid repo
    resp = httpx.post(
        f"{BACKEND}/api/submit",
        json={"repo": "https://gitlab.com/user/repo"},
        timeout=TIMEOUT,
    )
    data = resp.json()
    assert not data["success"]
    assert data["error"]["code"] == "INVALID_REPO"
    print("  Invalid repo rejected: PASS")


def test_ai_evaluation():
    """E2E: Submit code to AI engine → get evaluation result."""
    separator("TEST 4: AI Evaluation Flow")

    # Vulnerable Python code
    code = "import os\n\ndef run(cmd):\n    os.system(cmd)\n\ndef load(data):\n    import pickle\n    return pickle.loads(data)\n"

    resp = httpx.post(
        f"{AI_ENGINE}/evaluate",
        json={
            "submission_type": "file",
            "name": "vulnerable.py",
            "content": code,
            "mime": "text/plain",
        },
        timeout=TIMEOUT,
    )
    data = resp.json()

    assert resp.status_code == 200
    assert 0 <= data["score"] <= 100
    assert len(data["issues"]) > 0
    assert any("os.system" in i["message"] for i in data["issues"])

    print(f"  Score: {data['score']}/100")
    print(f"  Issues: {len(data['issues'])}")
    for i in data["issues"][:3]:
        print(f"    [{i['issue_type']}] {i['message']}")
    print("  PASS")

    # Clean Go code
    clean = 'package main\n\nimport "net/http"\n\nfunc handler(w http.ResponseWriter, r *http.Request) {\n\tw.WriteHeader(200)\n}\n'
    resp = httpx.post(
        f"{AI_ENGINE}/evaluate",
        json={"submission_type": "file", "name": "clean.go", "content": clean, "mime": "text/plain"},
        timeout=TIMEOUT,
    )
    data = resp.json()
    assert data["score"] >= 80
    print(f"  Clean code score: {data['score']}/100")
    print("  PASS")


def test_credential_verification():
    """E2E: Verify a CID against the backend."""
    separator("TEST 5: Credential Verification Flow")

    # Valid CID format but not on-chain
    resp = httpx.get(
        f"{BACKEND}/api/verify/QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco",
        timeout=TIMEOUT,
    )
    data = resp.json()
    assert resp.status_code == 200
    assert data["success"]
    print("  Valid CID query: PASS")

    # Invalid CID (injection attempt)
    resp = httpx.get(f"{BACKEND}/api/verify/abc;rm+-rf", timeout=TIMEOUT)
    data = resp.json()
    assert not data["success"]
    assert data["error"]["code"] == "INVALID_CID"
    print("  CID injection blocked: PASS")

    # Empty CID
    resp = httpx.get(f"{BACKEND}/api/verify/", timeout=TIMEOUT)
    assert resp.status_code != 200
    print("  Empty CID rejected: PASS")


def test_cors():
    """E2E: Verify CORS headers are present."""
    separator("TEST 6: CORS Headers")

    resp = httpx.get(
        f"{BACKEND}/health",
        headers={"Origin": "http://localhost:3000"},
        timeout=TIMEOUT,
    )
    assert resp.headers.get("access-control-allow-origin") == "http://localhost:3000"
    print("  Backend CORS: PASS")

    resp = httpx.get(
        f"{AI_ENGINE}/health",
        headers={"Origin": "http://localhost:3000"},
        timeout=TIMEOUT,
    )
    assert "access-control-allow-origin" in resp.headers
    print("  AI Engine CORS: PASS")

    # Preflight OPTIONS
    resp = httpx.options(
        f"{BACKEND}/api/submit",
        headers={"Origin": "http://localhost:3000", "Access-Control-Request-Method": "POST"},
        timeout=TIMEOUT,
    )
    assert resp.status_code == 204
    print("  Preflight OPTIONS: PASS")


def test_security_headers():
    """E2E: Verify security headers on backend responses."""
    separator("TEST 7: Security Headers")

    resp = httpx.get(f"{BACKEND}/health", timeout=TIMEOUT)
    headers = resp.headers

    checks = {
        "x-content-type-options": "nosniff",
        "x-frame-options": "DENY",
        "referrer-policy": "no-referrer",
        "cache-control": "no-store",
    }
    for key, expected in checks.items():
        actual = headers.get(key)
        assert actual == expected, f"{key}: expected {expected}, got {actual}"
        print(f"  {key}: {actual} PASS")

    assert headers.get("x-request-id")
    print(f"  x-request-id: {headers['x-request-id']} PASS")


if __name__ == "__main__":
    print("PoSA End-to-End Pipeline Test")
    print("=" * 60)

    if not check_services():
        print("\nServices not running. Start with: ./scripts/start.sh")
        sys.exit(1)

    passed = 0
    failed = 0
    tests = [
        test_file_upload,
        test_file_upload_validation,
        test_repo_submission,
        test_ai_evaluation,
        test_credential_verification,
        test_cors,
        test_security_headers,
    ]

    for test in tests:
        try:
            test()
            passed += 1
        except AssertionError as e:
            print(f"  FAILED: {e}")
            failed += 1
        except Exception as e:
            print(f"  ERROR: {type(e).__name__}: {e}")
            failed += 1

    separator("RESULTS")
    print(f"  Passed: {passed}/{len(tests)}")
    print(f"  Failed: {failed}/{len(tests)}")
    sys.exit(1 if failed > 0 else 0)
