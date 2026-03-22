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
│   ├── types.go                 # Type definitions (APIResponse, SubmitResponse, etc.)
│   ├── types_test.go            # Type serialization tests (envelope, MIME field)
│   ├── handlers.go              # Health, Verify, respondOK/respondError, writeJSON
│   ├── handlers_test.go         # Unit tests for Health, Verify, CID validation
│   ├── submit.go                # Submit handler (file upload + repo link)
│   ├── submit_test.go           # Unit tests for Submit (MIME, content scan, security)
│   └── integration_test.go      # Integration tests (mux routing, middleware, envelope)
├── validation/
│   ├── validation.go            # All validators: filename, MIME, content, repo URL, CID
│   └── validation_test.go       # Validator unit tests (filename, MIME, content, URL, CID)
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

### File Upload Validation Pipeline

File uploads pass through six validation steps in order. Each step must pass before the next runs:

```mmd
Size Limit → Filename Validation → Content Read → Empty Check → MIME Sniffing → Content Scan
```

| Step | Validator | Protects Against |
|---|---|---|
| 1. Size limit | `MaxBytesReader` (10MB) | Denial-of-service via large uploads |
| 2. Filename | `validation.Filename` | Path traversal, null bytes, blocked extensions, special chars |
| 3. Content read | `io.ReadAll` | — |
| 4. Empty check | Length check | No-op submissions |
| 5. MIME sniffing | `validation.ContentType` | Renamed binaries (ELF named `.go`, PNG named `.py`) |
| 6. Content scan | `validation.Content` | Binary magic bytes, null bytes, shebangs in non-script files |

### Validation Package

All validation logic lives in `backend/validation/` as pure functions:

| Validator | Protects Against |
|---|---|
| `Filename` | Path traversal (`../`), null bytes, backslashes, blocked extensions (`.exe`, `.sh`, `.bat`, `.dll`, `.bin`, `.elf`, `.class`, `.jar`, etc.), special characters |
| `ContentType` | Binary MIME types (images, archives, executables, PDFs). Whitelist: `text/plain`, `text/html`, `text/xml`, `application/json`, `application/xml` |
| `Content` | ELF/PE/Mach-O/Java class magic bytes, ZIP/gzip/RAR/PDF headers, embedded null bytes (first 8KB), shell shebangs in non-script files |
| `RepoURL` | Non-HTTPS, non-GitHub hosts, missing owner/repo, path traversal in URL, malformed URLs |
| `CID` | Injection characters (`;`, `\|`, `<`, `>`, `"`, `'`, `&`), null bytes, slashes, invalid format/length |

### Allowed MIME Types

| MIME Type | Allowed | Reason |
|---|---|---|
| `text/plain` | Yes | Source code, markdown, plain text |
| `text/html` | Yes | HTML files |
| `text/xml` | Yes | XML, SVG, config files |
| `application/json` | Yes | JSON config/data files |
| `application/xml` | Yes | XML variants |
| `application/octet-stream` | No | Generic binary |
| `application/zip` | No | Archives |
| `application/pdf` | No | PDFs |
| `image/*` | No | Images |
| `audio/*`, `video/*` | No | Media files |

### Dangerous Binary Headers

| Format | Magic Bytes | Detected By |
|---|---|---|
| ELF binary | `\x7fELF` | `validation.Content` |
| PE executable | `MZ` (`\x4d\x5a`) | `validation.Content` |
| Mach-O binary | `\xcf\xfa\xed\xfe` | `validation.Content` |
| Java class | `\xca\xfe\xba\xbe` | `validation.Content` |
| gzip archive | `\x1f\x8b` | `validation.Content` |
| ZIP archive | `PK\x03\x04` | `validation.Content` |
| PDF document | `%PDF` | `validation.Content` |
| RAR archive | `Rar!\x1a` | `validation.Content` |

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
  "data": {"type": "file", "name": "main.go", "size": 142, "mime": "text/plain", "message": "file received, pending evaluation"}
}
```

| Status | Code | Condition |
|---|---|---|
| `400` | `MISSING_CONTENT_TYPE` | No Content-Type header |
| `400` | `MISSING_FILE` | Missing `file` form field |
| `400` | `INVALID_FILENAME` | Path traversal, null bytes, blocked extension |
| `400` | `EMPTY_FILE` | 0-byte file |
| `400` | `INVALID_CONTENT_TYPE` | Binary MIME type detected (image, archive, executable) |
| `400` | `MALICIOUS_CONTENT` | Binary magic bytes, null bytes, or shebang in non-script file |
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
type SubmitResponse struct { Type string; Name string; Size int64; MIME string; Message string }
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
| `handlers` | 93.2% |
| `validation` | 97.0% |
| `middleware` | 100% |
| **Total** | **91.0%** |

## Next Steps

- **Issue #5:** AI evaluation engine integration
- **Issue #6:** IPFS/Filecoin storage integration
