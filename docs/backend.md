# Backend API

The Go backend is the central orchestrator of the PoSA pipeline — it receives user submissions, forwards them to the AI engine, stores reports on IPFS, and anchors proofs on-chain.

## Technology

- **Language:** Go 1.25+
- **Framework:** `net/http` standard library (no external dependencies)
- **Port:** `:8080`

## Directory Structure

```md
backend/
├── main.go                  # Server entrypoint, route registration
├── handlers/
│   ├── handlers.go          # Route handler functions
│   └── handlers_test.go     # Handler unit tests
├── services/
│   └── services.go          # Business logic (AI, IPFS, blockchain orchestration)
├── blockchain/
│   └── blockchain.go        # Chain interaction layer
└── go.mod                   # Module definition
```

## How It Works

1. `main.go` creates an `http.ServeMux`, registers all routes, and starts the server on `:8080`
2. Incoming requests are routed to handler functions in `handlers/handlers.go`
3. Each handler processes the request and returns a JSON response via the shared `writeJSON` helper
4. Business logic (AI calls, IPFS uploads, blockchain transactions) will live in `services/`
5. Direct chain interactions will live in `blockchain/`

## API Reference

### `GET /health`

Health check endpoint for uptime monitoring.

**Response:** `200 OK`

```json
{"status": "ok"}
```

### `POST /api/submit`

Submit work for AI evaluation. Accepts file uploads or repo links.

**Request formats:**

- `multipart/form-data` with a `file` field
- `application/json` with a `repo` field

**Response:** `501 Not Implemented` (placeholder)

```json
{"message": "submit endpoint not yet implemented"}
```

**Planned response (after Phase 1–4):**

```json
{
  "score": 82,
  "issues": [
    {"type": "security", "message": "Unvalidated user input at line 45"}
  ],
  "suggestions": ["Add input sanitization before DB query"],
  "cid": "QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco",
  "tx_hash": "0xabc123...def456"
}
```

### `GET /api/verify/{cid}`

Verify a credential by its IPFS CID.

**Path parameter:** `cid` (required) — the IPFS content identifier

**Response:** `501 Not Implemented` (placeholder)

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
| `Health` | `handlers/handlers.go` | Returns health status JSON |
| `Submit` | `handlers/handlers.go` | Handles work submissions |
| `Verify` | `handlers/handlers.go` | Handles credential verification by CID |
| `writeJSON` | `handlers/handlers.go` | Shared helper — sets Content-Type, status, encodes JSON |

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
# From project root
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
| `handlers` | 100% |
| `main` | 0% (server startup — not unit testable) |
| **Total** | **58.8%** |

### Test cases

| Test | What it verifies |
|---|---|
| `TestHealth` | Returns `200` with `{"status": "ok"}` and correct Content-Type |
| `TestSubmit` | Returns `501` with a non-empty placeholder message |
| `TestVerify/valid_cid` | Extracts CID from path, returns `501` with CID in response |
| `TestVerify/empty_cid` | Returns `400` with error message when CID is missing |
| `TestWriteJSON` | Correctly sets status code, Content-Type header, and encodes JSON body |

### Watch mode

```bash
./scripts/test-watch.sh backend
```

Automatically re-runs Go tests when `.go` files change.

## Next Steps

- **Issue #3:** Implement file upload endpoint with multipart parsing
- **Issue #4:** Add input validation and sanitization middleware
