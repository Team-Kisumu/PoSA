# AI Engine

> **Status:** Scaffold implemented (Phase 2)

The AI engine evaluates user-submitted work — code, writing, and tasks — and produces structured evaluation reports with scores, issues, and suggestions.

## Technology

- **Language:** Python 3.10+
- **Framework:** FastAPI 0.115+
- **Server:** Uvicorn
- **Testing:** pytest + httpx (via FastAPI TestClient)
- **Dependencies:** Defined in `ai/requirements.txt`

## Directory Structure

```md
ai/
├── __init__.py              # Package marker
├── evaluator.py             # FastAPI app with health and evaluation endpoints
├── requirements.txt         # Pinned Python dependencies
└── tests/
    ├── __init__.py          # Package marker
    └── test_evaluator.py    # Endpoint tests (13 test cases)
```

## API Reference

### `GET /health`

Health check endpoint for monitoring and load balancer probes.

**Response:** `200 OK`

```json
{"status": "ok"}
```

### `POST /evaluate`

Receives a submission payload from the Go backend and returns an evaluation result.

**Request:**

```json
{
  "submission_type": "file",
  "name": "main.go",
  "content": "package main\n\nfunc main() {}\n",
  "mime": "text/plain"
}
```

| Field | Type | Required | Constraints |
|---|---|---|---|
| `submission_type` | string | Yes | Must be `"file"` or `"repo"` |
| `name` | string | Yes | 1-500 characters |
| `content` | string | No | File content (required non-empty for file submissions) |
| `mime` | string | No | Detected MIME type from backend |

**Response:** `200 OK`

```json
{
  "score": 82,
  "issues": [
    {"issue_type": "security", "message": "Unvalidated user input at line 45", "line": 45}
  ],
  "suggestions": [
    "Add input sanitization before DB query"
  ]
}
```

| Field | Type | Description |
|---|---|---|
| `score` | int | 0-100 evaluation score |
| `issues` | list | Found issues with type, message, and optional line number |
| `suggestions` | list | Improvement suggestions |

**Error responses:**

| Status | Condition |
|---|---|
| `400` | File submission with empty content |
| `422` | Invalid submission_type, missing required fields, name too long, malformed JSON |

## Data Models

```python
class EvaluationRequest(BaseModel):
    submission_type: str  # "file" or "repo"
    name: str             # filename or repo URL (1-500 chars)
    content: str          # file content (empty for repo)
    mime: str             # detected MIME type

class Issue(BaseModel):
    issue_type: str       # "security", "quality", "logic", "style"
    message: str
    line: int | None      # line number, if applicable

class EvaluationResponse(BaseModel):
    score: int            # 0-100
    issues: list[Issue]
    suggestions: list[str]
```

## Running

```bash
# Set up virtual environment
cd ai
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt

# Start the server
python evaluator.py
# -> Uvicorn running on http://0.0.0.0:8000

# Or with uvicorn directly (auto-reload for development)
uvicorn ai.evaluator:app --reload --port 8000
```

### Interactive API Docs

FastAPI auto-generates interactive documentation:

- Swagger UI: `http://localhost:8000/docs`
- ReDoc: `http://localhost:8000/redoc`

## Testing

```bash
# From project root (with venv activated)
PYTHONPATH=. pytest ai/tests/ -v

# Or via the test script
./scripts/test.sh
```

### Test Cases

| Test | Verifies |
|---|---|
| `test_health` | 200 with `{"status": "ok"}` |
| `test_health_post_not_allowed` | POST to /health returns 405 |
| `test_evaluate_file_submission` | Valid file payload returns score + structure |
| `test_evaluate_repo_submission` | Valid repo payload accepted (empty content ok) |
| `test_evaluate_file_empty_content` | Empty content for file type returns 400 |
| `test_evaluate_file_whitespace_content` | Whitespace-only content returns 400 |
| `test_evaluate_invalid_submission_type` | Invalid type returns 422 |
| `test_evaluate_missing_required_fields` | Empty body returns 422 |
| `test_evaluate_missing_name` | Missing name returns 422 |
| `test_evaluate_name_too_long` | Name > 500 chars returns 422 |
| `test_evaluate_empty_body` | Empty request body returns 422 |
| `test_evaluate_invalid_json` | Malformed JSON returns 422 |
| `test_evaluate_response_structure` | Response has exactly score, issues, suggestions |

## Integration with Backend

The Go backend calls the AI engine via HTTP after receiving a user submission:

```mmd
User → Go Backend → AI Engine → Evaluation Result → IPFS → Blockchain
```

The backend will POST to `http://localhost:8000/evaluate` with the submission payload. The AI engine returns the evaluation, which the backend then stores on IPFS and anchors on-chain.

## Next Steps

- **Issue #8:** Implement code quality analysis (static analysis, bug detection)
- **Issue #9:** Implement writing evaluation module
- **Issue #10:** Scoring system with detailed explanations
