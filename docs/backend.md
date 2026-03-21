# Backend API

The Go backend is the central orchestrator of the PoSA pipeline — it receives user submissions, forwards them to the AI engine, stores reports on IPFS, and anchors proofs on-chain.

## Technology

- **Language:** Go 1.25+
- **Framework:** `net/http` standard library (no external dependencies)
- **Port:** `:8080`

## Directory Structure

```
backend/
├── main.go                  # Server entrypoint, route registration
├── handlers/
│   ├── handlers.go          # Health, Verify handlers + writeJSON helper
│   ├── handlers_test.go     # Unit tests for Health, Verify, writeJSON
│   ├── submit.go            # Submit handler (file upload + repo link)
│   ├── submit_test.go       # Unit tests for Submit (edge cases)
│   └── integration_test.go  # Integration tests (full mux routing)
├── services/
│   └── services.go          # Business logic (AI, IPFS, blockchain orchestration)
├── blockchain/
│   └── blockchain.go        # Chain interaction layer
└── go.mod                   # Module definition
```

## How It Works

1. `main.go` creates an `http.ServeMux`, registers all routes, and starts the server on `:8080`
2. Incoming requests are routed to handler functions by method and path
3. `Submit` detects Content-Type to route between file upload and repo link flows
4. Each handler returns a JSON response via the shared `writeJSON` helper
5. Business logic (AI calls, IPFS uploads, blockchain transactions) will live in `services/`
6. Direct chain interactions will live in `blockchain/`

## API Reference

### `GET /health`

Health check endpoint for uptime monitoring.

**Response:** `200 OK`
```json
{"status": "ok"}
```

### `POST /api/submit`

Submit work for AI evaluation. Supports two Content-Types:

#### File Upload — `multipart/form-data`

Upload a file via the `file` form field. Maximum size: **10MB**.

```bash
curl -X POST http://localhost:8080/api/submit \
  -F "file=@main.go"
```

**Response:** `200 OK`
```json
{
  "type": "file",
  "name": "main.go",
  "size": 142,
  "message": "file received, pending evaluation"
}
```

**Error responses:**

| Status | Condition | Body |
|---|---|---|
| `400` | Missing Content-Type header | `{"error": "Content-Type header is required"}` |
| `400` | Missing `file` form field | `{"error": "missing or invalid 'file' field"}` |
| `400` | Empty filename | `{"error": "filename is required"}` |
| `400` | Empty file (0 bytes) | `{"error": "file is empty"}` |
| `413` | File exceeds 10MB | `{"error": "file exceeds 10MB limit"}` |
| `415` | Unsupported Content-Type | `{"error": "Content-Type must be multipart/form-data or application/json"}` |

#### Repo Link — `application/json`

Submit a GitHub repository URL for analysis.

```bash
curl -X POST http://localhost:8080/api/submit \
  -H "Content-Type: application/json" \
  -d '{"repo": "https://github.com/user/project"}'
```

**Response:** `200 OK`
```json
{
  "type": "repo",
  "name": "https://github.com/user/project",
  "size": 0,
  "message": "repo link received, pending evaluation"
}
```

**Error responses:**

| Status | Condition | Body |
|---|---|---|
| `400` | Invalid/malformed JSON | `{"error": "invalid JSON body"}` |
| `400` | Missing or empty `repo` field | `{"error": "'repo' field is required"}` |
| `400` | Non-GitHub URL | `{"error": "repo must be a valid GitHub URL (https://github.com/...)"}` |

### `GET /api/verify/{cid}`

Verify a credential by its IPFS CID.

**Path parameter:** `cid` (required) — the IPFS content identifier

**Response:** `501 Not Implemented` (placeholder — pending Phase 3–4)
```json
{"message": "verify endpoint not yet implemented", "cid": "QmTestCid123"}
```

**Error — missing CID:** `400 Bad Request`
```json
{"error": "cid is required"}
```

## Key Functions

| Function | File | Purpose |
|---|---|---|
| `main` | `main.go` | Registers routes, starts HTTP server |
| `Health` | `handlers.go` | Returns health status JSON |
| `Submit` | `submit.go` | Routes by Content-Type to file upload or repo link handler |
| `handleFileUpload` | `submit.go` | Parses multipart form, enforces 10MB limit, validates file |
| `handleRepoLink` | `submit.go` | Decodes JSON, validates repo is a GitHub URL |
| `Verify` | `handlers.go` | Handles credential verification by CID |
| `writeJSON` | `handlers.go` | Shared helper — sets Content-Type, status, encodes JSON |

## Running

```bash
cd backend
go run main.go
```

The server starts on `http://localhost:8080`. Verify with:

```bash
curl http://localhost:8080/health
```

## Testing

### Run tests

```bash
# From project root — all stacks
./scripts/test.sh

# Backend only
./scripts/test-backend.sh

# Verbose output
./scripts/test-backend.sh -v

# With HTML coverage report
./scripts/test-backend.sh --html

# Or directly with go test
cd backend && go test -race -v ./...
```

### Test coverage

| Package | Coverage |
|---|---|
| `handlers` | 87.8% |
| `main` | 0% (server startup — not unit testable) |
| **Total** | **76.8%** |

Per-function breakdown:

| Function | Coverage |
|---|---|
| `Health` | 100% |
| `Verify` | 100% |
| `writeJSON` | 100% |
| `Submit` | 100% |
| `handleFileUpload` | 70% |
| `handleRepoLink` | 100% |

> `handleFileUpload` at 70% — the uncovered paths are `io.ReadAll` failure and the oversized file `MaxBytesReader` error, which are difficult to trigger in unit tests without mocking the reader.

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

### Watch mode

```bash
./scripts/test-watch.sh backend
```

Automatically re-runs Go tests when `.go` files change.

## Next Steps

- **Issue #4:** Add input validation and sanitization middleware
