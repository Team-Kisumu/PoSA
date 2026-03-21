# Testing

This document covers the test strategy, tooling, and instructions for running tests across all PoSA components.

## Test Strategy

Each component has its own test suite using the idiomatic testing framework for its language:

| Component | Framework | Coverage Target |
|---|---|---|
| Backend (Go) | `testing` + `httptest` | 80%+ |
| AI Engine (Python) | `pytest` | 80%+ |
| Frontend (Node) | Jest / Vitest | 70%+ |
| Smart Contracts | Flow Test / cargo test | 100% |

## Test Scripts

All scripts are in `scripts/` and executable from the project root.

### Run all tests

```bash
./scripts/test.sh
```

Runs tests for every stack that has a directory present. Skips missing components gracefully. Exits non-zero if any stack fails — suitable for CI.

**Output:**
```
═══════════════════════════════════
  PoSA Test Suite
═══════════════════════════════════

▶ Backend (Go)
ok   github.com/Murzuqisah/PoSA/handlers  1.015s  coverage: 100.0%
✓ Backend (Go) passed

⊘ AI Engine (Python) — skipped (directory not found)
⊘ Frontend (Node) — skipped (directory not found)

═══════════════════════════════════
  Results: 1 passed, 0 failed
═══════════════════════════════════
```

### Backend tests only

```bash
# Standard run
./scripts/test-backend.sh

# Verbose — shows individual test names
./scripts/test-backend.sh -v

# Generate HTML coverage report
./scripts/test-backend.sh --html
# Opens backend/coverage.html
```

What it runs:
1. `go vet ./...` — static analysis
2. `go test -race -count=1 -coverprofile=coverage.out ./...` — tests with race detection
3. `go tool cover -func=coverage.out` — per-function coverage breakdown

### Watch mode

```bash
# Watch a specific stack
./scripts/test-watch.sh backend
./scripts/test-watch.sh ai
./scripts/test-watch.sh frontend

# Watch everything
./scripts/test-watch.sh all
```

Monitors for file changes and re-runs the relevant test suite automatically. Uses `inotifywait` if available, otherwise falls back to 2-second polling.

## Backend Tests

Located in `backend/handlers/handlers_test.go`.

| Test | Endpoint | Verifies |
|---|---|---|
| `TestHealth` | `GET /health` | 200 status, `{"status":"ok"}`, JSON Content-Type |
| `TestSubmit` | `POST /api/submit` | 501 status, non-empty placeholder message |
| `TestVerify/valid_cid` | `GET /api/verify/{cid}` | CID path param extraction, 501 with CID in body |
| `TestVerify/empty_cid` | `GET /api/verify/` | 400 status, error message for missing CID |
| `TestWriteJSON` | (internal) | Status code, Content-Type header, JSON encoding |

### Running directly

```bash
cd backend

# All tests
go test ./...

# Verbose
go test -v ./...

# With race detection
go test -race ./...

# Single test
go test -run TestHealth ./handlers/

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Current coverage

```
handlers/handlers.go:8:   Health      100.0%
handlers/handlers.go:12:  Submit      100.0%
handlers/handlers.go:18:  Verify      100.0%
handlers/handlers.go:30:  writeJSON   100.0%
main.go:10:               main          0.0%
total:                    (statements)  58.8%
```

## CI Integration

Tests run automatically in GitHub Actions via `.github/workflows/ci-cd.yml`:

- **Backend:** `go vet` → `go test -race -coverprofile` → `go build`
- **AI Engine:** `flake8` lint → `pytest`
- **Frontend:** `npm ci` → `npm run lint` → `npm run build`

The `pre-push` git hook also runs lint and build checks before allowing a push. Install hooks with:

```bash
./scripts/setup-hooks.sh
```

## Adding New Tests

### Go

Create a `_test.go` file in the same package:

```go
func TestNewFeature(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/endpoint", nil)
    w := httptest.NewRecorder()

    HandlerFunc(w, req)

    assertStatus(t, w.Code, http.StatusOK)
}
```

### Python

Add test files under `ai/tests/` prefixed with `test_`:

```python
def test_evaluate_code():
    result = evaluate({"type": "code", "content": "print('hello')"})
    assert 0 <= result["score"] <= 100
```

### Node

Add test files under `frontend/src/__tests__/`:

```javascript
test('renders upload form', () => {
  render(<UploadForm />);
  expect(screen.getByText('Analyze')).toBeInTheDocument();
});
```
