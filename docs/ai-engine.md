# AI Engine

> **Status:** 📋 Planned (Phase 2)

The AI engine evaluates user-submitted work — code, writing, and tasks — and produces structured evaluation reports with scores, issues, and suggestions.

## Technology

- **Language:** Python 3.10+
- **Framework:** Flask or FastAPI
- **Dependencies:** Defined in `ai/requirements.txt`

## Planned Directory Structure

```zsh
ai/
├── evaluator.py             # Main evaluation engine
├── requirements.txt         # Python dependencies
└── tests/                   # Test suite
```

## How It Will Work

1. Backend sends a submission payload (code, text, or repo data) via HTTP
2. The engine runs analysis pipelines:
   - **Code analysis:** Static analysis for bugs, security vulnerabilities, logic correctness
   - **Writing evaluation:** Quality, originality, structure assessment
3. A score (0–100) is generated with detailed explanations
4. Structured JSON report is returned to the backend

## Planned API

### `POST /evaluate`

**Request:**

```json
{
  "type": "code",
  "language": "go",
  "content": "package main..."
}
```

**Response:**

```json
{
  "score": 82,
  "issues": [
    {"type": "security", "severity": "high", "message": "Unvalidated user input at line 45"}
  ],
  "suggestions": [
    "Add input sanitization before DB query"
  ]
}
```

## Testing

Tests will be runnable via:

```bash
# From project root
./scripts/test.sh

# Or directly
cd ai && python3 -m pytest --tb=short -q
```

## Related Issues

- **#5:** Set up Python evaluation engine scaffold
- **#6:** Implement code quality analysis
- **#7:** Implement writing evaluation module
