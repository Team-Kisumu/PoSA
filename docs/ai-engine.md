# AI Engine

> **Status:** Code and writing analysis implemented (Phase 2)

The AI engine evaluates user-submitted work using two evaluators:
- **Code analysis:** Pattern-based static analysis for Go, Python, and JavaScript
- **Writing evaluation:** Quality, structure, and readability scoring for text documents

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
├── analyzer.py              # Router: detects file type, delegates to code or writing evaluator
├── writing.py               # Writing quality evaluator (structure, readability, style)
├── requirements.txt         # Direct Python dependencies
├── rules/
│   ├── __init__.py          # Package marker
│   └── patterns.py          # Language-specific code analysis rules (Go, Python, JS)
└── tests/
    ├── __init__.py          # Package marker
    ├── test_analyzer.py     # Code analyzer unit tests (46 test cases)
    ├── test_writing.py      # Writing evaluator unit tests (29 test cases)
    └── test_evaluator.py    # Endpoint integration tests (19 test cases)
```

## Analysis Architecture

### Pipeline

```mmd
File Upload → File Type Detection → Code Analyzer / Writing Evaluator → Scoring → Response
```

The analyzer routes submissions based on file extension:
- Code extensions (.go, .py, .js, .ts) -> code analysis rules
- Writing extensions (.md, .txt, .rst, .html, .tex) -> writing quality checks
- Unknown extensions -> score 0 with suggestion

1. **Language detection** — File extension mapped to language (`EXTENSION_MAP`)
2. **Rule matching** — Source code scanned line-by-line against language-specific regex patterns
3. **Deduplication** — Same issue on the same line reported only once
4. **Scoring** — Each issue deducts points based on severity (high=15, medium=8, low=3)
5. **Suggestions** — Grouped by issue category (security, quality, style)

### Supported Languages

| Language | Extensions | Rules |
|---|---|---|
| Go | `.go` | 10 rules (5 security, 5 quality) |
| Python | `.py` | 11 rules (7 security, 3 quality, 1 style) |
| JavaScript | `.js`, `.jsx`, `.ts`, `.tsx`, `.mjs`, `.cjs` | 11 rules (6 security, 3 quality, 2 style) |
| Text/Writing | `.md`, `.txt`, `.rst`, `.html`, `.tex`, `.adoc` | 6 checks (length, structure, sentences, passive voice, weasel words, repetition) |

### Scoring System

| Severity | Point Deduction | Examples |
|---|---|---|
| High (-15) | Security vulnerabilities | eval(), exec.Command, SQL injection, innerHTML, pickle |
| Medium (-8) | Quality concerns | panic(), bare except, discarded errors, shell=True |
| Low (-3) | Style issues | fmt.Println, console.log, var, TODO comments |

Score = max(0, 100 - sum of deductions). Clean code scores 100.

### Security Rules

| Language | Pattern | Severity |
|---|---|---|
| Go | `exec.Command()` | High |
| Go | SQL query with string concatenation | High |
| Go | Unsanitized request data in response | High |
| Go | `http.ListenAndServe` with nil handler | Medium |
| Python | `eval()`, `exec()` | High |
| Python | `os.system()` | High |
| Python | `subprocess` with `shell=True` | High |
| Python | `pickle.load/loads` | High |
| Python | `yaml.load` without safe Loader | High |
| JavaScript | `eval()`, `new Function()` | High |
| JavaScript | `innerHTML` assignment | High |
| JavaScript | `document.write()` | High |
| JavaScript | `child_process` usage | High |
| JavaScript | `setTimeout` with string argument | Medium |

## API Reference

### `GET /health`

**Response:** `200 OK`

```json
{"status": "ok"}
```

### `POST /evaluate`

**Request:**

```json
{
  "submission_type": "file",
  "name": "main.go",
  "content": "package main\n\nimport \"os/exec\"\n\nfunc run(cmd string) {\n\texec.Command(cmd)\n}\n",
  "mime": "text/plain"
}
```

**Response:** `200 OK`

```json
{
  "score": 85,
  "issues": [
    {
      "issue_type": "security",
      "message": "Use of exec.Command — verify input is sanitized to prevent command injection",
      "line": 6
    }
  ],
  "suggestions": [
    "Security issues detected — review and fix before deployment"
  ]
}
```

**Error responses:**

| Status | Condition |
|---|---|
| `400` | File submission with empty content |
| `422` | Invalid submission_type, missing required fields, name too long, malformed JSON |

## Running

```bash
cd ai
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt

# Start the server
python evaluator.py
# -> Uvicorn running on http://0.0.0.0:8000

# Interactive docs
# Swagger UI: http://localhost:8000/docs
# ReDoc: http://localhost:8000/redoc
```

## Testing

```bash
# From project root (with venv activated)
PYTHONPATH=. pytest ai/tests/ -v

# 65 tests, ~0.8s
```

### Test Summary

| File | Tests | Covers |
|---|---|---|
| `test_analyzer.py` | 46 | Language detection (8), Go rules (7), Python rules (10), JS rules (10), scoring (4), suggestions (3), edge cases (4) |
| `test_writing.py` | 29 | File type detection (8), quality scoring (4), structure (4), passive voice (2), weasel words (2), sentence length (1), suggestions (3), analyzer routing (3), edge cases (2) |
| `test_evaluator.py` | 19 | Health (2), endpoint integration with analyzer (7), validation (8), response structure (2) |
| **Total** | **94** | |

## Next Steps

- Future: LLM-based analysis for deeper semantic understanding
- Future: Plagiarism detection for writing submissions
