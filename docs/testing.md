# Testing

This document covers the test strategy, tooling, and instructions for running tests across all PoSA components.

## Test Strategy

Each component has its own test suite using the idiomatic testing framework for its language:

| Component | Framework | Coverage Target | Current |
|---|---|---|---|
| Backend handlers (Go) | `testing` + `httptest` | 80%+ | 93.2% |
| Backend validation (Go) | `testing` | 95%+ | 97.0% |
| Backend middleware (Go) | `testing` + `httptest` | 100% | 100% |
| Backend storage (Go) | `testing` + `httptest` | 80%+ | 8 unit + 5 integration |
| AI analyzer (Python) | `pytest` | 80%+ | 150 tests |
| AI Impulse integration | `pytest` + e2e script | — | 3 e2e suites |
| Frontend (Node) | Jest / Vitest | 70%+ | — |
| Smart Contracts | Flow Test | 100% | — |

Tests are split into **validator tests** (input validation + content filtering), **type tests** (serialization), **unit tests** (isolated handler logic), **integration tests** (full mux routing + middleware + live API), **analyzer tests** (code + writing + quality rules), **storage tests** (IPFS mocks + Beryx live), and **e2e tests** (full pipeline).

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

### Validation tests — `validation/validation_test.go`

| Test | Cases |
|---|---|
| `TestFilename` | 6 valid + 17 invalid (empty, null byte, slashes, traversal, blocked extensions incl. .bin/.elf/.class/.jar, special chars) |
| `TestContentType` | 10 cases: plain text, HTML, JSON, empty, PNG, JPEG, GIF, ZIP, gzip, PDF |
| `TestContent` | 18 cases: valid Go/Python/Ruby, empty, ELF, PE, Mach-O, Java class, ZIP, gzip, PDF, RAR, null bytes (embedded + start), shebangs (Go/txt rejected, Python/JS/Perl/Ruby allowed) |
| `TestRepoURL` | 3 valid + 9 invalid (empty, whitespace, http, gitlab, no repo, traversal, bare string, empty owner) |
| `TestCID` | 2 valid + 9 invalid (empty, short, injection, null byte, pipe, brackets, quotes, slashes) |

### Type tests — `handlers/types_test.go`

| Test | Verifies |
|---|---|
| `TestAPIResponseEnvelopeJSON/success` | `data` present, `error` omitted |
| `TestAPIResponseEnvelopeJSON/error` | `error` present, `data` omitted |
| `TestSubmitResponseMIMEField/with_MIME` | `mime` field included when set |
| `TestSubmitResponseMIMEField/without_MIME` | `mime` field omitted when empty |

### Unit tests — `handlers/handlers_test.go`

| Test | Endpoint | Verifies |
|---|---|---|
| `TestHealth` | `GET /health` | 200, typed `HealthData` in envelope |
| `TestVerify/valid_cid` | `GET /api/verify/{cid}` | CID extraction, typed `VerifyResponse` |
| `TestVerify/empty_cid` | `GET /api/verify/` | 400, `INVALID_CID` error code |
| `TestVerify/cid_with_injection_chars` | `GET /api/verify/abc;rm+-rf` | 400, injection blocked |
| `TestVerify/cid_too_short` | `GET /api/verify/abc` | 400, format validation |
| `TestWriteJSON` | (internal) | Typed `APIResponse` envelope encoding |

### Unit tests — `handlers/submit_test.go`

| Test | Verifies |
|---|---|
| `TestSubmitFileUpload` | Valid file upload, typed `SubmitResponse` with MIME field |
| `TestSubmitFileUploadMissingField` | Wrong field name, `MISSING_FILE` |
| `TestSubmitFileUploadEmptyFile` | 0-byte file, `EMPTY_FILE` |
| `TestSubmitFileUploadPathTraversal/backslash` | Backslash traversal, `INVALID_FILENAME` |
| `TestSubmitFileUploadPathTraversal/dot_dot_slash` | `filepath.Base` sanitizes traversal |
| `TestSubmitFileUploadBlockedExtension` | `.exe`, `.bat`, `.sh`, `.ps1`, `.dll`, `.cmd` blocked |
| `TestSubmitFileUploadBinaryContent/PNG` | PNG disguised as `.go` rejected |
| `TestSubmitFileUploadBinaryContent/JPEG` | JPEG disguised as `.py` rejected |
| `TestSubmitFileUploadBinaryContent/ZIP` | ZIP disguised as `.txt` rejected |
| `TestSubmitFileUploadMaliciousContent/null_bytes` | Null bytes in source code rejected |
| `TestSubmitFileUploadMaliciousContent/shebang_go` | Shell shebang in `.go` file rejected |
| `TestSubmitFileUploadMaliciousContent/shebang_python` | Python shebang allowed |
| `TestSubmitRepoLink` | Valid GitHub URL, 200 |
| `TestSubmitRepoLinkEmpty` | Empty repo, `INVALID_REPO` |
| `TestSubmitRepoLinkWhitespace` | Whitespace-only, `INVALID_REPO` |
| `TestSubmitRepoLinkInvalidURL` | gitlab, http, bare string, path traversal, no repo, just domain |
| `TestSubmitInvalidJSON` | Malformed JSON, `INVALID_JSON` |
| `TestSubmitEmptyJSONBody` | `{}` body, `INVALID_REPO` |
| `TestSubmitUnknownJSONFields` | Extra fields rejected, `INVALID_JSON` |
| `TestSubmitUnsupportedContentType` | `text/plain`, 415 |
| `TestSubmitMissingContentType` | No header, 400 |

### Integration tests — `handlers/integration_test.go`

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

### Middleware tests — `middleware/middleware_test.go`

| Test | Verifies |
|---|---|
| `TestSecurityHeaders` | All 6 headers set correctly |
| `TestRequestID` | Present, 16 chars, unique across requests |
| `TestRecovery` | Panic caught, 500 JSON response |
| `TestRecoveryNoPanic` | Normal flow unaffected |

### Current coverage

```md
handlers/handlers.go:    Health            100.0%
handlers/handlers.go:    Verify            100.0%
handlers/handlers.go:    respondOK         100.0%
handlers/handlers.go:    respondError      100.0%
handlers/handlers.go:    writeJSON         100.0%
handlers/submit.go:      Submit            100.0%
handlers/submit.go:      handleFileUpload   85.7%
handlers/submit.go:      handleRepoLink    100.0%
validation/validation.go: Filename          92.3%
validation/validation.go: ContentType      100.0%
validation/validation.go: Content           94.1%
validation/validation.go: matchPrefix      100.0%
validation/validation.go: RepoURL          100.0%
validation/validation.go: CID              100.0%
middleware/middleware.go: SecurityHeaders   100.0%
middleware/middleware.go: RequestID         100.0%
middleware/middleware.go: Recovery          100.0%
middleware/middleware.go: generateID        100.0%
main.go:                 main                0.0%
total:                   (statements)       91.0%
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

# Only validation tests
go test -v ./validation/

# Only middleware tests
go test -v ./middleware/

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### E2E test — AI pipeline

```bash
source ai/.venv/bin/activate
PYTHONPATH=. python scripts/test_e2e_ai.py
```

Tests local analyzer + Impulse AI + combined analysis. Requires `AI_API_KEY` in `.env`.

### E2E test — Storage

```bash
./scripts/test_e2e_storage.sh
```

Tests storage unit tests, full backend suite, go-synapse dependency, and build.

### Beryx integration tests

```bash
export BERYX_API_TOKEN=your-jwt-token
cd backend && go test -tags=integration -v ./storage/
```

Tests live Filecoin RPC connection, chain queries, and SP Registry via Beryx.

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
