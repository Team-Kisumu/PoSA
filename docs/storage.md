# Decentralized Storage

> **Status:** Implemented (Phase 3)

The storage layer persists AI evaluation reports on IPFS and returns content-addressed CIDs for blockchain anchoring. Supports both a local IPFS node and the Pinata cloud pinning service.

## Technology

- **Storage:** IPFS (content-addressed)
- **Backends:** Local IPFS node (HTTP API) or [Pinata](https://www.pinata.cloud/) (cloud pinning)
- **Language:** Go (stdlib `net/http`, no external dependencies)
- **Package:** `backend/storage/`

## Directory Structure

```md
backend/storage/
├── ipfs.go          # Store interface, IPFSClient, PinataClient, factory
└── ipfs_test.go     # 12 tests with HTTP test server mocks
```

## Architecture

### Store Interface

```go
type Store interface {
    Upload(report []byte) (cid string, err error)
    Retrieve(cid string) (data []byte, err error)
}
```

Both backends implement this interface, making them swappable via the factory function.

### Backend Selection

```mmd
NewStore(ipfsURL, pinataURL, pinataKey)
  ├── pinataKey set? → PinataClient
  └── otherwise     → IPFSClient (default: localhost:5001)
```

### Upload Flow

```mmd
Report JSON → Multipart Form → IPFS/Pinata API → CID returned
```

1. Report bytes are wrapped in a multipart form upload (field: "file", name: "report.json")
2. Sent to the storage backend via HTTP POST
3. Backend returns a CID (content hash)
4. CID is used for blockchain anchoring and verification

### Retrieve Flow

```mmd
CID → IPFS cat / Pinata gateway → Report JSON returned
```

## API Reference (Internal)

### IPFSClient

| Method | Endpoint | Description |
|---|---|---|
| Upload | `POST /api/v0/add` | Multipart upload, returns `{"Hash": "Qm..."}` |
| Retrieve | `POST /api/v0/cat?arg={cid}` | Returns raw report bytes |

### PinataClient

| Method | Endpoint | Description |
|---|---|---|
| Upload | `POST /pinning/pinFileToIPFS` | Multipart upload with Bearer auth, returns `{"IpfsHash": "Qm..."}` |
| Retrieve | `GET gateway.pinata.cloud/ipfs/{cid}` | Public gateway fetch |

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `IPFS_API_URL` | Local IPFS node HTTP API URL | `http://localhost:5001` |
| `PINATA_API_URL` | Pinata API base URL | `https://api.pinata.cloud` |
| `PINATA_API_KEY` | Pinata JWT token (if set, Pinata is used instead of local IPFS) | — |

## Error Handling

| Error | Cause | Handling |
|---|---|---|
| Upload returns non-200 | IPFS node down, Pinata auth failure | Error with status code and body |
| Empty CID in response | Malformed IPFS/Pinata response | Error: "empty CID in response" |
| Retrieve returns non-200 | CID not found, network error | Error with status code and body |
| HTTP client timeout | Network issues (30s timeout) | Error wrapping the timeout |

## Running

```bash
# With local IPFS node
export IPFS_API_URL=http://localhost:5001
cd backend && go run main.go

# With Pinata
export PINATA_API_KEY=your-jwt-token
cd backend && go run main.go
```

## Testing

```bash
cd backend && go test -race -v ./storage/
# 12 tests, ~1s
```

All tests use `httptest.NewServer` to mock IPFS and Pinata APIs — no real IPFS node or Pinata account needed.

### Test Summary

| Test | Verifies |
|---|---|
| `TestIPFSUpload` | Multipart upload, CID extraction from response |
| `TestIPFSUploadError` | Error handling for 500 response |
| `TestIPFSUploadEmptyCID` | Error handling for empty Hash field |
| `TestIPFSRetrieve` | Fetch by CID, correct query parameter |
| `TestIPFSRetrieveError` | Error handling for 404 response |
| `TestPinataUpload` | Auth header, CID extraction from IpfsHash |
| `TestPinataUploadError` | Error handling for 401 response |
| `TestPinataUploadEmptyCID` | Error handling for empty IpfsHash |
| `TestPinataRetrieve` | Gateway fetch pattern |
| `TestNewStoreIPFS` | Factory returns IPFSClient when no Pinata key |
| `TestNewStorePinata` | Factory returns PinataClient when key is set |
| `TestNewStoreDefaultIPFS` | Factory uses default localhost:5001 URL |
