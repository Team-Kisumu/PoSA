# Testing

This document covers the test strategy, tooling, and instructions for running tests across all PoSA components.

## Test Strategy

Each component has its own test suite using the idiomatic testing framework for its language:

| Component | Framework | Coverage Target | Current |
|---|---|---|---|
| Backend (Go) | `testing` + `httptest` | 80%+ | 87.8% ✅ |
| AI Engine (Python) | `pytest` | 80%+ | — |
| Frontend (Node) | Jest / Vitest | 70%+ | — |
| Smart Contracts | Flow Test / cargo test | 100% | — |

Tests are split into **unit tests** (isolated handler logic) and **integration tests** (full mux routing with method enforcement).

## Test Scripts

All scripts are in `scripts/` and executable from the project root.

### Run all tests

```bash
./scripts/test.sh
```

Runs tests for every stack that has a directory present. Skips missing components gracefully. Exits non-zero if any stack fails — suitable for CI.

### Backend tests only

```bash
# Standard run
./scripts/test-backend.sh

# Verbose — shows individual test names
./scripts/test-backend.sh -v

# Generate HTML coverage report
./scripts/test-backend.sh --html
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

Monitors for file changes and re-runs the relevant test suite automatically.

## Backend Tests

### Unit tests — `handlers_test.go`

| Test | Endpoint | Verifies |
|---|---|---|
| `TestHealth` | `GET /health` | 200 status, `{"status":"ok"}`, JSON Content-Type |
| `TestVerify/valid_cid` | `GET /api/verify/{cid}` | CID extraction, 501 with CID in body |
| `TestVerify/empty_cid` | `GET /api/verify/` | 400 with error for missing CID |
| `TestWriteJSON` | (internal) | Status code, Content-Type, JSON encoding |

### Unit tests — `submit_test.go`

| Test | Verifies |
|---|---|
| `TestSubmitFileUpload` | Valid file upload → 200, correct type/name/size/message |
| `TestSubmitFileUploadMissingField` | Wrong form field name → 400 |
| `TestSubmitFileUploadEmptyFile` | 0-byte file → 400 |
| `TestSubmitRepoLink` | Valid GitHub URL → 200, correct type/name/message |
| `TestSubmitRepoLinkEmpty` | Empty repo string → 400 |
| `TestSubmitRepoLinkWhitespace` | Whitespace-only repo → 400 |
| `TestSubmitRepoLinkInvalidURL/gitlab` | GitLab URL → 400 |
| `TestSubmitRepoLinkInvalidURL/http` | HTTP (not HTTPS) → 400 |
| `TestSubmitRepoLinkInvalidURL/bare_string` | Non-URL string → 400 |
| `TestSubmitInvalidJSON` | Malformed JSON → 400 |
| `TestSubmitEmptyJSONBody` | Empty `{}` body → 400 |
| `TestSubmitUnsupportedContentType` | `text/plain` → 415 |
| `TestSubmitMissingContentType` | No Content-Type header → 400 |

### Integration tests — `integration_test.go`

Tests the full `http.ServeMux` routing — method enforcement, path matching, and end-to-end request flow.

| Test | Verifies |
|---|---|
| `TestIntegrationHealthEndpoint/GET_returns_200` | Full mux routing for health |
| `TestIntegrationHealthEndpoint/POST_not_allowed` | POST method rejected |
| `TestIntegrationSubmitEndpoint/POST_file_upload_through_mux` | File upload through full mux |
| `TestIntegrationSubmitEndpoint/POST_repo_link_through_mux` | Repo link through full mux |
| `TestIntegrationSubmitEndpoint/GET_not_allowed` | GET method rejected |
| `TestIntegrationVerifyEndpoint/GET_with_valid_CID_through_mux` | CID routing through full mux |
| `TestIntegrationVerifyEndpoint/POST_not_allowed` | POST method rejected |
| `TestIntegrationUnknownRoute` | Unknown path returns non-200 |

### Current coverage

```
handlers.go:8:   Health           100.0%
handlers.go:12:  Verify           100.0%
handlers.go:24:  writeJSON        100.0%
submit.go:23:    Submit           100.0%
submit.go:44:    handleFileUpload  70.0%
submit.go:93:    handleRepoLink   100.0%
main.go:10:      main               0.0%
total:           (statements)      76.8%
```

> Handler-only coverage: **87.8%**. The uncovered 12.2% is the `io.ReadAll` failure and oversized file `MaxBytesReader` error paths in `handleFileUpload`, which are difficult to trigger without mocking the reader.

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
go test -run TestSubmitFileUpload ./handlers/

# Only integration tests
go test -run TestIntegration ./handlers/

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
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
