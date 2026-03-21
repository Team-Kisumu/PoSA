# Backend API

The Go backend is the central orchestrator of the PoSA pipeline — it receives user submissions, forwards them to the AI engine, stores reports on IPFS, and anchors proofs on-chain.

## Technology

- **Language:** Go 1.25+
- **Framework:** `net/http` standard library (no external dependencies)
- **Port:** `:8080`

## Directory Structure

```md
backend/
├── main.go                      # Server entrypoint, middleware chain, route registration
├── handlers/
│   ├── types.go                 # Type definitions, validators (filename, URL, CID)
│   ├── types_test.go            # Validator unit tests
│   ├── handlers.go              # Health, Verify, respondOK/respondError, writeJSON
│   ├── handlers_test.go         # Unit tests for Health, Verify, CID validation
│   ├── submit.go                # Submit handler (file upload + repo link)
│   ├── submit_test.go           # Unit tests for Submit (security edge cases)
│   └── integration_test.go      # Integration tests (mux routing, middleware, envelope)
├── middleware/
│   ├── middleware.go             # SecurityHeaders, RequestID, Recovery
│   └── middleware_test.go        # Middleware unit tests
├── services/
│   └── services.go              # Business logic (AI, IPFS, blockchain orchestration)
├── blockchain/
│   └── blockchain.go            # Chain interaction layer
└── go.mod                       # Module definition
```

## Security Architecture

All requests pass through a middleware chain before reaching handlers:

```mmd
Request → Recovery → RequestID → SecurityHeaders → ServeMux → Handler
```

### Middleware

| Middleware | Purpose |
|---|---|
| `Recovery` | Catches panics, returns 500 JSON instead of crashing |
| `RequestID` | Generates unique 16-char hex ID per request (`X-Request-ID` header) |
| `SecurityHeaders` | Sets `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Content-Security-Policy: default-src 'none'`, `Referrer-Policy: no-referrer`, `Cache-Control: no-store` |

### Data Validation

All user input is validated through dedicated functions in `types.go`:

| Validator | Protects Against |
|---|---|
| `ValidateFilename` | Path traversal (`../`), null bytes, backslashes, blocked extensions (`.exe`, `.sh`, `.bat`, `.dll`, etc.), special characters |
| `ValidateRepoURL` | Non-HTTPS, non-GitHub hosts, missing owner/repo, path traversal in URL, malformed URLs |
| `ValidateCID` | Injection characters (`;`, `\|`, `<`, `>`, `"`, `'`, `&`), null bytes, slashes, invalid format/length |

### Input Limits

| Limit | Value | Enforced By |
|---|---|---|
| File upload size | 10MB | `http.MaxBytesReader` |
| JSON body size | 1MB | `http.MaxBytesReader` |
| Unknown JSON fields | Rejected | `decoder.DisallowUnknownFields()` |

## API Response Envelope

All responses use a consistent envelope:

**Success:**

```json
{
  "success": true,
  "data": { ... }
}
```

**Error:**

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable description"
  }
}
```

## API Reference

### `GET /health`

**Response:** `200 OK`

```json
{"success": true, "data": {"status": "ok"}}
```

### `POST /api/submit`

#### File Upload — `multipart/form-data`

```bash
curl -X POST http://localhost:8080/api/submit -F "file=@main.go"
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {"type": "file", "name": "main.go", "size": 142, "message": "file received, pending evaluation"}
}
```

| Status | Code | Condition |
|---|---|---|
| `400` | `MISSING_CONTENT_TYPE` | No Content-Type header |
| `400` | `MISSING_FILE` | Missing `file` form field |
| `400` | `INVALID_FILENAME` | Path traversal, null bytes, blocked extension |
| `400` | `EMPTY_FILE` | 0-byte file |
| `413` | `FILE_TOO_LARGE` | Exceeds 10MB |
| `415` | `UNSUPPORTED_MEDIA_TYPE` | Wrong Content-Type |

#### Repo Link — `application/json`

```bash
curl -X POST http://localhost:8080/api/submit \
  -H "Content-Type: application/json" \
  -d '{"repo": "https://github.com/user/project"}'
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {"type": "repo", "name": "https://github.com/user/project", "size": 0, "message": "repo link received, pending evaluation"}
}
```

| Status | Code | Condition |
|---|---|---|
| `400` | `INVALID_JSON` | Malformed JSON or unknown fields |
| `400` | `INVALID_REPO` | Empty, non-HTTPS, non-GitHub, missing owner/repo, path traversal |

### `GET /api/verify/{cid}`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {"cid": "QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco", "message": "verify endpoint not yet implemented"}
}
```

| Status | Code | Condition |
|---|---|---|
| `400` | `INVALID_CID` | Empty, injection characters, invalid format/length |

## Type Definitions

```go
// Response envelope
type APIResponse struct {
    Success bool   `json:"success"`
    Data    any    `json:"data,omitempty"`
    Error   *Error `json:"error,omitempty"`
}

type Error struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

// Domain types
type HealthData     struct { Status string `json:"status"` }
type SubmitResponse struct { Type string; Name string; Size int64; Message string }
type VerifyResponse struct { CID string; Message string }
type RepoRequest    struct { Repo string `json:"repo"` }
```

## Running

```bash
cd backend && go run main.go
curl http://localhost:8080/health
```

## Testing

```bash
./scripts/test-backend.sh -v      # verbose with coverage
./scripts/test-backend.sh --html  # HTML coverage report
cd backend && go test -race -v ./...
```

### Coverage

| Package | Coverage |
|---|---|
| `handlers` | 94.3% |
| `middleware` | 100% |
| **Total** | **89.1%** |

### Test Summary

| File | Tests | Covers |
|---|---|---|
| `types_test.go` | 31 | Filename, repo URL, CID validators (valid + invalid inputs) |
| `handlers_test.go` | 7 | Health, Verify (valid CID, empty, injection, too short), writeJSON |
| `submit_test.go` | 19 | File upload, repo link, path traversal, blocked extensions, unknown fields |
| `integration_test.go` | 13 | Mux routing, method enforcement, security headers, request ID, response envelope |
| `middleware_test.go` | 4 | SecurityHeaders, RequestID (uniqueness), Recovery (panic + no-panic) |

## Next Steps

- **Issue #4:** Add input validation and sanitization middleware
