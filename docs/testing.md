# Testing

This document covers the test strategy, tooling, and instructions for running tests across all PoSA components.

## Test Strategy

Each component has its own test suite using the idiomatic testing framework for its language:

| Component | Framework | Coverage Target | Current |
|---|---|---|---|
| Backend handlers (Go) | `testing` + `httptest` | 80%+ | 94.3% |
| Backend middleware (Go) | `testing` + `httptest` | 100% | 100% |
| AI Engine (Python) | `pytest` | 80%+ | — |
| Frontend (Node) | Jest / Vitest | 70%+ | — |
| Smart Contracts | Flow Test / cargo test | 100% | — |

Tests are split into **validator tests** (input validation), **unit tests** (isolated handler logic), **integration tests** (full mux routing + middleware), and **middleware tests**.

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
| `TestHealth` | `GET /health` | 200, typed `HealthData` in envelope |
| `TestVerify/valid_cid` | `GET /api/verify/{cid}` | CID extraction, typed `VerifyResponse` |
| `TestVerify/empty_cid` | `GET /api/verify/` | 400, `INVALID_CID` error code |
| `TestVerify/cid_with_injection_chars` | `GET /api/verify/abc;rm+-rf` | 400, injection blocked |
| `TestVerify/cid_too_short` | `GET /api/verify/abc` | 400, format validation |
| `TestWriteJSON` | (internal) | Typed `APIResponse` envelope encoding |

### Unit tests — `submit_test.go`

| Test | Verifies |
|---|---|
| `TestSubmitFileUpload` | Valid file upload, typed `SubmitResponse` |
| `TestSubmitFileUploadMissingField` | Wrong field name, `MISSING_FILE` |
| `TestSubmitFileUploadEmptyFile` | 0-byte file, `EMPTY_FILE` |
| `TestSubmitFileUploadPathTraversal/backslash` | Backslash traversal, `INVALID_FILENAME` |
| `TestSubmitFileUploadPathTraversal/dot_dot_slash` | `filepath.Base` sanitizes traversal |
| `TestSubmitFileUploadBlockedExtension` | `.exe`, `.bat`, `.sh`, `.ps1`, `.dll`, `.cmd` blocked |
| `TestSubmitRepoLink` | Valid GitHub URL, 200 |
| `TestSubmitRepoLinkEmpty` | Empty repo, `INVALID_REPO` |
| `TestSubmitRepoLinkWhitespace` | Whitespace-only, `INVALID_REPO` |
| `TestSubmitRepoLinkInvalidURL` | gitlab, http, bare string, path traversal, no repo, just domain |
| `TestSubmitInvalidJSON` | Malformed JSON, `INVALID_JSON` |
| `TestSubmitEmptyJSONBody` | `{}` body, `INVALID_REPO` |
| `TestSubmitUnknownJSONFields` | Extra fields rejected, `INVALID_JSON` |
| `TestSubmitUnsupportedContentType` | `text/plain`, 415 |
| `TestSubmitMissingContentType` | No header, 400 |

### Integration tests — `integration_test.go`

Tests the full `http.ServeMux` routing with middleware chain.

| Test | Verifies |
|---|---|
| `TestIntegrationHealthEndpoint/GET_returns_200` | Full mux routing |
| `TestIntegrationHealthEndpoint/POST_not_allowed` | Method enforcement |
| `TestIntegrationSubmitEndpoint/POST_file_upload_through_mux` | File upload end-to-end |
| `TestIntegrationSubmitEndpoint/POST_repo_link_through_mux` | Repo link end-to-end |
| `TestIntegrationSubmitEndpoint/GET_not_allowed` | Method enforcement |
| `TestIntegrationVerifyEndpoint/GET_with_valid_CID_through_mux` | CID routing |
| `TestIntegrationVerifyEndpoint/POST_not_allowed` | Method enforcement |
| `TestIntegrationUnknownRoute` | 404 for unknown paths |
| `TestIntegrationSecurityHeaders` | All 5 security headers present |
| `TestIntegrationRequestID` | Unique 16-char hex ID per request |
| `TestIntegrationResponseEnvelope/success` | `success` + `data` fields |
| `TestIntegrationResponseEnvelope/error` | `success` + `error` fields |

### Validator tests — `types_test.go`

| Test | Cases |
|---|---|
| `TestValidateFilename` | 6 valid + 13 invalid (empty, null byte, slashes, traversal, blocked extensions, special chars) |
| `TestValidateRepoURL` | 3 valid + 9 invalid (empty, http, gitlab, no repo, traversal, bare string) |
| `TestValidateCID` | 2 valid + 9 invalid (empty, short, injection, null byte, pipe, brackets, quotes, slashes) |

### Middleware tests — `middleware_test.go`

| Test | Verifies |
|---|---|
| `TestSecurityHeaders` | All 6 headers set correctly |
| `TestRequestID` | Present, 16 chars, unique across requests |
| `TestRecovery` | Panic caught, 500 JSON response |
| `TestRecoveryNoPanic` | Normal flow unaffected |

### Current coverage

```md
handlers.go:8:    Health           100.0%
handlers.go:12:   Verify           100.0%
handlers.go:24:   respondOK        100.0%
handlers.go:28:   respondError     100.0%
handlers.go:32:   writeJSON        100.0%
submit.go:11:     Submit           100.0%
submit.go:29:     handleFileUpload  81.0%
submit.go:69:     handleRepoLink   100.0%
types.go:63:      ValidateFilename  92.3%
types.go:84:      ValidateRepoURL  100.0%
types.go:109:     ValidateCID      100.0%
middleware.go:10: SecurityHeaders  100.0%
middleware.go:22: RequestID        100.0%
middleware.go:30: Recovery         100.0%
middleware.go:44: generateID       100.0%
main.go:11:       main               0.0%
total:            (statements)      89.1%
```

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

# Only validator tests
go test -run TestValidate ./handlers/

# Only middleware tests
go test -v ./middleware/

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## CI Integration

Tests run automatically in GitHub Actions via `.github/workflows/ci-cd.yml`:

- **Backend:** `go vet` then `go test -race -coverprofile` then `go build`
- **AI Engine:** `flake8` lint then `pytest`
- **Frontend:** `npm ci` then `npm run lint` then `npm run build`

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
    resp := decodeResponse(t, w)
    assertSuccess(t, resp)
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
